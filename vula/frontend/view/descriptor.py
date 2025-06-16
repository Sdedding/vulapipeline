"""Frame for displaying our descriptor QR codes."""

import json
from tkinter import Frame, ttk
from tkinter.ttk import Frame

from vula.frontend import DataProvider
from vula.frontend.components import QRCodeLabel
from vula.frontend.constants import (
    BACKGROUND_COLOR,
    FONT,
    FONT_SIZE_HEADER,
    FONT_SIZE_TEXT_XXL,
)
from ..style import configure_styles
from vula.peer import Descriptor


class DescriptorFrame(ttk.Notebook):
    """Display descriptor data inside the main window."""

    def __init__(self, parent: ttk.Notebook, controller: DataProvider) -> None:
        super().__init__(parent)
        self.style = configure_styles()
        self.controller = controller
        self._build_ui()

    def _build_ui(self) -> None:
        title_frame = ttk.Frame(self, padding=(10, 10), style="Vula.TFrame")
        text_frame = ttk.Frame(self, padding=(20, 0), style="Vula.TFrame")
        qr_frame: ttk.Frame = ttk.Frame(self, style="Vula.TFrame")

        self.grid_rowconfigure(2, weight=1)
        self.grid_columnconfigure(0, weight=1)

        title_frame.grid(row=0, sticky="nsew")
        text_frame.grid(row=1, sticky="nw")
        qr_frame.grid(row=2, sticky="nsew")

        ttk.Label(
            title_frame,
            text="Descriptor",
            style="Vula.Header.TLabel",
            font=(FONT, FONT_SIZE_HEADER),
        ).pack()

        descriptors = {
            ip: Descriptor.parse(d)
            for ip, d in json.loads(
                self.controller.our_latest_descriptors()
            ).items()
        }

        for ip, desc in descriptors.items():
            ip = str(desc.ip)
            ttk.Label(
                text_frame,
                text=ip,
                style="Vula.TLabel",
                font=(FONT, FONT_SIZE_TEXT_XXL),
            ).pack()

            qr_data = "local.vula:desc:" + str(desc)
            qr_code = QRCodeLabel(parent=qr_frame, qr_data=qr_data, resize=2)
            qr_code.configure(background=BACKGROUND_COLOR)
            qr_code.pack(pady=(0, 10))
