import { useEffect, useMemo, useState } from "react";
import { api } from "../api/client";
import { Card } from "../components/Card";
import { EChart } from "../components/charts/EChart";
import { EmptyState } from "../components/EmptyState";
import { formatNumber, percent } from "../utils";

export function PlayerDetailPage({
  battleId,
  statId,
  onImport
}: {
  battleId: number | null;
  statId: number | null;
  onImport: () => void;
}) {
  const [detail, setDetail] = useState<any | null>(null);

  useEffect(() => {
    if (!battleId || !statId) return;
    void api<any>(`/battles/${battleId}/players/${statId}`).then(setDetail);
  }, [battleId, statId]);

  const option = useMemo(() => {
    const dimensions = detail?.six_dimensions ?? [];
    return {
      color: ["#EF6F9F"],
      radar: {
        indicator: dimensions.map((item: any) => ({ name: item.label, max: 100 }))
      },
      series: [{ type: "radar", data: [{ name: "本人", value: dimensions.map((item: any) => item.score) }] }]
    };
  }, [detail]);

  if (!battleId) return <EmptyState onImport={onImport} />;
  if (!statId) return <div className="empty-state"><h2>请选择一名玩家</h2><p>从个人排名中点击玩家即可查看六维分析。</p></div>;
  if (!detail) return <div className="loading">正在读取玩家详情...</div>;

  return (
    <div className="player-detail-grid">
      <Card>
        <div className="player-head">
          <img className="avatar xl" src={detail.player.avatar_url} alt="" />
          <div>
            <h2>{detail.player.player_name}</h2>
            <p>{detail.player.guild_name} · {detail.player.career} · {detail.player.team_leader}</p>
            <div className="badge-row">
              <span className="badge">本场第 {detail.ranks.guild}</span>
              <span className="badge blue">同职业第 {detail.ranks.career}</span>
              <span className="badge green">团内第 {detail.ranks.team}</span>
            </div>
          </div>
          <div className="big-score">
            <span>综合分</span>
            <strong>{formatNumber(detail.player.composite_score)}</strong>
          </div>
        </div>
      </Card>
      <Card title="六维雷达分析">
        <EChart option={option} height={310} />
      </Card>
      <Card title="派生指标">
        <div className="kpi-grid">
          <div className="kpi"><span>KDA</span><strong>{formatNumber(detail.derived.kda_ratio)}</strong></div>
          <div className="kpi"><span>参团率</span><strong>{percent(detail.derived.participation_rate)}</strong></div>
          <div className="kpi"><span>玩家伤害占比</span><strong>{percent(detail.derived.player_damage_share)}</strong></div>
          <div className="kpi"><span>建筑伤害占比</span><strong>{percent(detail.derived.building_damage_share)}</strong></div>
        </div>
      </Card>
      <Card title="评分解释">
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>维度</th>
                <th>原始值</th>
                <th>范围</th>
                <th>标准分</th>
                <th>权重</th>
                <th>贡献</th>
              </tr>
            </thead>
            <tbody>
              {detail.six_dimensions.map((item: any) => (
                <tr key={item.metric}>
                  <td>{item.label}</td>
                  <td>{formatNumber(item.raw_value)}</td>
                  <td>{formatNumber(item.min)} - {formatNumber(item.max)}</td>
                  <td>{item.score}</td>
                  <td>{percent(item.ranking_weight)}</td>
                  <td><span className="score">{item.contribution}</span></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  );
}
