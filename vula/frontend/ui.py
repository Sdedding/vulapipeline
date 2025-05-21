import tkinter as tk
from tkinter import ttk

from vula.frontend.nview.preferences_view import PreferencesView as Prefs
from vula.frontend.nview.peers_view import PeersView as Peers
from vula.frontend.nview.status_view import StatusView as Status
from vula.frontend.nview.qr_view import QRView as QR


class App(tk.Tk):
    def __init__(self):
        super().__init__()
        self.title("Vula")
        self.geometry("800x600")

        # two equal columns
        self.columnconfigure(0, weight=1)
        self.columnconfigure(1, weight=1)
        self.rowconfigure(0, weight=1)
        self.rowconfigure(1, weight=0)

        # left + right
        self.peers = Peers(self)
        self.peers.grid(row=0, column=0, sticky="nsew")

        # right: split into prefs on top, QR on bottom
        right = ttk.Frame(self)
        right.grid(row=0, column=1, sticky="nsew")
        right.rowconfigure(0, weight=1)
        right.rowconfigure(1, weight=0)
        self.prefs = Prefs(right)
        self.prefs.grid(row=0, column=0, sticky="nsew")
        self.qr = QR(right)
        self.qr.grid(row=1, column=0, sticky="ew")
        ## self.qr.grid_remove()  # hidden until needed

        # bottom row
        self.status_bar = Status(self)
        self.status_bar.grid(row=1, column=0, columnspan=2, sticky="ew")



