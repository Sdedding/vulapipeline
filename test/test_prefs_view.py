# test_prefs_views.py
"""
GUI-facing tests for ``Prefs`` with zero global side-effects.

The heavy libraries that Prefs depends on are stubbed *temporarily*
(in a pytest fixture) so they never leak into the rest of the run.
"""
from __future__ import annotations

import importlib
import sys
import types
from types import ModuleType, SimpleNamespace
from typing import Any, Callable, cast

import pytest
import tkinter as tk
from unittest.mock import MagicMock

# --------------------------------------------------------------------------- #
# Constants describing the editable fields we poke at                         #
# --------------------------------------------------------------------------- #

BOOL_FIELDS = [
    "pin_new_peers",
    "accept_nonlocal",
    "auto_repair",
    "ephemeral_mode",
    "accept_default_route",
    "record_events",
    "overwrite_unpinned",
]

LIST_FIELDS = [
    "subnets_allowed",
    "subnets_forbidden",
    "iface_prefix_allowed",
    "local_domains",
]

# --------------------------------------------------------------------------- #
# Helpers – lightweight widget doubles that work without a running Tk         #
# --------------------------------------------------------------------------- #


class _FakeText(tk.Text):
    def __init__(self, value: str) -> None:  # noqa: D401
        # no call to tk.Text.__init__ – avoids opening a real window
        self._value = value
        self.tk = SimpleNamespace(call=lambda *_a, **_k: None)  # type: ignore[attr-defined]
        self._w = "widget"

    def get(self, *_a: Any, **_k: Any) -> str:  # type: ignore[override]
        return self._value

    def __getitem__(self, _key: str) -> "_FakeText":  # type: ignore[override]
        return self


class _FakeButton(tk.Button):
    def __init__(self, text_val: str) -> None:
        self._text_val = text_val
        self.tk = SimpleNamespace(call=lambda *_a, **_k: None)  # type: ignore[attr-defined]
        self._w = "widget"

    def __getitem__(self, key: str) -> str:  # type: ignore[override]
        if key == "text":
            return self._text_val
        raise KeyError(key)


# --------------------------------------------------------------------------- #
# Pytest fixture that builds *and cleans up* the required stubs               #
# --------------------------------------------------------------------------- #


@pytest.fixture
def prefs_factory(monkeypatch: pytest.MonkeyPatch) -> Callable[..., Any]:
    """
    Yield a callable that returns a ready-to-use Prefs instance,
    with all import-time heaviness stubbed only for the lifetime
    of the calling test.
    """
    # --- minimal stubs (only what Prefs touches during import) -------------
    def _new_mod(name: str) -> ModuleType:
        mod = ModuleType(name)
        monkeypatch.setitem(sys.modules, name, mod)
        return mod

    _backend = _new_mod("vula.backend")
    _backend.OrganizeBackend = type("OrganizeBackend", (), {})

    _overlay = _new_mod("vula.frontend.overlay")
    _overlay.PopupMessage = SimpleNamespace(showPopupMessage=lambda *_a, **_k: None)
    for cls in (
        "PeerDetailsOverlay",
        "DescriptorOverlay",
        "HelpOverlay",
        "VerificationKeyOverlay",
    ):
        setattr(_overlay, cls, object)

    _peers_view = _new_mod("vula.frontend.view.peers")
    _peers_view.Peers = object

    _pkg_resources = _new_mod("pkg_resources")
    _pkg_resources.resource_filename = lambda *_a, **_k: ""

    _pydbus = _new_mod("pydbus")
    _pydbus.SystemBus = lambda *_a, **_k: SimpleNamespace()

    _yaml = _new_mod("yaml")
    _yaml.safe_load = lambda *_a, **_k: {}
    _yaml.safe_dump = lambda *_a, **_k: ""

    _schema = _new_mod("schema")

    class _Dummy:
        def __init__(self, *_a: Any, **_k: Any) -> None: ...

    for attr in ("And", "Or", "Schema", "Optional", "Regex"):
        setattr(_schema, attr, _Dummy)

    class _SchemaError(Exception): ...

    _schema.SchemaError = _SchemaError
    _schema.Use = lambda f: f

    _nacl = _new_mod("nacl")
    _nacl_exc = _new_mod("nacl.exceptions")
    _nacl_sign = _new_mod("nacl.signing")
    _nacl_exc.BadSignatureError = type("BadSignatureError", (), {})
    _nacl_sign.SigningKey = object
    _nacl_sign.VerifyKey = object

    _pyroute2 = _new_mod("pyroute2")
    _pyroute2.IPRoute = type("IPRoute", (), {})
    _pyroute2.IPRSocket = type("IPRSocket", (), {})
    _pyroute2.WireGuard = type("WireGuard", (), {})
    _new_mod("pyroute2.netlink").nla = object

    _notclick = _new_mod("vula.notclick")

    class _DualUse:
        @staticmethod
        def object(_cls: Any | None = None, *_a: Any, **_k: Any) -> Any:
            return lambda f: f

        method = object  # same signature – not needed in the tested calls

    _notclick.DualUse = _DualUse
    for fn in ("blue", "green", "red", "yellow", "bold"):
        setattr(_notclick, fn, lambda s: s)
    _notclick.echo_maybepager = lambda *_a, **_k: None
    _notclick.shell_complete_helper = lambda _fn: {}

    _wg = _new_mod("vula.wg")
    _wg.Interface = type("Interface", (), {})

    _new_mod("vula.frontend.view.verification").VerificationKeyFrame = object
    _new_mod("vula.frontend.view.descriptor").DescriptorFrame = object

    # --- import Prefs now that its dependencies are satisfied --------------
    Prefs = importlib.import_module("vula.frontend.view.prefs").Prefs
    PrefsData = importlib.import_module("vula.frontend.dataprovider").Prefs

    # --------------------------------------------------------------------- #
    # Factory returned to the test functions                                #
    # --------------------------------------------------------------------- #

    def _make_instance(
        bool_values: dict[str, str] | None = None,
        int_val: str = "60\n",
    ) -> Any:
        """Return a Prefs instance whose widgets are pre-seeded with values."""
        bool_values = bool_values or {}

        prefs_obj = PrefsData(
            pin_new_peers=False,
            accept_nonlocal=False,
            auto_repair=False,
            subnets_allowed=[],
            subnets_forbidden=[],
            iface_prefix_allowed=[],
            local_domains=[],
            ephemeral_mode=False,
            accept_default_route=False,
            record_events=False,
            expire_time=60,
            overwrite_unpinned=False,
        )

        inst = cast(Any, Prefs.__new__(Prefs))
        inst.prefs = prefs_obj
        inst.data = cast(Any, MagicMock())
        inst.show_error = lambda *_a, **_k: 0
        inst.get_prefs = lambda *_a, **_k: None
        inst.hide_save_cancel = lambda *_a, **_k: None
        inst.update_all = lambda *_a, **_k: None
        inst.show_editable = True

        inst.widgets = {}
        for name in BOOL_FIELDS:
            inst.widgets[name] = _FakeButton(bool_values.get(name, "False"))
        for name in LIST_FIELDS:
            inst.widgets[name] = _FakeText("")
        inst.widgets["expire_time"] = _FakeText(int_val)
        return inst

    yield _make_instance
    # monkeypatch fixture automatically restores the original sys.modules


# --------------------------------------------------------------------------- #
# Actual tests                                                                #
# --------------------------------------------------------------------------- #


def test_save_boolean_pref_updates_value(prefs_factory: Callable[..., Any]) -> None:
    inst = prefs_factory({"pin_new_peers": "True"})
    inst.save_prefs()
    inst.data.set_pref.assert_any_call("pin_new_peers", "True")


def test_save_integer_pref_updates_value(prefs_factory: Callable[..., Any]) -> None:
    inst = prefs_factory(int_val="30\n")
    inst.save_prefs()
    inst.data.set_pref.assert_any_call("expire_time", "30\n")
