import tempfile
from tkinter import ttk
import tkinter as tk
from vula.frontend import dataprovider
from vula.peer import Descriptor
import json

try:
    import qrcode
except ImportError:
    qrcode = None


class DescriptorQR(ttk.Frame):
    def __init__(self, parent, **kwargs):
        super().__init__(parent)
        data = dataprovider.DataProvider()


        desc_frame = ttk.Frame(self)
        desc_frame.pack(fill="both", expand=True)
        my_descriptors = {
            ip: Descriptor(d)
            for ip, d in json.loads(data.our_latest_descriptors()).items()
        }

        for ip, desc in my_descriptors.items():
            # IP Label
            label_ip = ttk.Label(
                desc_frame,
                text=ip,
            )

            label_ip.pack()


            # Descriptor QR Code Image
            qr_data = "local.vula:desc:" + str(desc)


            with tempfile.NamedTemporaryFile(
                    prefix="vula-ui", suffix="qr-code"
            ) as tmp_file:
                tmp_file_name = tmp_file.name
                qr = qrcode.QRCode()
                qr.add_data(data=qr_data)
                img = qr.make_image()
                img.save(tmp_file_name)
                self.image = tk.PhotoImage(file=tmp_file_name)
                self.image = self.image.subsample(2, 2)
            qr_code = ttk.Label(desc_frame)
            qr_code.configure(image=self.image)
            qr_code.pack()