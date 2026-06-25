export type User = {
  id: number;
  username: string;
  is_admin: boolean;
  force_password_change: boolean;
  last_login_at: string | null;
};

export type Battle = {
  id: number;
  battle_at: string;
  source_filename: string;
  source_sha256: string;
  valid_row_count: number;
  original_row_count: number;
  scoring_rule_version: string;
  scoring_range_version: string;
  home_guild: string;
  opponent_guild: string;
  home_member_count: number;
  opponent_member_count: number;
  created_at: string;
};

export type PlayerSummary = {
  stat_id: number;
  player_id: number;
  player_name: string;
  guild_name: string;
  career: string;
  team_leader: string;
  composite_score: number;
  guild_rank?: number;
  career_rank?: number;
  team_rank?: number;
  kills: number;
  assists: number;
  kda_ratio: number;
  participation_rate: number;
  avatar_url: string;
  level?: number;
  player_damage?: number;
  building_damage?: number;
  healing?: number;
  damage_taken?: number;
  deaths?: number;
  qingdeng?: number;
  revive?: number;
  control?: number;
  player_damage_share?: number;
  building_damage_share?: number;
  player_damage_conversion_rate?: number;
  building_damage_conversion_rate?: number;
  [key: string]: unknown;
};

export type MetricRow = {
  metric: string;
  label: string;
  home: number;
  opponent: number;
  diff: number;
  diff_rate: number | null;
  direction: string;
  status: "advantage" | "weakness" | "balanced" | "neutral";
  home_average: number;
  opponent_average: number;
};

export type Insight = {
  type: "advantage" | "weakness" | "watch";
  title: string;
  basis: string;
};
