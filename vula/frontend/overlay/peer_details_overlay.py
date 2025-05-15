import gettext
import tkinter as tk
from tkinter import messagebox, ttk
from typing import Literal

from vula.frontend import DataProvider, PeerType
from vula.frontend.constants import (

    FONT,
    FONT_SIZE_HEADER,
    FONT_SIZE_TEXT_M,
    IMAGE_BASE_PATH,
)

from vula.frontend.overlay.popupMessage import PopupMessage

_ = gettext.gettext


class PeerDetailsOverlay(ttk.Frame):
    data = DataProvider()

    def __init__(self, parent: ttk.Frame, peer: PeerType) -> None:
        ttk.Frame.__init__(self, parent)
        super().__init__()
        self.app = parent
        self.peer = peer

        self.id = peer["id"]
        if peer["name"]:
            self.name = peer["name"]
        else:
            self.name = peer["other_names"]


        self.return_value: Literal[
            'delete', 'pin_and_verify', 'rename', 'additional_ip', None
        ] = None
        self.display_peer_details()

    def display_peer_details(self) -> None:
        self.top_frame = ttk.Frame(
            self,

            width=500,

        )

        self.top_canvas = tk.Canvas(
            self.top_frame,
            width=500,
            bd=0,
            highlightthickness=0,
            relief="ridge",
        )

        self.yscrollbar = ttk.Scrollbar(
            self.top_frame,
            orient="vertical",
            command=self.top_canvas.yview,
        )

        frame = ttk.Frame(
            self.top_canvas,
            width=600,
            height=870,
        )
        frame.grid(row=0, sticky="ew")
        self.canvas = tk.Canvas(
            frame,
            height=870,
            width=600,
            bd=0,
            highlightthickness=0,
            relief="ridge",
        )
        self.canvas.place(x=0, y=0)

        # Packing and configuring
        self.top_canvas.pack(side="left", fill="y", expand=1, anchor="nw")

        self.yscrollbar.pack(side="right", fill="y", expand=1)

        self.top_canvas.configure(yscrollcommand=self.yscrollbar.set)
        self.top_canvas.bind(
            '<Configure>',
            lambda e: self.top_canvas.configure(
                scrollregion=self.top_canvas.bbox('all')
            ),
        )

        self.top_canvas.create_window((0, 0), window=frame, anchor="nw")

        self.top_frame.pack(
            fill="both", expand=1, padx=(0, 0), pady=(0, 0), side="left"
        )

        self.top_frame.columnconfigure(0, weight=1)

        self.label = ttk.Label(frame, text="test")

        # Title
        self.canvas.create_text(
            33.0,
            26.0,
            anchor="nw",
            text="Peer",
            font=(FONT, FONT_SIZE_HEADER),
        )

        # Delete
        self.button_delete_image = tk.PhotoImage(
            file=IMAGE_BASE_PATH + 'delete.png'
        )
        button_delete = ttk.Button(
            frame,
            image=self.button_delete_image,

            command=lambda: self.delete_peer(self.peer["id"], self.name),

        )
        button_delete.place(x=450.0, y=34.0, width=39.0, height=31.0)

        # Name entry
        self.canvas.create_text(
            33.0,
            93.0,
            anchor="nw",
            text="name:",
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        self.entry_image_name = tk.PhotoImage(
            file=IMAGE_BASE_PATH + 'edit_name_entry.png'
        )
        self.entry_image_name_bg = self.canvas.create_image(
            110.0, 126.5, image=self.entry_image_name
        )
        self.entry_name = tk.Entry(
            frame,
            bd=0,
            highlightthickness=0,
        )
        self.entry_name.insert(0, self.peer["name"])
        self.entry_name.place(x=45.5, y=114.0, width=129.0, height=23.0)

        # name save
        self.button_image_name = tk.PhotoImage(
            file=IMAGE_BASE_PATH + 'save.png'
        )
        button_name = ttk.Button(
            frame,
            image=self.button_image_name,

            command=lambda: self.edit_peer(),

        )
        button_name.place(x=190.0, y=115.0, width=34.0, height=23.0)

        # Id
        self.canvas.create_text(
            33.0,
            168.0,
            anchor="nw",
            text="id:",
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        self.canvas.create_text(
            33.0,
            194.0,
            anchor="nw",
            text=self.peer["id"],
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        self.button_image_copy_id = tk.PhotoImage(
            file=IMAGE_BASE_PATH + 'clipboard.png'
        )
        button_copy_id = ttk.Button(
            frame,
            image=self.button_image_copy_id,

            command=lambda: self.add_to_clipbaord(self.peer["id"]),

        )
        button_copy_id.place(x=53.0, y=163.0, width=34.0, height=23.0)

        # Other names
        self.canvas.create_text(
            33.0,
            236.0,
            anchor="nw",
            text="other_names:",
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        self.canvas.create_text(
            33.0,
            262.0,
            anchor="nw",
            text=self.peer["other_names"],
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        # Status
        self.canvas.create_text(
            33.0,
            304.0,
            anchor="nw",
            text="status:",
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        self.canvas.create_text(
            33.0,
            330.0,
            anchor="nw",
            text=self.peer["status"],
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        self.button_image_pin_verify = tk.PhotoImage(
            file=IMAGE_BASE_PATH + 'pin_and_verify.png'
        )
        button_pin_verify = ttk.Button(
            frame,
            image=self.button_image_pin_verify,

            command=lambda: self.pin_verify(),

        )
        button_pin_verify.place(x=78.0, y=299.0, width=92.0, height=23.0)

        # Endpoint
        self.canvas.create_text(
            33.0,
            372.0,
            anchor="nw",
            text="endpoint:",
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        self.canvas.create_text(
            33.0,
            398.0,
            anchor="nw",
            text=self.peer["endpoint"],
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        # Allowed IP
        self.canvas.create_text(
            33.0,
            440.0,
            anchor="nw",
            text="allowed_ip:",
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        self.canvas.create_text(
            33.0,
            466.0,
            anchor="nw",
            text=str(self.peer["allowed_ip"]),
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        # Latest signature
        self.canvas.create_text(
            33.0,
            508.0,
            anchor="nw",
            text="latest_signature:",
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        self.canvas.create_text(
            33.0,
            534.0,
            anchor="nw",
            text=self.peer["latest_signature"],
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        # Latest handshake
        self.canvas.create_text(
            33.0,
            576.0,
            anchor="nw",
            text="latest_handshake:",
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        self.canvas.create_text(
            33.0,
            602.0,
            anchor="nw",
            text=self.peer["latest_handshake"],
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        # Wg pubkey
        self.canvas.create_text(
            33.0,
            644.0,
            anchor="nw",
            text="wg_pubkey:",
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        self.canvas.create_text(
            33.0,
            670.0,
            anchor="nw",
            text=self.peer["wg_pubkey"],
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        self.button_image_copy_wg_pubkey = tk.PhotoImage(
            file=IMAGE_BASE_PATH + 'clipboard.png'
        )
        button_copy_wg_pubkey = ttk.Button(
            frame,
            image=self.button_image_copy_wg_pubkey,

            command=lambda: self.add_to_clipbaord(self.peer["wg_pubkey"]),

        )
        button_copy_wg_pubkey.place(x=108.0, y=639.0, width=34.0, height=23.0)

        # Allowed IPs
        self.canvas.create_text(
            33.0,
            712.0,
            anchor="nw",
            text="allowed_ips:",
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        self.canvas.create_text(
            33.0,
            738.0,
            anchor="nw",
            text=self.peer["allowed_ip"],
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        # Name entry
        self.canvas.create_text(
            33.0,
            775.0,
            anchor="nw",
            text="Add additional IP to Peer:",
            font=(FONT, FONT_SIZE_TEXT_M),
        )

        self.entry_image_additional_ip = tk.PhotoImage(
            file=IMAGE_BASE_PATH + 'add_ip_to_peer_entry.png'
        )
        self.entry_image_additional_ip_bg = self.canvas.create_image(
            144.0, 808.5, image=self.entry_image_additional_ip
        )
        self.entry_additional_ip = ttk.Entry(
            frame,

        )
        self.entry_additional_ip.place(
            x=45.5, y=796.0, width=197.0, height=23.0
        )

        # name save
        self.button_image_additional_ip = tk.PhotoImage(
            file=IMAGE_BASE_PATH + 'save.png'
        )
        button_additional_ip = ttk.Button(
            frame,
            image=self.button_image_additional_ip,

            command=lambda: self.add_additional_ip(),

        )
        button_additional_ip.place(x=260.0, y=797.0, width=34.0, height=23.0)

    def delete_peer(self, peer_id: str, peer_name: str) -> None:
        # show message to ask user if they really
        # want to remove a peer

        box = messagebox.askyesno(
            _("Remove"),
            _("Remove this peer: ") + peer_name + " ?",
            parent=self,
        )

        if box:
            self.data.delete_peer(peer_id)
            self.return_value = "delete"
            self.destroy()

    def edit_peer(self) -> None:
        try:
            self.data.rename_peer(self.peer["id"], self.entry_name.get())
            self.return_value = "rename"
            self.destroy()
        except Exception:
            PopupMessage.showPopupMessage("Error", "Failed to edit peer")

    def pin_verify(self) -> None:
        # if name of peer was changed, we need other_names for pin_verify
        if self.peer["other_names"]:
            name = self.peer["other_names"]
        else:
            name = self.peer["name"]
        self.data.pin_and_verify(self.peer["id"], name)
        self.return_value = "pin_and_verify"
        self.destroy()

    def add_additional_ip(self) -> None:
        if len(self.entry_additional_ip.get()) == 0:
            return

        try:
            self.data.add_peer(self.peer["id"], self.entry_additional_ip.get())
            self.entry_additional_ip.delete(0, "end")
            self.return_value = "additional_ip"
            self.destroy()
        except Exception:
            PopupMessage.showPopupMessage(
                "Error", "Failed to add an additional ip address"
            )

    def add_to_clipbaord(self, text: str) -> None:
        self.clipboard_clear()
        self.clipboard_append(text)
        PopupMessage.showPopupMessage("Information", "Copied to clipboard")

    def show(
        self,
    ) -> Literal['delete', 'pin_and_verify', 'rename', 'additional_ip', None]:

        self.wait_window(self)
        return self.return_value
