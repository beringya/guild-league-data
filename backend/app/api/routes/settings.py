from __future__ import annotations

from fastapi import APIRouter, Depends, HTTPException, status
from pydantic import BaseModel
from sqlalchemy.orm import Session

from app.api.deps import get_current_user, require_csrf
from app.core.database import get_db
from app.core.models import AppUser
from app.services import analysis
from app.services.battle_service import BattleService
from app.services.settings_service import SettingsService

router = APIRouter(tags=["settings"], dependencies=[Depends(get_current_user)])


class SettingsPayload(BaseModel):
    values: dict


class RulePayload(BaseModel):
    name: str
    config: dict


class RangePayload(BaseModel):
    name: str
    config: dict
    source_battle_id: int | None = None
    sample_summary: dict | None = None


@router.get("/settings")
def get_settings(db: Session = Depends(get_db)) -> dict:
    return {"settings": SettingsService(db).get_settings()}


@router.put("/settings", dependencies=[Depends(require_csrf)])
def update_settings(payload: SettingsPayload, db: Session = Depends(get_db)) -> dict:
    return {"settings": SettingsService(db).update_settings(payload.values)}


@router.get("/scoring-rules")
def list_scoring_rules(db: Session = Depends(get_db)) -> dict:
    return {"items": SettingsService(db).list_scoring_rules()}


@router.post("/scoring-rules", dependencies=[Depends(require_csrf)])
def create_scoring_rule(payload: RulePayload, user: AppUser = Depends(get_current_user), db: Session = Depends(get_db)) -> dict:
    try:
        return SettingsService(db).create_scoring_rule(payload.config, payload.name, user.id)
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(exc)) from exc


@router.post("/scoring-rules/validate")
def validate_scoring_rule(payload: dict, db: Session = Depends(get_db)) -> dict:
    config = payload.get("config", payload)
    return SettingsService(db).validate_rule(config)


@router.post("/scoring-rules/range-suggestions")
def range_suggestions(payload: dict, db: Session = Depends(get_db)) -> dict:
    battle_id = payload.get("battle_id")
    if not battle_id:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="需要 battle_id")
    try:
        _battle, rows, _side_map = BattleService(db).rows_for_battle(int(battle_id))
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(exc)) from exc
    raw_rows = [
        {
            "guild_name": row["guild_name"],
            "player_name": row["player_name"],
            "level": row["level"],
            "career": row["career"],
            "team_leader": row["team_leader"],
            "kills": row["kills"],
            "assists": row["assists"],
            "logistics": row["logistics"],
            "player_damage": row["player_damage"],
            "building_damage": row["building_damage"],
            "healing": row["healing"],
            "damage_taken": row["damage_taken"],
            "deaths": row["deaths"],
            "qingdeng": row["qingdeng"],
            "revive": row["revive"],
            "control": row["control"],
        }
        for row in rows
    ]
    config, sample_summary = analysis.suggest_ranges(analysis.derive_metrics(raw_rows))
    return {"config": config, "sample_summary": sample_summary}


@router.get("/scoring-ranges")
def scoring_ranges(db: Session = Depends(get_db)) -> dict:
    return {"items": SettingsService(db).list_ranges()}


@router.post("/scoring-ranges", dependencies=[Depends(require_csrf)])
def publish_range(payload: RangePayload, user: AppUser = Depends(get_current_user), db: Session = Depends(get_db)) -> dict:
    return SettingsService(db).publish_range(payload.config, payload.name, user.id, payload.source_battle_id, payload.sample_summary)
