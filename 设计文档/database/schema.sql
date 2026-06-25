PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;

CREATE TABLE app_user (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  is_admin INTEGER NOT NULL DEFAULT 1,
  force_password_change INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  last_login_at TEXT
);

CREATE TABLE user_session (
  id TEXT PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL,
  expires_at TEXT NOT NULL,
  revoked_at TEXT
);

CREATE TABLE app_setting (
  key TEXT PRIMARY KEY,
  value_json TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE guild (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL
);

CREATE TABLE battle (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  battle_at TEXT NOT NULL,
  source_filename TEXT NOT NULL,
  source_sha256 TEXT NOT NULL UNIQUE,
  original_row_count INTEGER NOT NULL,
  valid_row_count INTEGER NOT NULL,
  scoring_rule_version TEXT NOT NULL,
  scoring_range_version TEXT NOT NULL,
  import_status TEXT NOT NULL,
  match_result_json TEXT,
  created_by INTEGER NOT NULL REFERENCES app_user(id),
  created_at TEXT NOT NULL
);

CREATE TABLE battle_guild (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  battle_id INTEGER NOT NULL REFERENCES battle(id) ON DELETE CASCADE,
  guild_id INTEGER NOT NULL REFERENCES guild(id),
  side TEXT NOT NULL CHECK(side IN ('home','opponent')),
  member_count INTEGER NOT NULL,
  UNIQUE(battle_id, side),
  UNIQUE(battle_id, guild_id)
);

CREATE TABLE player (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  guild_id INTEGER NOT NULL REFERENCES guild(id),
  canonical_name TEXT NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(guild_id, canonical_name)
);

CREATE TABLE career_avatar (
  career TEXT PRIMARY KEY,
  asset_path TEXT NOT NULL,
  content_sha256 TEXT,
  updated_by INTEGER REFERENCES app_user(id),
  updated_at TEXT NOT NULL
);

CREATE TABLE player_avatar (
  player_id INTEGER PRIMARY KEY REFERENCES player(id) ON DELETE CASCADE,
  source TEXT NOT NULL CHECK(source IN ('generated','uploaded','career_default')),
  seed TEXT,
  asset_path TEXT,
  content_sha256 TEXT,
  updated_by INTEGER REFERENCES app_user(id),
  updated_at TEXT NOT NULL
);

CREATE TABLE battle_player_stat (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  battle_id INTEGER NOT NULL REFERENCES battle(id) ON DELETE CASCADE,
  battle_guild_id INTEGER NOT NULL REFERENCES battle_guild(id) ON DELETE CASCADE,
  player_id INTEGER REFERENCES player(id),
  player_name_snapshot TEXT NOT NULL,
  level INTEGER,
  career TEXT NOT NULL,
  team_leader_snapshot TEXT NOT NULL,
  kills INTEGER NOT NULL DEFAULT 0,
  assists INTEGER NOT NULL DEFAULT 0,
  logistics INTEGER NOT NULL DEFAULT 0,
  player_damage INTEGER NOT NULL DEFAULT 0,
  building_damage INTEGER NOT NULL DEFAULT 0,
  healing INTEGER NOT NULL DEFAULT 0,
  damage_taken INTEGER NOT NULL DEFAULT 0,
  deaths INTEGER NOT NULL DEFAULT 0,
  qingdeng INTEGER NOT NULL DEFAULT 0,
  revive INTEGER NOT NULL DEFAULT 0,
  control INTEGER NOT NULL DEFAULT 0,
  kda_ratio REAL NOT NULL DEFAULT 0,
  player_damage_share REAL,
  building_damage_share REAL,
  player_damage_conversion_rate REAL,
  building_damage_conversion_rate REAL,
  participation_rate REAL,
  composite_score REAL,
  guild_rank INTEGER,
  career_rank INTEGER,
  team_rank INTEGER,
  six_dimension_json TEXT,
  score_detail_json TEXT,
  UNIQUE(battle_id, battle_guild_id, player_name_snapshot)
);

CREATE INDEX idx_stat_battle_side ON battle_player_stat(battle_id, battle_guild_id);
CREATE INDEX idx_stat_battle_career ON battle_player_stat(battle_id, career);
CREATE INDEX idx_stat_battle_team ON battle_player_stat(battle_id, team_leader_snapshot);
CREATE INDEX idx_stat_score ON battle_player_stat(battle_id, composite_score DESC);
CREATE INDEX idx_stat_player_history ON battle_player_stat(player_id, career, battle_id);


CREATE TABLE scoring_range_version (
  version TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  config_json TEXT NOT NULL,
  source_method TEXT NOT NULL,
  source_battle_id INTEGER REFERENCES battle(id) ON DELETE SET NULL,
  sample_summary_json TEXT NOT NULL,
  created_by INTEGER NOT NULL REFERENCES app_user(id),
  created_at TEXT NOT NULL,
  is_active INTEGER NOT NULL DEFAULT 0,
  is_frozen INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE scoring_rule (
  version TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','published','archived')),
  config_json TEXT NOT NULL,
  created_at TEXT NOT NULL,
  published_at TEXT,
  published_by INTEGER REFERENCES app_user(id),
  is_active INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE range_suggestion (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  battle_id INTEGER NOT NULL REFERENCES battle(id) ON DELETE CASCADE,
  career TEXT NOT NULL,
  metric TEXT NOT NULL,
  sample_size INTEGER NOT NULL,
  method TEXT NOT NULL CHECK(method IN ('p05_p95','min_max_margin','very_small_sample')),
  suggested_min REAL NOT NULL,
  suggested_max REAL NOT NULL,
  margin_ratio REAL,
  status TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','published','rejected')),
  metadata_json TEXT,
  created_at TEXT NOT NULL,
  UNIQUE(battle_id, career, metric)
);

CREATE TABLE import_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  battle_id INTEGER REFERENCES battle(id) ON DELETE SET NULL,
  level TEXT NOT NULL,
  code TEXT NOT NULL,
  row_number INTEGER,
  message TEXT NOT NULL,
  created_at TEXT NOT NULL
);
