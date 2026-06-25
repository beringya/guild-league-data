import { useEffect, useMemo, useState } from "react";
import { api } from "../api/client";
import { Card } from "../components/Card";
import { EChart } from "../components/charts/EChart";
import { EmptyState } from "../components/EmptyState";
import { PlayerTable } from "../components/PlayerTable";
import { TopList } from "../components/TopList";
import type { Battle, Insight, MetricRow, PlayerSummary } from "../types";
import { dateTime, formatNumber } from "../utils";

type Overview = {
  battle: Battle;
  guild_totals: Record<string, Record<string, number>>;
  guild_comparison: { home_guild: string; opponent_guild: string; rows: MetricRow[] };
  top_players: Record<string, PlayerSummary[]>;
  insights: Insight[];
};

export function OverviewPage({
  battleId,
  onImport,
  onOpenPlayer
}: {
  battleId: number | null;
  onImport: () => void;
  onOpenPlayer: (statId: number) => void;
}) {
  const [data, setData] = useState<Overview | null>(null);
  const [ranking, setRanking] = useState<PlayerSummary[]>([]);

  useEffect(() => {
    if (!battleId) return;
    void api<Overview>(`/battles/${battleId}/overview`).then(setData);
    void api<{ items: PlayerSummary[] }>(`/battles/${battleId}/rankings?side=home&page_size=8`).then((response) => setRanking(response.items));
  }, [battleId]);

  const chartOption = useMemo(() => {
    const rows = data?.guild_comparison.rows ?? [];
    return {
      color: ["#EF6F9F", "#70A8E7"],
      tooltip: {},
      legend: { bottom: 0 },
      grid: { left: 36, right: 18, top: 24, bottom: 42 },
      xAxis: { type: "category", data: rows.slice(0, 7).map((row) => row.label), axisLabel: { interval: 0 } },
      yAxis: { type: "value" },
      series: [
        { name: data?.guild_comparison.home_guild || "本帮", type: "bar", data: rows.slice(0, 7).map((row) => row.home) },
        { name: data?.guild_comparison.opponent_guild || "对手", type: "bar", data: rows.slice(0, 7).map((row) => row.opponent) }
      ]
    };
  }, [data]);

  const radarOption = useMemo(() => {
    const rows = data?.guild_comparison.rows.slice(0, 6) ?? [];
    return {
      color: ["#EF6F9F", "#70A8E7"],
      radar: {
        indicator: rows.map((row) => ({ name: row.label, max: Math.max(row.home, row.opponent, 1) }))
      },
      series: [
        {
          type: "radar",
          data: [
            { name: data?.guild_comparison.home_guild || "本帮", value: rows.map((row) => row.home) },
            { name: data?.guild_comparison.opponent_guild || "对手", value: rows.map((row) => row.opponent) }
          ]
        }
      ]
    };
  }, [data]);

  if (!battleId) return <EmptyState onImport={onImport} />;
  if (!data) return <div className="loading">正在读取比赛概览...</div>;

  const home = data.guild_comparison.home_guild;
  const opponent = data.guild_comparison.opponent_guild;
  const keyRows = data.guild_comparison.rows.slice(0, 7);

  return (
    <div className="page-grid">
      <section className="summary-strip">
        <Card className="guild-score">
          <div>
            <span>{home}</span>
            <strong>{data.battle.home_member_count} 人</strong>
          </div>
          <em>VS</em>
          <div>
            <span>{opponent}</span>
            <strong className="blue">{data.battle.opponent_member_count} 人</strong>
          </div>
        </Card>
        {keyRows.map((row) => (
          <Card key={row.metric} className="metric-card">
            <span>{row.label}</span>
            <strong>{formatNumber(row.home)}</strong>
            <small className={row.status}>{opponent} {formatNumber(row.opponent)}</small>
          </Card>
        ))}
      </section>
      <div className="hero-row">
        <Card title="本帮会 vs 对手帮会 · 核心指标" hint="总量对比">
          <EChart option={chartOption} height={250} />
        </Card>
        <Card title="能力维度雷达">
          <EChart option={radarOption} height={250} />
        </Card>
        <Card title="自动分析结论" actions={<span className="badge">规则生成</span>}>
          <div className="insight-list">
            {data.insights.map((insight) => (
              <div className={`insight ${insight.type}`} key={insight.title}>
                <b>{insight.type === "advantage" ? "+" : insight.type === "weakness" ? "!" : "i"}</b>
                <div>
                  <strong>{insight.title}</strong>
                  <p>{insight.basis}</p>
                </div>
              </div>
            ))}
          </div>
        </Card>
      </div>
      <div className="content-row">
        <Card title="个人综合排名 · 本帮会" hint={`比赛时间 ${dateTime(data.battle.battle_at)}`}>
          <PlayerTable players={ranking} onOpen={onOpenPlayer} />
        </Card>
        <div className="two-stack">
          <Card>
            <TopList title={`${home} TOP3`} players={data.top_players[home] ?? []} />
          </Card>
          <Card>
            <TopList title={`${opponent} TOP3`} players={data.top_players[opponent] ?? []} />
          </Card>
        </div>
      </div>
    </div>
  );
}
