from __future__ import annotations

from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session

from app.api.deps import get_current_user
from app.core.database import get_db
from app.services.battle_service import BattleService

router = APIRouter(prefix="/rankings", tags=["rankings"], dependencies=[Depends(get_current_user)])


@router.get("/players/aggregate")
@router.get("/history")
def aggregate_players(
    guild_id: int | None = None,
    career: str | None = None,
    min_matches: int = Query(3, ge=1),
    page: int = Query(1, ge=1),
    page_size: int = Query(50, ge=1, le=200),
    db: Session = Depends(get_db),
) -> dict:
    return BattleService(db).aggregate_rankings(guild_id, career, min_matches, page, page_size)
