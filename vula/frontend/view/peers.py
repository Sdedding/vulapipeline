from __future__ import annotations

"""
Peers list widget – Tkinter implementation.

Now based on the new dataclass model (Peer / PeerStatus) and the public
`get_provider()` entry-point exposed by `vula.frontend`.
"""

import gettext
import math
from tkinter import Button, Canvas, Frame, Label, PhotoImage
from typing import List

from vula.frontend import get_provider
from vula.frontend.datadomain import Peer, PeerStatus
from vula.frontend.constants import (
    BACKGROUND_COLOR,
    BACKGROUND_COLOR_CARD,
    FONT,
    FONT_SIZE_HEADER_2,
    FONT_SIZE_TEXT_L,
    FONT_SIZE_TEXT_M,
    FONT_SIZE_TEXT_S,
    FONT_SIZE_TEXT_XS,
    IMAGE_BASE_PATH,
    TEXT_COLOR_GREEN,
    TEXT_COLOR_GREY,
    TEXT_COLOR_HEADER_2,
    TEXT_COLOR_PURPLE,
    TEXT_COLOR_WHITE,
    TEXT_COLOR_YELLOW,
)
from vula.frontend.overlay import PeerDetailsOverlay, PopupMessage

_ = gettext.gettext


class Peers(Frame):
    """Scrollable list of peers with paging."""

    # All widgets share a *single* provider instance
    data = get_provider()

    peer_frames: List[Frame] = []

    # cached values for change detection
    num_peers = 0
    num_peers_after_remove = 0

    # pagination
    peer_page = 1
    peers_per_page = 5

    # ────────────────────────── Tk lifecycle ──────────────────────────

    def __init__(self, parent: Frame) -> None:
        super().__init__(parent)
        self.app = parent

        self.display_header()
        self.display_peers()
        self.display_buttons()

        # regular refresh
        self.after(5_000, self.update_loop)

    # ────────────────────────── UI builder ──────────────────────────

    def display_header(self) -> None:
        title_frame = Frame(self.app, bg=BACKGROUND_COLOR, height=40)
        title_frame.grid(row=0, column=0, pady=(10, 0), sticky="w")

        title = Canvas(
            title_frame,
            bg=BACKGROUND_COLOR,
            height=40,
            width=400,
            bd=0,
            highlightthickness=0,
            relief="ridge",
        )
        title.place(x=0, y=0)
        title.create_text(
            0,
            0,
            anchor="nw",
            text=_("Peers"),
            fill=TEXT_COLOR_HEADER_2,
            font=(FONT, FONT_SIZE_HEADER_2),
        )

    def display_buttons(self) -> None:
        self.buttons_frame = Frame(
            self.app, bg=BACKGROUND_COLOR, width=400, height=80
        )
        self.buttons_frame.grid(row=99, column=0, sticky="w", pady=10)

        input_canvas = Canvas(
            self.buttons_frame,
            bg=BACKGROUND_COLOR,
            height=80,
            width=400,
            bd=0,
            highlightthickness=0,
            relief="ridge",
        )
        input_canvas.place(x=0, y=0)

        self.button_image_previous_page = PhotoImage(
            file=IMAGE_BASE_PATH + "previous.png"
        )
        self.button_previous_page = Button(
            master=self.buttons_frame,
            image=self.button_image_previous_page,
            borderwidth=0,
            highlightthickness=0,
            command=self.previous_page,
            relief="flat",
        )

        self.button_image_next_page = PhotoImage(
            file=IMAGE_BASE_PATH + "next.png"
        )
        self.button_next_page = Button(
            master=self.buttons_frame,
            image=self.button_image_next_page,
            borderwidth=0,
            highlightthickness=0,
            command=self.next_page,
            relief="flat",
        )

        # only render if multiple pages
        if math.ceil(self.num_peers / self.peers_per_page) > 1:
            self.button_next_page.place(
                x=321.0, y=0.0, width=79.0, height=23.0
            )

    # ────────────────────────── Paging helpers ──────────────────────────

    def next_page(self) -> None:
        self.peer_page += 1
        self.button_previous_page.place(
            x=232.0, y=0.0, width=79.0, height=23.0
        )
        if self.peer_page == math.ceil(self.num_peers / self.peers_per_page):
            self.button_next_page.place_forget()

        self.refresh()

    def previous_page(self) -> None:
        self.peer_page -= 1
        self.button_next_page.place(x=321.0, y=0.0, width=79.0, height=23.0)
        if self.peer_page == 1:
            self.button_previous_page.place_forget()
        self.refresh()

    # ────────────────────────── Main list ──────────────────────────

    def display_peers(self) -> None:
        peers = self.data.get_peers()
        self.num_peers = len(peers)

        if self.num_peers == 0:
            Label(
                self.app,
                text=_("No peers to display"),
                bg=BACKGROUND_COLOR,
                fg=TEXT_COLOR_WHITE,
                font=(FONT, FONT_SIZE_TEXT_L),
            ).grid(row=1, column=0, sticky="w", pady=10)
            return

        # slice peers for current page
        start = (self.peer_page - 1) * self.peers_per_page
        stop = start + self.peers_per_page
        for idx, peer in enumerate(peers[start:stop], start=1):
            self._render_single_peer(idx, peer)

    def _render_single_peer(self, row: int, peer: Peer) -> None:
        """Render one peer entry."""
        frame = Frame(self.app, bg=BACKGROUND_COLOR, width=400, height=70)
        frame.grid(row=row, column=0, sticky="w", pady=10)

        canvas = Canvas(
            frame,
            bg=BACKGROUND_COLOR,
            height=70,
            width=400,
            bd=0,
            highlightthickness=0,
            relief="ridge",
        )
        canvas.place(x=0, y=0)

        self.round_rectangle(canvas, 0, 0, 400, 70, r=30, fill=BACKGROUND_COLOR_CARD)

        name = peer.name or peer.other_names or peer.id

        # peer name
        canvas.create_text(
            20.0,
            10.0,
            anchor="nw",
            text=name,
            fill=TEXT_COLOR_GREEN,
            font=(FONT, FONT_SIZE_TEXT_S),
        )

        # endpoint
        canvas.create_text(
            20.0,
            40.0,
            anchor="nw",
            text=peer.endpoint or "",
            fill=TEXT_COLOR_GREY,
            font=(FONT, FONT_SIZE_TEXT_XS),
        )

        # status labels
        if PeerStatus.ENABLED in peer.status:
            canvas.create_text(
                205.0,
                40.0,
                anchor="nw",
                text="enabled",
                fill=TEXT_COLOR_YELLOW,
                font=(FONT, FONT_SIZE_TEXT_M),
            )
        if PeerStatus.PINNED in peer.status:
            canvas.create_text(
                270.0,
                40.0,
                anchor="nw",
                text="pinned",
                fill=TEXT_COLOR_PURPLE,
                font=(FONT, FONT_SIZE_TEXT_M),
            )
        if PeerStatus.VERIFIED in peer.status:
            canvas.create_text(
                325.0,
                40.0,
                anchor="nw",
                text="verified",
                fill=TEXT_COLOR_GREEN,
                font=(FONT, FONT_SIZE_TEXT_M),
            )

        # correct late-binding capture of peer
        canvas.bind("<Button-1>", lambda _e, _p=peer: self.open_details(_p))

        self.peer_frames.append(frame)

    # ────────────────────────── Actions ──────────────────────────

    def open_details(self, peer: Peer) -> None:
        result = PeerDetailsOverlay(self.app, peer).show()
        if result in {"delete"}:
            self.num_peers_after_remove = self.num_peers - 1
            self.peer_page = 1
        self.refresh()

    # ────────────────────────── Refresh / polling ──────────────────────────

    def refresh(self) -> None:
        self.clear_peers()
        self.display_peers()

    def update_loop(self) -> None:
        """Periodic check whether peer count changed on backend."""
        peers = self.data.get_peers()
        if peers:
            if (
                len(peers) == self.num_peers_after_remove
                or len(peers) != self.num_peers
            ):
                try:
                    self.refresh()
                    self.num_peers_after_remove = 0
                except Exception:  # noqa: BLE001
                    PopupMessage.showPopupMessage("Error", "Could not update peers")
        self.after(5_000, self.update_loop)

    # ────────────────────────── Utils ──────────────────────────

    def clear_peers(self) -> None:
        for frame in self.peer_frames:
            frame.destroy()
        self.peer_frames.clear()

    @staticmethod
    def round_rectangle(canvas: Canvas, x: int, y: int, w: int, h: int, r: int = 25, **kw) -> None:
        """Draw a rounded rectangle on *canvas*."""
        points = [
            x + r, y,
            w - r, y,
            w, y,
            w, y + r,
            w, h - r,
            w, h,
            w - r, h,
            x + r, h,
            x, h,
            x, h - r,
            x, y + r,
            x, y,
        ]
        canvas.create_polygon(points, **kw, smooth=True)
