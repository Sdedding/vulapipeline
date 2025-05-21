from tkinter import ttk


class StatusView(ttk.Frame):
    def __init__(self, parent, **kwargs):
        super().__init__(parent, **kwargs)
        # DEBUG visual border:
        self.config(borderwidth=2, relief="ridge")
