import hashlib
import os
import sys
from typing import cast

import click

from vula.utils import optional_import

ggwave = optional_import("ggwave")
pyaudio = optional_import("pyaudio")

_GGWAVE_VOLUME: int = (
    100  # Full ggwave volume, use the os controls for adjustments
)


class VerifyAudio:
    def __init__(self, verbose: bool):
        self.verbose = verbose
        self.null_fds: list[int]
        self.store_fds: list[int]

        if not self.verbose:
            self.__mute()

        try:
            assert pyaudio is not None
            self.py_audio = pyaudio.PyAudio()
        except AssertionError:
            click.echo("pyaudio error")
            sys.exit(4)

        if not self.verbose:
            self.__unmute()

        try:
            assert ggwave is not None
            self.ggwave_instance = ggwave.init()
        except AssertionError:
            click.echo("ggwave error")
            sys.exit(5)

    def send_verification_key(self, vk: str) -> None:
        if self.verbose:
            click.echo(
                "Speaking hash: "
                + hashlib.sha256(vk.encode('utf-8')).hexdigest()
            )

        try:
            assert ggwave is not None
            waveform = ggwave.encode(
                hashlib.sha256(vk.encode('utf-8')).hexdigest(),
                protocolId=1,
                volume=_GGWAVE_VOLUME,
            )
        except AssertionError:
            click.echo("ggwave error")
            sys.exit(5)
        if pyaudio is not None:
            stream = self.py_audio.open(
                format=pyaudio.paFloat32,
                channels=1,
                rate=48000,
                output=True,
                frames_per_buffer=4096,
            )
            try:
                stream.write(waveform, len(waveform) // 4)
                stream.stop_stream()
            except SystemError as e:
                click.echo(f"pyaudio error: {e}")
                sys.exit(5)
            finally:
                stream.close()
                self.py_audio.terminate()

    def receive_verification_key(self) -> str | None:
        if pyaudio is not None:
            stream = self.py_audio.open(
                format=pyaudio.paFloat32,
                channels=1,
                rate=48000,
                input=True,
                frames_per_buffer=1024,
            )
        assert ggwave is not None
        try:
            count = 0
            while True:
                count += 1
                data = stream.read(1024, exception_on_overflow=False)
                click.echo(f"receiving audio data...{count}\r", nl=False)
                if not self.verbose:
                    self.__mute()
                res = ggwave.decode(self.ggwave_instance, data)
                if not self.verbose:
                    self.__unmute()
                if res is not None:
                    try:
                        return cast(str, res.decode("utf-8"))
                    except ValueError:
                        return None

        except KeyboardInterrupt:
            return None
        finally:
            ggwave.free(self.ggwave_instance)
            stream.stop_stream()
            stream.close()
            self.py_audio.terminate()
            return None

    def __mute(self) -> None:
        # Open a pair of null files
        self.null_fds = [os.open(os.devnull, os.O_RDWR) for x in range(2)]

        # Save the actual stdout (1) and stderr (2) file descriptors.
        self.store_fds = [os.dup(1), os.dup(2)]

        # Redirect descriptors
        os.dup2(self.null_fds[0], 1)
        os.dup2(self.null_fds[1], 2)

    def __unmute(self) -> None:
        # Reverse redirect
        os.dup2(self.store_fds[0], 1)
        os.dup2(self.store_fds[1], 2)

        # Make sure all fds are closed
        for fd in self.null_fds + self.store_fds:
            os.close(fd)
