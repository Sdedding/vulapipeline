from typing import List, TypedDict, Literal, cast, Optional, Any

import pydbus
import yaml

from vula.common import escape_ansi
from vula.constants import _ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH
from vula.organize import Organize


class PeerType(TypedDict):
    name: str
    id: str
    other_names: str
    status: str
    endpoint: str
    allowed_ips: str
    latest_signature: str
    latest_handshake: str
    wg_pubkey: str


class StatusType(TypedDict):
    publish: str
    discover: str
    organize: str


class PrefsType(TypedDict):
    pin_new_peers: bool
    accept_nonlocal: bool
    auto_repair: bool
    subnets_allowed: list[str]
    subnets_forbidden: list[str]
    iface_prefix_allowed: list[str]
    local_domains: list[str]
    ephemeral_mode: bool
    accept_default_route: bool
    record_events: bool
    expire_time: int
    overwrite_unpinned: bool


PrefsTypeKeys = Literal[
    "pin_new_peers",
    "accept_nonlocal",
    "auto_repair",
    "subnets_allowed",
    "subnets_forbidden",
    "iface_prefix_allowed",
    "local_domains",
    "ephemeral_mode",
    "accept_default_route",
    "record_events",
    "expire_time",
    "overwrite_unpinned",
]


class DataProvider:
    def get_peers(self) -> List[PeerType]:
        organize = pydbus.SystemBus().get(
            _ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH
        )

        # Get all peer ids from the dbus
        ids = organize.peer_ids("enabled")

        peers: List[PeerType] = []

        # Loop over all ids in the list
        for id in ids:
            # Create empty dict for peer
            peer_dict: dict[str, Optional[str]] = {
                "name": None,
                "id": None,
                "other_names": None,
                "status": None,
                "endpoint": None,
                "allowed_ips": None,
                "latest_signature": None,
                "latest_handshake": None,
                "wg_pubkey": None,
            }

            # Fill in the data
            peer_raw = organize.show_peer(id)
            peer_clear = escape_ansi(peer_raw)
            peer_lines = peer_clear.split("\n")

            peer_dict["name"] = peer_lines[0].lstrip().split(": ")[1]
            peer_dict["id"] = peer_lines[1].lstrip().split(": ")[1]
            if "other names" in peer_lines[2]:
                peer_dict["other_names"] = (
                    peer_lines[2].lstrip().split(": ")[1]
                )
                peer_dict["status"] = peer_lines[3].lstrip().split(": ")[1]
                peer_dict["endpoint"] = peer_lines[4].lstrip().split(": ")[1]
                peer_dict["allowed_ips"] = (
                    peer_lines[5].lstrip().split(": ")[1]
                )
                peer_dict["latest_signature"] = (
                    peer_lines[6].lstrip().split(": ")[1]
                )
                peer_dict["latest_handshake"] = (
                    peer_lines[7].lstrip().split(": ")[1]
                )
                peer_dict["wg_pubkey"] = peer_lines[8].lstrip().split(": ")[1]
            else:
                peer_dict["status"] = peer_lines[2].lstrip().split(": ")[1]
                peer_dict["endpoint"] = peer_lines[3].lstrip().split(": ")[1]
                peer_dict["allowed_ips"] = (
                    peer_lines[4].lstrip().split(": ")[1]
                )
                peer_dict["latest_signature"] = (
                    peer_lines[5].lstrip().split(": ")[1]
                )
                peer_dict["latest_handshake"] = (
                    peer_lines[6].lstrip().split(": ")[1]
                )
                peer_dict["wg_pubkey"] = peer_lines[7].lstrip().split(": ")[1]

            peers.append(cast(PeerType, peer_dict))

        return peers

    def get_prefs(self) -> PrefsType:
        organize = pydbus.SystemBus().get(
            _ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH
        )

        # Get the data from the organize dbus
        data = organize.show_prefs()
        items: PrefsType = yaml.safe_load(data)

        return items

    def get_status(self) -> Optional[StatusType]:
        # Fetch the data from the systemd dbus
        bus = pydbus.SystemBus()
        systemd = bus.get(
            "org.freedesktop.systemd1", "/org/freedesktop/systemd1"
        )

        # Create an empty dict for the status
        status = StatusType(publish="", discover="", organize="")

        # Define names to consider in the result
        for name in status.keys():
            name = cast(Literal["publish", "discover", "organize"], name)
            # Template string for service name
            unit_name = "vula-%s.service" % (name,)
            try:
                unit = pydbus.SystemBus().get(
                    ".systemd1", systemd.GetUnit(unit_name)
                )
                status[name] = unit.ActiveState
            except Exception as ex:
                print(ex)
                return None

        return status

    def our_latest_descriptors(self) -> str:
        organize: Organize = pydbus.SystemBus().get(
            _ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH
        )
        return organize.our_latest_descriptors()

    def delete_peer(self, peer_vk: str) -> None:
        organize: Organize = pydbus.SystemBus().get(
            _ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH
        )
        organize.remove_peer(peer_vk)

    def rename_peer(self, peer_vk: str, name: str) -> None:
        organize: Organize = pydbus.SystemBus().get(
            _ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH
        )
        organize.set_peer(peer_vk, ["petname"], name)

    def pin_and_verify(self, peer_vk: str, peer_name: str) -> None:
        organize: Organize = pydbus.SystemBus().get(
            _ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH
        )
        organize.verify_and_pin_peer(peer_vk, peer_name)

    def add_peer(self, peer_vk: str, ip: str) -> None:
        organize: Organize = pydbus.SystemBus().get(
            _ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH
        )
        organize.peer_addr_add(peer_vk, ip)

    def set_pref(self, pref: str, value: Any) -> str:
        organize: Organize = pydbus.SystemBus().get(
            _ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH
        )
        return organize.set_pref(pref, value)

    def add_pref(self, pref: str, value: Any) -> str:
        organize: Organize = pydbus.SystemBus().get(
            _ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH
        )
        return organize.add_pref(pref, value)

    def remove_pref(self, pref: str, value: Any) -> str:
        organize: Organize = pydbus.SystemBus().get(
            _ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH
        )
        return organize.remove_pref(pref, value)
