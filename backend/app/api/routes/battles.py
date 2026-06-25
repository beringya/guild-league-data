from __future__ import annotations

from datetime import datetime

from fastapi import APIRouter, Depends, File, Form, HTTPException, Query, UploadFile, status
from sqlalchemy.orm import Session

from app.api.deps import get_current_user, require_csrf
from app.core.config import get_settings
from app.core.database import get_db
from app.core.models import AppUser
from app.services.battle_service import BattleService

router = APIRouter(prefix="/battles", tags=["battles"], dependencies=[Depends(get_current_user)])


@router.post("/import/preview", dependencies=[Depends(require_csrf)])
async def preview_import(file: UploadFile = File(...), db: Session = Depends(get_db)) -> dict:
    content = await _read_limited(file)
    try:
        return BattleService(db).preview(content, file.filename or "upload.csv")
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(exc)) from exc


@router.post("/import/confirm", dependencies=[Depends(require_csrf)])
async def confirm_import(
    file: UploadFile = File(...),
    home_guild: str = Form(...),
    battle_at: str | None = Form(None),
    user: AppUser = Depends(get_current_user),
    db: Session = Depends(get_db),
) -> dict:
    content = await _read_limited(file)
    battle_time = datetime.fromisoformat(battle_at) if battle_at else None
    try:
        battle = BattleService(db).confirm(content, file.filename or "upload.csv", home_guild, battle_time, user.id)
        return {"battle": BattleService(db).battle_summary(battle)}
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(exc)) from exc


@router.get("")
def list_battles(db: Session = Depends(get_db)) -> dict:
    return {"items": BattleService(db).list_battles()}


@router.get("/{battle_id}")
def get_battle(battle_id: int, db: Session = Depends(get_db)) -> dict:
    service = BattleService(db)
    return {"battle": service.battle_summary(service.get_battle(battle_id))}


@router.delete("/{battle_id}", dependencies=[Depends(require_csrf)])
def delete_battle(battle_id: int, db: Session = Depends(get_db)) -> dict:
    try:
        BattleService(db).delete_battle(battle_id)
        return {"ok": True}
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(exc)) from exc


@router.post("/{battle_id}/reanalyze", dependencies=[Depends(require_csrf)])
def reanalyze_battle(battle_id: int, db: Session = Depends(get_db)) -> dict:
    try:
        return BattleService(db).reanalyze(battle_id)
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(exc)) from exc


@router.get("/{battle_id}/overview")
def overview(battle_id: int, db: Session = Depends(get_db)) -> dict:
    try:
        return BattleService(db).overview(battle_id)
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(exc)) from exc


@router.get("/{battle_id}/rankings")
def rankings(
    battle_id: int,
    side: str = Query("home", pattern="^(home|opponent|all)$"),
    career: str | None = None,
    team: str | None = None,
    search: str | None = None,
    page: int = Query(1, ge=1),
    page_size: int = Query(50, ge=1, le=200),
    db: Session = Depends(get_db),
) -> dict:
    try:
        return BattleService(db).rankings(battle_id, side, career, team, search, page, page_size)
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(exc)) from exc


@router.get("/{battle_id}/players/{stat_id}")
def player_detail(battle_id: int, stat_id: int, db: Session = Depends(get_db)) -> dict:
    try:
        return BattleService(db).player_detail(battle_id, stat_id)
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(exc)) from exc


@router.get("/{battle_id}/team-top3")
def team_top3(
    battle_id: int,
    side: str = Query("home", pattern="^(home|opponent)$"),
    metric: str = "composite_score",
    db: Session = Depends(get_db),
) -> dict:
    try:
        return BattleService(db).team_top3(battle_id, side, metric)
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(exc)) from exc


@router.get("/{battle_id}/guild-comparison")
def guild_comparison(battle_id: int, db: Session = Depends(get_db)) -> dict:
    try:
        return BattleService(db).guild_comparison(battle_id)
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(exc)) from exc


@router.get("/{battle_id}/squad-comparison")
def squad_comparison(battle_id: int, scope: str = "all", db: Session = Depends(get_db)) -> dict:
    try:
        return BattleService(db).squad_comparison(battle_id, scope)
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(exc)) from exc


async def _read_limited(file: UploadFile) -> bytes:
    content = await file.read()
    if len(content) > get_settings().upload_max_bytes:
        raise HTTPException(status_code=status.HTTP_413_REQUEST_ENTITY_TOO_LARGE, detail="上传文件过大")
    if file.filename and not file.filename.lower().endswith(".csv"):
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="仅支持 CSV 文件")
    return content
