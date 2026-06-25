from __future__ import annotations

import sqlite3
from datetime import datetime
from pathlib import Path

from app.core.config import get_settings


class BackupService:
    def __init__(self):
        self.settings = get_settings()

    def create_backup(self) -> dict:
        if not self.settings.database_url.startswith("sqlite:///"):
            raise ValueError("当前仅支持 SQLite 备份")
        source_path = Path(self.settings.database_url.replace("sqlite:///", "", 1))
        if not source_path.is_absolute():
            source_path = Path.cwd() / source_path
        if not source_path.exists():
            raise ValueError("数据库文件不存在，无法备份")
        backup_path = self.settings.backup_dir / f"app-{datetime.now():%Y%m%d-%H%M%S}.db"
        with sqlite3.connect(source_path) as source:
            with sqlite3.connect(backup_path) as target:
                source.backup(target)
                integrity = target.execute("PRAGMA integrity_check").fetchone()[0]
        return {"path": str(backup_path), "integrity_check": integrity, "created_at": datetime.now().isoformat()}
