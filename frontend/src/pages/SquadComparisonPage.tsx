import { useEffect, useMemo, useState } from "react";
import { api } from "../api/client";
import { Card } from "../components/Card";
import { EChart } from "../components/charts/EChart";
import { EmptyState } from "../components/EmptyState";
import { formatNumber } from "../utils";

export function SquadComparisonPage({ battleId, onImport }: { battleId: number | null; onImport: () => void }) {
  const [data, setData] = useState<any | null>(null);
  useEffect(() => {
    if (!battleId) return;
    void api<any>(`/battles/${battleId}/squad-comparison`).then(setData);
  }, [battleId]);

  const option = useMemo(() => {
    const squads = data?.squads ?? [];
    return {
      color: ["#EF6F9F"],
      tooltip: {},
      grid: { left: 80, right: 20, top: 20, bottom: 20 },
      xAxis: { type: "value" },
      yAxis: { type: "category", data: squads.map((item: any) => `${item.guild_name} · ${item.team_leader}`) },
      series: [{ type: "bar", data: squads.map((item: any) => item.averages.player_damage) }]
    };
  }, [data]);

  if (!battleId) return <EmptyState onImport={onImport} />;
  if (!data) return <div className="loading">正在读取分团对比...</div>;

  return (
    <div className="page-grid">
      <Card title="分团玩家伤害人均">
        <EChart option={option} height={380} />
      </Card>
      <Card title="全部分团明细">
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>帮会</th>
                <th>分团</th>
                <th>人数</th>
                <th>人均玩家伤害</th>
                <th>人均建筑伤害</th>
                <th>人均治疗</th>
                <th>人均控制</th>
              </tr>
            </thead>
            <tbody>
              {data.squads.map((row: any) => (
                <tr key={`${row.guild_name}-${row.team_leader}`}>
                  <td>{row.guild_name}</td>
                  <td>{row.team_leader}</td>
                  <td>{row.member_count}</td>
                  <td>{formatNumber(row.averages.player_damage)}</td>
                  <td>{formatNumber(row.averages.building_damage)}</td>
                  <td>{formatNumber(row.averages.healing)}</td>
                  <td>{formatNumber(row.averages.control)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  );
}
