from __future__ import annotations

import hashlib
import hmac
import secrets
from datetime import datetime, timedelta, timezone

from argon2 import PasswordHasher
from argon2.exceptions import VerifyMismatchError
from sqlalchemy import select, update
from sqlalchemy.orm import Session

from app.core.config import get_settings
from app.core.models import AppUser, UserSession

password_hasher = PasswordHasher()


def utcnow() -> datetime:
    return datetime.now(timezone.utc).replace(tzinfo=None)


def hash_password(password: str) -> str:
    return password_hasher.hash(password)


def verify_password(password: str, password_hash: str) -> bool:
    try:
        return password_hasher.verify(password_hash, password)
    except VerifyMismatchError:
        return False


def token_digest(token: str) -> str:
    settings = get_settings()
    return hmac.new(settings.session_secret.encode("utf-8"), token.encode("utf-8"), hashlib.sha256).hexdigest()


def create_session(db: Session, user: AppUser) -> tuple[str, str, UserSession]:
    settings = get_settings()
    token = secrets.token_urlsafe(48)
    csrf_token = secrets.token_urlsafe(32)
    now = utcnow()
    session = UserSession(
        id=secrets.token_hex(16),
        user_id=user.id,
        token_hash=token_digest(token),
        csrf_token_hash=token_digest(csrf_token),
        created_at=now,
        expires_at=now + timedelta(hours=settings.session_hours),
    )
    db.add(session)
    return token, csrf_token, session


def get_session_by_token(db: Session, token: str | None) -> UserSession | None:
    if not token:
        return None
    stmt = (
        select(UserSession)
        .where(UserSession.token_hash == token_digest(token))
        .where(UserSession.revoked_at.is_(None))
        .where(UserSession.expires_at > utcnow())
    )
    return db.execute(stmt).scalar_one_or_none()


def revoke_session(db: Session, session: UserSession) -> None:
    session.revoked_at = utcnow()


def revoke_other_sessions(db: Session, user_id: int, keep_session_id: str | None = None) -> None:
    stmt = update(UserSession).where(UserSession.user_id == user_id, UserSession.revoked_at.is_(None))
    if keep_session_id:
        stmt = stmt.where(UserSession.id != keep_session_id)
    db.execute(stmt.values(revoked_at=utcnow()))


def generate_password(length: int = 24) -> str:
    alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789!@#$%^&*"
    return "".join(secrets.choice(alphabet) for _ in range(length))
