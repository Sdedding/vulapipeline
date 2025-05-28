from typing import Optional, TypeVar, cast
import importlib

T = TypeVar("T")


def optional_import(name: str) -> Optional[T]:
    try:
        module = importlib.import_module(name)
        return cast(T, module)
    except ImportError:
        return None
