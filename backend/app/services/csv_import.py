from __future__ import annotations

import csv
import hashlib
import io
import re
from collections import Counter, defaultdict
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path

from app.core.defaults import DEFAULT_CAREER_PROFILES

REQUIRED_COLUMNS = [
    "帮会名",
    "玩家",
    "等级",
    "职业",
    "所在团长",
    "击败",
    "助攻",
    "战备资源",
    "对玩家伤害",
    "对建筑伤害",
    "治疗值",
    "承受伤害",
    "重伤",
    "青灯焚骨",
    "化羽",
    "控制",
]

COLUMN_ALIASES = {
    "击杀": "击败",
    "团长": "所在团长",
    "玩家名": "玩家",
    "帮会": "帮会名",
    "拆塔伤害": "对建筑伤害",
    "玩家伤害": "对玩家伤害",
    "死亡": "重伤",
}

COLUMN_TO_FIELD = {
    "帮会名": "guild_name",
    "玩家": "player_name",
    "等级": "level",
    "职业": "career",
    "所在团长": "team_leader",
    "击败": "kills",
    "助攻": "assists",
    "战备资源": "logistics",
    "对玩家伤害": "player_damage",
    "对建筑伤害": "building_damage",
    "治疗值": "healing",
    "承受伤害": "damage_taken",
    "重伤": "deaths",
    "青灯焚骨": "qingdeng",
    "化羽": "revive",
    "控制": "control",
}

TEXT_FIELDS = {"guild_name", "player_name", "career", "team_leader"}
NUMERIC_FIELDS = set(COLUMN_TO_FIELD.values()) - TEXT_FIELDS
ENCODINGS = ["utf-8-sig", "utf-8", "gb18030"]
BATTLE_TIME_RE = re.compile(
    r"(?P<year>\d{4})_(?P<month>\d{2})_(?P<day>\d{2})_(?P<hour>\d{2})_(?P<minute>\d{2})_(?P<second>\d{2})"
)


@dataclass
class ImportIssue:
    level: str
    code: str
    message: str
    row_number: int | None = None
    field: str | None = None

    def as_dict(self) -> dict:
        return {
            "level": self.level,
            "code": self.code,
            "message": self.message,
            "row_number": self.row_number,
            "field": self.field,
        }


@dataclass
class ParsedImport:
    filename: str
    sha256: str
    encoding: str
    rows: list[dict]
    physical_rows: int
    repeated_header_rows_removed: int
    warnings: list[ImportIssue]
    errors: list[ImportIssue]
    detected_battle_time: datetime | None
    columns: list[str]

    def summary(self, default_home_guild: str | None = None) -> dict:
        guild_counter = Counter(row["guild_name"] for row in self.rows)
        team_counts: dict[str, Counter] = defaultdict(Counter)
        career_counts: dict[str, Counter] = defaultdict(Counter)
        for row in self.rows:
            team_counts[row["guild_name"]][row["team_leader"]] += 1
            career_counts[row["guild_name"]][row["career"]] += 1
        guilds = list(guild_counter.keys())
        unknown_careers = sorted({row["career"] for row in self.rows if row["career"] not in DEFAULT_CAREER_PROFILES})
        default_home = default_home_guild if default_home_guild in guild_counter else None
        return {
            "filename": self.filename,
            "sha256": self.sha256,
            "encoding": self.encoding,
            "physical_rows": self.physical_rows,
            "valid_player_rows": len(self.rows),
            "repeated_header_rows_removed": self.repeated_header_rows_removed,
            "columns": self.columns,
            "guilds": dict(guild_counter),
            "teams": {guild: dict(counter) for guild, counter in team_counts.items()},
            "careers": sorted({row["career"] for row in self.rows}),
            "career_counts": {guild: dict(counter) for guild, counter in career_counts.items()},
            "unknown_careers": unknown_careers,
            "detected_battle_time": self.detected_battle_time.isoformat() if self.detected_battle_time else None,
            "suggested_home_guild": default_home or (guilds[0] if len(guilds) == 2 else None),
            "warnings": [issue.as_dict() for issue in self.warnings],
            "errors": [issue.as_dict() for issue in self.errors],
            "can_confirm": not self.errors and len(guild_counter) == 2 and not unknown_careers,
        }


def parse_csv_bytes(content: bytes, filename: str) -> ParsedImport:
    sha256 = hashlib.sha256(content).hexdigest()
    text, encoding = _decode(content)
    stream = io.StringIO(text, newline="")
    reader = csv.reader(stream)
    raw_rows = list(reader)
    warnings: list[ImportIssue] = []
    errors: list[ImportIssue] = []

    raw_rows = [row for row in raw_rows if any(cell.strip() for cell in row)]
    physical_rows = max(len(raw_rows) - 1, 0)
    if not raw_rows:
        errors.append(ImportIssue("error", "empty_file", "CSV 文件为空"))
        return ParsedImport(filename, sha256, encoding, [], 0, 0, warnings, errors, None, [])

    columns = [_normalize_column(cell) for cell in raw_rows[0]]
    missing = [column for column in REQUIRED_COLUMNS if column not in columns]
    if missing:
        errors.append(ImportIssue("error", "missing_columns", f"缺少必需字段：{', '.join(missing)}"))
        return ParsedImport(filename, sha256, encoding, [], physical_rows, 0, warnings, errors, _parse_battle_time(filename), columns)

    index = {column: columns.index(column) for column in REQUIRED_COLUMNS}
    parsed_rows: list[dict] = []
    repeated_headers = 0
    for row_number, raw in enumerate(raw_rows[1:], start=2):
        padded = raw + [""] * (len(columns) - len(raw))
        required_values = [padded[index[column]].strip() if index[column] < len(padded) else "" for column in REQUIRED_COLUMNS]
        if _is_repeated_header(required_values):
            repeated_headers += 1
            continue

        record: dict = {}
        for column in REQUIRED_COLUMNS:
            field = COLUMN_TO_FIELD[column]
            value = padded[index[column]].strip() if index[column] < len(padded) else ""
            if field in TEXT_FIELDS:
                if not value:
                    errors.append(ImportIssue("error", "required_text_blank", f"{column} 不能为空", row_number, column))
                record[field] = value
            else:
                record[field] = _parse_int(value, column, row_number, warnings, errors)

        if record.get("career") and record["career"] not in DEFAULT_CAREER_PROFILES:
            errors.append(ImportIssue("error", "career_not_configured", f"职业规则未配置：{record['career']}", row_number, "职业"))

        parsed_rows.append(record)

    guild_count = Counter(row["guild_name"] for row in parsed_rows if row.get("guild_name"))
    if len(guild_count) != 2:
        errors.append(ImportIssue("error", "guild_count_invalid", f"MVP 要求文件恰好包含两个帮会，当前识别到 {len(guild_count)} 个"))

    return ParsedImport(
        filename=filename,
        sha256=sha256,
        encoding=encoding,
        rows=parsed_rows,
        physical_rows=physical_rows,
        repeated_header_rows_removed=repeated_headers,
        warnings=warnings,
        errors=errors,
        detected_battle_time=_parse_battle_time(filename),
        columns=columns,
    )


def parse_csv_file(path: Path) -> ParsedImport:
    return parse_csv_bytes(path.read_bytes(), path.name)


def _decode(content: bytes) -> tuple[str, str]:
    last_error: UnicodeDecodeError | None = None
    for encoding in ENCODINGS:
        try:
            return content.decode(encoding), encoding
        except UnicodeDecodeError as exc:
            last_error = exc
    if last_error:
        raise ValueError("无法识别 CSV 编码，请使用 UTF-8 或 GB18030") from last_error
    raise ValueError("无法识别 CSV 编码")


def _normalize_column(value: str) -> str:
    normalized = value.strip().replace("\ufeff", "")
    return COLUMN_ALIASES.get(normalized, normalized)


def _is_repeated_header(values: list[str]) -> bool:
    if not values:
        return False
    matches = 0
    for value, column in zip(values, REQUIRED_COLUMNS, strict=False):
        if _normalize_column(value) == column:
            matches += 1
    return matches / len(REQUIRED_COLUMNS) >= 0.8


def _parse_int(
    value: str,
    column: str,
    row_number: int,
    warnings: list[ImportIssue],
    errors: list[ImportIssue],
) -> int:
    cleaned = value.replace(",", "").replace(" ", "").strip()
    if cleaned == "":
        warnings.append(ImportIssue("warning", "empty_number_as_zero", f"{column} 为空，已按 0 处理", row_number, column))
        return 0
    try:
        return int(cleaned)
    except ValueError:
        errors.append(ImportIssue("error", "invalid_number", f"{column} 不是合法整数：{value}", row_number, column))
        return 0


def _parse_battle_time(filename: str) -> datetime | None:
    match = BATTLE_TIME_RE.search(filename)
    if not match:
        return None
    parts = {key: int(value) for key, value in match.groupdict().items()}
    return datetime(parts["year"], parts["month"], parts["day"], parts["hour"], parts["minute"], parts["second"])
