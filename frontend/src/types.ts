export interface User {
  id: number
  username: string
  force_password_change: boolean
}

export interface UpdateInfo {
  current_version: string
  latest_version: string
  update_available: boolean
  channel: string
  release_url?: string
  download_url?: string
  checksum?: string
  notes?: string
  install_command?: string
  apply_enabled: boolean
  apply_command?: string
  source: string
  checked_at: string
  error?: string
}

export interface UpdateApplyResult {
  started: boolean
  latest_version?: string
  command?: string
  message?: string
  error?: string
  started_at: string
}

export interface BattleSummary {
  id: number
  battle_at: string
  source_filename: string
  source_sha256: string
  home_guild: string
  opponent_guild: string
  valid_row_count: number
  created_at: string
}

export interface Totals {
  member_count: number
  kills: number
  assists: number
  player_damage: number
  building_damage: number
  healing: number
  damage_taken: number
  deaths: number
  control: number
  composite_avg: number
}

export interface ScoredStat {
  row_number: number
  side: 'home' | 'opponent'
  guild_name: string
  player_name: string
  level: number
  career: string
  team_leader: string
  kills: number
  assists: number
  logistics: number
  player_damage: number
  building_damage: number
  healing: number
  damage_taken: number
  deaths: number
  qingdeng: number
  revive: number
  control: number
  kda_ratio: number
  participation_rate: number
  player_damage_share: number
  building_damage_share: number
  player_damage_conversion_rate: number
  building_damage_conversion_rate: number
  composite_score: number
  guild_rank: number
  career_rank: number
  team_rank: number
  six_dimensions: DimensionScore[]
  score_detail: Record<string, unknown>
}

export interface RawStat {
  row_number: number
  guild_name: string
  player_name: string
  level: number
  career: string
  team_leader: string
  kills: number
  assists: number
  logistics: number
  player_damage: number
  building_damage: number
  healing: number
  damage_taken: number
  deaths: number
  qingdeng: number
  revive: number
  control: number
}

export interface DimensionScore {
  slot: number
  metric: string
  label: string
  value: number
  score: number
  weight: number
  contribution: number
  percentile: number
  range_min: number
  range_max: number
  range_source: string
}

export interface Insight {
  metric: string
  label: string
  home: number
  opponent: number
  delta: number
  message: string
}

export interface ImportPreview {
  token: string
  source_filename: string
  source_sha256: string
  original_row_count: number
  valid_row_count: number
  inferred_battle_at?: string
  guilds: Array<{ name: string; member_count: number; careers: Record<string, number>; teams: Record<string, number> }>
  preview_rows: RawStat[]
  warnings: Array<{ code: string; message: string; row_number?: number }>
  errors: Array<{ code: string; message: string; row_number?: number }>
  unknown_careers: string[]
  range_suggestions?: Array<{ career: string; metric: string; sample_size: number; method: string; suggested_min: number; suggested_max: number; warning?: string }>
}
