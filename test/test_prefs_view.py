import sys
import types
import tkinter as tk
from unittest.mock import MagicMock
from importlib import import_module

# Stub heavy backend and overlay modules before importing Prefs
backend_stub = types.ModuleType('vula.backend')
backend_stub.OrganizeBackend = type('OrganizeBackend', (), {})
sys.modules['vula.backend'] = backend_stub

overlay_stub = types.ModuleType('vula.frontend.overlay')
overlay_stub.PopupMessage = types.SimpleNamespace(showPopupMessage=lambda *a, **k: None)
overlay_stub.PeerDetailsOverlay = object
overlay_stub.DescriptorOverlay = object
overlay_stub.HelpOverlay = object
overlay_stub.VerificationKeyOverlay = object
sys.modules['vula.frontend.overlay'] = overlay_stub

peers_stub = types.ModuleType('vula.frontend.view.peers')
peers_stub.Peers = object
sys.modules['vula.frontend.view.peers'] = peers_stub

Prefs = import_module('vula.frontend.view.prefs').Prefs
PrefsData = import_module('vula.frontend.dataprovider').Prefs

class FakeText(tk.Text):
    def __init__(self, val):
        self.val = val
    def get(self, a, b):
        return self.val

class FakeButton(tk.Button):
    def __init__(self, text):
        self._text = text
    def __getitem__(self, key):
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

def create_prefs_instance():
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
    inst = Prefs.__new__(Prefs)
    inst.prefs = prefs
    inst.data = MagicMock()
    inst.show_error = lambda res: 0
    inst.get_prefs = lambda: None
    inst.hide_save_cancel = lambda: None
    inst.update_all = lambda: None
    inst.show_editable = True
    return inst

def prepare_widgets(inst, bool_values=None, int_value='60\n'):
    bool_values = bool_values or {}
    inst.widgets = {}
    for name in BOOL_FIELDS:
        inst.widgets[name] = FakeButton(bool_values.get(name, 'False'))
    for name in LIST_FIELDS:
        inst.widgets[name] = FakeText('')
    inst.widgets['expire_time'] = FakeText(int_value)


def test_save_boolean_pref_updates_value():
    inst = create_prefs_instance()
    prepare_widgets(inst, {'pin_new_peers': 'True'})
    inst.save_prefs()
    inst.data.set_pref.assert_any_call('pin_new_peers', 'True')


def test_save_integer_pref_updates_value():
    inst = create_prefs_instance()
    prepare_widgets(inst, int_value='30\n')
    inst.save_prefs()
    inst.data.set_pref.assert_any_call('expire_time', '30\n')
