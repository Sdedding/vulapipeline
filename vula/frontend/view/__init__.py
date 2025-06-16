import gettext

import pkg_resources

from .peers import Peers
from .prefs import Prefs
from .verification import VerificationKeyFrame
from .descriptor import DescriptorFrame

__all__ = [
    "Peers",
    "Prefs",
    "VerificationKeyFrame",
    "DescriptorFrame",
    "locale_path",
    "lang_translations",
]
