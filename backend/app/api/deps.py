from __future__ import annotations

from fastapi import Depends, Header, HTTPException, Request, status
from sqlalchemy.orm import Session

from app.core.config import get_settings
from app.core.database import get_db
from app.core.models import AppUser, UserSession
from app.core.security import get_session_by_token, token_digest


def get_current_session(request: Request, db: Session = Depends(get_db)) -> UserSession:
    settings = get_settings()
    token = request.cookies.get(settings.cookie_name)
    session = get_session_by_token(db, token)
    if not session:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="请先登录")
    return session


def get_current_user(session: UserSession = Depends(get_current_session), db: Session = Depends(get_db)) -> AppUser:
    user = db.get(AppUser, session.user_id)
    if not user:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="请先登录")
    return user


def require_csrf(
    request: Request,
    session: UserSession = Depends(get_current_session),
    x_csrf_token: str | None = Header(default=None, alias="X-CSRF-Token"),
) -> None:
    if request.method in {"GET", "HEAD", "OPTIONS"}:
        return
    if not x_csrf_token or token_digest(x_csrf_token) != session.csrf_token_hash:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="CSRF 校验失败")
