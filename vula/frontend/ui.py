"""Tk‑based main window for the Vula GUI.

Refactor notes
--------------
* Uses *provider_factory.get_provider()* (no direct `DataProvider()`).
* All other logic unchanged – only the backend wiring is different.
"""
from __future__ import annotations

import gettext
import tkinter as tk
from tkinter import Button, Canvas, Frame, PhotoImage

from vula import common
from vula.frontend.provider_factory import get_provider  # NEW
from vula.frontend.constants import (
    BACKGROUND_COLOR,
    FONT,
    FONT_SIZE_HEADER,
    FONT_SIZE_TEXT_L,
    FONT_SIZE_TEXT_XL,
    HEIGHT,
    IMAGE_BASE_PATH,
    TEXT_COLOR_HEADER,
    TEXT_COLOR_WHITE,
    WIDTH,
)
from vula.frontend.overlay import (
    DescriptorOverlay,
    HelpOverlay,
    VerificationKeyOverlay,
)
from vula.frontend.view import Peers, Prefs

_ = gettext.gettext


class App(tk.Tk):
    """Main dashboard window."""

    def __init__(self, *args, **kwargs) -> None:  # noqa: D401
        # ---------------- backend provider ----------------
        self.provider = get_provider()  # replaces direct DataProvider()

        # ---------------- Tk root init --------------------
        super().__init__(*args, **kwargs)
        self.geometry(f"{WIDTH}x{HEIGHT}")
        self.config(bg=BACKGROUND_COLOR)

        # ========== layout frames ==========
        header_frame = Frame(self, bg=BACKGROUND_COLOR, height=50, pady=3)
        content_frame = Frame(self, bg=BACKGROUND_COLOR, height=600, padx=3, pady=3)
        footer_frame = Frame(
            self, bg=BACKGROUND_COLOR, height=150, pady=30, padx=30
        )
        bottom_frame = Frame(self, bg=BACKGROUND_COLOR, height=50, pady=3)

        self.grid_rowconfigure(1, weight=1)
        self.grid_columnconfigure(0, weight=1)
        header_frame.grid(row=0, sticky="ew")
        content_frame.grid(row=1, sticky="nsew")
        footer_frame.grid(row=2, sticky="e")
        bottom_frame.grid(row=3, sticky="s")

        # ========== content panes ==========
        content_frame.grid_rowconfigure(0, weight=1)
        content_frame.grid_columnconfigure(1, weight=1)

        pref_frame = Frame(content_frame, bg=BACKGROUND_COLOR, width=600, height=600)
        pref_frame.grid(row=0, column=0, sticky="ns")
        pref_frame.grid_propagate(False)
        self.prefs = Prefs(pref_frame)

        peers_frame = Frame(content_frame, bg=BACKGROUND_COLOR, width=600, height=600)
        peers_frame.grid(row=0, column=1, sticky="nsew")
        peers_frame.grid_propagate(False)
        self.peers_view = Peers(peers_frame)
        peers_frame.grid_columnconfigure(0, weight=1)

        # ========== header ==========
        header_canvas = Canvas(
            header_frame,
            bg=BACKGROUND_COLOR,
            height=50,
            width=1200,
            bd=0,
            highlightthickness=0,
        )
        header_canvas.place(x=0, y=0)
        header_canvas.create_text(
            30.0,
            10.0,
            anchor="nw",
            text="Dashboard",
            fill=TEXT_COLOR_HEADER,
            font=(FONT, FONT_SIZE_HEADER),
        )

        # ========== footer: verification key & descriptor ==========
        footer_frame.grid_rowconfigure(1, weight=1)
        footer_frame.grid_columnconfigure(1, weight=1)

        tk.Label(
            footer_frame,
            text="Verification Key:",
            bg=BACKGROUND_COLOR,
            fg=TEXT_COLOR_WHITE,
            font=(FONT, FONT_SIZE_TEXT_L),
        ).grid(row=0, column=0, sticky="w")

        icon_qr = PhotoImage(file=IMAGE_BASE_PATH + "show_qr.png")
        tk.Button(
            footer_frame,
            image=icon_qr,
            command=self.open_vk_qr_code,
            borderwidth=0,
            highlightthickness=0,
            relief="sunken",
            background=BACKGROUND_COLOR,
            activebackground=BACKGROUND_COLOR,
        ).grid(row=0, column=1)

        tk.Label(
            footer_frame,
            text="Descriptor:",
            bg=BACKGROUND_COLOR,
            fg=TEXT_COLOR_WHITE,
            font=(FONT, FONT_SIZE_TEXT_L),
        ).grid(row=1, column=0, sticky="w")

        DescriptorOverlayBtn = DescriptorOverlay(self)
        tk.Button(
            footer_frame,
            image=icon_qr,
            command=DescriptorOverlayBtn.openNewWindow,
            borderwidth=0,
            highlightthickness=0,
            relief="sunken",
            background=BACKGROUND_COLOR,
            activebackground=BACKGROUND_COLOR,
        ).grid(row=1, column=1)

        # ========== status bar ==========
        status = self.provider.get_status()
        tk.Label(
            bottom_frame,
            text=f"Publish: {_(status.publish)} \t Discover: {_(status.discover)} \t Organize: {_(status.organize)}",
            bg=BACKGROUND_COLOR,
            fg=TEXT_COLOR_WHITE,
            font=(FONT, FONT_SIZE_TEXT_XL),
        ).grid(row=0, column=0)

        # ========== bottom commands ==========
        btn_cfg = dict(
            borderwidth=0,
            highlightthickness=0,
            relief="sunken",
            background=BACKGROUND_COLOR,
            activebackground=BACKGROUND_COLOR,
            activeforeground=BACKGROUND_COLOR,
        )
        img_rediscover = PhotoImage(file=IMAGE_BASE_PATH + "rediscover.png")
        Button(bottom_frame, image=img_rediscover, command=self.rediscover, **btn_cfg).grid(
            row=0, column=1, pady=20, padx=(40, 20)
        )

        img_repair = PhotoImage(file=IMAGE_BASE_PATH + "repair.png")
        Button(bottom_frame, image=img_repair, command=self.repair, **btn_cfg).grid(
            row=0, column=2, pady=20, padx=20
        )

        img_gate = PhotoImage(file=IMAGE_BASE_PATH + "release_gateway.png")
        Button(bottom_frame, image=img_gate, command=self.release_gateway, **btn_cfg).grid(
            row=0, column=3, pady=20, padx=20
        )

        HelpOverlayBtn = HelpOverlay(self)
        img_help = PhotoImage(file=IMAGE_BASE_PATH + "help.png")
        Button(bottom_frame, image=img_help, command=HelpOverlayBtn.openNewWindow, **btn_cfg).grid(
            row=0, column=4, pady=20, padx=20
        )

        # keep references to images (Tk GC)
        self._img_refs = [icon_qr, img_rediscover, img_repair, img_gate, img_help]

    # ───────────────────── helper callbacks ─────────────────────
    def rediscover(self) -> None:
        common.organize_dbus_if_active().rediscover()

    def release_gateway(self) -> None:
        common.organize_dbus_if_active().release_gateway()

    def repair(self) -> None:
        common.organize_dbus_if_active().sync(True)

    def open_vk_qr_code(self) -> None:
        VerificationKeyOverlay(self)


# ───────────────────────── entrypoint ─────────────────────────

def main() -> None:  # noqa: D401
    try:
        app = App()
        app.title("Vula")
        app.mainloop()
    except RuntimeError as exc:
        print(exc)
