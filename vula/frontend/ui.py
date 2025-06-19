import gettext
import tkinter as tk
from tkinter import Button, Frame, PhotoImage, ttk
from typing import Any

from vula import common
from vula.frontend import DataProvider
from vula.frontend.constants import (
    BACKGROUND_COLOR,
    FONT,
    FONT_SIZE_TEXT_L,
    FONT_SIZE_TEXT_XL,
    HEIGHT,
    IMAGE_BASE_PATH,
    TEXT_COLOR_WHITE,
    WIDTH,
)
from vula.frontend.overlay import HelpOverlay
from vula.frontend.view import (
    Peers,
    Prefs,
    VerificationKeyFrame,
    DescriptorFrame,
)

_ = gettext.gettext


class App(tk.Tk):
    def __init__(self, *args: Any, **kwargs: Any) -> None:
        self.data = DataProvider()
        tk.Tk.__init__(self, *args, **kwargs)

        self.geometry("{}x{}".format(WIDTH, HEIGHT))
        self.config(bg=BACKGROUND_COLOR)

        # create all of the main containers

        content_frame = Frame(
            self, bg=BACKGROUND_COLOR, width=1200, height=600, padx=3, pady=3
        )
        footer_frame = Frame(
            self, bg=BACKGROUND_COLOR, width=1200, height=150, pady=30, padx=30
        )
        bottom_frame = Frame(
            self, bg=BACKGROUND_COLOR, width=600, height=50, pady=3
        )

        # layout all of the main containers
        self.grid_rowconfigure(1, weight=1)
        self.grid_columnconfigure(0, weight=1)

        content_frame.grid(row=1, sticky="nsew")
        footer_frame.grid(row=2, sticky="e")
        bottom_frame.grid(row=3, sticky="s")

        # create the center widgets
        content_frame.grid_rowconfigure(0, weight=1)
        content_frame.grid_columnconfigure(0, weight=1)

        self.notebook = ttk.Notebook(content_frame, style="TNotebook")

        self.notebook.grid(row=0, column=0, columnspan=2, sticky="nsew")

        peers_frame = Frame(self.notebook, bg=BACKGROUND_COLOR)
        pref_frame = Frame(self.notebook, bg=BACKGROUND_COLOR)
        verification_frame = VerificationKeyFrame(self.notebook, self.data)
        descriptor_frame = DescriptorFrame(self.notebook, self.data)

        self.notebook.add(peers_frame, text="Peers")
        self.notebook.add(pref_frame, text="Settings")
        self.notebook.add(verification_frame, text="Verification")
        self.notebook.add(descriptor_frame, text="Descriptor")
        self.verification_frame = verification_frame
        self.descriptor_frame = descriptor_frame

        peers_frame.grid_columnconfigure(0, weight=1)
        pref_frame.grid_columnconfigure(0, weight=1)
        verification_frame.grid_columnconfigure(0, weight=1)
        descriptor_frame.grid_columnconfigure(0, weight=1)

        self.peers_new = Peers(peers_frame, self.data)
        self.prefs = Prefs(pref_frame, self.data)

        footer_frame.grid_rowconfigure(1, weight=1)
        footer_frame.grid_columnconfigure(1, weight=1)

        vk_label = tk.Label(
            footer_frame,
            text="Verification Key:",
            bg=BACKGROUND_COLOR,
            fg=TEXT_COLOR_WHITE,
            font=(FONT, FONT_SIZE_TEXT_L),
        )
        vk_label.grid(row=0, column=0, sticky="w")

        self.button_image = PhotoImage(file=IMAGE_BASE_PATH + 'show_qr.png')
        vk_button = tk.Button(
            footer_frame,
            text="Show QR",
            image=self.button_image,
            borderwidth=0,
            highlightthickness=0,
            relief="sunken",
            background=BACKGROUND_COLOR,
            activebackground=BACKGROUND_COLOR,
            activeforeground=BACKGROUND_COLOR,
            command=lambda: self.open_vk_qr_code(),
        )
        vk_button.grid(row=0, column=1)

        desc_label = tk.Label(
            footer_frame,
            text="Descriptor:",
            bg=BACKGROUND_COLOR,
            fg=TEXT_COLOR_WHITE,
            font=(FONT, FONT_SIZE_TEXT_L),
        )
        desc_label.grid(row=1, column=0, sticky="w")
        desc_button = tk.Button(
            footer_frame,
            text="Show QR",
            image=self.button_image,
            borderwidth=0,
            highlightthickness=0,
            relief="sunken",
            background=BACKGROUND_COLOR,
            activebackground=BACKGROUND_COLOR,
            activeforeground=BACKGROUND_COLOR,
            command=lambda: self.open_descriptor_qr_code(),
        )
        desc_button.grid(row=1, column=1)

        # Get the status of the different vula processes
        state = self.data.get_status()

        # @TODO: The case where state is None might have to be handled
        # @TODO: differently
        if state is None:
            state = {
                "publish": "no status available",
                "discover": "no status available",
                "organize": "no status available",
            }

        # Display the status at the bottom
        status_label = tk.Label(
            bottom_frame,
            text=f'Publish: {_(state["publish"])} '
            f'\t Discover: {_(state["discover"])} '
            f'\t Organize: {_(state["organize"])}',
            bg=BACKGROUND_COLOR,
            fg=TEXT_COLOR_WHITE,
            font=(FONT, FONT_SIZE_TEXT_XL),
        )
        status_label.grid(row=0, column=0)

        # Add different command
        self.button_image_rediscover = PhotoImage(
            file=IMAGE_BASE_PATH + 'rediscover.png'
        )
        btn_Rediscover = Button(
            bottom_frame,
            text="Rediscover",
            image=self.button_image_rediscover,
            borderwidth=0,
            highlightthickness=0,
            relief="sunken",
            background=BACKGROUND_COLOR,
            activebackground=BACKGROUND_COLOR,
            activeforeground=BACKGROUND_COLOR,
            command=lambda: self.rediscover(),
        )
        self.button_image_repair = PhotoImage(
            file=IMAGE_BASE_PATH + 'repair.png'
        )
        btn_Repair = Button(
            bottom_frame,
            text="Repair",
            image=self.button_image_repair,
            command=lambda: self.repair(),
            borderwidth=0,
            highlightthickness=0,
            relief="sunken",
            background=BACKGROUND_COLOR,
            activebackground=BACKGROUND_COLOR,
            activeforeground=BACKGROUND_COLOR,
        )
        self.button_image_gate = PhotoImage(
            file=IMAGE_BASE_PATH + 'release_gateway.png'
        )
        btn_Release_Gateway = Button(
            bottom_frame,
            text="Release Gateway",
            image=self.button_image_gate,
            command=lambda: self.release_gateway(),
            borderwidth=0,
            highlightthickness=0,
            relief="sunken",
            background=BACKGROUND_COLOR,
            activebackground=BACKGROUND_COLOR,
            activeforeground=BACKGROUND_COLOR,
        )

        info_Help = HelpOverlay(self)

        self.button_image_help = PhotoImage(file=IMAGE_BASE_PATH + 'help.png')
        btn_Help = Button(
            bottom_frame,
            text="Help",
            image=self.button_image_help,
            command=lambda: info_Help.openNewWindow(),
            borderwidth=0,
            highlightthickness=0,
            relief="sunken",
            background=BACKGROUND_COLOR,
            activebackground=BACKGROUND_COLOR,
            activeforeground=BACKGROUND_COLOR,
        )
        btn_Rediscover.grid(row=0, column=1, pady=(20, 20), padx=(40, 20))
        btn_Repair.grid(row=0, column=2, pady=(20, 20), padx=(20, 20))
        btn_Release_Gateway.grid(row=0, column=3, pady=(20, 20), padx=(20, 20))
        btn_Help.grid(row=0, column=4, pady=(20, 20), padx=(20, 20))

    def rediscover(self) -> None:
        common.organize_dbus_if_active().rediscover()

    def release_gateway(self) -> None:
        common.organize_dbus_if_active().release_gateway()

    def repair(self) -> None:
        common.organize_dbus_if_active().sync(True)

    def open_vk_qr_code(self) -> None:
        self.notebook.select(self.verification_frame)

    def open_descriptor_qr_code(self) -> None:
        self.notebook.select(self.descriptor_frame)


def main() -> None:
    try:
        app = App()
        app.title("Vula")
        app.mainloop()
    except RuntimeError as e:
        print(e)
