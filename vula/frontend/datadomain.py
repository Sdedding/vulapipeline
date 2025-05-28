from __future__ import annotations

"""Domain objects + thin DBus accessor for *vula* front‑end.

This module now exposes **only** the new dataclass‑based model.
The legacy `TypedDict` shims were removed (Task 1, sub‑step 1.1).
"""

from dataclasses import dataclass, asdict
from datetime import datetime
from enum import Enum
from ipaddress import IPv4Address, IPv6Address, ip_address
from typing import Any, Dict, List, Optional, Set, Union

import pydbus
import yaml

from vula.common import escape_ansi
from vula.constants import _ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH

__all__ = [
    "PeerStatus",
    "Peer",
    "Prefs",
    "ServiceStatus",
    "SystemState",
    "DataDomain",
]


# ───────────────────────────────── enums & dataclasses ─────────────────────────────────

class PeerStatus(str, Enum):
    ENABLED = "enabled"
    DISABLED = "disabled"
    PINNED = "pinned"
    UNPINNED = "unpinned"
    VERIFIED = "verified"
    UNVERIFIED = "unverified"


@dataclass
class Peer:
    id: str
    name: Optional[str]
    other_names: Optional[str]
    status: Set[PeerStatus]
    endpoint: Optional[str]
    allowed_ips: List[str]
    latest_signature: Optional[str]
    latest_handshake: Optional[datetime]
    wg_pubkey: Optional[str]

    # ––––––––––––––– mapping helpers –––––––––––––––
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "Peer":
        status_set: Set[PeerStatus] = {
            PeerStatus(flag.strip())
            for flag in (data.get("status") or "").split(",")
            if flag
        }
        handshake: Optional[datetime] = None
        if ts := data.get("latest_handshake"):
            try:
                handshake = datetime.fromisoformat(ts)
            except ValueError:
                pass
        allowed = (
            data.get("allowed_ips", "").split(",")
            if isinstance(data.get("allowed_ips"), str)
            else data.get("allowed_ips", [])
        )
        return cls(
            id=data.get("id", ""),
            name=data.get("name"),
            other_names=data.get("other_names"),
            status=status_set,
            endpoint=data.get("endpoint"),
            allowed_ips=allowed,
            latest_signature=data.get("latest_signature"),
            latest_handshake=handshake,
            wg_pubkey=data.get("wg_pubkey"),
        )

    def to_dict(self) -> Dict[str, Any]:
        result = asdict(self)
        result["status"] = ",".join(s.value for s in self.status)
        if self.latest_handshake:
            result["latest_handshake"] = self.latest_handshake.isoformat()
        result["allowed_ips"] = ",".join(self.allowed_ips)
        return result


@dataclass
class Prefs:
    pin_new_peers: bool
    accept_nonlocal: bool
    auto_repair: bool
    ephemeral_mode: bool
    accept_default_route: bool
    record_events: bool
    overwrite_unpinned: bool
    enable_ipv4: bool
    enable_ipv6: bool
    subnets_allowed: List[str]
    subnets_forbidden: List[str]
    iface_prefix_allowed: List[str]
    local_domains: List[str]
    expire_time: int
    primary_ip: Optional[str]

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "Prefs":
        """Robust dict → Prefs conversion with sane fallbacks."""
        d = {**data}  # shallow copy
        for key, default in (
            ("pin_new_peers", False),
            ("accept_nonlocal", False),
            ("auto_repair", True),
            ("ephemeral_mode", False),
            ("accept_default_route", False),
            ("record_events", True),
            ("overwrite_unpinned", False),
            ("enable_ipv4", True),
            ("enable_ipv6", True),
            ("subnets_allowed", []),
            ("subnets_forbidden", []),
            ("iface_prefix_allowed", []),
            ("local_domains", []),
            ("expire_time", 0),
            ("primary_ip", None),
        ):
            d.setdefault(key, default)
        return cls(**d)  # type: ignore[arg-type]


@dataclass
class ServiceStatus:
    publish: str
    discover: str
    organize: str

    @classmethod
    def from_dict(cls, data: Dict[str, str]) -> "ServiceStatus":
        return cls(
            publish=data.get("publish", "unknown"),
            discover=data.get("discover", "unknown"),
            organize=data.get("organize", "unknown"),
        )


@dataclass
class SystemState:
    current_subnets: Dict[str, List[str]]
    current_interfaces: Dict[str, List[str]]
    our_wg_pk: bytes
    gateways: List[Union[IPv4Address, IPv6Address]]
    has_v6: bool

    @property
    def current_ips(self):  # noqa: D401
        """Flat list of currently configured IP addresses."""
        return [
            ip_address(ip)
            for ips in self.current_subnets.values()
            for ip in ips
        ]

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "SystemState":
        gws: List[Union[IPv4Address, IPv6Address]] = []
        for gw in data.get("gateways", []):
            try:
                gws.append(ip_address(gw))
            except ValueError:
                pass
        return cls(
            current_subnets=data.get("current_subnets", {}),
            current_interfaces=data.get("current_interfaces", {}),
            our_wg_pk=data.get("our_wg_pk", b""),
            gateways=gws,
            has_v6=data.get("has_v6", False),
        )


# ───────────────────────────────── thin DBus wrapper ─────────────────────────────────

class DataDomain:
    """Synchronous helper that talks to *vula‑organize* via DBus."""

    def __init__(self):
        self._organize = None

    # ––––– lazy DBus proxy –––––
    @property
    def organize(self):
        if self._organize is None:
            try:
                bus = pydbus.SystemBus()
                self._organize = bus.get(
                    _ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH
                )
            except Exception:
                self._organize = None
        return self._organize

    # ––––– public API –––––
    def get_peers(self) -> List[Peer]:
        if not self.organize:
            return []
        try:
            raw: List[str] = self.organize.GetPeers()  # type: ignore[attr-defined]
            return [Peer.from_dict(yaml.safe_load(p)) for p in raw]
        except Exception:
            return []

    def get_service_status(self) -> ServiceStatus:
        if not self.organize:
            return ServiceStatus("unknown", "unknown", "unknown")
        try:
            data: Dict[str, str] = self.organize.GetStatus()  # type: ignore[attr-defined]
            return ServiceStatus.from_dict(data)
        except Exception:
            return ServiceStatus("unknown", "unknown", "unknown")

    def get_preferences(self) -> Prefs:
        if not self.organize:
            return Prefs.from_dict({})
        try:
            yaml_str: str = self.organize.GetPrefs()  # type: ignore[attr-defined]
            data: Dict[str, Any] = yaml.safe_load(yaml_str) or {}
            return Prefs.from_dict(data)
        except Exception:
            return Prefs.from_dict({})


# Convenience singleton (kept for backward compatibility)
data_domain = DataDomain()
