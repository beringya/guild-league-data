from __future__ import annotations

import hashlib
import secrets
from pathlib import Path

from fastapi import APIRouter, Depends, File, HTTPException, Response, UploadFile, status
from sqlalchemy.orm import Session

from app.api.deps import get_current_user, require_csrf
from app.core.config import get_settings
from app.core.database import get_db
from app.core.models import CareerAvatar, PlayerAvatar
from app.core.security import utcnow

router = APIRouter(tags=["avatars"])


@router.get("/avatars/generated/{seed}")
def generated_avatar(seed: int, career: str = "默认") -> Response:
    colors = ["#EF6F9F", "#70A8E7", "#62B992", "#F2A55F", "#9F8FE8", "#C9497A"]
    digest = hashlib.sha256(f"{seed}:{career}".encode("utf-8")).digest()
    bg = colors[digest[0] % len(colors)]
    accent = colors[digest[1] % len(colors)]
    eye = "#493543"
    svg = f"""<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96">
<rect width="96" height="96" rx="48" fill="#fff0f5"/>
<circle cx="48" cy="50" r="34" fill="{bg}" opacity=".92"/>
<circle cx="34" cy="44" r="5" fill="{eye}"/>
<circle cx="62" cy="44" r="5" fill="{eye}"/>
<path d="M35 61c7 7 19 7 26 0" fill="none" stroke="{eye}" stroke-width="5" stroke-linecap="round"/>
<path d="M24 26c4-15 15-17 22-3M50 23c8-14 22-10 22 7" fill="none" stroke="{accent}" stroke-width="8" stroke-linecap="round"/>
<text x="48" y="88" text-anchor="middle" font-size="11" font-family="sans-serif" fill="#493543">{career[:2]}</text>
</svg>"""
    return Response(svg, media_type="image/svg+xml")


@router.get("/avatars/uploaded/{path:path}")
def uploaded_avatar(path: str) -> Response:
    avatar_dir = get_settings().data_dir / "avatars"
    target = (avatar_dir / path).resolve()
    if avatar_dir.resolve() not in target.parents and target != avatar_dir.resolve():
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="非法头像路径")
    if not target.exists():
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="头像不存在")
    media_type = _media_type(target.suffix)
    return Response(target.read_bytes(), media_type=media_type)


@router.put("/players/{player_id}/avatar", dependencies=[Depends(get_current_user), Depends(require_csrf)])
async def replace_player_avatar(player_id: int, file: UploadFile = File(...), db: Session = Depends(get_db)) -> dict:
    path, digest = await _save_avatar(file, f"players/{player_id}")
    avatar = db.get(PlayerAvatar, player_id)
    if avatar:
        avatar.source = "uploaded"
        avatar.asset_path = path
        avatar.content_sha256 = digest
        avatar.updated_at = utcnow()
    else:
        db.add(PlayerAvatar(player_id=player_id, source="uploaded", asset_path=path, content_sha256=digest, updated_at=utcnow()))
    db.commit()
    return {"ok": True, "avatar_url": f"/api/avatars/uploaded/{path}", "player_id": player_id}


@router.delete("/players/{player_id}/avatar", dependencies=[Depends(get_current_user), Depends(require_csrf)])
def reset_player_avatar(player_id: int, db: Session = Depends(get_db)) -> dict:
    avatar = db.get(PlayerAvatar, player_id)
    if avatar:
        db.delete(avatar)
        db.commit()
    return {"ok": True, "message": "已恢复为稳定随机头像", "player_id": player_id}


@router.put("/careers/{career}/avatar", dependencies=[Depends(get_current_user), Depends(require_csrf)])
async def replace_career_avatar(career: str, file: UploadFile = File(...), db: Session = Depends(get_db)) -> dict:
    path, digest = await _save_avatar(file, f"careers/{career}")
    avatar = db.get(CareerAvatar, career)
    if avatar:
        avatar.asset_path = path
        avatar.content_sha256 = digest
        avatar.updated_at = utcnow()
    else:
        db.add(CareerAvatar(career=career, asset_path=path, content_sha256=digest, updated_at=utcnow()))
    db.commit()
    return {"ok": True, "avatar_url": f"/api/avatars/uploaded/{path}", "career": career}


@router.delete("/careers/{career}/avatar", dependencies=[Depends(get_current_user), Depends(require_csrf)])
def reset_career_avatar(career: str, db: Session = Depends(get_db)) -> dict:
    avatar = db.get(CareerAvatar, career)
    if avatar:
        db.delete(avatar)
        db.commit()
    return {"ok": True, "message": "已删除职业默认头像", "career": career}


async def _save_avatar(file: UploadFile, prefix: str) -> tuple[str, str]:
    suffix = Path(file.filename or "").suffix.lower()
    if suffix not in {".png", ".jpg", ".jpeg", ".webp"}:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="头像仅支持 PNG/JPG/WebP")
    content = await file.read()
    if len(content) > 2 * 1024 * 1024:
        raise HTTPException(status_code=status.HTTP_413_REQUEST_ENTITY_TOO_LARGE, detail="头像文件不能超过 2MB")
    digest = hashlib.sha256(content).hexdigest()
    relative = f"{prefix}/{secrets.token_hex(8)}{suffix}"
    target = get_settings().data_dir / "avatars" / relative
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_bytes(content)
    return relative, digest


def _media_type(suffix: str) -> str:
    return {
        ".png": "image/png",
        ".jpg": "image/jpeg",
        ".jpeg": "image/jpeg",
        ".webp": "image/webp",
    }.get(suffix.lower(), "application/octet-stream")
