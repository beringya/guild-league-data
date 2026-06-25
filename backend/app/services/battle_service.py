from __future__ import annotations

from datetime import datetime
from pathlib import Path

from sqlalchemy import desc, select, update
from sqlalchemy.orm import Session

from app.core.config import get_settings
from app.core.defaults import default_rule_copy
from app.core.models import (
    AppSetting,
    Battle,
    BattleGuild,
    BattlePlayerStat,
    CareerAvatar,
    Guild,
    ImportLog,
    Player,
    PlayerAvatar,
    ScoringRangeVersion,
    ScoringRule,
)
from app.core.security import utcnow
from app.services import analysis
from app.services.csv_import import parse_csv_bytes
from app.services.serialization import dumps, loads


class BattleService:
    def __init__(self, db: Session):
        self.db = db
        self.settings = get_settings()

    def preview(self, content: bytes, filename: str) -> dict:
        parsed = parse_csv_bytes(content, filename)
        duplicate = self.db.execute(select(Battle).where(Battle.source_sha256 == parsed.sha256)).scalar_one_or_none()
        default_home = self.get_setting("default_home_guild")
        payload = parsed.summary(default_home)
        payload["duplicate_battle_id"] = duplicate.id if duplicate else None
        payload["can_confirm"] = payload["can_confirm"] and duplicate is None
        if duplicate:
            payload["warnings"].append(
                {"level": "warning", "code": "duplicate_file", "message": "该文件摘要已导入，默认不重复入库", "row_number": None, "field": None}
            )
        return payload

    def confirm(
        self,
        content: bytes,
        filename: str,
        home_guild: str,
        battle_at: datetime | None,
        user_id: int,
    ) -> Battle:
        parsed = parse_csv_bytes(content, filename)
        if parsed.errors:
            raise ValueError(parsed.errors[0].message)
        if self.db.execute(select(Battle).where(Battle.source_sha256 == parsed.sha256)).scalar_one_or_none():
            raise ValueError("该文件已导入，不能重复入库")

        guild_names = sorted({row["guild_name"] for row in parsed.rows})
        if home_guild not in guild_names:
            raise ValueError("选择的本帮会不在 CSV 中")
        if len(guild_names) != 2:
            raise ValueError("MVP 要求文件恰好包含两个帮会")

        enriched = analysis.derive_metrics(parsed.rows)
        range_config, sample_summary = analysis.suggest_ranges(enriched)
        rule_config = default_rule_copy()
        scored_rows = analysis.score_rows(enriched, rule_config, range_config)

        now = utcnow()
        battle_time = battle_at or parsed.detected_battle_time or now
        range_version = f"range-{now.strftime('%Y%m%d%H%M%S')}"
        battle = Battle(
            battle_at=battle_time,
            source_filename=filename,
            source_sha256=parsed.sha256,
            original_row_count=parsed.physical_rows,
            valid_row_count=len(parsed.rows),
            scoring_rule_version=rule_config["version"],
            scoring_range_version=range_version,
            import_status="completed",
            match_result_json=None,
            created_by=user_id,
            created_at=now,
        )
        self.db.add(battle)
        self.db.flush()

        opponent = next(name for name in guild_names if name != home_guild)
        side_by_guild = {home_guild: "home", opponent: "opponent"}
        battle_guilds: dict[str, BattleGuild] = {}
        for guild_name in guild_names:
            guild = self._get_or_create_guild(guild_name)
            bg = BattleGuild(
                battle_id=battle.id,
                guild_id=guild.id,
                side=side_by_guild[guild_name],
                member_count=sum(1 for row in parsed.rows if row["guild_name"] == guild_name),
            )
            self.db.add(bg)
            self.db.flush()
            battle_guilds[guild_name] = bg

        self.db.add(
            ScoringRangeVersion(
                version=range_version,
                name=f"{battle_time:%Y-%m-%d %H:%M} 职业范围",
                config_json=dumps(range_config),
                source_method="same_battle_same_career_both_guilds",
                source_battle_id=battle.id,
                sample_summary_json=dumps(sample_summary),
                created_by=user_id,
                created_at=now,
                is_active=1,
                is_frozen=1,
            )
        )
        self.db.execute(update(ScoringRangeVersion).where(ScoringRangeVersion.version != range_version).values(is_active=0))

        for row in scored_rows:
            guild = battle_guilds[row["guild_name"]].guild
            player = self._get_or_create_player(guild.id, row["player_name"])
            stat = BattlePlayerStat(
                battle_id=battle.id,
                battle_guild_id=battle_guilds[row["guild_name"]].id,
                player_id=player.id,
                player_name_snapshot=row["player_name"],
                level=row["level"],
                career=row["career"],
                team_leader_snapshot=row["team_leader"],
                kills=row["kills"],
                assists=row["assists"],
                logistics=row["logistics"],
                player_damage=row["player_damage"],
                building_damage=row["building_damage"],
                healing=row["healing"],
                damage_taken=row["damage_taken"],
                deaths=row["deaths"],
                qingdeng=row["qingdeng"],
                revive=row["revive"],
                control=row["control"],
                kda_ratio=row["kda_ratio"],
                player_damage_share=row["player_damage_share"],
                building_damage_share=row["building_damage_share"],
                player_damage_conversion_rate=row["player_damage_conversion_rate"],
                building_damage_conversion_rate=row["building_damage_conversion_rate"],
                participation_rate=row["participation_rate"],
                composite_score=row["composite_score"],
                guild_rank=row["guild_rank"],
                career_rank=row["career_rank"],
                team_rank=row["team_rank"],
                six_dimension_json=dumps(row["six_dimensions"]),
                score_detail_json=dumps(row["score_detail"]),
            )
            self.db.add(stat)

        for issue in parsed.warnings:
            self.db.add(ImportLog(battle_id=battle.id, level=issue.level, code=issue.code, row_number=issue.row_number, message=issue.message, created_at=now))

        self.set_setting("default_home_guild", home_guild)
        self._save_upload(content, parsed.sha256, filename)
        self.db.commit()
        self.db.refresh(battle)
        return battle

    def list_battles(self) -> list[dict]:
        battles = self.db.execute(select(Battle).order_by(desc(Battle.battle_at))).scalars().all()
        return [self.battle_summary(battle) for battle in battles]

    def get_battle(self, battle_id: int) -> Battle:
        battle = self.db.get(Battle, battle_id)
        if not battle:
            raise ValueError("比赛不存在")
        return battle

    def delete_battle(self, battle_id: int) -> None:
        battle = self.get_battle(battle_id)
        self.db.delete(battle)
        self.db.commit()

    def battle_summary(self, battle: Battle) -> dict:
        guilds = self.db.execute(select(BattleGuild).where(BattleGuild.battle_id == battle.id)).scalars().all()
        home = next((item for item in guilds if item.side == "home"), None)
        opponent = next((item for item in guilds if item.side == "opponent"), None)
        return {
            "id": battle.id,
            "battle_at": battle.battle_at.isoformat(),
            "source_filename": battle.source_filename,
            "source_sha256": battle.source_sha256,
            "valid_row_count": battle.valid_row_count,
            "original_row_count": battle.original_row_count,
            "scoring_rule_version": battle.scoring_rule_version,
            "scoring_range_version": battle.scoring_range_version,
            "home_guild": home.guild.name if home else None,
            "opponent_guild": opponent.guild.name if opponent else None,
            "home_member_count": home.member_count if home else 0,
            "opponent_member_count": opponent.member_count if opponent else 0,
            "created_at": battle.created_at.isoformat(),
        }

    def rows_for_battle(self, battle_id: int) -> tuple[Battle, list[dict], dict[str, str]]:
        battle = self.get_battle(battle_id)
        stats = self.db.execute(select(BattlePlayerStat).where(BattlePlayerStat.battle_id == battle_id)).scalars().all()
        side_map: dict[str, str] = {}
        guild_by_bg = {}
        for bg in self.db.execute(select(BattleGuild).where(BattleGuild.battle_id == battle_id)).scalars():
            guild_by_bg[bg.id] = bg.guild.name
            side_map[bg.guild.name] = bg.side
        rows = [self.stat_to_row(stat, guild_by_bg[stat.battle_guild_id]) for stat in stats]
        return battle, rows, side_map

    def overview(self, battle_id: int) -> dict:
        battle, rows, side_map = self.rows_for_battle(battle_id)
        return analysis.overview_payload(self.battle_summary(battle), rows, side_map)

    def rankings(
        self,
        battle_id: int,
        side: str = "home",
        career: str | None = None,
        team: str | None = None,
        search: str | None = None,
        page: int = 1,
        page_size: int = 50,
    ) -> dict:
        _battle, rows, side_map = self.rows_for_battle(battle_id)
        selected_guilds = {guild for guild, value in side_map.items() if value == side} if side in {"home", "opponent"} else set(side_map)
        filtered = [row for row in rows if row["guild_name"] in selected_guilds]
        if career:
            filtered = [row for row in filtered if row["career"] == career]
        if team:
            filtered = [row for row in filtered if row["team_leader"] == team]
        if search:
            filtered = [row for row in filtered if search.lower() in row["player_name"].lower()]
        ordered = analysis.sorted_for_ranking(filtered)
        total = len(ordered)
        start = max(page - 1, 0) * page_size
        return {
            "total": total,
            "page": page,
            "page_size": page_size,
            "items": [self.row_public(row) for row in ordered[start : start + page_size]],
        }

    def player_detail(self, battle_id: int, stat_id: int) -> dict:
        _battle, rows, _side_map = self.rows_for_battle(battle_id)
        row = next((item for item in rows if item["stat_id"] == stat_id), None)
        if not row:
            raise ValueError("玩家统计不存在")
        return analysis.player_detail(row, rows)

    def team_top3(self, battle_id: int, side: str = "home", metric: str = "composite_score") -> dict:
        _battle, rows, side_map = self.rows_for_battle(battle_id)
        guild = next((name for name, value in side_map.items() if value == side), None)
        if not guild:
            raise ValueError("帮会不存在")
        return {"guild_name": guild, "teams": analysis.team_top3(rows, guild, metric)}

    def guild_comparison(self, battle_id: int) -> dict:
        _battle, rows, side_map = self.rows_for_battle(battle_id)
        totals = analysis.aggregate_by(rows, "guild_name")
        comparison = analysis.compare_guilds(totals, side_map)
        comparison["career_averages"] = analysis.career_averages(rows)
        comparison["insights"] = analysis.build_insights(comparison)
        return comparison

    def squad_comparison(self, battle_id: int, scope: str = "all") -> dict:
        _battle, rows, _side_map = self.rows_for_battle(battle_id)
        return {"scope": scope, "squads": analysis.squad_comparison(rows, scope)}

    def reanalyze(self, battle_id: int) -> dict:
        battle, rows, _side_map = self.rows_for_battle(battle_id)
        rule = self.db.execute(select(ScoringRule).where(ScoringRule.is_active == 1)).scalar_one_or_none()
        if not rule:
            raise ValueError("没有可用的评分规则")
        range_version = self.db.execute(select(ScoringRangeVersion).where(ScoringRangeVersion.is_active == 1)).scalar_one_or_none()
        if not range_version:
            range_version = self.db.get(ScoringRangeVersion, battle.scoring_range_version)
        if not range_version:
            raise ValueError("没有可用的职业范围版本")

        rule_config = loads(rule.config_json)
        range_config = loads(range_version.config_json)
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
        scored = analysis.score_rows(analysis.derive_metrics(raw_rows), rule_config, range_config)
        by_key = {(row["guild_name"], row["player_name"]): row for row in scored}
        stats = self.db.execute(select(BattlePlayerStat).where(BattlePlayerStat.battle_id == battle_id)).scalars().all()
        guild_by_bg = {bg.id: bg.guild.name for bg in self.db.execute(select(BattleGuild).where(BattleGuild.battle_id == battle_id)).scalars()}
        for stat in stats:
            scored_row = by_key[(guild_by_bg[stat.battle_guild_id], stat.player_name_snapshot)]
            stat.kda_ratio = scored_row["kda_ratio"]
            stat.player_damage_share = scored_row["player_damage_share"]
            stat.building_damage_share = scored_row["building_damage_share"]
            stat.player_damage_conversion_rate = scored_row["player_damage_conversion_rate"]
            stat.building_damage_conversion_rate = scored_row["building_damage_conversion_rate"]
            stat.participation_rate = scored_row["participation_rate"]
            stat.composite_score = scored_row["composite_score"]
            stat.guild_rank = scored_row["guild_rank"]
            stat.career_rank = scored_row["career_rank"]
            stat.team_rank = scored_row["team_rank"]
            stat.six_dimension_json = dumps(scored_row["six_dimensions"])
            stat.score_detail_json = dumps(scored_row["score_detail"])
        battle.scoring_rule_version = rule.version
        battle.scoring_range_version = range_version.version
        self.db.commit()
        return {"battle": self.battle_summary(battle), "updated_stats": len(stats)}

    def aggregate_rankings(
        self,
        guild_id: int | None = None,
        career: str | None = None,
        min_matches: int = 3,
        page: int = 1,
        page_size: int = 50,
    ) -> dict:
        stmt = select(BattlePlayerStat, Battle, Player, Guild).join(Battle).join(Player, BattlePlayerStat.player_id == Player.id).join(Guild, Player.guild_id == Guild.id)
        if guild_id:
            stmt = stmt.where(Player.guild_id == guild_id)
        if career:
            stmt = stmt.where(BattlePlayerStat.career == career)
        rows = self.db.execute(stmt).all()
        grouped: dict[tuple[int, str, str], list[tuple[BattlePlayerStat, Battle, Player, Guild]]] = {}
        for item in rows:
            stat, _battle, player, _guild = item
            grouped.setdefault((player.id, player.canonical_name, stat.career), []).append(item)

        summaries = []
        for (_player_id, player_name, item_career), items in grouped.items():
            if len(items) < min_matches:
                continue
            scores = [stat.composite_score or 0 for stat, _battle, _player, _guild in items]
            latest = max(items, key=lambda item: item[1].battle_at)
            summaries.append(
                {
                    "player_id": latest[2].id,
                    "player_name": player_name,
                    "guild_name": latest[3].name,
                    "career": item_career,
                    "match_count": len(items),
                    "average_composite_score": round(sum(scores) / len(scores), 2),
                    "cumulative_contribution_score": round(sum(scores), 2),
                    "best_composite_score": round(max(scores), 2),
                    "latest_composite_score": round(latest[0].composite_score or 0, 2),
                    "latest_battle_at": latest[1].battle_at.isoformat(),
                    "trend": [round(stat.composite_score or 0, 2) for stat, _battle, _player, _guild in sorted(items, key=lambda item: item[1].battle_at)[-5:]],
                }
            )
        ordered = sorted(summaries, key=lambda item: (-item["average_composite_score"], -item["match_count"], item["player_name"]))
        start = max(page - 1, 0) * page_size
        return {"total": len(ordered), "page": page, "page_size": page_size, "items": ordered[start : start + page_size]}

    def stat_to_row(self, stat: BattlePlayerStat, guild_name: str) -> dict:
        avatar_url = self._avatar_url(stat.player_id, stat.career)
        return {
            "stat_id": stat.id,
            "player_id": stat.player_id,
            "guild_name": guild_name,
            "player_name": stat.player_name_snapshot,
            "level": stat.level,
            "career": stat.career,
            "team_leader": stat.team_leader_snapshot,
            "kills": stat.kills,
            "assists": stat.assists,
            "logistics": stat.logistics,
            "player_damage": stat.player_damage,
            "building_damage": stat.building_damage,
            "healing": stat.healing,
            "damage_taken": stat.damage_taken,
            "deaths": stat.deaths,
            "qingdeng": stat.qingdeng,
            "revive": stat.revive,
            "control": stat.control,
            "kda_ratio": stat.kda_ratio,
            "player_damage_share": stat.player_damage_share,
            "building_damage_share": stat.building_damage_share,
            "player_damage_conversion_rate": stat.player_damage_conversion_rate,
            "building_damage_conversion_rate": stat.building_damage_conversion_rate,
            "participation_rate": stat.participation_rate,
            "composite_score": stat.composite_score,
            "guild_rank": stat.guild_rank,
            "career_rank": stat.career_rank,
            "team_rank": stat.team_rank,
            "six_dimensions": loads(stat.six_dimension_json, []),
            "score_detail": loads(stat.score_detail_json, {}),
            "primary_metric_score": (loads(stat.six_dimension_json, [{}],) or [{}])[0].get("score", 0),
            "avatar_url": avatar_url,
        }

    def row_public(self, row: dict) -> dict:
        payload = analysis.player_summary(row)
        payload.update({metric: row.get(metric, 0) for metric in ["level", "kills", "assists", "player_damage", "building_damage", "healing", "damage_taken", "deaths", "qingdeng", "revive", "control"]})
        payload.update(
            {
                "kda_ratio": row["kda_ratio"],
                "participation_rate": row["participation_rate"],
                "player_damage_share": row["player_damage_share"],
                "building_damage_share": row["building_damage_share"],
                "player_damage_conversion_rate": row["player_damage_conversion_rate"],
                "building_damage_conversion_rate": row["building_damage_conversion_rate"],
                "six_dimensions": row["six_dimensions"],
            }
        )
        return payload

    def get_setting(self, key: str):
        setting = self.db.get(AppSetting, key)
        return loads(setting.value_json) if setting else None

    def set_setting(self, key: str, value) -> None:
        setting = self.db.get(AppSetting, key)
        if setting:
            setting.value_json = dumps(value)
            setting.updated_at = utcnow()
        else:
            self.db.add(AppSetting(key=key, value_json=dumps(value), updated_at=utcnow()))

    def _get_or_create_guild(self, name: str) -> Guild:
        guild = self.db.execute(select(Guild).where(Guild.name == name)).scalar_one_or_none()
        if guild:
            return guild
        guild = Guild(name=name, created_at=utcnow())
        self.db.add(guild)
        self.db.flush()
        return guild

    def _get_or_create_player(self, guild_id: int, name: str) -> Player:
        player = self.db.execute(select(Player).where(Player.guild_id == guild_id, Player.canonical_name == name)).scalar_one_or_none()
        if player:
            return player
        player = Player(guild_id=guild_id, canonical_name=name, created_at=utcnow())
        self.db.add(player)
        self.db.flush()
        return player

    def _save_upload(self, content: bytes, sha256: str, filename: str) -> None:
        suffix = Path(filename).suffix or ".csv"
        target = self.settings.upload_dir / f"{sha256}{suffix}"
        target.write_bytes(content)

    def _avatar_url(self, player_id: int | None, career: str) -> str | None:
        if player_id:
            player_avatar = self.db.get(PlayerAvatar, player_id)
            if player_avatar and player_avatar.asset_path:
                return f"/api/avatars/uploaded/{player_avatar.asset_path}"
        career_avatar = self.db.get(CareerAvatar, career)
        if career_avatar:
            return f"/api/avatars/uploaded/{career_avatar.asset_path}"
        return None
