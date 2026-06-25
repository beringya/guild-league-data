import type { PlayerSummary } from "../types";
import { formatNumber } from "../utils";

export function TopList({ title, players }: { title: string; players: PlayerSummary[] }) {
  return (
    <div>
      <div className="mini-title">{title}</div>
      <div className="top-list">
        {players.slice(0, 3).map((player, index) => (
          <div className="top-item" key={player.stat_id}>
            <div className={`medal m${index + 1}`}>{index + 1}</div>
            <img className="avatar" src={player.avatar_url} alt="" />
            <div>
              <strong>{player.player_name}</strong>
              <span>{player.career} · {player.team_leader}</span>
            </div>
            <b className="score">{formatNumber(player.composite_score)}</b>
          </div>
        ))}
      </div>
    </div>
  );
}
