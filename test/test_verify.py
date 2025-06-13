import json
from unittest.mock import MagicMock, Mock, patch
import sys

from _pytest.capture import CaptureFixture

from vula.prefs import Prefs
from vula.verify import VerifyCommands
from vula.peer import Descriptor
import vula.verify
from vula.common import yamlrepr


class TestVerifyCommands:
    def create_verify_commands(self) -> VerifyCommands:
        test_descriptors = {
            'eth0': str(
                Descriptor(
                    {
                        'r': '',
                        'pk': 'Y' * 43 + '=',
                        'c': 'Z' * 86 + '==',
                        'v4a': '192.168.122.140',
                        'vk': 'A' * 43 + '=',
                        'vf': '1638114786',
                        'dt': '86400',
                        'port': '5354',
                        'hostname': 'vula.local.',
                        'e': 'False',
                        's': 'X' * 86 + '==',
                    }
                )
            ),
            'eth1': str(
                Descriptor(
                    {
                        'r': '',
                        'pk': 'G' * 43 + '=',
                        'c': 'H' * 86 + '==',
                        'v4a': '192.168.122.150',
                        'vk': 'A' * 43 + '=',
                        'vf': '1638114786',
                        'dt': '86400',
                        'port': '5354',
                        'hostname': 'vula.local.',
                        'e': 'False',
                        's': 'X' * 86 + '==',
                    }
                )
            ),
        }

        mockdbus = Mock()
        mockorganize = Mock()
        mockorganize.our_latest_descriptors.return_value = json.dumps(
            test_descriptors
        )
        mockorganize.show_prefs.return_value = repr(yamlrepr(Prefs()))
        mockdbus.SystemBus.return_value.get.return_value = mockorganize

        mockclickctx = MagicMock()
        mockclickctx.meta.get.return_value.get.return_value = None

        with patch("vula.verify.pydbus", mockdbus):
            # the __wrapped__ call gets rid of the click decorators.
            verify: VerifyCommands = vula.verify.VerifyCommands.__wrapped__(mockclickctx)  # type: ignore[attr-defined]  # noqa: E501

        return verify

    def test_init_loads_descriptors(self) -> None:
        verify = self.create_verify_commands()
        assert verify.vk is not None
        assert verify.my_descriptors is not None

    def test_my_vk(
        self, monkeypatch: MagicMock, capsys: CaptureFixture[str]
    ) -> None:
        monkeypatch.setattr(sys.stdout, 'isatty', lambda: True)
        verify = self.create_verify_commands()
        verify.my_vk()
        output = capsys.readouterr().out
        assert 'A' * 43 + '=' in output
        assert 'Y' * 43 + '=' not in output
        assert 'X' * 86 + '=' not in output
        assert 'Z' * 86 + '=' not in output

    def test_my_descriptor(
        self, monkeypatch: MagicMock, capsys: CaptureFixture[str]
    ) -> None:
        monkeypatch.setattr(sys.stdout, 'isatty', lambda: True)
        verify = self.create_verify_commands()
        verify.my_descriptor()
        output = capsys.readouterr().out
        assert 'Descriptor for eth0:' in output
        assert 'Descriptor for eth1:' in output
