import tkinter as tk
from tkinter import Canvas
from vula.frontend.constants import (
    BACKGROUND_COLOR,
    FONT,
    FONT_SIZE_HEADER,
    TEXT_COLOR_HEADER,
    WIDTH,
)


class Header(tk.Frame):
    def __init__(self, parent: tk.Misc, title: str = "Dashboard") -> None:
        super().__init__(parent, bg=BACKGROUND_COLOR, height=50)

        canvas = Canvas(
            self,
            bg=BACKGROUND_COLOR,
            height=50,
            width=WIDTH,
            bd=0,
            highlightthickness=0,
        )
        canvas.pack(fill="both", expand=True)
        canvas.create_text(
            30,
            10,
            anchor="nw",
            text=title,
            fill=TEXT_COLOR_HEADER,
            font=(FONT, FONT_SIZE_HEADER),
        )
