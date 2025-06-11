from dataclasses import dataclass
from typing import List, Literal, Optional, TypedDict, cast

from logging import getLogger

from vula.backend import OrganizeBackend


@dataclass
class Peer:
    name: Optional[str]
    id: Optional[str]
    other_names: Optional[str]
    status: Optional[str]
    endpoint: Optional[str]
    allowed_ips: Optional[str]
    latest_signature: Optional[str]
    latest_handshake: Optional[str]
    wg_pubkey: Optional[str]


class StatusType(TypedDict):
    publish: str
    discover: str
    organize: str


@dataclass
class Prefs:
    pin_new_peers: bool
    accept_nonlocal: bool
    auto_repair: bool
    subnets_allowed: list
    subnets_forbidden: list
    iface_prefix_allowed: list
    local_domains: list
    ephemeral_mode: bool
    accept_default_route: bool
    record_events: bool
    expire_time: int
    overwrite_unpinned: bool


def peer_from_dict(data: dict) -> Peer:
    """Convert a raw dict from organize into :class:`Peer`."""
    return Peer(
        name=data.get("name"),
        id=data.get("id"),
        other_names=data.get("other_names"),
        status=data.get("status"),
        endpoint=data.get("endpoint"),
        allowed_ips=data.get("allowed_ips"),
        latest_signature=data.get("latest_signature"),
        latest_handshake=data.get("latest_handshake"),
        wg_pubkey=data.get("wg_pubkey"),
    )


def prefs_from_dict(data: dict) -> Prefs:
    """Convert a raw dict from organize into :class:`Prefs`."""
    return Prefs(
        pin_new_peers=data.get("pin_new_peers", False),
        accept_nonlocal=data.get("accept_nonlocal", False),
        auto_repair=data.get("auto_repair", False),
        subnets_allowed=data.get("subnets_allowed", []),
        subnets_forbidden=data.get("subnets_forbidden", []),
        iface_prefix_allowed=data.get("iface_prefix_allowed", []),
        local_domains=data.get("local_domains", []),
        ephemeral_mode=data.get("ephemeral_mode", False),
        accept_default_route=data.get("accept_default_route", False),
        record_events=data.get("record_events", False),
        expire_time=data.get("expire_time", 0),
        overwrite_unpinned=data.get("overwrite_unpinned", False),
    )


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
    def __init__(self) -> None:
        self.log = getLogger(__name__)
        self.backend = OrganizeBackend()

    def get_peers(self) -> List[Peer]:
        ids = self.backend.peer_ids("enabled")
        peers: List[Peer] = []
        for peer_id in ids:
            info = self.backend.get_peer_info(peer_id)
            peers.append(peer_from_dict(info))
        return peers

    def get_prefs(self) -> Prefs:
        return prefs_from_dict(self.backend.get_prefs())

    def get_peer(self, peer_id: str) -> Optional[Peer]:
        """Return a single peer by id or ``None`` if unavailable."""
        info = self.backend.get_peer_info(peer_id)
        return peer_from_dict(info) if info else None

    def get_status(self) -> Optional[StatusType]:
        # Fetch the data from the systemd dbus
        if not self.backend.connect(service=False):
            return None
        bus = self.backend.bus
        # Explicitly retrieve the systemd manager object to avoid interface
        # resolution issues when calling methods such as ``GetUnit``.
        systemd = bus.get(".systemd1", "/org/freedesktop/systemd1")

        # Create an empty dict for the status
        status = StatusType(publish="", discover="", organize="")

        # Define names to consider in the result
        for name in status.keys():
            name = cast(Literal["publish", "discover", "organize"], name)
            # Template string for service name
            unit_name = "vula-%s.service" % (name,)
            try:
                unit = bus.get(".systemd1", systemd.GetUnit(unit_name))
                status[name] = unit.ActiveState
            except Exception as ex:
                self.log.error("Failed to get unit %s: %s", unit_name, ex)
                return None

        return status

    def our_latest_descriptors(self):
        return self.backend.our_latest_descriptors()

    def delete_peer(self, peer_vk):
        self.backend.remove_peer(peer_vk)

    def rename_peer(self, peer_vk, name):
        self.backend.rename_peer(peer_vk, name)

    def pin_and_verify(self, peer_vk, peer_name):
        self.backend.verify_and_pin_peer(peer_vk, peer_name)

    def add_peer(self, peer_vk, ip):
        self.backend.add_peer_ip(peer_vk, ip)

    def set_pref(self, pref, value):
        return self.backend.set_pref(pref, value)

    def add_pref(self, pref, value):
        return self.backend.add_pref(pref, value)

    def remove_pref(self, pref, value):
        self.backend.remove_pref(pref, value)


def get_provider() -> DataProvider:
    """Return a singleton instance of :class:`DataProvider`."""
    if not hasattr(get_provider, "_instance"):
        get_provider._instance = DataProvider()  # type: ignore[attr-defined]
    return get_provider._instance  # type: ignore[attr-defined]
