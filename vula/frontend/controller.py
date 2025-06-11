from __future__ import annotations

from typing import List, Optional

from vula import common

from .dataprovider import DataProvider, Peer, get_provider


class Controller:
    """Mediator between the Tkinter views and the :class:`DataProvider`."""

    def __init__(self, provider: Optional[DataProvider] = None) -> None:
        self.provider = provider or get_provider()

    # retrieval helpers -------------------------------------------------
    def get_peers(self) -> List[Peer]:
        return self.provider.get_peers()

    def get_peer(self, peer_id: str) -> Optional[Peer]:
        return self.provider.get_peer(peer_id)

    def get_prefs(self):
        return self.provider.get_prefs()

    def get_status(self):
        return self.provider.get_status()

    def our_latest_descriptors(self):
        return self.provider.our_latest_descriptors()

    # peer actions ------------------------------------------------------
    def delete_peer(self, peer_id: str) -> bool:
        peer = self.get_peer(peer_id)
        if peer and peer.status and "unpinned" not in peer.status:
            return False
        try:
            self.provider.delete_peer(peer_id)
        except Exception:
            return False
        return True

    def rename_peer(self, peer_id: str, name: str) -> None:
        self.provider.rename_peer(peer_id, name)

    def pin_and_verify(self, peer_id: str, name: str) -> None:
        self.provider.pin_and_verify(peer_id, name)

    def add_peer_ip(self, peer_id: str, ip: str) -> None:
        self.provider.add_peer(peer_id, ip)

    # preference actions ------------------------------------------------
    def set_pref(self, pref, value):
        return self.provider.set_pref(pref, value)

    def add_pref(self, pref, value):
        return self.provider.add_pref(pref, value)

    def remove_pref(self, pref, value):
        self.provider.remove_pref(pref, value)

    # system actions ---------------------------------------------------
    def rediscover(self) -> None:
        """Trigger rediscover via the organize DBus service."""
        common.organize_dbus_if_active().rediscover()

    def release_gateway(self) -> None:
        """Release the current gateway via DBus."""
        common.organize_dbus_if_active().release_gateway()

    def repair(self) -> None:
        """Synchronize organize's state via DBus."""
        common.organize_dbus_if_active().sync(True)
