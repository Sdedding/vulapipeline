import gettext
import tkinter as tk
from tkinter import Frame
from typing import Any

from vula import common
from vula.frontend import DataProvider
from vula.frontend.components.header import Header
from vula.frontend.components.footer import Footer
from vula.frontend.components.command_bar import CommandBar
from vula.frontend.constants import (
    BACKGROUND_COLOR,
    HEIGHT,
    WIDTH,
)
from vula.frontend.overlay import (
    HelpOverlay,
    VerificationKeyOverlay,
)
from vula.frontend.components.prefs_panel import PrefsPanel
from vula.frontend.components.peers_panel import PeersPanel

_ = gettext.gettext


class App(tk.Tk):
    def __init__(self, *args: Any, **kwargs: Any) -> None:
        super().__init__(*args, **kwargs)

        self.geometry("{}x{}".format(WIDTH, HEIGHT))
        self.config(bg=BACKGROUND_COLOR)

        data = DataProvider()

        # create all of the main containers
        header_frame = Header(self)

        content_frame = Frame(
            self, bg=BACKGROUND_COLOR, width=1200, height=600, padx=3, pady=3
        )
        footer_frame = Footer(
            self,
            on_show_vk=self.open_vk_qr_code,
        )

        # layout all of the main containers
        self.grid_rowconfigure(1, weight=1)
        self.grid_rowconfigure(3, weight=0)
        self.grid_columnconfigure(0, weight=1)

        header_frame.grid(row=0, sticky="ew")
        content_frame.grid(row=1, sticky="nsew")
        footer_frame.grid(row=2, sticky="e")

        # Display the status at the bottom
        state = data.get_status() or {
            "publish": "no status available",
            "discover": "no status available",
            "organize": "no status available",
        }
        status_text = (
            f'Publish: {_(state["publish"])}   '
            f'Discover: {_(state["discover"])}   '
            f'Organize: {_(state["organize"])}'
        )

        cmd_bar = CommandBar(
            self,
            status_text=status_text,
            on_rediscover=self.rediscover,
            on_repair=self.repair,
            on_release_gate=self.release_gateway,
            on_help=HelpOverlay(
                self
            ).openNewWindow,
        )
        cmd_bar.grid(row=3, sticky="s")

        # create the center widgets
        content_frame.grid_rowconfigure(0, weight=1)
        content_frame.grid_columnconfigure(1, weight=1)

        self.prefs = PrefsPanel(content_frame)
        self.prefs.grid(row=0, column=0, sticky="ns", padx=3, pady=3)

        self.peers_new = PeersPanel(content_frame)
        self.peers_new.grid(row=0, column=1, sticky="nsew", padx=3, pady=3)
        self.peers_new.grid_columnconfigure(0, weight=1)

    def rediscover(self) -> None:
        common.organize_dbus_if_active().rediscover()

    def release_gateway(self) -> None:
        common.organize_dbus_if_active().release_gateway()

    def repair(self) -> None:
        common.organize_dbus_if_active().sync(True)

    def open_vk_qr_code(self) -> None:
        VerificationKeyOverlay(self)


def main() -> None:
    try:
        app = App()
        app.title("Vula")
        app.mainloop()
    except RuntimeError as e:
        print(e)
