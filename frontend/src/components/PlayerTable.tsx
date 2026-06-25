import type { PlayerSummary } from "../types";
import { formatNumber, percent } from "../utils";

export function PlayerTable({ players, onOpen }: { players: PlayerSummary[]; onOpen?: (statId: number) => void }) {
  return (
    <div className="table-wrap">
      <table className="table">
        <thead>
          <tr>
            <th>排名</th>
            <th>玩家</th>
            <th>职业</th>
            <th>所在团</th>
            <th>击败</th>
            <th>助攻</th>
            <th>KDA</th>
            <th>玩家伤害</th>
            <th>建筑伤害</th>
            <th>参团率</th>
            <th>综合分</th>
          </tr>
        </thead>
        <tbody>
          {players.map((player, index) => (
            <tr key={player.stat_id} onClick={() => onOpen?.(player.stat_id)} className={onOpen ? "clickable-row" : ""}>
              <td>
                <span className="rank">{player.guild_rank ?? index + 1}</span>
              </td>
              <td className="player-cell">
                <img className="avatar" src={player.avatar_url} alt="" />
                {player.player_name}
              </td>
              <td>
                <span className="badge">{player.career}</span>
              </td>
              <td>{player.team_leader}</td>
              <td>{formatNumber(Number(player.kills))}</td>
              <td>{formatNumber(Number(player.assists))}</td>
              <td>{formatNumber(Number(player.kda_ratio))}</td>
              <td>{formatNumber(Number(player.player_damage || 0))}</td>
              <td>{formatNumber(Number(player.building_damage || 0))}</td>
              <td>{percent(Number(player.participation_rate || 0))}</td>
              <td>
                <span className="score">{formatNumber(Number(player.composite_score))}</span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
