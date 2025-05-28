#!/usr/bin/env python3
"""Run the Tk GUI with the in-memory MockDataProvider (no DBus required)."""

from __future__ import annotations

import os
import sys
from pathlib import Path

# 1. Always force the mock back-end
os.environ["VULA_MOCK"] = "1"

# 2. Make repo root importable when running from a git checkout
ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from vula.frontend import constants as _c   # ← neue Zeile
_c.IMAGE_BASE_PATH = "/workspaces/vula/misc/images/"

# 3. Import and start the GUI
try:
    from vula.frontend.ui import App  # noqa: E402

    app = App()
    app.title("Vula GUI — Mock mode")
    app.mainloop()
except ImportError:
    print("⚠️  Tk / gi not available – running headless mock.")
    from vula.frontend.provider_factory import get_provider  # noqa: E402

    prov = get_provider()
    print("Provider OK:", type(prov).__name__)
