from __future__ import annotations

import gettext
import pkg_resources

# re-export widgets
from .peers import Peers  # noqa: F401
from .prefs import Prefs  # noqa: F401

# install translations for view namespace
_locale_path = pkg_resources.resource_filename("vula", "locale")
_gettext = gettext.translation(
    domain="ui.view", localedir=_locale_path, fallback=True
)
_gettext.install()  # type: ignore[attr-defined]

__all__: list[str] = ["Peers", "Prefs"]
