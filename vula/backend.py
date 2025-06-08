import json
from logging import getLogger

import pydbus

from .constants import _ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH


class OrganizeBackend:
    """Helper maintaining a persistent DBus connection to organize."""

    def __init__(self) -> None:
        self.log = getLogger(__name__)
        self.bus = None
        self.organize = None

    def connect(self, service: bool = True) -> bool:
        """Connect to DBus and optionally to the organize service."""
        if not self.bus:
            try:
                self.bus = pydbus.SystemBus()
            except Exception as exc:  # pragma: no cover - bus not available
                self.log.error("Failed to connect to system DBus: %s", exc)
                self.bus = None
                self.organize = None
                return False
        if service and not self.organize:
            try:
                self.organize = self.bus.get(
                    _ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH
                )
            except Exception as exc:  # pragma: no cover - service missing
                self.log.error("Failed to connect to organize DBus: %s", exc)
                self.organize = None
                return False
        return bool(self.bus and (self.organize or not service))

    def peer_ids(self, which: str = "enabled") -> list[str]:
        if not self.connect():
            return []
        try:
            return list(self.organize.peer_ids(which))
        except Exception as exc:
            self.log.error("DBus error retrieving peer ids: %s", exc)
            return []

    def get_peer_info(self, peer_id: str) -> dict:
        if not self.connect():
            return {}
        try:
            data = self.organize.get_peer_info(peer_id)
            return json.loads(data)
        except Exception as exc:
            self.log.error("DBus error retrieving peer %s: %s", peer_id, exc)
            return {}

    def get_prefs(self) -> dict:
        if not self.connect():
            return {}
        try:
            return json.loads(self.organize.get_prefs())
        except Exception as exc:
            self.log.error("DBus error retrieving prefs: %s", exc)
            return {}

    def our_latest_descriptors(self) -> str:
        if not self.connect():
            return "{}"
        try:
            return self.organize.our_latest_descriptors()
        except Exception as exc:
            self.log.error("DBus error retrieving descriptors: %s", exc)
            return "{}"

    def remove_peer(self, peer_vk: str) -> None:
        if self.connect():
            try:
                self.organize.remove_peer(peer_vk)
            except Exception as exc:
                self.log.error("DBus error removing peer %s: %s", peer_vk, exc)

    def rename_peer(self, peer_vk: str, name: str) -> None:
        if self.connect():
            try:
                self.organize.set_peer(peer_vk, ["petname"], name)
            except Exception as exc:
                self.log.error("DBus error renaming peer %s: %s", peer_vk, exc)

    def verify_and_pin_peer(self, peer_vk: str, peer_name: str) -> None:
        if self.connect():
            try:
                self.organize.verify_and_pin_peer(peer_vk, peer_name)
            except Exception as exc:
                self.log.error("DBus error verifying peer %s: %s", peer_vk, exc)

    def add_peer_ip(self, peer_vk: str, ip: str) -> None:
        if self.connect():
            try:
                self.organize.peer_addr_add(peer_vk, ip)
            except Exception as exc:
                self.log.error("DBus error adding ip for %s: %s", peer_vk, exc)

    def set_pref(self, pref: str, value):
        if not self.connect():
            return None
        try:
            return self.organize.set_pref(pref, value)
        except Exception as exc:
            self.log.error("DBus error setting pref %s: %s", pref, exc)
            return None

    def add_pref(self, pref: str, value):
        if not self.connect():
            return None
        try:
            return self.organize.add_pref(pref, value)
        except Exception as exc:
            self.log.error("DBus error adding pref %s: %s", pref, exc)
            return None

    def remove_pref(self, pref: str, value):
        if self.connect():
            try:
                self.organize.remove_pref(pref, value)
            except Exception as exc:
                self.log.error("DBus error removing pref %s: %s", pref, exc)
