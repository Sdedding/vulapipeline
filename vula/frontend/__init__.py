from __future__ import annotations

"""
Vula frontend package.

Public surface:
    * get_provider()  – returns a fully-initialised implementation of
      `vula.frontend.api.DataProvider` (real or mock).

All GUI code should obtain the provider like this::

    from vula.frontend import get_provider
    provider = get_provider()

Everything else imported here (constants, translations) is forwarded unchanged
for backwards compatibility.
"""

# ─────────────────────────────────────────────────────────────────────────────
# 1. Provider factory – the single point of truth
# ─────────────────────────────────────────────────────────────────────────────
from .provider_factory import get_provider  # noqa: F401  (re-exported below)

# ─────────────────────────────────────────────────────────────────────────────
# 2. UI constants (color palette, fonts, sizes…)
# ─────────────────────────────────────────────────────────────────────────────
from .constants import (  # noqa: F401
    BACKGROUND_COLOR,
    BACKGROUND_COLOR_CARD,
    BACKGROUND_COLOR_ENTRY,
    BACKGROUND_COLOR_ERROR,
    FONT,
    FONT_SIZE_HEADER,
    FONT_SIZE_HEADER_2,
    FONT_SIZE_TEXT_L,
    FONT_SIZE_TEXT_M,
    FONT_SIZE_TEXT_S,
    FONT_SIZE_TEXT_XL,
    FONT_SIZE_TEXT_XS,
    FONT_SIZE_TEXT_XXL,
    HEIGHT,
    IMAGE_BASE_PATH,
    TEXT_COLOR_BLACK,
    TEXT_COLOR_GREEN,
    TEXT_COLOR_GREY,
    TEXT_COLOR_HEADER,
    TEXT_COLOR_HEADER_2,
    TEXT_COLOR_ORANGE,
    TEXT_COLOR_PURPLE,
    TEXT_COLOR_RED,
    TEXT_COLOR_WHITE,
    TEXT_COLOR_YELLOW,
    WIDTH,
)

# ─────────────────────────────────────────────────────────────────────────────
# 3. Internationalisation (gettext)
# ─────────────────────────────────────────────────────────────────────────────
import gettext
import pkg_resources

_locale_path = pkg_resources.resource_filename("vula", "locale")
_translation = gettext.translation(
    domain="ui",
    localedir=_locale_path,
    fallback=True,
)
_translation.install()  # type: ignore[attr-defined]

# ─────────────────────────────────────────────────────────────────────────────
# 4. What this module exports
# ─────────────────────────────────────────────────────────────────────────────
__all__: list[str] = [
    # factory
    "get_provider",
    # constants
    "BACKGROUND_COLOR",
    "BACKGROUND_COLOR_CARD",
    "BACKGROUND_COLOR_ENTRY",
    "BACKGROUND_COLOR_ERROR",
    "FONT",
    "FONT_SIZE_HEADER",
    "FONT_SIZE_HEADER_2",
    "FONT_SIZE_TEXT_L",
    "FONT_SIZE_TEXT_M",
    "FONT_SIZE_TEXT_S",
    "FONT_SIZE_TEXT_XL",
    "FONT_SIZE_TEXT_XS",
    "FONT_SIZE_TEXT_XXL",
    "HEIGHT",
    "IMAGE_BASE_PATH",
    "TEXT_COLOR_BLACK",
    "TEXT_COLOR_GREEN",
    "TEXT_COLOR_GREY",
    "TEXT_COLOR_HEADER",
    "TEXT_COLOR_HEADER_2",
    "TEXT_COLOR_ORANGE",
    "TEXT_COLOR_PURPLE",
    "TEXT_COLOR_RED",
    "TEXT_COLOR_WHITE",
    "TEXT_COLOR_YELLOW",
    "WIDTH",
]
