from typing import List, TypedDict, Literal, cast, Optional

from logging import getLogger
import pydbus

from vula.backend import OrganizeBackend


class PeerType(TypedDict):
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


class PrefsType(TypedDict):
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

    def get_peers(self) -> List[PeerType]:
        ids = self.backend.peer_ids("enabled")
        peers: List[PeerType] = []
        for peer_id in ids:
            info = self.backend.get_peer_info(peer_id)
            peers.append(cast(PeerType, info))
        return peers

    def get_prefs(self) -> PrefsType:
        return cast(PrefsType, self.backend.get_prefs())

    def get_status(self) -> Optional[StatusType]:
        # Fetch the data from the systemd dbus
        if not self.backend.connect(service=False):
            return None
        bus = self.backend.bus
        systemd = bus.get(".systemd1", "/")

        # Create an empty dict for the status
        status = StatusType(publish="", discover="", organize="")

        # Define names to consider in the result
        for name in status.keys():
            name = cast(Literal["publish", "discover", "organize"], name)
            # Template string for service name
            unit_name = "vula-%s.service" % (name,)
            try:
                unit = bus.get(
                    ".systemd1", systemd.GetUnit(unit_name)
                )
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
