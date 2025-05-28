"""DBus‑backed DataProvider – production‑ready implementation."""

from __future__ import annotations
from datetime import datetime
from typing import List, Dict, Set, Callable, Any

import logging
import re
import yaml
import pydbus
from gi.repository import GLib

from .datadomain import Peer, Prefs, ServiceStatus, SystemState, PeerStatus
from .api import DataProvider, Callback

log = logging.getLogger("vula.frontend.dbus_provider")

# ───────────────────────── helpers ─────────────────────────

def _parse_status(raw: str) -> Set[PeerStatus]:
    """Convert space-separated string of status flags to a set of PeerStatus enums."""
    return {PeerStatus(flag) for flag in raw.split() if flag}

# robust peer‑field regex:  <key>:<space><value>
_KV = re.compile(r"^([^:]+):\s+(.*)$")

_DEF_BOOL_KEYS = {
    "pin_new_peers",
    "accept_nonlocal",
    "auto_repair",
    "ephemeral_mode",
    "accept_default_route",
    "record_events",
    "overwrite_unpinned",
    "enable_ipv4",
    "enable_ipv6",
}

_LIST_KEYS = {
    "subnets_allowed",
    "subnets_forbidden",
    "iface_prefix_allowed",
    "local_domains",
}

# ───────────────────────── provider ───────────────────────

class RealDataProvider(DataProvider):
    """Production data‑source that talks to *vula‑organize* over DBus."""
    
    _SUBS: Dict[str, List[Callback]] = {k: [] for k in (
        "peers_changed", "prefs_changed", "status_changed")}

    def __init__(self) -> None:
        self._bus = pydbus.SystemBus()
        from vula.constants import _ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH
        self._organize = self._bus.get(_ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH)
        self._attach_signals()

    # ───── signal -> observer ─────
    def _attach_signals(self):
        """Hook up DBus signals to trigger observer callbacks."""
        try:
            self._bus.subscribe(
                object="/local/vula/organize",
                signal="PrefsChanged",
                signal_fired=lambda *_: self._emit("prefs_changed"),
            )
            self._bus.subscribe(
                object="/local/vula/organize",
                signal="PeersChanged",
                signal_fired=lambda *_: self._emit("peers_changed"),
            )
        except Exception as exc:  # noqa: BLE001
            log.warning("DBus signal hookup failed: %s", exc)

    # ───── read ─────
    def get_peers(self) -> List[Peer]:
        """Fetch and parse all peers from the DBus service."""
        try:
            raw_ids: List[str] = self._organize.peer_ids("all")
            return [self._map_peer(self._organize.show_peer(pid)) for pid in raw_ids]
        except Exception as exc:
            log.error("get_peers DBus error: %s", exc)
            return []

    def get_prefs(self) -> Prefs:
        """Get and parse preferences from the DBus service."""
        yaml_str: str = self._safe_call(self._organize.show_prefs, "")
        try:
            data: Dict[str, Any] = yaml.safe_load(yaml_str) or {}
        except yaml.YAMLError as exc:
            log.error("Cannot parse prefs YAML: %s", exc)
            data = {}
        for k in _DEF_BOOL_KEYS:
            data.setdefault(k, False)
        for k in _LIST_KEYS:
            data.setdefault(k, [])
        data.setdefault("expire_time", 0)
        data.setdefault("primary_ip", None)
        return Prefs(
            pin_new_peers=data["pin_new_peers"],
            accept_nonlocal=data["accept_nonlocal"],
            auto_repair=data["auto_repair"],
            ephemeral_mode=data["ephemeral_mode"],
            accept_default_route=data["accept_default_route"],
            record_events=data["record_events"],
            overwrite_unpinned=data["overwrite_unpinned"],
            enable_ipv4=data.get("enable_ipv4", True),
            enable_ipv6=data.get("enable_ipv6", True),
            subnets_allowed=data["subnets_allowed"],
            subnets_forbidden=data["subnets_forbidden"],
            iface_prefix_allowed=data["iface_prefix_allowed"],
            local_domains=data["local_domains"],
            expire_time=int(data["expire_time"]),
            primary_ip=data.get("primary_ip"),
        )

    def get_status(self) -> ServiceStatus:
        """Get the status of Vula services."""
        return self._safe_call(self._raw_status, ServiceStatus("?", "?", "?"))

    def _raw_status(self):
        """Get the raw systemd service status."""
        systemd = self._bus.get(".systemd1", "/")
        def _state(unit: str) -> str:
            try:
                return self._bus.get(".systemd1", systemd.GetUnit(unit)).ActiveState
            except Exception:
                return "unknown"
        return ServiceStatus(
            publish=_state("vula-publish.service"),
            discover=_state("vula-discover.service"),
            organize=_state("vula-organize.service"),
        )

    def our_latest_descriptors(self) -> str:
        """Get the latest peer descriptors for this node."""
        return self._safe_call(self._organize.our_latest_descriptors, "{}")

    def get_system_state(self) -> SystemState:
        """Get detailed system state information."""
        raw = self._safe_call(lambda: yaml.safe_load(self._organize.dump_state(False)), {})
        sys_part = raw.get("system_state", {}) if isinstance(raw, dict) else {}
        return SystemState(
            current_subnets=sys_part.get("current_subnets", {}),
            current_interfaces=sys_part.get("current_interfaces", {}),
            our_wg_pk=bytes(32),
            gateways=[],
            has_v6=sys_part.get("has_v6", True),
        )

    # ───── peer / pref ops ─────
    def rename_peer(self, peer_id: str, new_name: str) -> None:
        """Rename a peer in the system."""
        self._safe_call(self._organize.set_peer, None, peer_id, ["petname"], new_name)
        self._emit("peers_changed")

    def delete_peer(self, peer_id: str) -> None:
        """Delete a peer from the system."""
        self._safe_call(self._organize.remove_peer, None, peer_id)
        self._emit("peers_changed")

    def verify_and_pin_peer(self, peer_id: str, hostname: str) -> None:
        """Mark a peer as verified and pinned."""
        self._safe_call(self._organize.verify_and_pin_peer, None, peer_id, hostname)
        self._emit("peers_changed")

    def peer_addr_add(self, peer_id: str, ip: str) -> None:
        """Add an IP address to a peer."""
        self._safe_call(self._organize.peer_addr_add, None, peer_id, ip)
        self._emit("peers_changed")

    def peer_addr_del(self, peer_id: str, ip: str) -> None:
        """Remove an IP address from a peer."""
        self._safe_call(self._organize.peer_addr_del, None, peer_id, ip)
        self._emit("peers_changed")

    def set_pref(self, key: str, value):
        """Set a preference value."""
        self._safe_call(self._organize.set_pref, None, key, str(value))
        self._emit("prefs_changed")

    def add_pref(self, key: str, value: str):
        """Add a value to a preference list."""
        self._safe_call(self._organize.add_pref, None, key, value)
        self._emit("prefs_changed")

    def remove_pref(self, key: str, value: str):
        """Remove a value from a preference list."""
        self._safe_call(self._organize.remove_pref, None, key, value)
        self._emit("prefs_changed")

    # ───── observer logic ─────
    def subscribe(self, event: str, callback: Callback):
        """Subscribe to changes in data."""
        if event not in self._SUBS:
            raise ValueError(event)
        self._SUBS[event].append(callback)

    def _emit(self, event: str):
        """Emit an event to all subscribers."""
        for cb in list(self._SUBS.get(event, [])):
            try:
                cb()
            except Exception as exc:  # noqa: BLE001
                log.warning("observer callback failed: %s", exc)

    # ───── internal parse ─────
    def _map_peer(self, raw: str) -> Peer:
        """Map raw peer output from DBus to a Peer object."""
        kv = {m.group(1).strip(): m.group(2).strip() for line in raw.splitlines() if (m := _KV.match(line))}
        name = kv.get("name")
        pid = kv.get("id") or ""  # should never be empty
        other = kv.get("other names")
        status = _parse_status(kv.get("status", ""))
        allowed_ips = kv.get("allowed_ips", "").split(", ") if kv.get("allowed_ips") else []
        latest_hs = None
        if ts := kv.get("latest_handshake"):
            try:
                latest_hs = datetime.fromisoformat(ts)
            except ValueError:
                pass
        return Peer(
            id=pid,
            name=name or None,
            other_names=other,
            status=status,
            endpoint=kv.get("endpoint"),
            allowed_ips=allowed_ips,
            latest_signature=kv.get("latest_signature"),
            latest_handshake=latest_hs,
            wg_pubkey=kv.get("wg_pubkey"),
        )

    # ───── generic safe wrapper ─────
    def _safe_call(self, fn: Callable, default, *args):
        """Wrap a function call with error handling."""
        try:
            return fn(*args)
        except Exception as exc:  # noqa: BLE001
            log.error("DBus call failed: %s", exc)
            return default
