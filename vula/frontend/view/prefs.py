from __future__ import annotations

"""
Preferences widget – Tkinter implementation.

Uses the dataclass-based :class:`vula.frontend.datadomain.Prefs`
instead of the old dict/TypedDict representation.
"""

import gettext
from dataclasses import asdict
from typing import Dict, Union, Any

import tkinter as tk
from tkinter import (
    Button,
    Canvas,
    Event,
    Frame,
    Label,
    PhotoImage,
    Scrollbar,
    Text,
)
from tkinter.constants import W

from vula.frontend import get_provider
from vula.frontend.datadomain import Prefs as PrefsModel
from vula.frontend.constants import (
    BACKGROUND_COLOR,
    BACKGROUND_COLOR_CARD,
    FONT,
    FONT_SIZE_HEADER_2,
    FONT_SIZE_TEXT_XL,
    IMAGE_BASE_PATH,
    TEXT_COLOR_BLACK,
    TEXT_COLOR_GREEN,
    TEXT_COLOR_GREY,
    TEXT_COLOR_HEADER_2,
    TEXT_COLOR_RED,
    TEXT_COLOR_WHITE,
)
from vula.frontend.overlay import PopupMessage

_ = gettext.gettext


class Prefs(Frame):
    """Editable preferences view."""

    data = get_provider()

    # map pref-name ➜ widget that holds the edited value
    widgets: Dict[str, Union[Text, Button]]

    def __init__(self, frame: Frame) -> None:
        super().__init__(frame)
        self.frame = frame

        self.show_editable: bool = False
        self.prefs: PrefsModel
        self.widgets = {}

        self.display_header()
        self.display_frames()

        self.reload_prefs()
        self.display_prefs()
        self.hide_save_cancel()

    # ──────────────────────────── layout scaffolding ────────────────────────────

    def display_header(self) -> None:
        title_frame = Frame(
            self.frame, bg=BACKGROUND_COLOR, padx=30, width=400, height=40
        )
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
            text=_("Settings"),
            fill=TEXT_COLOR_HEADER_2,
            font=(FONT, FONT_SIZE_HEADER_2),
        )

    def display_frames(self) -> None:
        # containers
        self.top_frame = Frame(
            self.frame,
            bg=BACKGROUND_COLOR,
            width=597,
            padx=30,
            pady=10,
            highlightbackground=BACKGROUND_COLOR,
            highlightcolor=BACKGROUND_COLOR,
            highlightthickness=1,
        )
        self.bottom_frame = Frame(
            self.frame,
            bg=BACKGROUND_COLOR,
            width=597,
            height=97,
            padx=30,
            pady=10,
            highlightbackground=BACKGROUND_COLOR,
            highlightcolor=BACKGROUND_COLOR,
            highlightthickness=1,
        )

        self.pref_canvas = Canvas(
            self.top_frame,
            background=BACKGROUND_COLOR_CARD,
            highlightbackground=BACKGROUND_COLOR_CARD,
            highlightcolor=BACKGROUND_COLOR_CARD,
            highlightthickness=1,
        )
        self.yscrollbar = Scrollbar(
            self.top_frame,
            orient="vertical",
            command=self.pref_canvas.yview,
            relief="flat",
        )
        self.pref_content_frame = Frame(
            self.pref_canvas,
            bg=BACKGROUND_COLOR_CARD,
            highlightbackground=BACKGROUND_COLOR_CARD,
            highlightcolor=BACKGROUND_COLOR_CARD,
            highlightthickness=1,
        )

        # geometry / scrolling
        self.pref_canvas.pack(side="left", fill="y", expand=True)
        self.yscrollbar.pack(side="right", fill="y")
        self.pref_canvas.configure(yscrollcommand=self.yscrollbar.set)
        self.pref_canvas.bind(
            "<Configure>",
            lambda _e: self.pref_canvas.configure(
                scrollregion=self.pref_canvas.bbox("all")
            ),
        )
        self.pref_canvas.create_window(
            (0, 0), window=self.pref_content_frame, anchor="nw"
        )

        self.top_frame.pack(
            fill="both", expand=True, padx=(0, 50), pady=(50, 0), side="top"
        )
        self.bottom_frame.pack(side="left")

        # buttons
        self._build_action_buttons()

    def _build_action_buttons(self) -> None:
        self.button_image_edit = PhotoImage(file=IMAGE_BASE_PATH + "edit.png")
        self.button_image_save_blue = PhotoImage(
            file=IMAGE_BASE_PATH + "save_blue.png"
        )
        self.button_image_cancel = PhotoImage(
            file=IMAGE_BASE_PATH + "cancel.png"
        )

        self.btn_edit = Button(
            master=self.bottom_frame,
            image=self.button_image_edit,
            command=self.set_editable,
            borderwidth=0,
            highlightthickness=0,
            relief="sunken",
            background=BACKGROUND_COLOR,
            activebackground=BACKGROUND_COLOR,
        )
        self.btn_save = Button(
            self.bottom_frame,
            image=self.button_image_save_blue,
            command=self.save_prefs,
            borderwidth=0,
            highlightthickness=0,
            relief="sunken",
            background=BACKGROUND_COLOR,
            activebackground=BACKGROUND_COLOR,
        )
        self.btn_cancel = Button(
            self.bottom_frame,
            image=self.button_image_cancel,
            command=self.cancel,
            borderwidth=0,
            highlightthickness=0,
            relief="sunken",
            background=BACKGROUND_COLOR,
            activebackground=BACKGROUND_COLOR,
        )
        self.btn_edit.pack(side="left", padx=10, pady=10)

    # ──────────────────────────── backend interaction ────────────────────────────

    def reload_prefs(self) -> None:
        """Fetch preferences from backend."""
        self.prefs = self.data.get_prefs()

    # ──────────────────────────── editing helpers ────────────────────────────

    @staticmethod
    def _toggle_bool_button(event: Event) -> None:  # noqa: D401
        """Toggle a boolean button value."""
        widget: Button = event.widget  # type: ignore[assignment]
        if widget["text"] == "True":
            widget.config(text="False", bg=TEXT_COLOR_RED)
        else:
            widget.config(text="True", bg=TEXT_COLOR_GREEN)

    @staticmethod
    def _bool_on_enter(event: Event) -> None:
        event.widget.config(bg=TEXT_COLOR_BLACK)

    @staticmethod
    def _bool_on_leave(event: Event) -> None:
        widget: Button = event.widget  # type: ignore[assignment]
        widget.config(bg=TEXT_COLOR_GREEN if widget["text"] == "True" else TEXT_COLOR_RED)

    # ──────────────────────────── save / cancel ────────────────────────────

    def save_prefs(self) -> None:
        """Write modified preferences back through the provider API."""
        orig = asdict(self.prefs)

        for key, original_value in orig.items():
            widget = self.widgets.get(key)
            if widget is None:
                continue

            # list preferences (Text widget with newline-separated items)
            if isinstance(original_value, list) and isinstance(widget, Text):
                current_list = widget.get("1.0", "end").strip().splitlines()
                # additions
                for val in current_list:
                    if val and val not in original_value:
                        if self._check_err(self.data.add_pref(key, val)):
                            return
                # removals
                for val in original_value:
                    if val not in current_list:
                        if self._check_err(self.data.remove_pref(key, val)):
                            return

            # boolean preference (Button toggle)
            elif isinstance(original_value, bool) and isinstance(widget, Button):
                new_val = widget["text"] == "True"
                if new_val != original_value:
                    if self._check_err(self.data.set_pref(key, new_val)):
                        return

            # int preference (single-line Text)
            elif isinstance(original_value, int) and isinstance(widget, Text):
                try:
                    new_int = int(widget.get("1.0", "end").strip())
                except ValueError:
                    PopupMessage.showPopupMessage("Error", "Invalid integer value")
                    return
                if new_int != original_value:
                    if self._check_err(self.data.set_pref(key, new_int)):
                        return

        # success – reload + redraw
        self.reload_prefs()
        self.show_editable = False
        self.hide_save_cancel()
        self.refresh_view()

    def _check_err(self, res: Any) -> bool:
        """Return True if *res* indicates an error and show popup."""
        if isinstance(res, str) and res.lower().startswith("error"):
            PopupMessage.showPopupMessage("Error", str(res))
            return True
        return False

    def cancel(self) -> None:
        self.reload_prefs()
        self.show_editable = False
        self.hide_save_cancel()
        self.refresh_view()

    # ──────────────────────────── view refresh helpers ────────────────────────────

    def set_editable(self) -> None:
        self.show_editable = True
        self.show_save_cancel()
        self.refresh_view()

    def refresh_view(self) -> None:
        """Clear and rebuild preference table."""
        for widget in self.pref_content_frame.grid_slaves():
            widget.grid_forget()
        self.pref_canvas.destroy()
        self.yscrollbar.destroy()
        self.top_frame.destroy()
        self.bottom_frame.destroy()

        self.display_frames()
        self.display_prefs()

    # ──────────────────────────── save / cancel button visibility ────────────────────────────

    def show_save_cancel(self) -> None:
        self.btn_edit.pack_forget()
        self.btn_cancel.pack(side="left", padx=10, pady=10)
        self.btn_save.pack(side="left", padx=10, pady=10)

    def hide_save_cancel(self) -> None:
        self.btn_edit.pack(side="left", padx=10, pady=10)
        self.btn_save.pack_forget()
        self.btn_cancel.pack_forget()

    # ──────────────────────────── preference table ────────────────────────────

    def display_prefs(self) -> None:
        """Render the preference list (read-only or editable)."""
        pref_dict = asdict(self.prefs)
        self.widgets.clear()
        row = 1

        for key, value in pref_dict.items():
            # left-hand label
            Label(
                self.pref_content_frame,
                text=_(key) + ":",
                font=(FONT, FONT_SIZE_TEXT_XL),
                fg=TEXT_COLOR_WHITE,
                bg=BACKGROUND_COLOR_CARD,
                anchor="nw",
            ).grid(row=row, column=0, padx=2, pady=2, sticky="nw")

            # ----- list preferences -----
            if isinstance(value, list):
                if self.show_editable:
                    txt = Text(
                        self.pref_content_frame,
                        height=max(len(value), 1),
                        width=20,
                        bg=BACKGROUND_COLOR_CARD,
                        fg=TEXT_COLOR_GREY,
                        highlightbackground=BACKGROUND_COLOR_CARD,
                        insertbackground=TEXT_COLOR_WHITE,
                    )
                    txt.insert(tk.END, "\n".join(value))
                    txt.grid(row=row, column=1, padx=1, pady=1, sticky=W)
                    self.widgets[key] = txt
                else:
                    if not value:
                        value = ["None"]
                    for element in value:
                        Label(
                            self.pref_content_frame,
                            text=str(element),
                            font=(FONT, FONT_SIZE_TEXT_XL),
                            bg=BACKGROUND_COLOR_CARD,
                            fg=TEXT_COLOR_WHITE,
                        ).grid(row=row, column=1, padx=1, pady=1, sticky=W)
                        row += 1
                    continue  # already advanced row

            # ----- boolean preference -----
            elif isinstance(value, bool):
                if self.show_editable:
                    color = TEXT_COLOR_GREEN if value else TEXT_COLOR_RED
                    btn = Button(
                        self.pref_content_frame,
                        text=str(value),
                        width=5,
                        height=1,
                        bg=color,
                        fg=TEXT_COLOR_WHITE,
                        highlightbackground=BACKGROUND_COLOR_CARD,
                        borderwidth=0,
                        highlightthickness=0,
                        relief="flat",
                    )
                    btn.grid(row=row, column=1, padx=1, pady=1, sticky=W)
                    btn.bind("<Leave>", self._bool_on_leave)
                    btn.bind("<Enter>", self._bool_on_enter)
                    btn.bind("<Button-1>", self._toggle_bool_button)
                    self.widgets[key] = btn
                else:
                    Label(
                        self.pref_content_frame,
                        text=str(value),
                        font=(FONT, FONT_SIZE_TEXT_XL),
                        bg=BACKGROUND_COLOR_CARD,
                        fg=TEXT_COLOR_GREEN if value else TEXT_COLOR_RED,
                    ).grid(row=row, column=1, padx=1, pady=1, sticky=W)

            # ----- int preference -----
            elif isinstance(value, int):
                if self.show_editable:
                    txt = Text(
                        self.pref_content_frame,
                        height=1,
                        width=20,
                        bg=BACKGROUND_COLOR_CARD,
                        fg=TEXT_COLOR_GREY,
                        highlightbackground=BACKGROUND_COLOR_CARD,
                        insertbackground=TEXT_COLOR_WHITE,
                    )
                    txt.insert(tk.END, str(value))
                    txt.grid(row=row, column=1, padx=1, pady=1, sticky=W)
                    self.widgets[key] = txt
                else:
                    Label(
                        self.pref_content_frame,
                        text=str(value),
                        font=(FONT, FONT_SIZE_TEXT_XL),
                        bg=BACKGROUND_COLOR_CARD,
                        fg=TEXT_COLOR_WHITE,
                    ).grid(row=row, column=1, padx=1, pady=1, sticky=W)

            row += 1


if __name__ == "__main__":
    import doctest

    doctest.testmod()
