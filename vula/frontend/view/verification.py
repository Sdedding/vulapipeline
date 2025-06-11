import json
from tkinter import Frame, PhotoImage, ttk

from vula.frontend import Controller
from vula.frontend.components import QRCodeLabel
from vula.frontend.constants import (
    BACKGROUND_COLOR,
    FONT,
    FONT_SIZE_HEADER,
    FONT_SIZE_TEXT_L,
    IMAGE_BASE_PATH,
)
from vula.frontend.overlay import PopupMessage
from vula.peer import Descriptor
from ..style import configure_styles


class VerificationKeyFrame(Frame):
    """Display our verification key inside the main window."""

    def __init__(self, parent: Frame, controller: Controller) -> None:
        super().__init__(parent, bg=BACKGROUND_COLOR)
        self.style = configure_styles()
        self.controller = controller
        self._build_ui()

    def _build_ui(self) -> None:
        title_frame = ttk.Frame(self, padding=(10, 10), style="Vula.TFrame")
        text_frame = ttk.Frame(self, padding=(20, 0), style="Vula.TFrame")
        qr_frame = ttk.Frame(self, style="Vula.TFrame")

        self.grid_rowconfigure(2, weight=1)
        self.grid_columnconfigure(0, weight=1)

        title_frame.grid(row=0, sticky="nw")
        text_frame.grid(row=1, sticky="nw")
        qr_frame.grid(row=2, sticky="nsew")

        ttk.Label(
            title_frame,
            text="Verification Key",
            style="Vula.Header.TLabel",
            font=(FONT, FONT_SIZE_HEADER),
        ).pack()

        descriptors = {
            ip: Descriptor.parse(d)
            for ip, d in json.loads(
                self.controller.our_latest_descriptors()
            ).items()
        }

        text_frame.grid_rowconfigure(1, weight=1)
        text_frame.grid_columnconfigure(2, weight=1)
        self.button_image = PhotoImage(file=IMAGE_BASE_PATH + "clipboard.png")

        for desc in descriptors.values():
            vk = str(desc.vk)
            ttk.Label(
                text_frame,
                text=vk,
                style="Vula.TLabel",
                font=(FONT, FONT_SIZE_TEXT_L),
            ).grid(row=0, column=0, sticky="w")

            def command(key: str = vk) -> None:
                self._add_to_clipboard(key)

            ttk.Button(
                text_frame,
                image=self.button_image,
                text="Copy",
                command=command,
                style="Vula.TButton",
            ).grid(row=0, column=1, padx=5)

            qr_data = "local.vula:vk:" + vk
            ttk.Label(
                qr_frame, text="Scan this QR code", style="Vula.TLabel"
            ).pack()
            qr_code = QRCodeLabel(parent=qr_frame, qr_data=qr_data, resize=1)
            qr_code.configure(background=BACKGROUND_COLOR)
            qr_code.pack(pady=(0, 10))

    def _add_to_clipboard(self, text: str) -> None:
        self.clipboard_clear()
        self.clipboard_append(text)
        PopupMessage.showPopupMessage(
            "Information", "Verification key copied to clipboard"
        )
