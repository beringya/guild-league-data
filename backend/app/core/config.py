from functools import lru_cache
from pathlib import Path

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

    app_name: str = "逆水寒帮会联赛数据分析平台"
    app_version: str = "1.0.0"
    database_url: str = "sqlite:///./data/app.db"
    data_dir: Path = Path("./data")
    backup_dir: Path = Path("./backups")
    session_secret: str = Field(default="dev-only-change-me")
    cookie_name: str = "nsh_session"
    csrf_cookie_name: str = "nsh_csrf"
    cookie_secure: bool = False
    session_hours: int = 8
    admin_username: str = "admin"
    upload_max_bytes: int = 5 * 1024 * 1024
    cors_origins: str = ""

    @property
    def upload_dir(self) -> Path:
        return self.data_dir / "uploads"


@lru_cache
def get_settings() -> Settings:
    settings = Settings()
    settings.data_dir.mkdir(parents=True, exist_ok=True)
    settings.upload_dir.mkdir(parents=True, exist_ok=True)
    settings.backup_dir.mkdir(parents=True, exist_ok=True)
    return settings
