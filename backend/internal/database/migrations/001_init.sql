CREATE TABLE IF NOT EXISTS app_user (
  id BIGSERIAL PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  is_admin BOOLEAN NOT NULL DEFAULT TRUE,
  force_password_change BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_login_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS user_session (
  id TEXT PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  csrf_token_hash TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS app_setting (
  key TEXT PRIMARY KEY,
  value_json JSONB NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS guild (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS battle (
  id BIGSERIAL PRIMARY KEY,
  battle_at TIMESTAMPTZ NOT NULL,
  source_filename TEXT NOT NULL,
  source_sha256 TEXT NOT NULL UNIQUE,
  original_row_count INT NOT NULL,
  valid_row_count INT NOT NULL,
  scoring_rule_version TEXT NOT NULL,
  scoring_range_version TEXT NOT NULL,
  import_status TEXT NOT NULL,
  match_result_json JSONB,
  created_by BIGINT REFERENCES app_user(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS battle_guild (
  id BIGSERIAL PRIMARY KEY,
  battle_id BIGINT NOT NULL REFERENCES battle(id) ON DELETE CASCADE,
  guild_id BIGINT NOT NULL REFERENCES guild(id),
  side TEXT NOT NULL CHECK(side IN ('home','opponent')),
  member_count INT NOT NULL,
  UNIQUE(battle_id, side),
  UNIQUE(battle_id, guild_id)
);

CREATE TABLE IF NOT EXISTS player (
  id BIGSERIAL PRIMARY KEY,
  guild_id BIGINT NOT NULL REFERENCES guild(id),
  canonical_name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(guild_id, canonical_name)
);

CREATE TABLE IF NOT EXISTS career_avatar (
  career TEXT PRIMARY KEY,
  asset_path TEXT NOT NULL,
  content_sha256 TEXT,
  updated_by BIGINT REFERENCES app_user(id),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS player_avatar (
  player_id BIGINT PRIMARY KEY REFERENCES player(id) ON DELETE CASCADE,
  source TEXT NOT NULL CHECK(source IN ('generated','uploaded','career_default')),
  seed TEXT,
  asset_path TEXT,
  content_sha256 TEXT,
  updated_by BIGINT REFERENCES app_user(id),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS battle_player_stat (
  id BIGSERIAL PRIMARY KEY,
  battle_id BIGINT NOT NULL REFERENCES battle(id) ON DELETE CASCADE,
  battle_guild_id BIGINT NOT NULL REFERENCES battle_guild(id) ON DELETE CASCADE,
  player_id BIGINT REFERENCES player(id),
  guild_name_snapshot TEXT NOT NULL,
  player_name_snapshot TEXT NOT NULL,
  level INT,
  career TEXT NOT NULL,
  team_leader_snapshot TEXT NOT NULL,
  kills INT NOT NULL DEFAULT 0,
  assists INT NOT NULL DEFAULT 0,
  logistics BIGINT NOT NULL DEFAULT 0,
  player_damage BIGINT NOT NULL DEFAULT 0,
  building_damage BIGINT NOT NULL DEFAULT 0,
  healing BIGINT NOT NULL DEFAULT 0,
  damage_taken BIGINT NOT NULL DEFAULT 0,
  deaths INT NOT NULL DEFAULT 0,
  qingdeng INT NOT NULL DEFAULT 0,
  revive INT NOT NULL DEFAULT 0,
  control INT NOT NULL DEFAULT 0,
  kda_ratio DOUBLE PRECISION NOT NULL DEFAULT 0,
  player_damage_share DOUBLE PRECISION,
  building_damage_share DOUBLE PRECISION,
  player_damage_conversion_rate DOUBLE PRECISION,
  building_damage_conversion_rate DOUBLE PRECISION,
  participation_rate DOUBLE PRECISION,
  composite_score DOUBLE PRECISION,
  guild_rank INT,
  career_rank INT,
  team_rank INT,
  six_dimension_json JSONB,
  score_detail_json JSONB,
  UNIQUE(battle_id, battle_guild_id, player_name_snapshot)
);

CREATE INDEX IF NOT EXISTS idx_stat_battle_side ON battle_player_stat(battle_id, battle_guild_id);
CREATE INDEX IF NOT EXISTS idx_stat_battle_career ON battle_player_stat(battle_id, career);
CREATE INDEX IF NOT EXISTS idx_stat_battle_team ON battle_player_stat(battle_id, team_leader_snapshot);
CREATE INDEX IF NOT EXISTS idx_stat_score ON battle_player_stat(battle_id, composite_score DESC);
CREATE INDEX IF NOT EXISTS idx_stat_player_history ON battle_player_stat(player_id, career, battle_id);

CREATE TABLE IF NOT EXISTS scoring_rule (
  version TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','published','archived')),
  config_json JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  published_at TIMESTAMPTZ,
  published_by BIGINT REFERENCES app_user(id),
  is_active BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS scoring_range_version (
  version TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  config_json JSONB NOT NULL,
  source_method TEXT NOT NULL,
  source_battle_id BIGINT REFERENCES battle(id) ON DELETE SET NULL,
  sample_summary_json JSONB NOT NULL,
  created_by BIGINT REFERENCES app_user(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  is_active BOOLEAN NOT NULL DEFAULT FALSE,
  is_frozen BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS range_suggestion (
  id BIGSERIAL PRIMARY KEY,
  battle_id BIGINT REFERENCES battle(id) ON DELETE CASCADE,
  career TEXT NOT NULL,
  metric TEXT NOT NULL,
  sample_size INT NOT NULL,
  method TEXT NOT NULL CHECK(method IN ('p05_p95','min_max_margin','very_small_sample')),
  suggested_min DOUBLE PRECISION NOT NULL,
  suggested_max DOUBLE PRECISION NOT NULL,
  margin_ratio DOUBLE PRECISION,
  status TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','published','rejected')),
  metadata_json JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(battle_id, career, metric)
);

CREATE TABLE IF NOT EXISTS import_log (
  id BIGSERIAL PRIMARY KEY,
  battle_id BIGINT REFERENCES battle(id) ON DELETE SET NULL,
  level TEXT NOT NULL,
  code TEXT NOT NULL,
  row_number INT,
  message TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO app_setting(key, value_json)
VALUES
  ('default_home_guild', '""'::jsonb),
  ('aggregate_min_matches', '3'::jsonb),
  ('comparison_threshold', '0.05'::jsonb),
  ('auto_backup_enabled', 'true'::jsonb)
ON CONFLICT (key) DO NOTHING;
