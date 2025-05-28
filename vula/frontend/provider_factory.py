"""Return the right DataProvider (real DBus or mock)."""
from __future__ import annotations

import importlib
import os
from types import ModuleType
from typing import TYPE_CHECKING

from .api import DataProvider

# --------------------------------------------------------------------------- #
# 1.  Try to import both implementations very defensively                      #
# --------------------------------------------------------------------------- #
RealDataProvider: type[DataProvider] | None
MockDataProvider: type[DataProvider] | None

try:
    from .dbus_provider import RealDataProvider  # type: ignore
except Exception:  # gi / dbus not available
    RealDataProvider = None  # type: ignore[assignment]

# primary path: package‐relative (works when repo is installed)
try:
    from vula.tools.mock_provider import MockDataProvider  # type: ignore
except Exception:
    # fallback: *top-level* tools/ folder (works in dev checkouts)
    try:
        MockDataProvider = importlib.import_module("tools.mock_provider").MockDataProvider  # type: ignore
    except Exception:
        MockDataProvider = None  # type: ignore[assignment]

# --------------------------------------------------------------------------- #
# 2.  Public helper                                                            #
# --------------------------------------------------------------------------- #
def get_provider(*, mock: bool | None = None) -> DataProvider:
    """Return a fully-initialised backend implementation.

    Resolution order
    ----------------
    1.  Caller forces `mock=True`   → always return mock or raise.
    2.  Caller forces `mock=False`  → try real, else mock.
    3.  Caller passes `None`        → env var ``VULA_MOCK`` decides.
    """
    want_mock = (
        mock
        if mock is not None
        else os.getenv("VULA_MOCK", "0").lower() in {"1", "true", "yes"}
    )

    if want_mock:
        if MockDataProvider is None:
            raise RuntimeError("Mock backend not available in this checkout.")
        return MockDataProvider()

    # prefer real provider if DBus deps are present
    if RealDataProvider is not None:
        return RealDataProvider()

    # fall back to mock in degraded envs
    if MockDataProvider is not None:
        return MockDataProvider()

    raise RuntimeError("No DataProvider backend could be instantiated.")
