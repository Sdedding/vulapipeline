import tempfile
import tkinter as tk
from tkinter import ttk

from vula.frontend.constants import BACKGROUND_COLOR, TEXT_COLOR_WHITE
from vula.utils import optional_import

qrcode = optional_import("qrcode")


class QRCodeLabel(ttk.Label):
    def __init__(self, parent: ttk.Frame, qr_data: str, resize: int) -> None:
        assert qrcode
        with tempfile.NamedTemporaryFile(
            prefix="vula-ui", suffix="qr-code"
        ) as tmp_file:
            tmp_file_name = tmp_file.name
            super().__init__(parent)
            qr = qrcode.QRCode()
            qr.add_data(data=qr_data)
            img = qr.make_image(
                fill_color=TEXT_COLOR_WHITE, back_color=BACKGROUND_COLOR
            )
            img.save(tmp_file_name)
            self.image = tk.PhotoImage(file=tmp_file_name)
            self.image = self.image.subsample(
                resize, resize  # type: ignore[arg-type, unused-ignore]
            )
            self.configure(image=self.image)
