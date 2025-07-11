"""
vula-discover is a stateless program that prints each WireGuard mDNS
service and formats the service parameters into a single easy-to-parse line.
This program's output is intended to be fed into a daemon that configures
WireGuard peers discovered by vula to configure the local vula
interface.

The output of this program may be written to a pipe, a log file, a unix
socket, the vula-organize DBus interface for processing descriptors, or any
other place. It should run with the lowest privileges possible. The
output is not filtered and so adversaries may attempt to inject unreasonable
hosts such as `127.0.0.1` or other addresses. Care should be taken that only
addresses for the local network segment are used as WireGuard peers.
"""

from ipaddress import ip_address as ip_addr_parser
from logging import Logger, getLogger
from typing import Optional, Callable, Any
import click
import pydbus
from click.exceptions import Exit
from gi.repository import GLib
from pyroute2 import IPRoute
from zeroconf import ServiceBrowser, ServiceInfo, ServiceListener, Zeroconf

from .constants import (
    _DISCOVER_DBUS_NAME,
    _LABEL,
    _ORGANIZE_DBUS_NAME,
    _ORGANIZE_DBUS_PATH,
)
from .peer import Descriptor


class VulaServiceListener(ServiceListener):
    """
    *VulaServiceListener* is for use with *zeroconf*'s *ServiceBrowser*.

    The key-value pairs conform to
    https://tools.ietf.org/html/rfc6763#section-6.4.
    """

    def __init__(
        self,
        callback: Callable[[Descriptor], None],
        our_wg_pk: Optional[str] = None,
    ) -> None:
        """
        Specifying our_wg_pk is optional, and allows discover to drop our own
        local descriptors before they get to organize. This reduces log noise
        and was helpful for debugging.
        """
        super(VulaServiceListener, self).__init__()
        self.log: Logger = getLogger()
        self.callback = callback
        self.our_wg_pk = our_wg_pk

    def add_service(self, zeroconf: Zeroconf, s_type: str, name: str) -> None:
        """
        When *zeroconf* discovers a new WireGuard service it calls
        *add_service*, which passes an instantiated descriptor object to the
        callback (unless its pk is our our_wg_pk).
        """

        # Typing note:
        # 'Any' works here and while 'Optional[ServiceInfo]' should, it does
        # not unless mypy is called with --no-strict-optional like so:
        #
        #   mypy --ignore-missing-imports  --no-strict-optional discover.py
        info: Optional[ServiceInfo] = zeroconf.get_service_info(s_type, name)

        if info is None:
            return

        try:
            desc = Descriptor.from_zeroconf_properties(info.properties)
        except Exception as ex:
            self.log.debug(
                "discover dropped invalid descriptor: %r (%r)"
                % (info.properties, ex)
            )
            return

        if str(desc.pk) == self.our_wg_pk:
            self.log.debug("discover ignored descriptor with our_wg_pk")

        else:
            self.callback(desc)

    def update_service(self, *a: Any, **kw: Any) -> None:
        return self.add_service(*a, **kw)


class Discover(object):
    dbus = '''
    <node>
      <interface name='local.vula.discover1.Listen'>
        <method name='listen'>
          <arg type='as' name='ip_addrs' direction='in'/>
          <arg type='s' name='our_wg_pk' direction='in'/>
        </method>
      </interface>
    </node>
    '''

    def __init__(self) -> None:
        self.callbacks: list[Callable[[Descriptor], Any]] = []
        self.browsers: dict[str, tuple[Zeroconf, ServiceBrowser]] = {}
        self.log: Logger = getLogger()

    def callback(self, value: Descriptor) -> None:
        for callback in self.callbacks:
            callback(value)

    def listen_on_ip_or_if(self, ip_address: str, interface: str) -> None:
        """
        Deprecated.

        This is for the CLI to accept interface names.
        """

        if interface and ip_address:
            self.log.info("Must pick interface or IP address")
            raise Exit(1)

        ip_addr: Optional[str] = None

        if ip_address:
            try:
                # @TODO assignment has no effect and could be removed
                ip_addr = str(ip_addr_parser(ip_address))
            except:  # noqa: E722
                self.log.info("Invalid IP address argument")
                raise Exit(3)
            ip_addr = ip_address
        elif interface:
            with IPRoute() as ipr:
                index = ipr.link_lookup(ifname=interface)[0]
                a = ipr.get_addr(match=lambda x: x['index'] == index)
                ip_addr = dict(a[0]['attrs'])['IFA_ADDRESS']

        if ip_addr:
            self.listen([ip_addr])

    def listen(
        self, ip_addrs: list[str], our_wg_pk: Optional[str] = None
    ) -> None:
        for ip_addr in ip_addrs:
            if ip_addr in self.browsers:
                self.log.info("Not launching a second browser for %r", ip_addr)
                continue
            zeroconf: Zeroconf = Zeroconf(interfaces=[ip_addr])
            self.log.debug("Starting ServiceBrowser for %r", ip_addr)
            browser: ServiceBrowser = ServiceBrowser(
                zeroconf, _LABEL, VulaServiceListener(self.callback, our_wg_pk)
            )
            self.browsers[ip_addr] = (zeroconf, browser)
        for old_ip in list(self.browsers):
            if old_ip not in ip_addrs:
                self.log.info(
                    "Removing old service browser for %r (new ip_addrs=%r)",
                    old_ip,
                    ip_addrs,
                )
                self.browsers[old_ip][1].cancel()
                self.browsers[old_ip][0].close()
                del self.browsers[old_ip]

    def shutdown(self) -> None:
        for ip, browser in list(self.browsers.items()):
            del self.browsers[ip]
            browser[1].cancel()

    def is_alive(self) -> bool:
        return any(browser[1].is_alive() for browser in self.browsers.values())

    @classmethod
    def daemon(cls, use_dbus: bool, ip_address: str, interface: str) -> None:
        """
        This method implements the non-monolithic daemon mode where we run
        Discover in its own process (as deployed on GNU/systemd).
        """

        loop = GLib.MainLoop()

        discover = cls()

        discover.callbacks.append(lambda d: discover.log.debug("%s", d))

        if use_dbus:
            discover.log.debug("dbus enabled")
            system_bus = pydbus.SystemBus()
            process = system_bus.get(
                _ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH
            ).process_descriptor_string
            discover.callbacks.append(lambda d: process(str(d)))
            system_bus.publish(_DISCOVER_DBUS_NAME, discover)

        discover.listen_on_ip_or_if(ip_address, interface)

        loop.run()  # type: ignore[no-untyped-call]


# FIXME: should we shutdown zeroconf objects upon glib shutdown? probably.
#        try:
#            while True:
#                sleep(1)
#        except KeyboardInterrupt:
#            pass
#        finally:
#            discover.shutdown()
#        return 0


@click.command(short_help="Layer 3 mDNS discovery daemon")
@click.option(
    "-d",
    "--dbus/--no-dbus",
    'use_dbus',
    default=True,
    is_flag=True,
    help="use dbus for IPC",
)
@click.option(
    "-I",
    "--ip-address",
    type=str,
    help="bind this IP address instead of automatically choosing which IP "
    "to bind",
)
@click.option(
    #    "-i",
    "--interface",
    type=str,
    help="bind to the primary IP address for the given interface, "
    "automatically choosing which IP to announce",
)
def main(**kwargs: Any) -> None:
    Discover.daemon(**kwargs)


if __name__ == "__main__":
    main()
