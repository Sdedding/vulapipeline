from tkinter import ttk
from vula.frontend import dataprovider


class PeersView(ttk.Frame):
    def __init__(self, parent, **kwargs):
        super().__init__(parent, **kwargs)
        self.data = dataprovider.DataProvider()
        # DEBUG visual border:
        self.config(borderwidth=2, relief="ridge")


        self.peers_list = ttk.Frame(self)

        peers = self.data.get_peers()
        self.num_peers = len(peers)

        for peer in peers:
            name = peer.get("name")
            name_label = ttk.Label(self.peers_list, text=name)
            name_label.pack(side="top")

        self.peers_list.pack(side="top")


