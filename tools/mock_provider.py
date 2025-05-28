"""
In-memory **MockDataProvider** – zero external dependencies.

Implements the full :class:`vula.frontend.api.DataProvider` contract and is
used by tests, CI screenshots and local development.
"""
from __future__ import annotations

from pathlib import Path
from typing import Callable, Dict, List, Set

import sys

# ensure project root importable in loose dev checkouts
ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT)) if str(ROOT) not in sys.path else None

from vula.frontend.api import Callback, DataProvider
from vula.frontend.datadomain import (
    Peer,
    PeerStatus,
    Prefs,
    ServiceStatus,
    SystemState,
)

__all__ = ["MockDataProvider"]


class MockDataProvider(DataProvider):
    """A fully-featured, mutable in-memory backend for the GUI."""

    _SUBS: Dict[str, List[Callback]] = {k: [] for k in ("peers_changed", "prefs_changed", "status_changed")}

    def __init__(self) -> None:
        self.peers: List[Peer] = []
        self.prefs = Prefs(
            pin_new_peers=False,
            accept_nonlocal=False,
            auto_repair=True,
            ephemeral_mode=False,
            accept_default_route=False,
            record_events=True,
            overwrite_unpinned=False,
            enable_ipv4=True,
            enable_ipv6=True,
            subnets_allowed=[],
            subnets_forbidden=[],
            iface_prefix_allowed=[],
            local_domains=[],
            expire_time=0,
            primary_ip=None,
        )
        self.status = ServiceStatus("-", "-", "-")
        self.system = SystemState({}, {}, bytes(32), [], True)

    # ───────────────────────── read ─────────────────────────
    def get_peers(self) -> List[Peer]:  # noqa: D401
        return list(self.peers)

    def get_prefs(self) -> Prefs:
        return self.prefs

    def get_status(self) -> ServiceStatus:
        return self.status

    def our_latest_descriptors(self) -> str:
        return "{}"

    def get_system_state(self) -> SystemState:
        return self.system

    # ───────────────────────── mutating ops ─────────────────────────
    def rename_peer(self, peer_id: str, new_name: str) -> None:
        for p in self.peers:
            if p.id == peer_id:
                p.name = new_name
                break
        self._emit("peers_changed")

    def delete_peer(self, peer_id: str) -> None:
        self.peers = [p for p in self.peers if p.id != peer_id]
        self._emit("peers_changed")

    def verify_and_pin_peer(self, peer_id: str, hostname: str) -> None:  # hostname ignored here
        for p in self.peers:
            if p.id == peer_id:
                p.status |= {PeerStatus.PINNED, PeerStatus.VERIFIED}
                break
        self._emit("peers_changed")

    def peer_addr_add(self, peer_id: str, ip: str) -> None:
        for p in self.peers:
            if p.id == peer_id and ip not in p.allowed_ips:
                p.allowed_ips.append(ip)
                break
        self._emit("peers_changed")

    def peer_addr_del(self, peer_id: str, ip: str) -> None:
        for p in self.peers:
            if p.id == peer_id and ip in p.allowed_ips:
                p.allowed_ips.remove(ip)
                break
        self._emit("peers_changed")

    def set_pref(self, key: str, value) -> None:
        if hasattr(self.prefs, key):
            setattr(self.prefs, key, value)
        self._emit("prefs_changed")

    def add_pref(self, key: str, value: str) -> None:
        lst = getattr(self.prefs, key, None)
        if isinstance(lst, list) and value not in lst:
            lst.append(value)
            self._emit("prefs_changed")

    def remove_pref(self, key: str, value: str) -> None:
        lst = getattr(self.prefs, key, None)
        if isinstance(lst, list) and value in lst:
            lst.remove(value)
            self._emit("prefs_changed")

    # ───────────────────────── observer ─────────────────────────
    def subscribe(self, event: str, callback: Callback) -> None:
        if event not in self._SUBS:
            raise ValueError(event)
        self._SUBS[event].append(callback)

    def _emit(self, event: str) -> None:
        for cb in list(self._SUBS.get(event, [])):
            try:
                cb()
            except Exception as exc:  # noqa: BLE001
                print("observer callback failed:", exc)
