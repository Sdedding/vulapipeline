from tkinter import ttk

from PIL.ImageOps import expand

from vula.frontend import dataprovider


class PreferencesView(ttk.Frame):
    def __init__(self, parent, **kwargs):
        super().__init__(parent, **kwargs)
        data = dataprovider.DataProvider()
        # DEBUG visual border:
        self.config(borderwidth = 2, relief = "ridge")

        pref_frame = ttk.Frame(self)
        pref_frame.pack(side = "top", expand = True, fill = "both")
        prefs = data.get_prefs()
        local_domains = prefs.get("local_domains")
        local_domains_string = ""
        while local_domains:
            local_domains_string += local_domains.pop()
        local_domains_label = ttk.Label(pref_frame, text = local_domains_string)
        local_domains_label.pack(side = "top", fill = "both")
