import tkinter as tk
from typing import Callable, cast

from vula.frontend.constants import (
    BACKGROUND_COLOR,
    FONT,
    FONT_SIZE_TEXT_L,
    IMAGE_BASE_PATH,
    TEXT_COLOR_WHITE,
)

from vula.frontend.components.icon_button import IconButton
from vula.frontend.overlay import DescriptorOverlay


class Footer(tk.Frame):
    def __init__(
        self,
        parent: tk.Misc,
        on_show_vk: Callable[[], None],
    ) -> None:
        super().__init__(parent, bg=BACKGROUND_COLOR, height=150)

        tk.Label(
            self,
            text="Verification Key:",
            bg=BACKGROUND_COLOR,
            fg=TEXT_COLOR_WHITE,
            font=(FONT, FONT_SIZE_TEXT_L),
        ).grid(row=0, column=0, sticky="w")

        IconButton(
            self,
            IMAGE_BASE_PATH + "show_qr.png",
            on_show_vk,
            padx=10,
        ).grid_at(0, 1)

        tk.Label(
            self,
            text="Descriptor:",
            bg=BACKGROUND_COLOR,
            fg=TEXT_COLOR_WHITE,
            font=(FONT, FONT_SIZE_TEXT_L),
        ).grid(row=1, column=0, sticky="w")

        IconButton(
            self,
            IMAGE_BASE_PATH + "show_qr.png",
            lambda: DescriptorOverlay(
                cast(tk.Tk, self.winfo_toplevel())
            ).openNewWindow(),
            padx=10,
        ).grid_at(1, 1)
