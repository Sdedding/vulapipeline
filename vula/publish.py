"""
 vula-publish is a program that announces a WireGuard mDNS service as
 informed by organize over DBus or as controlled by organize in monolith mode.

>>> p = Publish()
>>> type(p.zeroconfs)
<class 'dict'>
>>> type(p.log)
<class 'logging.RootLogger'>

"""

from ipaddress import IPv4Address, IPv6Address
from logging import Logger, getLogger
from platform import node as hostname
from typing import Any

import click
import pydbus
from gi.repository import GLib
from zeroconf import NonUniqueNameException, ServiceInfo, Zeroconf

from .constants import _LABEL, _PUBLISH_DBUS_NAME, _VULA_ULA_SUBNET
from .peer import Descriptor


class Publish(object):
    dbus = '''
    <node>
      <interface name='local.vula.publish1.Listen'>
        <method name='listen'>
          <arg type='a{ss}' name='new_announcements' direction='in'/>
        </method>
      </interface>
    </node>
    '''

    def __init__(self) -> None:
        self.log: Logger = getLogger()
        self.zeroconfs: dict[str, Zeroconf] = {}

    def listen(self, new_announcements: dict[str, str]) -> None:
        # First we remove all old zeroconf listeners that are not in our new
        # instructions
        for iface, zc in list(self.zeroconfs.items()):
            if iface not in new_announcements:
                self.log.info(
                    "Removing old service announcement for %r", iface
                )
                zc.close()
                del self.zeroconfs[iface]
        # Now we add a zeroconf listener for each new IP and ServiceInfo if it
        # is not already existing, else we update the old zc object with the
        # new desc
        for iface, desc_string in new_announcements.items():
            desc = Descriptor.parse(desc_string)
            listen_IPs: list[IPv4Address | IPv6Address] = [
                a for a in desc.all_addrs if a not in _VULA_ULA_SUBNET
            ]
            self.log.debug(
                "Starting mDNS service announcement for %r with listen_IPs %r",
                iface,
                listen_IPs,
            )
            name: str = hostname() + "." + _LABEL
            service_info: ServiceInfo = ServiceInfo(
                _LABEL,
                name=name,
                addresses=[ip.packed for ip in listen_IPs],
                port=desc.port,
                properties=desc.as_zeroconf_properties,
                server=desc.hostname,
            )
            zeroconf = self.zeroconfs.get(iface)
            if zeroconf:
                # Do update dance
                self.log.debug("Updating vula service: %s", service_info)
                zeroconf.update_service(service_info)
                self.log.debug("Updating vula service.")

            else:
                zeroconf = self.zeroconfs[iface] = Zeroconf(
                    # note that the "interfaces" argument to zeroconf is a list
                    # of IPs
                    interfaces=[str(ip) for ip in listen_IPs],
                )
                self.log.debug("Registering vula service: %s", service_info)
                try:
                    zeroconf.register_service(service_info)
                    self.log.debug("Registered vula mDNS publishing service.")
                except NonUniqueNameException:
                    self.log.debug(
                        "Unable to register vula mDNS publishing service."
                    )

    @classmethod
    def daemon(cls) -> None:
        """
        This method implements the non-monolithic daemon mode where we run
        publish in its own process (as we deploy on GNU/systemd).
        """
        loop = GLib.MainLoop()
        publish = cls()
        publish.log.debug("Debug level logging enabled")
        pydbus.SystemBus().publish(_PUBLISH_DBUS_NAME, publish)
        publish.log.debug("dbus enabled")
        loop.run()  # type: ignore[no-untyped-call]


@click.command(short_help="Layer 3 mDNS publish daemon")
def main(**kwargs: Any) -> None:
    Publish.daemon(**kwargs)


if __name__ == "__main__":
    main()
