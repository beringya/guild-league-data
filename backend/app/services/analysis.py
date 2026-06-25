from __future__ import annotations

import math
from collections import defaultdict
from statistics import mean, median
from typing import Iterable

from app.core.defaults import COMPARISON_METRICS, DEFAULT_CAREER_PROFILES, METRIC_LABELS, NEGATIVE_METRICS, NEUTRAL_METRICS

EPSILON = 1e-9


def derive_metrics(rows: list[dict]) -> list[dict]:
    guild_totals = aggregate_by(rows, "guild_name")
    enriched: list[dict] = []
    for row in rows:
        item = dict(row)
        totals = guild_totals[row["guild_name"]]
        item["kda_ratio"] = round((item["kills"] + item["assists"]) / max(item["deaths"], 1), 4)
        item["no_death_badge"] = item["deaths"] == 0
        item["participation_rate"] = safe_div(item["kills"] + item["assists"], totals["kills"])
        item["player_damage_share"] = safe_div(item["player_damage"], totals["player_damage"])
        item["building_damage_share"] = safe_div(item["building_damage"], totals["building_damage"])
        enriched.append(item)
    return enriched


def aggregate_by(rows: Iterable[dict], key: str) -> dict[str, dict]:
    result: dict[str, dict] = defaultdict(lambda: {"member_count": 0})
    for row in rows:
        name = row[key]
        result[name]["member_count"] += 1
        for metric in COMPARISON_METRICS:
            result[name][metric] = result[name].get(metric, 0) + row.get(metric, 0)
    return dict(result)


def suggest_ranges(rows: list[dict]) -> tuple[dict, dict]:
    by_career: dict[str, list[dict]] = defaultdict(list)
    for row in rows:
        by_career[row["career"]].append(row)

    range_config: dict[str, dict] = {}
    sample_summary: dict[str, dict] = {}
    for career, profile in DEFAULT_CAREER_PROFILES.items():
        career_rows = by_career.get(career, [])
        range_config[career] = {}
        sample_summary[career] = {}
        metrics = {dimension["metric"] for dimension in profile["dimensions"]}
        for metric in metrics:
            values = sorted(float(row.get(metric, 0) or 0) for row in career_rows)
            suggestion = _suggest_metric_range(values)
            range_config[career][metric] = {"min": suggestion["min"], "max": suggestion["max"]}
            sample_summary[career][metric] = suggestion
    return range_config, sample_summary


def score_rows(rows: list[dict], rule_config: dict, range_config: dict) -> list[dict]:
    by_career = defaultdict(list)
    for row in rows:
        by_career[row["career"]].append(row)

    scored: list[dict] = []
    for row in rows:
        profile = rule_config["career_profiles"][row["career"]]
        career_ranges = range_config.get(row["career"], {})
        dimensions = []
        composite = 0.0
        first_primary_score = 0.0
        for dimension in profile["dimensions"]:
            metric = dimension["metric"]
            metric_range = career_ranges.get(metric, {"min": 0, "max": 1})
            raw_value = float(row.get(metric, 0) or 0)
            score = normalize(raw_value, metric_range["min"], metric_range["max"], dimension.get("direction", "higher"))
            weight = float(dimension.get("ranking_weight", 0) or 0)
            contribution = score * weight
            if weight > 0 and first_primary_score == 0.0:
                first_primary_score = score
            composite += contribution
            values = [float(peer.get(metric, 0) or 0) for peer in by_career[row["career"]]]
            dimensions.append(
                {
                    "slot": dimension["slot"],
                    "metric": metric,
                    "label": dimension["label"],
                    "raw_value": raw_value,
                    "min": metric_range["min"],
                    "max": metric_range["max"],
                    "direction": dimension.get("direction", "higher"),
                    "score": round(score, 2),
                    "ranking_weight": weight,
                    "contribution": round(contribution, 2),
                    "enabled": dimension.get("enabled", True),
                    "percentile": percentile_rank(values, raw_value),
                    "sample_size": len(values),
                }
            )

        item = dict(row)
        item["six_dimensions"] = dimensions
        item["composite_score"] = round(composite, 2)
        item["primary_metric_score"] = round(first_primary_score, 4)
        item["player_damage_conversion_rate"] = _conversion(dimensions, "player_damage")
        item["building_damage_conversion_rate"] = _conversion(dimensions, "building_damage")
        item["score_detail"] = {
            "rule_version": rule_config["version"],
            "composite_score": item["composite_score"],
            "dimensions": dimensions,
            "advantages": top_dimensions(dimensions, reverse=True),
            "weaknesses": top_dimensions(dimensions, reverse=False),
        }
        scored.append(item)

    assign_ranks(scored)
    return scored


def assign_ranks(rows: list[dict]) -> None:
    for key, rank_field in [
        ("guild_name", "guild_rank"),
        ("career", "career_rank"),
        ("team_leader", "team_rank"),
    ]:
        grouped: dict[str, list[dict]] = defaultdict(list)
        for row in rows:
            group_key = row[key] if key != "team_leader" else f"{row['guild_name']}::{row['team_leader']}"
            grouped[group_key].append(row)
        for group_rows in grouped.values():
            for index, row in enumerate(sorted_for_ranking(group_rows), start=1):
                row[rank_field] = index


def sorted_for_ranking(rows: Iterable[dict]) -> list[dict]:
    return sorted(
        rows,
        key=lambda row: (
            -(row.get("composite_score") or 0),
            -(row.get("primary_metric_score") or 0),
            -(row.get("participation_rate") or 0),
            -(row.get("kda_ratio") or 0),
            row.get("player_name", ""),
        ),
    )


def overview_payload(battle: dict, rows: list[dict], side_map: dict[str, str]) -> dict:
    guild_totals = aggregate_by(rows, "guild_name")
    guilds = list(guild_totals.keys())
    comparison = compare_guilds(guild_totals, side_map)
    rankings = {guild: sorted_for_ranking([row for row in rows if row["guild_name"] == guild])[:3] for guild in guilds}
    return {
        "battle": battle,
        "guild_totals": guild_totals,
        "guild_comparison": comparison,
        "top_players": {guild: [player_summary(row) for row in players] for guild, players in rankings.items()},
        "insights": build_insights(comparison),
    }


def team_top3(rows: list[dict], guild_name: str, metric: str = "composite_score") -> list[dict]:
    grouped: dict[str, list[dict]] = defaultdict(list)
    for row in rows:
        if row["guild_name"] == guild_name:
            grouped[row["team_leader"]].append(row)
    result = []
    for team, team_rows in grouped.items():
        totals = aggregate_totals(team_rows)
        ordered = sorted(team_rows, key=lambda item: (-(item.get(metric) or 0), item["player_name"]))
        result.append(
            {
                "team_leader": team,
                "member_count": len(team_rows),
                "totals": totals,
                "averages": average_totals(totals, len(team_rows)),
                "top_players": [player_summary(row) for row in ordered[:3]],
            }
        )
    return sorted(result, key=lambda item: item["team_leader"])


def squad_comparison(rows: list[dict], scope: str = "all") -> list[dict]:
    grouped: dict[tuple[str, str], list[dict]] = defaultdict(list)
    for row in rows:
        grouped[(row["guild_name"], row["team_leader"])].append(row)
    result = []
    for (guild, team), team_rows in grouped.items():
        totals = aggregate_totals(team_rows)
        result.append(
            {
                "guild_name": guild,
                "team_leader": team,
                "member_count": len(team_rows),
                "totals": totals,
                "averages": average_totals(totals, len(team_rows)),
                "top_players": [player_summary(row) for row in sorted_for_ranking(team_rows)[:3]],
            }
        )
    return sorted(result, key=lambda item: (item["guild_name"], item["team_leader"]))


def compare_guilds(guild_totals: dict[str, dict], side_map: dict[str, str]) -> dict:
    home = next((guild for guild, side in side_map.items() if side == "home"), None)
    opponent = next((guild for guild, side in side_map.items() if side == "opponent"), None)
    if not home or not opponent:
        return {"rows": []}
    home_totals = guild_totals[home]
    opponent_totals = guild_totals[opponent]
    rows = []
    for metric in COMPARISON_METRICS:
        home_value = home_totals.get(metric, 0)
        opponent_value = opponent_totals.get(metric, 0)
        diff = home_value - opponent_value
        diff_rate = diff / max(abs(opponent_value), EPSILON) if opponent_value else None
        direction = metric_direction(metric)
        status = "neutral"
        if direction == "higher":
            status = _status_from_rate(diff_rate)
        elif direction == "lower":
            status = _status_from_rate(-diff_rate if diff_rate is not None else None)
        rows.append(
            {
                "metric": metric,
                "label": METRIC_LABELS[metric],
                "home": home_value,
                "opponent": opponent_value,
                "diff": diff,
                "diff_rate": diff_rate,
                "direction": direction,
                "status": status,
                "home_average": safe_div(home_value, home_totals["member_count"]),
                "opponent_average": safe_div(opponent_value, opponent_totals["member_count"]),
            }
        )
    return {"home_guild": home, "opponent_guild": opponent, "rows": rows}


def build_insights(comparison: dict) -> list[dict]:
    actionable = [row for row in comparison.get("rows", []) if row["direction"] != "neutral" and row["diff_rate"] is not None]
    advantages = sorted([row for row in actionable if row["status"] == "advantage"], key=lambda row: row["diff_rate"], reverse=True)[:3]
    weaknesses = sorted([row for row in actionable if row["status"] == "weakness"], key=lambda row: row["diff_rate"])[:3]
    insights = []
    for row in advantages:
        insights.append(
            {
                "type": "advantage",
                "title": f"{row['label']}优势明显",
                "basis": f"本帮 {format_number(row['home'])}，对手 {format_number(row['opponent'])}，差异 {format_percent(row['diff_rate'])}",
            }
        )
    for row in weaknesses:
        insights.append(
            {
                "type": "weakness",
                "title": f"{row['label']}低于对手",
                "basis": f"本帮 {format_number(row['home'])}，对手 {format_number(row['opponent'])}，差异 {format_percent(row['diff_rate'])}",
            }
        )
    neutral = next((row for row in comparison.get("rows", []) if row["metric"] == "damage_taken"), None)
    if neutral:
        insights.append(
            {
                "type": "watch",
                "title": "承受伤害需结合阵容解释",
                "basis": f"本帮 {format_number(neutral['home'])}，对手 {format_number(neutral['opponent'])}，该指标不直接判定优劣",
            }
        )
    return insights


def aggregate_totals(rows: Iterable[dict]) -> dict:
    totals = {"member_count": 0}
    for row in rows:
        totals["member_count"] += 1
        for metric in COMPARISON_METRICS:
            totals[metric] = totals.get(metric, 0) + row.get(metric, 0)
    return totals


def average_totals(totals: dict, count: int) -> dict:
    return {metric: safe_div(totals.get(metric, 0), count) for metric in COMPARISON_METRICS}


def career_averages(rows: list[dict]) -> list[dict]:
    grouped: dict[tuple[str, str], list[dict]] = defaultdict(list)
    for row in rows:
        grouped[(row["guild_name"], row["career"])].append(row)
    result = []
    for (guild, career), group_rows in grouped.items():
        totals = aggregate_totals(group_rows)
        result.append(
            {
                "guild_name": guild,
                "career": career,
                "member_count": len(group_rows),
                "averages": average_totals(totals, len(group_rows)),
            }
        )
    return sorted(result, key=lambda item: (item["career"], item["guild_name"]))


def player_detail(row: dict, rows: list[dict]) -> dict:
    same_career = [item for item in rows if item["career"] == row["career"]]
    same_home = [item for item in same_career if item["guild_name"] == row["guild_name"]]
    other = [item for item in same_career if item["guild_name"] != row["guild_name"]]
    dimensions = row["six_dimensions"]
    return {
        "player": player_summary(row),
        "raw": {metric: row.get(metric, 0) for metric in COMPARISON_METRICS},
        "derived": {
            "kda_ratio": row["kda_ratio"],
            "participation_rate": row["participation_rate"],
            "player_damage_share": row["player_damage_share"],
            "building_damage_share": row["building_damage_share"],
            "player_damage_conversion_rate": row["player_damage_conversion_rate"],
            "building_damage_conversion_rate": row["building_damage_conversion_rate"],
        },
        "ranks": {"guild": row.get("guild_rank"), "career": row.get("career_rank"), "team": row.get("team_rank")},
        "six_dimensions": dimensions,
        "career_benchmarks": {
            "same_guild_average": dimension_average(dimensions, same_home),
            "opponent_average": dimension_average(dimensions, other),
            "same_career_median": dimension_stat(dimensions, same_career, median),
        },
        "score_detail": row["score_detail"],
        "history": [],
    }


def dimension_average(dimensions: list[dict], rows: list[dict]) -> dict[str, float]:
    return dimension_stat(dimensions, rows, mean)


def dimension_stat(dimensions: list[dict], rows: list[dict], fn) -> dict[str, float]:
    result = {}
    for dimension in dimensions:
        values = [float(row.get(dimension["metric"], 0) or 0) for row in rows]
        result[dimension["metric"]] = round(fn(values), 4) if values else 0
    return result


def player_summary(row: dict) -> dict:
    avatar_url = row.get("avatar_url") or generated_avatar_url(row.get("player_id"), row["player_name"], row["career"])
    return {
        "stat_id": row.get("stat_id") or row.get("id"),
        "player_id": row.get("player_id"),
        "player_name": row["player_name"],
        "guild_name": row["guild_name"],
        "career": row["career"],
        "team_leader": row["team_leader"],
        "composite_score": row.get("composite_score"),
        "guild_rank": row.get("guild_rank"),
        "career_rank": row.get("career_rank"),
        "team_rank": row.get("team_rank"),
        "kills": row.get("kills", 0),
        "assists": row.get("assists", 0),
        "kda_ratio": row.get("kda_ratio", 0),
        "participation_rate": row.get("participation_rate", 0),
        "avatar_url": avatar_url,
    }


def generated_avatar_url(player_id: int | None, player_name: str, career: str) -> str:
    seed = player_id or abs(hash(f"{career}:{player_name}"))
    return f"/api/avatars/generated/{seed % 24}?career={career}"


def normalize(value: float, min_value: float, max_value: float, direction: str = "higher") -> float:
    if math.isclose(max_value, min_value):
        return 100.0 if value >= max_value else 0.0
    if direction == "lower":
        raw = (max_value - value) / (max_value - min_value)
    else:
        raw = (value - min_value) / (max_value - min_value)
    return max(0.0, min(1.0, raw)) * 100


def percentile_rank(values: list[float], value: float) -> float:
    if not values:
        return 0
    lower_or_equal = sum(1 for item in values if item <= value)
    return round(lower_or_equal / len(values) * 100, 2)


def top_dimensions(dimensions: list[dict], reverse: bool) -> list[dict]:
    ordered = sorted(dimensions, key=lambda item: item["score"], reverse=reverse)
    return [{"metric": item["metric"], "label": item["label"], "score": item["score"], "raw_value": item["raw_value"]} for item in ordered[:2]]


def safe_div(value: float, denominator: float) -> float:
    return value / denominator if denominator else 0.0


def metric_direction(metric: str) -> str:
    if metric in NEGATIVE_METRICS:
        return "lower"
    if metric in NEUTRAL_METRICS:
        return "neutral"
    return "higher"


def _conversion(dimensions: list[dict], metric: str) -> float | None:
    target = next((dimension for dimension in dimensions if dimension["metric"] == metric), None)
    return round(target["score"] / 100, 4) if target else None


def _status_from_rate(rate: float | None) -> str:
    if rate is None:
        return "neutral"
    if rate >= 0.05:
        return "advantage"
    if rate <= -0.05:
        return "weakness"
    return "balanced"


def _suggest_metric_range(values: list[float]) -> dict:
    if not values:
        return {"min": 0, "max": 1, "sample_size": 0, "method": "very_small_sample", "warning": "无样本"}
    n = len(values)
    if n >= 20:
        min_value = percentile(values, 5)
        max_value = percentile(values, 95)
        method = "p05_p95"
    elif n >= 3:
        min_value = min(values)
        max_value = max(values)
        span = max(max_value - min_value, 1)
        min_value -= span * 0.1
        max_value += span * 0.1
        method = "min_max_margin"
    else:
        min_value = min(values)
        max_value = max(values)
        step = max(max(abs(min_value), abs(max_value)) * 0.1, 1)
        min_value -= step
        max_value += step
        method = "very_small_sample"
    min_value = max(0, min_value)
    if math.isclose(min_value, max_value):
        max_value = min_value + 1
    return {"min": round(min_value, 4), "max": round(max_value, 4), "sample_size": n, "method": method}


def percentile(values: list[float], percent: float) -> float:
    if not values:
        return 0
    if len(values) == 1:
        return values[0]
    rank = (len(values) - 1) * percent / 100
    lower = math.floor(rank)
    upper = math.ceil(rank)
    if lower == upper:
        return values[int(rank)]
    weight = rank - lower
    return values[lower] * (1 - weight) + values[upper] * weight


def format_percent(value: float | None) -> str:
    if value is None:
        return "对手为 0"
    return f"{value * 100:.1f}%"


def format_number(value: float) -> str:
    if abs(value) >= 100000000:
        return f"{value / 100000000:.2f}亿"
    if abs(value) >= 10000:
        return f"{value / 10000:.1f}万"
    return f"{value:.0f}"
