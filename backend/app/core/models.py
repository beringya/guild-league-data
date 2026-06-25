from datetime import datetime

from sqlalchemy import (
    CheckConstraint,
    BigInteger,
    DateTime,
    Float,
    ForeignKey,
    Index,
    Integer,
    String,
    Text,
    UniqueConstraint,
)
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.core.database import Base


class AppUser(Base):
    __tablename__ = "app_user"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    username: Mapped[str] = mapped_column(String(80), unique=True, nullable=False)
    password_hash: Mapped[str] = mapped_column(Text, nullable=False)
    is_admin: Mapped[int] = mapped_column(Integer, nullable=False, default=1)
    force_password_change: Mapped[int] = mapped_column(Integer, nullable=False, default=1)
    created_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)
    last_login_at: Mapped[datetime | None] = mapped_column(DateTime)

    sessions: Mapped[list["UserSession"]] = relationship(back_populates="user", cascade="all, delete-orphan")


class UserSession(Base):
    __tablename__ = "user_session"

    id: Mapped[str] = mapped_column(String(64), primary_key=True)
    user_id: Mapped[int] = mapped_column(ForeignKey("app_user.id", ondelete="CASCADE"), nullable=False)
    token_hash: Mapped[str] = mapped_column(String(128), unique=True, nullable=False)
    csrf_token_hash: Mapped[str] = mapped_column(String(128), nullable=False)
    created_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)
    expires_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)
    revoked_at: Mapped[datetime | None] = mapped_column(DateTime)

    user: Mapped[AppUser] = relationship(back_populates="sessions")


class AppSetting(Base):
    __tablename__ = "app_setting"

    key: Mapped[str] = mapped_column(String(120), primary_key=True)
    value_json: Mapped[str] = mapped_column(Text, nullable=False)
    updated_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)


class Guild(Base):
    __tablename__ = "guild"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    name: Mapped[str] = mapped_column(String(120), unique=True, nullable=False)
    created_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)


class Battle(Base):
    __tablename__ = "battle"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    battle_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)
    source_filename: Mapped[str] = mapped_column(String(255), nullable=False)
    source_sha256: Mapped[str] = mapped_column(String(64), unique=True, nullable=False)
    original_row_count: Mapped[int] = mapped_column(Integer, nullable=False)
    valid_row_count: Mapped[int] = mapped_column(Integer, nullable=False)
    scoring_rule_version: Mapped[str] = mapped_column(String(80), nullable=False)
    scoring_range_version: Mapped[str] = mapped_column(String(80), nullable=False)
    import_status: Mapped[str] = mapped_column(String(40), nullable=False, default="completed")
    match_result_json: Mapped[str | None] = mapped_column(Text)
    created_by: Mapped[int] = mapped_column(ForeignKey("app_user.id"), nullable=False)
    created_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)

    guilds: Mapped[list["BattleGuild"]] = relationship(back_populates="battle", cascade="all, delete-orphan")
    stats: Mapped[list["BattlePlayerStat"]] = relationship(back_populates="battle", cascade="all, delete-orphan")


class BattleGuild(Base):
    __tablename__ = "battle_guild"
    __table_args__ = (UniqueConstraint("battle_id", "side"), UniqueConstraint("battle_id", "guild_id"))

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    battle_id: Mapped[int] = mapped_column(ForeignKey("battle.id", ondelete="CASCADE"), nullable=False)
    guild_id: Mapped[int] = mapped_column(ForeignKey("guild.id"), nullable=False)
    side: Mapped[str] = mapped_column(String(20), nullable=False)
    member_count: Mapped[int] = mapped_column(Integer, nullable=False)

    battle: Mapped[Battle] = relationship(back_populates="guilds")
    guild: Mapped[Guild] = relationship()


class Player(Base):
    __tablename__ = "player"
    __table_args__ = (UniqueConstraint("guild_id", "canonical_name"),)

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    guild_id: Mapped[int] = mapped_column(ForeignKey("guild.id"), nullable=False)
    canonical_name: Mapped[str] = mapped_column(String(120), nullable=False)
    created_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)

    guild: Mapped[Guild] = relationship()


class CareerAvatar(Base):
    __tablename__ = "career_avatar"

    career: Mapped[str] = mapped_column(String(40), primary_key=True)
    asset_path: Mapped[str] = mapped_column(Text, nullable=False)
    content_sha256: Mapped[str | None] = mapped_column(String(64))
    updated_by: Mapped[int | None] = mapped_column(ForeignKey("app_user.id"))
    updated_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)


class PlayerAvatar(Base):
    __tablename__ = "player_avatar"
    __table_args__ = (CheckConstraint("source IN ('generated','uploaded','career_default')"),)

    player_id: Mapped[int] = mapped_column(ForeignKey("player.id", ondelete="CASCADE"), primary_key=True)
    source: Mapped[str] = mapped_column(String(30), nullable=False)
    seed: Mapped[str | None] = mapped_column(Text)
    asset_path: Mapped[str | None] = mapped_column(Text)
    content_sha256: Mapped[str | None] = mapped_column(String(64))
    updated_by: Mapped[int | None] = mapped_column(ForeignKey("app_user.id"))
    updated_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)


class BattlePlayerStat(Base):
    __tablename__ = "battle_player_stat"
    __table_args__ = (
        UniqueConstraint("battle_id", "battle_guild_id", "player_name_snapshot"),
        Index("idx_stat_battle_side", "battle_id", "battle_guild_id"),
        Index("idx_stat_battle_career", "battle_id", "career"),
        Index("idx_stat_battle_team", "battle_id", "team_leader_snapshot"),
        Index("idx_stat_score", "battle_id", "composite_score"),
        Index("idx_stat_player_history", "player_id", "career", "battle_id"),
    )

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    battle_id: Mapped[int] = mapped_column(ForeignKey("battle.id", ondelete="CASCADE"), nullable=False)
    battle_guild_id: Mapped[int] = mapped_column(ForeignKey("battle_guild.id", ondelete="CASCADE"), nullable=False)
    player_id: Mapped[int | None] = mapped_column(ForeignKey("player.id"))
    player_name_snapshot: Mapped[str] = mapped_column(String(120), nullable=False)
    level: Mapped[int | None] = mapped_column(Integer)
    career: Mapped[str] = mapped_column(String(40), nullable=False)
    team_leader_snapshot: Mapped[str] = mapped_column(String(120), nullable=False)
    kills: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    assists: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    logistics: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    player_damage: Mapped[int] = mapped_column(BigInteger, nullable=False, default=0)
    building_damage: Mapped[int] = mapped_column(BigInteger, nullable=False, default=0)
    healing: Mapped[int] = mapped_column(BigInteger, nullable=False, default=0)
    damage_taken: Mapped[int] = mapped_column(BigInteger, nullable=False, default=0)
    deaths: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    qingdeng: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    revive: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    control: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    kda_ratio: Mapped[float] = mapped_column(Float, nullable=False, default=0)
    player_damage_share: Mapped[float | None] = mapped_column(Float)
    building_damage_share: Mapped[float | None] = mapped_column(Float)
    player_damage_conversion_rate: Mapped[float | None] = mapped_column(Float)
    building_damage_conversion_rate: Mapped[float | None] = mapped_column(Float)
    participation_rate: Mapped[float | None] = mapped_column(Float)
    composite_score: Mapped[float | None] = mapped_column(Float)
    guild_rank: Mapped[int | None] = mapped_column(Integer)
    career_rank: Mapped[int | None] = mapped_column(Integer)
    team_rank: Mapped[int | None] = mapped_column(Integer)
    six_dimension_json: Mapped[str | None] = mapped_column(Text)
    score_detail_json: Mapped[str | None] = mapped_column(Text)

    battle: Mapped[Battle] = relationship(back_populates="stats")
    battle_guild: Mapped[BattleGuild] = relationship()
    player: Mapped[Player | None] = relationship()


class ScoringRangeVersion(Base):
    __tablename__ = "scoring_range_version"

    version: Mapped[str] = mapped_column(String(80), primary_key=True)
    name: Mapped[str] = mapped_column(String(120), nullable=False)
    config_json: Mapped[str] = mapped_column(Text, nullable=False)
    source_method: Mapped[str] = mapped_column(String(120), nullable=False)
    source_battle_id: Mapped[int | None] = mapped_column(ForeignKey("battle.id", ondelete="SET NULL"))
    sample_summary_json: Mapped[str] = mapped_column(Text, nullable=False)
    created_by: Mapped[int] = mapped_column(ForeignKey("app_user.id"), nullable=False)
    created_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)
    is_active: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    is_frozen: Mapped[int] = mapped_column(Integer, nullable=False, default=1)


class ScoringRule(Base):
    __tablename__ = "scoring_rule"
    __table_args__ = (CheckConstraint("status IN ('draft','published','archived')"),)

    version: Mapped[str] = mapped_column(String(80), primary_key=True)
    name: Mapped[str] = mapped_column(String(120), nullable=False)
    status: Mapped[str] = mapped_column(String(20), nullable=False, default="draft")
    config_json: Mapped[str] = mapped_column(Text, nullable=False)
    created_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)
    published_at: Mapped[datetime | None] = mapped_column(DateTime)
    published_by: Mapped[int | None] = mapped_column(ForeignKey("app_user.id"))
    is_active: Mapped[int] = mapped_column(Integer, nullable=False, default=0)


class RangeSuggestion(Base):
    __tablename__ = "range_suggestion"
    __table_args__ = (
        UniqueConstraint("battle_id", "career", "metric"),
        CheckConstraint("method IN ('p05_p95','min_max_margin','very_small_sample')"),
        CheckConstraint("status IN ('draft','published','rejected')"),
    )

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    battle_id: Mapped[int] = mapped_column(ForeignKey("battle.id", ondelete="CASCADE"), nullable=False)
    career: Mapped[str] = mapped_column(String(40), nullable=False)
    metric: Mapped[str] = mapped_column(String(60), nullable=False)
    sample_size: Mapped[int] = mapped_column(Integer, nullable=False)
    method: Mapped[str] = mapped_column(String(40), nullable=False)
    suggested_min: Mapped[float] = mapped_column(Float, nullable=False)
    suggested_max: Mapped[float] = mapped_column(Float, nullable=False)
    margin_ratio: Mapped[float | None] = mapped_column(Float)
    status: Mapped[str] = mapped_column(String(20), nullable=False, default="draft")
    metadata_json: Mapped[str | None] = mapped_column(Text)
    created_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)


class ImportLog(Base):
    __tablename__ = "import_log"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    battle_id: Mapped[int | None] = mapped_column(ForeignKey("battle.id", ondelete="SET NULL"))
    level: Mapped[str] = mapped_column(String(20), nullable=False)
    code: Mapped[str] = mapped_column(String(80), nullable=False)
    row_number: Mapped[int | None] = mapped_column(Integer)
    message: Mapped[str] = mapped_column(Text, nullable=False)
    created_at: Mapped[datetime] = mapped_column(DateTime, nullable=False)
