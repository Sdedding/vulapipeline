import tkinter as tk
from typing import Callable

from vula.frontend.constants import (
    BACKGROUND_COLOR,
    FONT,
    FONT_SIZE_TEXT_XL,
    IMAGE_BASE_PATH,
    TEXT_COLOR_WHITE,
)
from vula.frontend.components.icon_button import IconButton


class CommandBar(tk.Frame):
    def __init__(
        self,
        parent: tk.Misc,
        status_text: str,
        on_rediscover: Callable[[], None],
        on_repair: Callable[[], None],
        on_release_gate: Callable[[], None],
        on_help: Callable[[], None],
    ) -> None:
        super().__init__(parent, bg=BACKGROUND_COLOR, height=50)

        tk.Label(
            self,
            text=status_text,
            bg=BACKGROUND_COLOR,
            fg=TEXT_COLOR_WHITE,
            font=(FONT, FONT_SIZE_TEXT_XL),
        ).grid(row=0, column=0, sticky="w")

        IconButton(
            self, IMAGE_BASE_PATH + "rediscover.png", on_rediscover, padx=30
        ).grid_at(0, 1)
        IconButton(self, IMAGE_BASE_PATH + "repair.png", on_repair).grid_at(
            0, 2
        )
        IconButton(
            self, IMAGE_BASE_PATH + "release_gateway.png", on_release_gate
        ).grid_at(0, 3)
        IconButton(self, IMAGE_BASE_PATH + "help.png", on_help).grid_at(0, 4)
