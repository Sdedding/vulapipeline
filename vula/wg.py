"""
This module provides the interface for configuring WireGuard interfaces.

It currently requires PyRoute2 and only works on GNU/Linux; in the future it
should contain implementations for other platforms, too.

A goal of the Feb12 2021 refactor is that sys_pyroute2 will be the only user of
this module.
"""

from __future__ import annotations

import time
from base64 import b64decode, b64encode  # noqa: F401
from datetime import timedelta
from ipaddress import ip_address, ip_network
from logging import Logger, getLogger
from typing import (
    Self,
    Optional,
    Any,
    Type,
    TypeVar,
    Union,
    cast,
)

import click
from pyroute2 import IPRoute
from pyroute2 import WireGuard as PyRoute2WireGuard
from pyroute2.netlink import nla as netlink_atom
from schema import And, Optional as Optional_, Or, Schema, Use


from .common import (
    DualUse,
    attrdict,
    b64_bytes,
    comma_separated_Nets,
    format_byte_stats,
    jsonrepr_hl,
    schemattrdict,
    serializable,
    yamlrepr_hl,
)


def _wg_interface_list() -> list[str]:
    """
    This returns a list of names of current WireGuard interfaces.

    There must be a better way to do this!
    >>> type(_wg_interface_list())
    <class 'list'>
    """
    interfaces = []
    links = IPRoute().get_links()
    for link in links:
        linkinfo = link.get_attr('IFLA_LINKINFO')
        if linkinfo is not None:
            if linkinfo.get_attr('IFLA_INFO_KIND') == "wireguard":
                interfaces.append(link.get_attr('IFLA_IFNAME'))
    return interfaces


class PeerConfig(schemattrdict, serializable):
    schema = Schema(
        {
            Optional_('unspec'): object,
            Optional_('remove'): bool,
            Optional_('public_key'): And(b64_bytes.with_len(32), Use(str)),
            Optional_('preshared_key'): And(b64_bytes.with_len(32), Use(str)),
            Optional_('endpoint_addr'): And(ip_address, Use(str)),
            Optional_('endpoint_port'): Use(int),
            Optional_('persistent_keepalive'): Use(int),
            Optional_('allowed_ips'): Or(
                [And(ip_network, Use(str))],
                And(Use(comma_separated_Nets), Use(list)),
            ),
            Optional_('stats'): dict(
                rx_bytes=int, tx_bytes=int, latest_handshake=int
            ),
            Optional_('latest_handshake'): Use(int),
        }
    )

    default = dict(remove=False)

    T = TypeVar("T", bound="PeerConfig")

    @classmethod
    def from_netlink(cls: Type[T], peer: dict[str, Any]) -> T:
        """
        This converts approximately from what pyroute2 produces to what
        pyroute2 consumes (plus the extra key "stats" with two keys).

        >>> p = PeerConfig.from_netlink({'attrs': dict(
        ...     public_key=b64encode(b'A'*32),
        ...     preshared_key=b64encode(b'A'*32),
        ...     persistent_keepalive_interval=666,
        ...     rx_bytes=2,
        ...     tx_bytes=3,
        ...     protocol_version=99,
        ...     endpoint={'addr': '192.168.0.0', 'port': 1000},
        ...     last_handshake_time={'tv_sec':567})})
        >>> type(p)
        <class 'vula.wg.PeerConfig'>
        >>> p['persistent_keepalive']
        666
        >>> p['endpoint_addr']
        '192.168.0.0'
        >>> p['endpoint_port']
        1000
        >>> p['stats']['latest_handshake']
        567
        >>> p['protocol_version'] # doctest: +IGNORE_EXCEPTION_DETAIL
        Traceback (most recent call last):
        KeyError:
        >>> p = PeerConfig.from_netlink({'attrs': dict(
        ...     public_key=b64encode(b'A'*32),
        ...     preshared_key=b64encode(b'A'*32),
        ...     persistent_keepalive_interval=666,
        ...     rx_bytes=2,
        ...     tx_bytes=3,
        ...     protocol_version=99,
        ...     endpoint={'addr': 'FE80::FFFF:FFFF:FFFF:FFFE', 'port': 1000},
        ...     last_handshake_time={'tv_sec':567})})
        >>> type(p)
        <class 'vula.wg.PeerConfig'>
        >>> p['persistent_keepalive']
        666
        >>> p['endpoint_addr']
        'FE80::FFFF:FFFF:FFFF:FFFE'
        >>> p['endpoint_port']
        1000
        >>> p['stats']['latest_handshake']
        567
        >>> p['protocol_version'] # doctest: +IGNORE_EXCEPTION_DETAIL
        Traceback (most recent call last):
        KeyError:
        """
        res: dict[str, Any]
        res = {
            k.replace('WGPEER_A_', '').lower(): (
                dict(v) if isinstance(v, netlink_atom) else v
            )
            for k, v in dict(peer['attrs']).items()
        }
        res['allowed_ips'] = [
            ip_network(
                (
                    bytes.fromhex(
                        net['WGALLOWEDIP_A_IPADDR'].replace(':', '')
                    ),
                    net['WGALLOWEDIP_A_CIDR_MASK'],
                ),
            )
            for net in (
                dict(atom['attrs']) for atom in res.pop('allowedips', ())
            )
        ]
        res['public_key'] = res['public_key'].decode()
        res['preshared_key'] = res['preshared_key'].decode()
        if b64decode(res['preshared_key']) == bytes([0]) * 32:
            del res['preshared_key']

        if 'endpoint' in res:
            res['endpoint_addr'] = res['endpoint']['addr']
            res['endpoint_port'] = res['endpoint']['port']
            res.pop('endpoint')
        res['persistent_keepalive'] = res.pop('persistent_keepalive_interval')
        res['stats'] = dict(
            rx_bytes=res['rx_bytes'],
            tx_bytes=res['tx_bytes'],
            latest_handshake=res.pop('last_handshake_time')['tv_sec'],
        )
        res.pop('rx_bytes')
        res.pop('tx_bytes')
        res.pop('protocol_version')
        return cls(res)

    @property
    def wg_show(self: PeerConfig) -> str:
        return "\n  ".join(
            click.style(label, bold=True) + ': ' + str(value)
            for label, value in {
                click.style('peer', fg="yellow"): click.style(
                    self['public_key'], fg="yellow"
                ),
                'preshared key': (self.get('preshared_key') and '(hidden)'),
                'endpoint': (
                    'endpoint_addr' in self
                    and 'endpoint_port' in self
                    and "{endpoint_addr}:{endpoint_port}".format(**self)
                ),
                'allowed ips': (", ".join(self['allowed_ips']) or "(none)"),
                'latest handshake': (
                    self['stats']['latest_handshake']
                    and str(
                        timedelta(
                            seconds=int(
                                time.time() - self['stats']['latest_handshake']
                            )
                        ),
                    )
                    + ' ago'
                ),
                'transfer': (
                    sum(self['stats'].values())
                    and "{rx_bytes} received, {tx_bytes} sent".format(
                        **format_byte_stats(self['stats'])
                    )
                ),
                'persistent keepalive': (
                    self.get('persistent_keepalive')
                    and "every %s seconds" % (self['persistent_keepalive'],)
                ),
            }.items()
            if value is not False and value is not None
        )

    @property
    def wg_showconf(self: PeerConfig) -> str:
        return "[Peer]\n" + "\n".join(
            label + ' = ' + value
            for label, value in {
                'PublicKey': self['public_key'],
                'PresharedKey': self.get('preshared_key'),
                'AllowedIPs': ", ".join(self['allowed_ips']),
                'Endpoint': (
                    'endpoint_addr' in self
                    and 'endpoint_port' in self
                    and "{endpoint_addr}:{endpoint_port}".format(**self)
                ),
                'PersistentKeepalive': (
                    self.get('persistent_keepalive')
                    and str(self['persistent_keepalive'])
                ),
            }.items()
            if value not in (None, False, '')
        )


@DualUse.object()
@click.argument('name', type=str)
class Interface(attrdict, yamlrepr_hl):
    """
    This is a wrapper for pyroute2's WireGuard interface.

    Give it an interface name, and it will give you a dict with the data in the
    same shape as the pyroute2 structure but with less irritating names.

    But wait, there's more...it also has a "peers" attribute which gives you
    the peers in the shape that pyroute2's WireGuard module expects to
    pass to the "set" method (note: round trips not yet tested).

    Because this is a DualUse.object, you can see the data on the commandline
    with commands like these:

    sudo vula wg Interface vula query

    sudo vula wg Interface vula peers

    sudo vula wg Interface vula _is_up

    etc.
    """

    # Type informations
    cli: click.Group

    def __init__(self, name: str, ipr: Optional[IPRoute] = None):
        """
        >>> iface = Interface("test interface")
        >>> iface.name
        'test interface'
        >>> iface._ipr # doctest: +ELLIPSIS
        <pyroute2.iproute.linux.IPRoute object at 0x...>
        >>> len(Interface(42))
        0
        """
        self.log: Logger = getLogger()
        self.name = name
        self._wg = PyRoute2WireGuard()
        if ipr is None:
            ipr = IPRoute()
        self._if_index = None
        self._ipr = ipr
        self.query()

    @property
    def _get_link(self) -> Union[Any, dict[Any, Any]]:
        """
        >>> custom_ipr = IPRoute()
        >>> link_name = custom_ipr.get_links(0)[0].get_attr('IFLA_IFNAME')
        >>> int = Interface(link_name, custom_ipr)
        >>> link_name == int._get_link.get_attr('IFLA_IFNAME')
        True

        >>> int = Interface("random interface name")
        >>> int._get_link
        {}
        """
        res = self._ipr.get_links(ifname=self.name)
        if res:
            return res[0]
        else:
            return {}

    @property
    def _exists(self) -> bool:
        return bool(self._ipr.link_lookup(ifname=self.name))

    @property
    def _is_up(self) -> bool:
        return bool(self._get_link.get('state') == 'up')

    @DualUse.method()
    @click.option('-n', '--dryrun', is_flag=True)
    @click.argument('private_key', type=str)
    @click.argument('listen_port', type=int)
    @click.argument('fwmark', type=int)
    def sync_interface(
        self, private_key: str, listen_port: int, fwmark: str, dryrun: bool
    ) -> list[str]:
        """
        Creates, brings up, and configures an interface.

        Returns a list of the necessary steps.
        """
        res = []

        private_key_enc: bytes = private_key.encode()

        if not self._exists:
            if not dryrun:
                self._ipr.link('add', ifname=self.name, kind="wireguard")
            res.append('# create interface')
            res.append('ip link add %s type wireguard' % (self.name,))

        if not self._is_up:
            if not dryrun:
                if_index = self._ipr.link_lookup(ifname=self.name)[0]
                self._ipr.link('set', index=if_index, state='up')
            res.append('# bring up interface')
            res.append('ip link set up %s' % (self.name,))

        self.query()

        data = dict(
            private_key=private_key_enc, listen_port=listen_port, fwmark=fwmark
        )
        todo = {}
        for k, v in data.items():
            if self.get(k) != v:
                todo[k] = v
        if todo:
            if not dryrun:
                self.set(**dict(todo))
            if 'private_key' in todo:
                todo['private_key'] = '<redacted private key>'
            res.append('# configure interface')
            res.append("WireGuard.set(%r, **%r)" % (self.name, todo))
        return res

    @DualUse.method()
    def query(self) -> Self:
        """
        This calls "info" for the interface via pyroute2, and (re-)populates
        our dictionary. (We're a dict subclass, recall). It returns self.
        """
        self.clear()
        self.log.debug("Fetching interface info for %s", self.name)
        try:
            res: tuple[dict[str, Any]] = self._wg.info(self.name)
        except Exception as ex:
            self.log.warn("Failed to query interface %r: %r", self.name, ex)
            return self

        data = {
            k.replace('WGDEVICE_A_', '').lower(): (
                dict(v) if isinstance(v, netlink_atom) else v
            )
            for k, v in dict(res[0]['attrs']).items()
        }
        data['peers'] = list(
            map(PeerConfig.from_netlink, data.get('peers', ()))
        )
        self.clear()
        self.update(data)
        return self

    def set(self, **kwargs: Any) -> list[str]:
        self.log.debug("Calling WireGuard.set(%r, **%r)", self.name, kwargs)
        res: list[str] = self._wg.set(self.name, **kwargs)
        self.log.debug("WireGuard.set(%r, **%r) -> %r", self.name, kwargs, res)
        return res

    def apply_peerconfig(self, new: attrdict, dryrun: bool = False) -> str:
        """
        This sets only the keys that have changed, and returns a list of the
        new keys that needed to be set. Due to a bug in PyRoute2 and/or Linux,
        it is necessary to always set the allowed_ips if anything is set, so
        this does that.
        """
        self.query()
        cur = self._peers_by_pubkey.get(new["public_key"])
        res: list[str] = []
        if cur:
            if new.get('remove'):
                res.append(
                    '# removing wireguard peer %s' % (new['public_key'],)
                )

            elif 'allowed_ips' not in new:
                # pyroute2/wg bug workaround
                new['allowed_ips'] = cur['allowed_ips']

            for key in list(new):
                if cur.get(key) == new[key]:
                    if key in ("allowed_ips", 'public_key'):
                        # allowed_ips is here for pyroute2/wg bug workaround
                        continue
                    self.log.debug(
                        "not resetting %r=%r as it is unchanged",
                        key,
                        cur.get(key),
                    )
                    del new[key]
                elif key != 'remove':
                    res.append(
                        "# {key} is {cur}; should be {new}".format(
                            key=key, cur=cur.get(key), new=new[key]
                        )
                        if key != 'preshared_key'
                        else f'# {key} (redacted) is incorrect'
                    )
            # workaround for pyroute2 irritatingly handling endpoint addr
            # and port separately (while they actually need to be set
            # together)
            if 'endpoint_addr' in new and 'endpoint_port' not in new:
                new['endpoint_port'] = cur['endpoint_port']
            elif 'endpoint_port' in new and 'endpoint_addr' not in new:
                new['endpoint_addr'] = cur['endpoint_addr']
        else:
            if new.get('remove'):
                return "# can't remove non-existent wireguard peer %s" % (
                    new['public_key'],
                )

        if (
            cur
            and len(new) == 2
            and new.get('allowed_ips') == cur['allowed_ips']
        ):
            # pyroute2/wg bug workaround
            self.log.debug("apply_peerconfig: no wg update necessary")

        else:
            if cur:
                res.append(
                    '# reconfigure wireguard peer %s' % (new['public_key'],)
                )
            else:
                res.append(
                    '# configure new wireguard peer %s' % (new['public_key'],)
                )

            res.append(
                "vula wg set {interface} peer {pk} "
                "{remove}{endpoint}{args}{allowed_ips}".format(
                    remove="remove " if new.get('remove') else "",
                    endpoint=(
                        "endpoint %s:%s "
                        % (new['endpoint_addr'], new['endpoint_port'])
                        if (
                            new.get('endpoint_addr')
                            and new.get('endpoint_port')
                        )
                        else ''
                    ),
                    args="".join(
                        f"{k}"
                        f" {'<redacted psk>' if k == 'preshared_key' else v} "
                        for k, v in new.items()
                        if k in ('persistent_keepalive', 'preshared_key')
                    ),
                    allowed_ips=(
                        'allowed-ips %s '
                        % ",".join(ip for ip in new.get('allowed_ips', ()))
                        if 'allowed_ips' in new
                        else ""
                    ),
                    interface=self.name,
                    pk=new['public_key'],
                )
            )

            for line in res:
                self.log.info("[#] %s", line)

            if not dryrun:
                self.set(peer=new)

        return "\n".join(filter(None, res))

    @property
    def peers(self) -> list[PeerConfig]:
        """
        Returns list of peer structures which should be identical to those
        passed to the pyroute2 WireGuard set method, except with an extra
        "stats" key.

        And also except for that it won't do IPv6 correctly.

        This would be a good function to write tests for (and perhaps send
        upstream to pyroute2).
        """
        return cast(list[PeerConfig], self.get('peers', []))

    @property
    def _peers_by_pubkey(self) -> dict[str, PeerConfig]:
        return {peer['public_key']: peer for peer in self.peers}

    @property
    def wg_show(self) -> str:
        """
        This returns output similar to the "wg show" command.
        """
        peers: list[PeerConfig] = list(self.peers)
        return (
            "\n  ".join(
                click.style(label, bold=True) + ': ' + str(value)
                for label, value in {
                    click.style(
                        'interface', fg="green", bold=True
                    ): click.style(self.name, fg="green"),
                    'public key': (
                        'public_key' in self and self['public_key'].decode()
                    ),
                    'private key': ('public_key' in self and '(hidden)'),
                    'listening port': self.get('listen_port'),
                    'fwmark': (
                        self.get('fwmark') and "0x%x" % (self['fwmark'],)
                    ),
                }.items()
                if value not in (None, False)
            )
            + ("\n\n" if peers else '')
            + "\n\n".join(peer.wg_show for peer in peers)
        )

    @property
    def wg_showconf(self) -> str:
        """
        Shows the current configuration of a given WireGuard interface, for use
        with `setconf'.
        """
        peers = list(self.peers)
        return "[Interface]\n" + (
            "\n".join(
                "%s = %s" % (label, value)
                for label, value in {
                    'ListenPort': self.get('listen_port'),
                    'FwMark': (
                        self.get('fwmark') and "0x%x" % (self['fwmark'],)
                    ),
                    'PrivateKey': self.get('private_key', b'').decode(),
                }.items()
                if value not in (None, False, '')
            )
            + ("\n\n" if peers else '')
            + "\n\n".join(peer.wg_showconf for peer in peers)
        )


@DualUse.object(
    invoke_without_command=True,
    short_help="set and retrieve configuration of WireGuard interfaces",
)
@click.pass_context
class wg(object):
    """
    The subcommands here exist mostly to aid in the development of the
    wg.Interface class, and aren't currently intended for normal use.
    """

    # Type information for @DualUse.method
    cli: click.Group

    def __init__(self, ctx: click.Context, *a: Any, **kw: Any):
        if ctx.invoked_subcommand is None:
            click.echo(self.show())

    @DualUse.method(
        short_help="Shows the current configuration and device information"
    )
    @click.option(
        '-f',
        '--format',
        'fmt',
        type=click.Choice(["wireguard", "json", "yaml"]),
        default='wireguard',
        show_default=True,
    )
    @click.argument('interfaces', type=str, nargs=-1)
    def show(
        self, fmt: str = "wireguard", interfaces: Optional[list[str]] = None
    ) -> Optional[str]:
        """
        Produces output very similar to the "wg show" command.

        (The transfer counters, last handshake time, and keepalive interval are
        all formatted differently.)
        """
        if interfaces is None or len(interfaces) == 0:
            interfaces = _wg_interface_list()

        _interfaces = [Interface(name).query() for name in interfaces]

        if fmt == 'wireguard':
            return "\n\n".join(iface.wg_show for iface in _interfaces)

        elif fmt == 'yaml':
            return str(
                yamlrepr_hl({iface.name: iface for iface in _interfaces})
            )

        elif fmt == 'json':
            return str(
                jsonrepr_hl({iface.name: iface for iface in _interfaces})
            )

        return None

    @DualUse.method()
    @click.argument('interface', type=str)
    def showconf(self, interface: str) -> str:
        """
        Shows the current configuration of a given WireGuard interface, for use
        with "setconf".
        """
        return Interface(interface).query().wg_showconf

    @staticmethod
    @DualUse.method(
        short_help="Change the current configuration, add peers, remove "
        "peers, or change peers."
    )
    @click.option('--private-key', type=str)
    @click.option('--listen-port', type=int)
    @click.option('--fwmark', type=int)
    @click.argument('interface', type=str)
    @click.argument('args', type=str, nargs=-1)
    @click.pass_context
    def set(
        ctx: click.Context,
        self: Any,
        interface: str,
        args: tuple[str, ...] = (),
        /,
        **kwargs: Any,
    ) -> str:
        """
        Change the current configuration, add peers, remove peers, or change
        peers.

        Usage: wg set <interface> [listen-port <port>] [fwmark <mark>]
        [private-key <base64 private key>] [peer <base64 public key> [remove]
        [preshared-key <base64 preshared key>] [ endpoint <ip>:<port>]
        [persistent-keepalive <interval seconds>] [allowed-ips
        <ip1>/<cidr1>[,<ip2>/<cidr2>]...] ]...

        This is intended to behave very similarly to the normal wg tool, except
        that it takes private keys on the commandline (and does not yet support
        reading them from files, as wg does). Yes, this is not a great idea,
        but it makes testing easier.
        """
        dev = Interface(interface)
        _kwargs: dict[str, Any] = {
            k: v for k, v in kwargs.items() if v not in (None, ())
        }
        current: dict[str, Optional[str | bool]] = {}
        kwargss = []
        _args = list(sum(_kwargs.items(), ())) + list(args)
        allowed = {"private_key", "listen_port", "fwmark"}
        try:
            while _args:
                key = _args.pop(0)
                key = key.replace('-', '_')
                if key in allowed:
                    allowed.remove(key)
                    if current.get(key) is not None:
                        raise Exception("duplicate argument: %r" % (key,))
                    current[key] = _args.pop(0)
                elif key == "peer":
                    if current:
                        kwargss.append(current)
                    pk = _args.pop(0)
                    current = dict(public_key=pk)
                    allowed = {
                        "preshared_key",
                        "endpoint",
                        "persistent_keepalive",
                        "allowed_ips",
                    }
                elif key == "remove":
                    if not kwargss:
                        raise Exception("'remove' requires 'peer' first")
                    current[key] = True
                else:
                    raise Exception("bad key: %r" % (key,))
            kwargss.append(current)
        except Exception as ex:
            raise Exception("Failed to parse arguments: %r" % (ex,))
        res: list[Any] = []

        if 'public_key' not in kwargss[0]:
            res.append(dev.set(**kwargss.pop(0)))
        for peer in kwargss:
            if 'endpoint' in peer:
                assert isinstance(peer["endpoint"], str)
                peer['endpoint_addr'], _, peer['endpoint_port'] = peer[
                    'endpoint'
                ].rpartition(':')
                assert isinstance(peer["endpoint_addr"], str)
                peer['endpoint_addr'] = peer['endpoint_addr'].strip("[]")
                del peer['endpoint']
            peer = PeerConfig(peer)
            res.append(dev.apply_peerconfig(peer))
        return "\n".join(map(str, res))

    @DualUse.method()
    @click.argument('interface', type=str)
    @click.argument('public_key', type=str)
    @click.option('--preshared-key', type=str)
    @click.option('--endpoint-addr', type=lambda x: str(ip_address(x)))
    @click.option('--endpoint-port', type=int)
    @click.option('--persistent-keepalive', type=int)
    @click.option(
        '--allowed-ip',
        'allowed_ips',
        type=lambda x: str(ip_network(x)),
        multiple=True,
    )
    @click.option('--remove', is_flag=True, help="Remove this peer")
    def set_peer(self, interface: str, **kwargs: Any) -> str:
        """
        This allows setting peer information.
        """
        _kwargs = {k: v for k, v in kwargs.items() if v not in (None, ())}
        if 'allowed_ips' in _kwargs:
            kwargs['allowed_ips'] = list(_kwargs['allowed_ips'])
        wg = Interface(interface)
        return wg.apply_peerconfig(attrdict(kwargs))


@click.group()
def link() -> None:
    """
    Link commands
    """


@link.command(name='del', short_help="Delete an interface")
@click.argument('name', type=str)
def del_(name: str) -> None:
    """
    Delete an interface.

    Note: this is not WireGuard-specific.
    """
    IPRoute().link('del', ifname=name)


@link.command()
@click.argument('name', type=str)
def add(name: str) -> None:
    """
    Add a new WireGuard interface
    """
    IPRoute().link('add', ifname=name, kind="wireguard")


wg.cli.add_command(link, name="link")
wg.cli.add_command(Interface.cli, name="Interface")

main = wg.cli
