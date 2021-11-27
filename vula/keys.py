from logging import Logger, getLogger

import click

from .common import organize_dbus_if_active
from .notclick import DualUse


@DualUse.object(
    short_help="Modify used keys",
    invoke_without_command=True,
)
@click.pass_context
class KeysCommands(object):
    """
    Commands to modify currently used keys
    """

    def __init__(self, ctx):
        self.organize = (
            ctx.meta.get('Organize', {}).get('magic_instance')
            or organize_dbus_if_active()
        )
        self.log: Logger = getLogger()

    @DualUse.method(short_help="Rotate used keys")
    @click.option(
        '-vk',
        '--verification-key',
        is_flag=True,
        help="Rotate the verification keypair",
    )
    @click.option(
        '-wg',
        '--wireguard-key',
        is_flag=True,
        help="Rotate the wireguard keypair",
    )
    @click.option(
        '-csidh', '--csidh-key', is_flag=True, help="Rotate the csidh keypair"
    )
    def rotate(
        self, csidh_key=False, verification_key=False, wireguard_key=False
    ) -> str:
        return self.organize.rotate_keys(
            csidh_key, verification_key, wireguard_key
        )


main = KeysCommands.cli
