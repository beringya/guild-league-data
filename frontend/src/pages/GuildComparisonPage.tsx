import { useEffect, useMemo, useState } from "react";
import { api } from "../api/client";
import { Card } from "../components/Card";
import { EChart } from "../components/charts/EChart";
import { EmptyState } from "../components/EmptyState";
import type { MetricRow } from "../types";
import { formatNumber, percent } from "../utils";

export function GuildComparisonPage({ battleId, onImport }: { battleId: number | null; onImport: () => void }) {
  const [data, setData] = useState<any | null>(null);
  useEffect(() => {
    if (!battleId) return;
    void api<any>(`/battles/${battleId}/guild-comparison`).then(setData);
  }, [battleId]);

  const option = useMemo(() => {
    const rows: MetricRow[] = data?.rows ?? [];
    return {
      color: ["#EF6F9F", "#70A8E7"],
      tooltip: {},
      legend: { bottom: 0 },
      xAxis: { type: "category", data: rows.map((row) => row.label) },
      yAxis: { type: "value" },
      series: [
        { name: data?.home_guild || "本帮", type: "bar", data: rows.map((row) => row.home_average) },
        { name: data?.opponent_guild || "对手", type: "bar", data: rows.map((row) => row.opponent_average) }
      ]
    };
  }, [data]);

  if (!battleId) return <EmptyState onImport={onImport} />;
  if (!data) return <div className="loading">正在读取帮会对比...</div>;

  return (
    <div className="compare-grid">
      <Card title="双方人均指标对比">
        <EChart option={option} height={300} />
      </Card>
      <Card title="规则化结论">
        {data.insights.map((item: any) => (
          <div className={`insight ${item.type}`} key={item.title}>
            <b>{item.type === "advantage" ? "+" : item.type === "weakness" ? "!" : "i"}</b>
            <div><strong>{item.title}</strong><p>{item.basis}</p></div>
          </div>
        ))}
      </Card>
      <Card title="总量、人均和差异" className="span-2">
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>指标</th>
                <th>{data.home_guild}</th>
                <th>{data.opponent_guild}</th>
                <th>差异</th>
                <th>差异百分比</th>
                <th>判断</th>
              </tr>
            </thead>
            <tbody>
              {data.rows.map((row: MetricRow) => (
                <tr key={row.metric}>
                  <td>{row.label}</td>
                  <td>{formatNumber(row.home)}</td>
                  <td>{formatNumber(row.opponent)}</td>
                  <td>{formatNumber(row.diff)}</td>
                  <td>{percent(row.diff_rate)}</td>
                  <td><span className={`badge ${row.status}`}>{row.status}</span></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  );
}
