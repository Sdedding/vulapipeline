import tkinter as tk
from typing import Any

from vula.frontend.constants import BACKGROUND_COLOR
from vula.frontend.view import Peers


class PeersPanel(tk.Frame):
    def __init__(self, parent: tk.Misc, *a: Any, **kw: Any) -> None:
        super().__init__(parent, bg=BACKGROUND_COLOR, width=600, height=600)
        self.grid_propagate(False)
        self._peers = Peers(self)
