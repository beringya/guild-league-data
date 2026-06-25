from __future__ import annotations

import json
from datetime import date, datetime
from decimal import Decimal
from typing import Any


def dumps(value: Any) -> str:
    return json.dumps(value, ensure_ascii=False, default=_json_default)


def loads(value: str | None, default: Any = None) -> Any:
    if value is None:
        return default
    return json.loads(value)


def _json_default(value: Any) -> Any:
    if isinstance(value, (datetime, date)):
        return value.isoformat()
    if isinstance(value, Decimal):
        return float(value)
    raise TypeError(f"Object of type {type(value).__name__} is not JSON serializable")
