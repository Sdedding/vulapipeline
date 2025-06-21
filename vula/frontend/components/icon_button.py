import tkinter as tk
from tkinter import PhotoImage
from typing import Callable

from vula.frontend.constants import BACKGROUND_COLOR


class IconButton(tk.Button):
    def __init__(
        self,
        parent: tk.Misc,
        image_path: str,
        callback: Callable[[], None],
        *,
        padx: int = 20,
    ) -> None:
        self._img = PhotoImage(file=image_path)
        super().__init__(
            parent,
            image=self._img,
            command=callback,
            borderwidth=0,
            relief="flat",
            highlightthickness=0,
            background=BACKGROUND_COLOR,
            activebackground=BACKGROUND_COLOR,
        )
        self._grid_padx = padx

    def grid_at(self, row: int, column: int) -> None:
        self.grid(row=row, column=column, padx=self._grid_padx)
