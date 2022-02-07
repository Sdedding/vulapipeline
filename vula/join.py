"""
  *vula-join* is a daemon thread that uses the REUNION protocol to find and pin peers with a matching passphrase.
  After it has been started, it sends a multicast request to the entered address and waits for a response
  according to the REUNION protocol.
"""

import threading
import click
import time

import pydbus

from .organize import Organize

from .constants import (
    _ORGANIZE_DBUS_NAME,
    _ORGANIZE_DBUS_PATH,
)

from reunion.multicast import UDPListener
from .notclick import DualUse, yellow


@DualUse.object(
    short_help="Use the reunion protocol to find and pin peers using a common passphrase",
    invoke_without_command=True,
)
@click.option(
    "--interval",
    "-I",
    default=60,
    help="Interval at which to start new sessions",
    show_default=True,
)
@click.option("--multicast-group", default="224.3.29.71", show_default=True)
@click.option("--bind-addr", default="0.0.0.0", show_default=True)
@click.option("--port", default=9005, show_default=True)
@click.option(
    "--reveal-once",
    is_flag=True,
    help="Only reveal the message to the first person with the correct passphrase",
)
@click.option("--passphrase", prompt=True, type=str, help="The passphrase")
@click.option("--message", prompt=True, type=str, help="The message")
@click.pass_context
class JoinCommand(object):

    """
    The join command starts the REUNION protocol.
    It also starts the UDP listener daemon thread to listen to any messages regarding the protocol.
    If the passphrase should be changed, the thread/command has to be stopped and started again.
    """

    def __init__(self, ctx, **kw):

        click.echo("join command called")

        assert not kw[
            "reveal_once"
        ], "sorry, the --reveal-once feature has bitrotted at the moment. It should be reimplemented in the ReunionSession object."

        self.update(**kw)

        organize: Organize = ctx.meta.get('Organize', {}).get('magic_instance')

        if not organize:
            bus = pydbus.SessionBus()
            organize = bus.get(_ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH)

        self.organize = organize

        if (
            self._listening_thread is not None
            and self._listening_thread.is_running is True
        ):
            click.echo(yellow("Join is already running"))
            return None

        if (
            self._listening_thread is not None
            and self._listening_thread.is_running is False
        ):
            click.echo(yellow("Join is currently stopping"))
            return None

        self._listening_thread = threading.Thread(target=self.run, kwargs=kw)
        self._listening_thread.is_running = True
        self._listening_thread.daemon = True  # Should be a non blocking thread

        if ctx.invoked_subcommand is None:
            click.echo("join thread started")
            self._listening_thread.start()

    def participant_found(addr: str):

        print("participant_found")

    def run(
        self,
        passphrase,
        message,
        interval,
        multicast_group,
        port,
        reveal_once,
        bind_addr,
    ) -> None:

        passphrase = passphrase.encode()
        message = message.encode()

        udp = UDPListener(bind_addr, 0)
        udp.bind_multicast(multicast_group, port)
        udp.bind_callback(self.participant_found)

        started = None

        while self._listening_thread.is_running is True:
            if started is None or time.time() - started > interval:
                started = time.time()
                udp.new_session(passphrase, message)
                udp.send(b"t1_", udp.session.t1, (multicast_group, port))

            udp.poll()
            udp.poll_multicast()

        self._listening_thread = None
        click.echo("Stopped")

    @DualUse.method()
    def stop(self) -> None:
        """
        This join sub-command stops the UDP listening daemon thread, if there is one
        """

        click.echo("join stop called")

        if self._listening_thread is None:
            click.echo(yellow("Join is not running"))
            return None

        if self._listening_thread.is_running is False:
            click.echo(yellow("Is already stopping"))
            return None

        click.echo("Stopping")
        self._listening_thread.is_running = False


main = JoinCommand.cli

if __name__ == "__main__":
    main()
