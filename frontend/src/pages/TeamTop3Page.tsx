import { useEffect, useState } from "react";
import { api } from "../api/client";
import { Card } from "../components/Card";
import { EmptyState } from "../components/EmptyState";
import { TopList } from "../components/TopList";
import { formatNumber } from "../utils";

export function TeamTop3Page({ battleId, onImport }: { battleId: number | null; onImport: () => void }) {
  const [side, setSide] = useState("home");
  const [data, setData] = useState<any | null>(null);

  useEffect(() => {
    if (!battleId) return;
    void api<any>(`/battles/${battleId}/team-top3?side=${side}`).then(setData);
  }, [battleId, side]);

  if (!battleId) return <EmptyState onImport={onImport} />;

  return (
    <Card
      title="团内 TOP3"
      actions={
        <div className="tabs">
          <button className={side === "home" ? "active" : ""} onClick={() => setSide("home")}>本帮会</button>
          <button className={side === "opponent" ? "active" : ""} onClick={() => setSide("opponent")}>对手帮会</button>
        </div>
      }
    >
      <div className="team-grid">
        {(data?.teams ?? []).map((team: any) => (
          <section className="team-card" key={team.team_leader}>
            <header>
              <strong>{team.team_leader}</strong>
              <span>{team.member_count} 人</span>
            </header>
            <div className="team-summary">
              <span>击败 {formatNumber(team.totals.kills)}</span>
              <span>助攻 {formatNumber(team.totals.assists)}</span>
              <span>控制 {formatNumber(team.totals.control)}</span>
            </div>
            <TopList title="" players={team.top_players} />
          </section>
        ))}
      </div>
    </Card>
  );
}
