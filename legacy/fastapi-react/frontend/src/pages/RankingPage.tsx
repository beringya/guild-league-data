import { useEffect, useState } from "react";
import { api } from "../api/client";
import { Card } from "../components/Card";
import { EmptyState } from "../components/EmptyState";
import { PlayerTable } from "../components/PlayerTable";
import type { PlayerSummary } from "../types";

export function RankingPage({
  battleId,
  onImport,
  onOpenPlayer
}: {
  battleId: number | null;
  onImport: () => void;
  onOpenPlayer: (statId: number) => void;
}) {
  const [mode, setMode] = useState<"single" | "history">("single");
  const [side, setSide] = useState("home");
  const [items, setItems] = useState<PlayerSummary[]>([]);
  const [history, setHistory] = useState<any[]>([]);

  useEffect(() => {
    if (!battleId) return;
    if (mode === "single") {
      void api<{ items: PlayerSummary[] }>(`/battles/${battleId}/rankings?side=${side}&page_size=100`).then((response) => setItems(response.items));
    } else {
      void api<{ items: any[] }>("/rankings/players/aggregate?min_matches=1&page_size=100").then((response) => setHistory(response.items));
    }
  }, [battleId, side, mode]);

  if (!battleId) return <EmptyState onImport={onImport} />;

  return (
    <Card
      title="个人综合排名"
      actions={
        <div className="tabs">
          <button className={mode === "single" ? "active" : ""} onClick={() => setMode("single")}>单场排名</button>
          <button className={mode === "history" ? "active" : ""} onClick={() => setMode("history")}>多场总榜</button>
        </div>
      }
    >
      {mode === "single" ? (
        <>
          <div className="filters">
            {[
              ["home", "本帮会"],
              ["opponent", "对手帮会"],
              ["all", "双方合并"]
            ].map(([value, label]) => (
              <button className={side === value ? "seg active" : "seg"} key={value} onClick={() => setSide(value)}>
                {label}
              </button>
            ))}
          </div>
          <PlayerTable players={items} onOpen={onOpenPlayer} />
        </>
      ) : (
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>玩家</th>
                <th>帮会</th>
                <th>职业</th>
                <th>参赛场次</th>
                <th>场均综合分</th>
                <th>累计贡献分</th>
                <th>最高单场</th>
                <th>最近一场</th>
              </tr>
            </thead>
            <tbody>
              {history.map((row) => (
                <tr key={`${row.player_id}-${row.career}`}>
                  <td>{row.player_name}</td>
                  <td>{row.guild_name}</td>
                  <td><span className="badge">{row.career}</span></td>
                  <td>{row.match_count}</td>
                  <td><span className="score">{row.average_composite_score}</span></td>
                  <td>{row.cumulative_contribution_score}</td>
                  <td>{row.best_composite_score}</td>
                  <td>{row.latest_composite_score}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </Card>
  );
}
