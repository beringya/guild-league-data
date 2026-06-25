from pathlib import Path

from app.services import analysis
from app.services.csv_import import parse_csv_file


ROOT = Path(__file__).resolve().parents[2]
SAMPLE = ROOT / "设计文档" / "data" / "sample_battle.csv"
ORIGINAL = ROOT / "联赛初始数据" / "banghuiliansai2026_06_05_20_10_32(1).csv"


def test_sample_csv_profile_matches_spec():
    parsed = parse_csv_file(SAMPLE)

    assert parsed.physical_rows == 180
    assert len(parsed.rows) == 179
    assert parsed.repeated_header_rows_removed == 1
    assert parsed.detected_battle_time is None

    summary = parsed.summary()
    assert summary["guilds"] == {"满月": 90, "星河": 89}
    assert len(summary["careers"]) == 13
    assert not summary["unknown_careers"]


def test_original_filename_can_infer_battle_time():
    parsed = parse_csv_file(ORIGINAL)

    assert parsed.detected_battle_time is not None
    assert parsed.detected_battle_time.strftime("%Y-%m-%d %H:%M:%S") == "2026-06-05 20:10:32"


def test_kda_participation_and_scoring_are_finite():
    parsed = parse_csv_file(SAMPLE)
    rows = analysis.derive_metrics(parsed.rows)
    range_config, _sample = analysis.suggest_ranges(rows)
    from app.core.defaults import default_rule_copy

    scored = analysis.score_rows(rows, default_rule_copy(), range_config)

    assert all(row["kda_ratio"] >= 0 for row in scored)
    assert all(0 <= row["participation_rate"] for row in scored)
    assert all(0 <= row["composite_score"] <= 100 for row in scored)
    suwen = next(row for row in scored if row["career"] == "素问")
    weighted = [item for item in suwen["six_dimensions"] if item["ranking_weight"] > 0]
    assert {item["metric"] for item in weighted} == {"healing", "damage_taken", "revive"}
