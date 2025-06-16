from .help_overlay import HelpOverlay
from .peer_details_overlay import PeerDetailsOverlay
from .popupMessage import PopupMessage

import pkg_resources
import gettext

locale_path = pkg_resources.resource_filename('vula', 'locale')
lang_translations = gettext.translation(
    domain="ui.overlay", localedir=locale_path, fallback=True
)
lang_translations.install()

__all__ = [
    "HelpOverlay",
    "PeerDetailsOverlay",
    "PopupMessage",
    "locale_path",
    "lang_translations",
]
