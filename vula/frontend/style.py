from tkinter import ttk

from .constants import (
    BACKGROUND_COLOR,
    BACKGROUND_COLOR_CARD,
    TEXT_COLOR_HEADER,
    TEXT_COLOR_WHITE, TEXT_COLOR_HEADER_2,
)


def configure_styles() -> ttk.Style:
    """Configure ttk styles for a consistent look."""
    style = ttk.Style()
    # Use the current theme but ensure our custom styles exist
    style.theme_use(style.theme_use())
    style.configure("TNotebook",
                    borderwidth=0)
    style.configure(
        "Vula.Vertical.TScrollbar",
        gripcount=0,
        background=BACKGROUND_COLOR_CARD,
        troughcolor=BACKGROUND_COLOR,
        bordercolor=BACKGROUND_COLOR,
        lightcolor=BACKGROUND_COLOR_CARD,
        darkcolor=BACKGROUND_COLOR_CARD,)

    style.configure("Vula.TFrame", background=BACKGROUND_COLOR)
    style.configure("Vula.TLabel", background=BACKGROUND_COLOR, foreground=TEXT_COLOR_WHITE)
    style.configure(
        "Vula.Header.TLabel",
        background=BACKGROUND_COLOR,
        foreground=TEXT_COLOR_HEADER_2,
    )
    style.configure("Vula.TButton", background=BACKGROUND_COLOR)
    return style
