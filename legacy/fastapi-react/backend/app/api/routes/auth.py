from __future__ import annotations

from collections import defaultdict, deque
from datetime import timedelta

from fastapi import APIRouter, Depends, HTTPException, Request, Response, status
from pydantic import BaseModel, Field
from sqlalchemy import select
from sqlalchemy.orm import Session

from app.api.deps import get_current_session, get_current_user, require_csrf
from app.core.config import get_settings
from app.core.database import get_db
from app.core.models import AppUser, UserSession
from app.core.security import create_session, hash_password, revoke_other_sessions, revoke_session, utcnow, verify_password

router = APIRouter(prefix="/auth", tags=["auth"])
_failures: dict[str, deque] = defaultdict(deque)


class LoginPayload(BaseModel):
    username: str = Field(min_length=1)
    password: str = Field(min_length=1)


class ChangePasswordPayload(BaseModel):
    old_password: str = Field(min_length=1)
    new_password: str = Field(min_length=10)


@router.post("/login")
def login(payload: LoginPayload, request: Request, response: Response, db: Session = Depends(get_db)) -> dict:
    ip = request.client.host if request.client else "unknown"
    _check_rate_limit(ip)
    user = db.execute(select(AppUser).where(AppUser.username == payload.username)).scalar_one_or_none()
    if not user or not verify_password(payload.password, user.password_hash):
        _record_failure(ip)
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="账号或密码错误")

    token, csrf_token, session = create_session(db, user)
    user.last_login_at = utcnow()
    db.commit()
    _failures.pop(ip, None)

    settings = get_settings()
    response.set_cookie(
        settings.cookie_name,
        token,
        httponly=True,
        secure=settings.cookie_secure,
        samesite="lax",
        max_age=settings.session_hours * 3600,
    )
    response.set_cookie(
        settings.csrf_cookie_name,
        csrf_token,
        httponly=False,
        secure=settings.cookie_secure,
        samesite="lax",
        max_age=settings.session_hours * 3600,
    )
    return {"user": public_user(user), "csrf_token": csrf_token}


@router.post("/logout", dependencies=[Depends(require_csrf)])
def logout(response: Response, session: UserSession = Depends(get_current_session), db: Session = Depends(get_db)) -> dict:
    revoke_session(db, session)
    db.commit()
    settings = get_settings()
    response.delete_cookie(settings.cookie_name)
    response.delete_cookie(settings.csrf_cookie_name)
    return {"ok": True}


@router.get("/me")
def me(user: AppUser = Depends(get_current_user)) -> dict:
    return {"user": public_user(user)}


@router.post("/change-password", dependencies=[Depends(require_csrf)])
def change_password(
    payload: ChangePasswordPayload,
    session: UserSession = Depends(get_current_session),
    user: AppUser = Depends(get_current_user),
    db: Session = Depends(get_db),
) -> dict:
    if not verify_password(payload.old_password, user.password_hash):
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="旧密码不正确")
    if payload.old_password == payload.new_password:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="新密码不能与旧密码相同")
    user.password_hash = hash_password(payload.new_password)
    user.force_password_change = 0
    revoke_other_sessions(db, user.id, keep_session_id=session.id)
    db.commit()
    return {"ok": True, "user": public_user(user)}


def public_user(user: AppUser) -> dict:
    return {
        "id": user.id,
        "username": user.username,
        "is_admin": bool(user.is_admin),
        "force_password_change": bool(user.force_password_change),
        "last_login_at": user.last_login_at.isoformat() if user.last_login_at else None,
    }


def _check_rate_limit(ip: str) -> None:
    now = utcnow()
    attempts = _failures[ip]
    while attempts and attempts[0] < now - timedelta(minutes=15):
        attempts.popleft()
    if len(attempts) >= 5:
        raise HTTPException(status_code=status.HTTP_429_TOO_MANY_REQUESTS, detail="登录失败次数过多，请稍后再试")


def _record_failure(ip: str) -> None:
    _failures[ip].append(utcnow())
