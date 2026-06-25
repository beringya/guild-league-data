from __future__ import annotations

from copy import deepcopy

METRIC_LABELS = {
    "kills": "击败",
    "assists": "助攻",
    "logistics": "战备资源",
    "player_damage": "对玩家伤害",
    "building_damage": "对建筑伤害",
    "healing": "治疗值",
    "damage_taken": "承受伤害",
    "deaths": "重伤",
    "qingdeng": "青灯焚骨",
    "revive": "化羽",
    "control": "控制",
    "kda_ratio": "KDA",
    "participation_rate": "参团率",
}

RAW_METRICS = [
    "kills",
    "assists",
    "logistics",
    "player_damage",
    "building_damage",
    "healing",
    "damage_taken",
    "deaths",
    "qingdeng",
    "revive",
    "control",
]

COMPARISON_METRICS = [
    "kills",
    "assists",
    "player_damage",
    "building_damage",
    "healing",
    "damage_taken",
    "deaths",
    "qingdeng",
    "revive",
    "control",
]

POSITIVE_METRICS = {
    "kills",
    "assists",
    "player_damage",
    "building_damage",
    "healing",
    "qingdeng",
    "revive",
    "control",
    "participation_rate",
}
NEGATIVE_METRICS = {"deaths"}
NEUTRAL_METRICS = {"damage_taken"}


def dimension(slot: int, metric: str, weight: float = 0, direction: str = "higher") -> dict:
    return {
        "slot": slot,
        "metric": metric,
        "label": METRIC_LABELS[metric],
        "enabled": True,
        "direction": direction,
        "range": {"min": None, "max": None, "source": "published_range_version"},
        "ranking_weight": weight,
    }


DEFAULT_CAREER_PROFILES: dict[str, dict] = {
    "素问": {
        "dimensions": [
            dimension(1, "healing", 0.55),
            dimension(2, "damage_taken", 0.25),
            dimension(3, "revive", 0.20),
            dimension(4, "assists"),
            dimension(5, "participation_rate"),
            dimension(6, "kda_ratio"),
        ]
    },
    "铁衣": {
        "dimensions": [
            dimension(1, "control", 0.60),
            dimension(2, "damage_taken", 0.40),
            dimension(3, "kda_ratio"),
            dimension(4, "participation_rate"),
            dimension(5, "player_damage"),
            dimension(6, "building_damage"),
        ]
    },
    "神相": {
        "dimensions": [
            dimension(1, "player_damage", 0.50),
            dimension(2, "building_damage", 0.50),
            dimension(3, "kills"),
            dimension(4, "kda_ratio"),
            dimension(5, "participation_rate"),
            dimension(6, "control"),
        ]
    },
    "血河": {
        "dimensions": [
            dimension(1, "player_damage", 0.50),
            dimension(2, "building_damage", 0.50),
            dimension(3, "kills"),
            dimension(4, "kda_ratio"),
            dimension(5, "participation_rate"),
            dimension(6, "damage_taken"),
        ]
    },
    "沧澜": {
        "dimensions": [
            dimension(1, "player_damage", 0.50),
            dimension(2, "building_damage", 0.50),
            dimension(3, "kills"),
            dimension(4, "kda_ratio"),
            dimension(5, "participation_rate"),
            dimension(6, "control"),
        ]
    },
    "玄机": {
        "dimensions": [
            dimension(1, "player_damage", 0.50),
            dimension(2, "building_damage", 0.50),
            dimension(3, "kills"),
            dimension(4, "kda_ratio"),
            dimension(5, "participation_rate"),
            dimension(6, "control"),
        ]
    },
    "云瑶": {
        "dimensions": [
            dimension(1, "player_damage", 0.50),
            dimension(2, "building_damage", 0.50),
            dimension(3, "kda_ratio"),
            dimension(4, "participation_rate"),
            dimension(5, "control"),
            dimension(6, "healing"),
        ]
    },
    "碎梦": {
        "dimensions": [
            dimension(1, "kills", 0.50),
            dimension(2, "player_damage", 0.30),
            dimension(3, "building_damage", 0.20),
            dimension(4, "kda_ratio"),
            dimension(5, "assists"),
            dimension(6, "participation_rate"),
        ]
    },
    "九灵": {
        "dimensions": [
            dimension(1, "qingdeng", 0.60),
            dimension(2, "player_damage", 0.20),
            dimension(3, "building_damage", 0.20),
            dimension(4, "kda_ratio"),
            dimension(5, "participation_rate"),
            dimension(6, "control"),
        ]
    },
    "鸿音": {
        "dimensions": [
            dimension(1, "control", 0.55),
            dimension(2, "healing", 0.45),
            dimension(3, "revive"),
            dimension(4, "kda_ratio"),
            dimension(5, "participation_rate"),
            dimension(6, "damage_taken"),
        ]
    },
    "潮光": {
        "template_reference": "神相",
        "dimensions": [
            dimension(1, "player_damage", 0.50),
            dimension(2, "building_damage", 0.50),
            dimension(3, "kills"),
            dimension(4, "kda_ratio"),
            dimension(5, "participation_rate"),
            dimension(6, "control"),
        ],
    },
    "荒羽": {
        "template_reference": "神相",
        "dimensions": [
            dimension(1, "player_damage", 0.50),
            dimension(2, "building_damage", 0.50),
            dimension(3, "kills"),
            dimension(4, "kda_ratio"),
            dimension(5, "participation_rate"),
            dimension(6, "control"),
        ],
    },
    "龙吟": {
        "template_reference": "神相",
        "dimensions": [
            dimension(1, "player_damage", 0.50),
            dimension(2, "building_damage", 0.50),
            dimension(3, "kills"),
            dimension(4, "kda_ratio"),
            dimension(5, "participation_rate"),
            dimension(6, "control"),
        ],
    },
}

for _career, _profile in DEFAULT_CAREER_PROFILES.items():
    _profile.setdefault("status", "confirmed")
    _profile.setdefault("ranking_enabled", True)


DEFAULT_SCORING_RULE = {
    "version": "v1.5-final",
    "status": "published",
    "six_dimension_schema": {
        "required_enabled_slots": 6,
        "ranking_weight_sum_for_active_profile": 1.0,
        "zero_weight_slots_are_analysis_only": True,
    },
    "career_profiles": DEFAULT_CAREER_PROFILES,
}

DEFAULT_SETTINGS = {
    "default_home_guild": None,
    "advantage_threshold": 0.05,
    "timezone": "Asia/Shanghai",
    "backup_retention_days": 14,
    "session_hours": 8,
    "multi_match_min_matches": 3,
}


def default_rule_copy() -> dict:
    return deepcopy(DEFAULT_SCORING_RULE)
