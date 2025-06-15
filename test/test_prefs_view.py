import sys
import types
from types import ModuleType, SimpleNamespace
import tkinter as tk
from unittest.mock import MagicMock
from importlib import import_module
from typing import Any, TYPE_CHECKING, cast
import builtins

# Stub heavy backend and overlay modules before importing Prefs
backend_stub = ModuleType('vula.backend')
setattr(backend_stub, 'OrganizeBackend', type('OrganizeBackend', (), {}))
sys.modules['vula.backend'] = backend_stub

overlay_stub = ModuleType('vula.frontend.overlay')
setattr(
    overlay_stub,
    'PopupMessage',
    SimpleNamespace(showPopupMessage=lambda *a, **k: None),
)
setattr(overlay_stub, 'PeerDetailsOverlay', object)
setattr(overlay_stub, 'DescriptorOverlay', object)
setattr(overlay_stub, 'HelpOverlay', object)
setattr(overlay_stub, 'VerificationKeyOverlay', object)
sys.modules['vula.frontend.overlay'] = overlay_stub

peers_stub = ModuleType('vula.frontend.view.peers')
setattr(peers_stub, 'Peers', object)
sys.modules['vula.frontend.view.peers'] = peers_stub

pkg_resources_stub = types.ModuleType('pkg_resources')
setattr(pkg_resources_stub, 'resource_filename', lambda *a, **k: '')
sys.modules['pkg_resources'] = pkg_resources_stub

setattr(builtins, 'pkg_resources', pkg_resources_stub)
setattr(builtins, 'gettext', import_module('gettext'))

pydbus_stub = types.ModuleType('pydbus')
setattr(pydbus_stub, 'SystemBus', lambda *a, **k: types.SimpleNamespace())
sys.modules['pydbus'] = pydbus_stub

yaml_stub = types.ModuleType('yaml')
setattr(yaml_stub, 'safe_load', lambda *a, **k: {})
setattr(yaml_stub, 'safe_dump', lambda *a, **k: '')
sys.modules['yaml'] = yaml_stub

schema_stub = types.ModuleType('schema')


class DummySchema:
    def __init__(self, *a: Any, **k: Any) -> None:
        pass


setattr(schema_stub, 'And', DummySchema)
setattr(schema_stub, 'Or', DummySchema)
setattr(schema_stub, 'Schema', DummySchema)


class SchemaError(Exception):
    pass


setattr(schema_stub, 'SchemaError', SchemaError)
setattr(schema_stub, 'Use', lambda f: f)
setattr(schema_stub, 'Optional', DummySchema)
setattr(schema_stub, 'Regex', DummySchema)
sys.modules['schema'] = schema_stub

nacl_stub = types.ModuleType('nacl')
nacl_exceptions_stub = types.ModuleType('nacl.exceptions')
setattr(
    nacl_exceptions_stub,
    'BadSignatureError',
    type('BadSignatureError', (), {}),
)
nacl_signing_stub = types.ModuleType('nacl.signing')
setattr(nacl_signing_stub, 'SigningKey', object)
setattr(nacl_signing_stub, 'VerifyKey', object)
sys.modules['nacl'] = nacl_stub
sys.modules['nacl.exceptions'] = nacl_exceptions_stub
sys.modules['nacl.signing'] = nacl_signing_stub

pyroute2_stub = types.ModuleType('pyroute2')
IPRoute = type('IPRoute', (), {})
IPRSocket = type('IPRSocket', (), {})
WireGuard = type('WireGuard', (), {})
setattr(pyroute2_stub, 'IPRoute', IPRoute)
setattr(pyroute2_stub, 'IPRSocket', IPRSocket)
setattr(pyroute2_stub, 'WireGuard', WireGuard)
sys.modules['pyroute2'] = pyroute2_stub
netlink_stub = types.ModuleType('pyroute2.netlink')
setattr(netlink_stub, 'nla', object)
sys.modules['pyroute2.netlink'] = netlink_stub

notclick_stub = types.ModuleType('vula.notclick')


class DualUse:
    @staticmethod
    def object(cls: Any | None = None, *a: Any, **k: Any) -> Any:
        def wrapper(f: Any) -> Any:
            return f

        return wrapper

    @staticmethod
    def method(name: str | None = None, *a: Any, **k: Any) -> Any:
        def wrapper(f: Any) -> Any:
            return f

        return wrapper


setattr(notclick_stub, 'DualUse', DualUse)
setattr(notclick_stub, 'blue', lambda s: s)
setattr(notclick_stub, 'green', lambda s: s)
setattr(notclick_stub, 'red', lambda s: s)
setattr(notclick_stub, 'yellow', lambda s: s)
setattr(notclick_stub, 'bold', lambda s: s)
setattr(notclick_stub, 'echo_maybepager', lambda s: None)
setattr(notclick_stub, 'shell_complete_helper', lambda fn: {})
sys.modules['vula.notclick'] = notclick_stub

wg_stub = types.ModuleType('vula.wg')


class Interface:
    pass


setattr(wg_stub, 'Interface', Interface)
sys.modules['vula.wg'] = wg_stub

verification_stub = types.ModuleType('vula.frontend.view.verification')
setattr(verification_stub, 'VerificationKeyFrame', object)
sys.modules['vula.frontend.view.verification'] = verification_stub

descriptor_stub = types.ModuleType('vula.frontend.view.descriptor')
setattr(descriptor_stub, 'DescriptorFrame', object)
sys.modules['vula.frontend.view.descriptor'] = descriptor_stub

if TYPE_CHECKING:
    from vula.frontend.view.prefs import Prefs as PrefsType
else:
    PrefsType = Any

Prefs = import_module('vula.frontend.view.prefs').Prefs
PrefsData = import_module('vula.frontend.dataprovider').Prefs


class FakeText(tk.Text):
    def __init__(self, val: str) -> None:
        self.val = val

        def dummy(*_a: Any, **_k: Any) -> None:
            pass

        self.tk = types.SimpleNamespace(call=dummy)  # type: ignore[assignment]
        self._w = 'widget'

    def get(self, a: Any, b: Any) -> str:  # type: ignore[override]
        return self.val

    def __getitem__(self, key: str) -> "FakeText":
        return self


class FakeButton(tk.Button):
    def __init__(self, text: str) -> None:
        self._text = text

        def dummy(*_a: Any, **_k: Any) -> None:
            pass

        self.tk = types.SimpleNamespace(call=dummy)  # type: ignore[assignment]
        self._w = 'widget'

    def __getitem__(self, key: str) -> str:
        if key == 'text':
            return self._text
        raise KeyError(key)


BOOL_FIELDS = [
    'pin_new_peers',
    'accept_nonlocal',
    'auto_repair',
    'ephemeral_mode',
    'accept_default_route',
    'record_events',
    'overwrite_unpinned',
]

LIST_FIELDS = [
    'subnets_allowed',
    'subnets_forbidden',
    'iface_prefix_allowed',
    'local_domains',
]


def create_prefs_instance() -> PrefsType:
    prefs = PrefsData(
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
    inst = cast(PrefsType, Prefs.__new__(Prefs))
    inst.prefs = prefs
    inst.data = cast(Any, MagicMock())
    setattr(inst, 'show_error', lambda res: 0)
    setattr(inst, 'get_prefs', lambda: None)
    setattr(inst, 'hide_save_cancel', lambda: None)
    setattr(inst, 'update_all', lambda: None)
    inst.show_editable = True
    return inst


def prepare_widgets(
    inst: PrefsType,
    bool_values: dict[str, str] | None = None,
    int_value: str = "60\n",
) -> None:
    bool_values = bool_values or {}
    inst.widgets = {}
    for name in BOOL_FIELDS:
        inst.widgets[name] = FakeButton(bool_values.get(name, 'False'))
    for name in LIST_FIELDS:
        inst.widgets[name] = FakeText('')
    inst.widgets['expire_time'] = FakeText(int_value)


def test_save_boolean_pref_updates_value() -> None:
    inst = create_prefs_instance()
    prepare_widgets(inst, {'pin_new_peers': 'True'})
    inst.save_prefs()
    cast(MagicMock, inst.data).set_pref.assert_any_call(
        'pin_new_peers', 'True'
    )


def test_save_integer_pref_updates_value() -> None:
    inst = create_prefs_instance()
    prepare_widgets(inst, int_value='30\n')
    inst.save_prefs()
    cast(MagicMock, inst.data).set_pref.assert_any_call('expire_time', '30\n')
