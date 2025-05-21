from tkinter import ttk
from vula.frontend.ncomponents import descriptor_qr as descriptor

class QRView(ttk.Frame):
    def __init__(self, parent, **kwargs):
        super().__init__(parent, **kwargs)
        self.qr_frame = ttk.Frame(self)
        qr_image = descriptor.DescriptorQR(self.qr_frame)
        # DEBUG visual border:
        qr_image.pack()
        self.config(borderwidth=2, relief="ridge")
        self.qr_frame.pack(fill="both", expand=True)