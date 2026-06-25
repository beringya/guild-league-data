from __future__ import annotations

import json
import logging
from datetime import datetime

from sqlalchemy import select
from sqlalchemy.orm import Session

from app.core.config import get_settings
from app.core.database import create_all
from app.core.defaults import DEFAULT_SETTINGS, default_rule_copy
from app.core.models import AppSetting, AppUser, ScoringRule
from app.core.security import generate_password, hash_password, utcnow

logger = logging.getLogger(__name__)


def bootstrap_database(db: Session) -> None:
    create_all()
    _ensure_settings(db)
    _ensure_default_scoring_rule(db)
    _ensure_admin(db)
    db.commit()


def _ensure_settings(db: Session) -> None:
    now = utcnow()
    for key, value in DEFAULT_SETTINGS.items():
        exists = db.get(AppSetting, key)
        if not exists:
            db.add(AppSetting(key=key, value_json=json.dumps(value, ensure_ascii=False), updated_at=now))


def _ensure_default_scoring_rule(db: Session) -> None:
    rule = db.get(ScoringRule, "v1.5-final")
    if rule:
        return
    now = utcnow()
    db.add(
        ScoringRule(
            version="v1.5-final",
            name="默认十三职业评分规则",
            status="published",
            config_json=json.dumps(default_rule_copy(), ensure_ascii=False),
            created_at=now,
            published_at=now,
            is_active=1,
        )
    )


def _ensure_admin(db: Session) -> None:
    settings = get_settings()
    user = db.execute(select(AppUser).where(AppUser.username == settings.admin_username)).scalar_one_or_none()
    if user:
        return

    password = generate_password()
    db.add(
        AppUser(
            username=settings.admin_username,
            password_hash=hash_password(password),
            is_admin=1,
            force_password_change=1,
            created_at=utcnow(),
        )
    )
    logger.warning("首次初始化管理员账号完成：admin / %s", password)
