import os  # noqa: F401
import sys
from logging import DEBUG, INFO, WARN, basicConfig, getLogger  # noqa: F401

import click
import dbus
import pydbus

from . import (  # noqa: F40
    common,
    configure,
    discover,
    discover_alt,
    engine,
    organize,
    peer,
    prefs,
    publish,
    publish_alt,
    status,
    tray,
    verify,
    wg,
)
from .constants import (
    _DATE_FMT,
    _LOG_FMT,
    _ORGANIZE_DBUS_NAME,
    _ORGANIZE_DBUS_PATH,
    _PUBLISH_DBUS_NAME,
    _PUBLISH_DBUS_PATH,
)
from .frontend import ui
from .notclick import Debuggable
from typing import Any, Optional


@click.version_option()
@click.option(
    "-v",
    "--verbose",
    'log_level',
    flag_value=DEBUG,
    help="Set log level DEBUG",
)
@click.option(
    "-q", "--quiet", 'log_level', flag_value=WARN, help="Set log level WARN"
)
@click.option(
    "-i",
    "--info",
    'log_level',
    flag_value=INFO,
    help="Set log level INFO",
)
@click.option(
    '--no-pager',
    is_flag=True,
    help="Disable automatic pager use for commands with long output",
)
@click.group(cls=Debuggable, scope=globals(), invoke_without_command=True)
@click.pass_context
def main(
    ctx: click.Context,
    log_level: Optional[int] = None,
    /,
    *args: Any,
    **kwargs: Any,
) -> None:
    """
    vula tools

    With no arguments, runs "peer show".
    """

    if not log_level:
        log_level = INFO

    basicConfig(
        stream=sys.stdout, level=log_level, datefmt=_DATE_FMT, format=_LOG_FMT
    )

    if ctx.invoked_subcommand is None:
        ctx.invoke(status.main)


@main.command()
@click.option(
    "-q",
    "--quick",
    is_flag=True,
    help="Only wait for services, not configuration",
)
def start(quick: bool) -> None:
    "Activate organize daemon via dbus, and report its status"
    bus = pydbus.SystemBus()
    if bus.dbus.NameHasOwner(_ORGANIZE_DBUS_NAME):  # type: ignore[attr-defined]  # noqa: E501
        click.echo("start: vula d-bus service is already active")
    else:
        click.echo('start: activating vula organize service via dbus')
        bus.get(_ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH)
        # activate publish here so that it will be done by the time status
        # runs.
        bus.get(_PUBLISH_DBUS_NAME, _PUBLISH_DBUS_PATH)
    status.main(
        args=(('--only-systemd',) if quick else ()), standalone_mode=False
    )


@main.command()
def stop() -> Any:
    """Stop vula

    This is a shortcut for systemctl stop vula.slice
    """
    return dbus.Interface(
        dbus.SystemBus().get_object(
            "org.freedesktop.systemd1", "/org/freedesktop/systemd1"
        ),
        "org.freedesktop.systemd1.Manager",
    ).StopUnit('vula.slice', 'replace')


@main.command(short_help="Starts the graphical user interface")
def gui() -> None:
    ui.main()


@main.command(short_help="Ensure that system is configured correctly")
@click.option(
    '-n',
    '--dry-run',
    is_flag=True,
    help="Print what would be done, without doing it",
)
def repair(dry_run: bool) -> None:
    """
    This checks if the system is configured correctly, and (re)configures it if
    it isn't.
    """
    res = common.organize_dbus_if_active().sync(dry_run)
    if res:
        click.echo("\n".join(res))


@main.command()
def rediscover() -> None:
    """
    Tell organize to ask discover for more peers.
    """
    organize = common.organize_dbus_if_active()  # noqa: F811
    click.echo('Discovering on ' + organize.rediscover())


@main.command()
def release_gateway() -> None:
    """
    Stop using vula for the default route.

    This command must be run to roam to a non-vula gateway after using a pinned
    peer as the gateway.

    When the system gateway changes to the IP of a pinned peer, it will
    automatically re-enable that peer as the gateway.
    """
    click.echo(common.organize_dbus_if_active().release_gateway())


for name, value in list(globals().items()):
    if name == 'wg':
        wg = wg.main  # type: ignore
        # wg is accessible in debug mode via Debuggable's magic interface
        # but we don't want it in our command list in the GUI because it isn't
        # supported (it is mostly an incomplete replica of the normal wg tools;
        # users should use those instead).
        continue
    cmd = getattr(value, 'main', None)
    if isinstance(cmd, click.Command):
        main.add_command(cmd, name=name)
try:
    # this gives us a very nice "repl" subcommand.
    # apt install python3-click-repl to enable it.
    from click_repl import register_repl

    register_repl(main)
except ImportError:
    pass

if __name__ == "__main__":
    main()
