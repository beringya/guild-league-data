from __future__ import annotations

from sqlalchemy import select
from sqlalchemy.orm import Session

from app.core.defaults import default_rule_copy
from app.core.models import AppSetting, ScoringRangeVersion, ScoringRule
from app.core.security import utcnow
from app.services.serialization import dumps, loads


class SettingsService:
    def __init__(self, db: Session):
        self.db = db

    def get_settings(self) -> dict:
        rows = self.db.execute(select(AppSetting)).scalars().all()
        return {row.key: loads(row.value_json) for row in rows}

    def update_settings(self, values: dict) -> dict:
        now = utcnow()
        for key, value in values.items():
            setting = self.db.get(AppSetting, key)
            if setting:
                setting.value_json = dumps(value)
                setting.updated_at = now
            else:
                self.db.add(AppSetting(key=key, value_json=dumps(value), updated_at=now))
        self.db.commit()
        return self.get_settings()

    def list_scoring_rules(self) -> list[dict]:
        rules = self.db.execute(select(ScoringRule).order_by(ScoringRule.created_at.desc())).scalars().all()
        return [
            {
                "version": rule.version,
                "name": rule.name,
                "status": rule.status,
                "is_active": bool(rule.is_active),
                "created_at": rule.created_at.isoformat(),
                "published_at": rule.published_at.isoformat() if rule.published_at else None,
                "config": loads(rule.config_json),
            }
            for rule in rules
        ]

    def active_scoring_rule(self) -> dict:
        rule = self.db.execute(select(ScoringRule).where(ScoringRule.is_active == 1)).scalar_one_or_none()
        if not rule:
            return default_rule_copy()
        return loads(rule.config_json)

    def validate_rule(self, config: dict) -> dict:
        errors = []
        for career, profile in config.get("career_profiles", {}).items():
            dimensions = [item for item in profile.get("dimensions", []) if item.get("enabled", True)]
            if len(dimensions) != 6:
                errors.append(f"{career} 必须正好启用 6 个维度")
            weight_sum = round(sum(float(item.get("ranking_weight", 0) or 0) for item in dimensions), 6)
            if abs(weight_sum - 1.0) > 0.0001:
                errors.append(f"{career} 参与排名权重合计必须为 100%，当前为 {weight_sum * 100:.1f}%")
            for item in dimensions:
                range_config = item.get("range", {})
                min_value = range_config.get("min")
                max_value = range_config.get("max")
                if min_value is not None and max_value is not None and min_value >= max_value:
                    errors.append(f"{career} / {item.get('label')} 的下限必须小于上限")
        return {"valid": not errors, "errors": errors}

    def create_scoring_rule(self, config: dict, name: str, user_id: int) -> dict:
        validation = self.validate_rule(config)
        if not validation["valid"]:
            raise ValueError("; ".join(validation["errors"]))
        version = f"rule-{utcnow().strftime('%Y%m%d%H%M%S')}"
        config["version"] = version
        self.db.add(
            ScoringRule(
                version=version,
                name=name,
                status="published",
                config_json=dumps(config),
                created_at=utcnow(),
                published_at=utcnow(),
                published_by=user_id,
                is_active=1,
            )
        )
        for rule in self.db.execute(select(ScoringRule).where(ScoringRule.version != version)).scalars():
            rule.is_active = 0
        self.db.commit()
        return {"version": version, "config": config}

    def list_ranges(self) -> list[dict]:
        ranges = self.db.execute(select(ScoringRangeVersion).order_by(ScoringRangeVersion.created_at.desc())).scalars().all()
        return [
            {
                "version": item.version,
                "name": item.name,
                "is_active": bool(item.is_active),
                "is_frozen": bool(item.is_frozen),
                "source_method": item.source_method,
                "source_battle_id": item.source_battle_id,
                "created_at": item.created_at.isoformat(),
                "config": loads(item.config_json),
                "sample_summary": loads(item.sample_summary_json),
            }
            for item in ranges
        ]

    def publish_range(self, config: dict, name: str, user_id: int, source_battle_id: int | None = None, sample_summary: dict | None = None) -> dict:
        version = f"range-{utcnow().strftime('%Y%m%d%H%M%S')}"
        self.db.add(
            ScoringRangeVersion(
                version=version,
                name=name,
                config_json=dumps(config),
                source_method="manual_or_admin_confirmed_suggestion",
                source_battle_id=source_battle_id,
                sample_summary_json=dumps(sample_summary or {}),
                created_by=user_id,
                created_at=utcnow(),
                is_active=1,
                is_frozen=1,
            )
        )
        for item in self.db.execute(select(ScoringRangeVersion).where(ScoringRangeVersion.version != version)).scalars():
            item.is_active = 0
        self.db.commit()
        return {"version": version}
