import { useEffect, useState } from "react";
import { api, postJson } from "../api/client";
import { Card } from "../components/Card";
import type { Battle } from "../types";
import { dateTime } from "../utils";

export function HistoryPage({ onOpen }: { onOpen: (battleId: number) => void }) {
  const [items, setItems] = useState<Battle[]>([]);
  const [message, setMessage] = useState("");

  const load = async () => {
    const response = await api<{ items: Battle[] }>("/battles");
    setItems(response.items);
  };

  useEffect(() => {
    void load();
  }, []);

  const backup = async () => {
    const response = await postJson<{ backup: { path: string; integrity_check: string } }>("/backups", {});
    setMessage(`备份完成：${response.backup.integrity_check}`);
  };

  const reanalyze = async (battleId: number) => {
    await postJson(`/battles/${battleId}/reanalyze`, {});
    setMessage("已按当前规则重新分析比赛");
    await load();
  };

  const remove = async (battleId: number) => {
    await api(`/battles/${battleId}`, { method: "DELETE" });
    setMessage("比赛记录已删除");
    await load();
  };

  return (
    <Card title="历史记录" actions={<button className="btn primary" onClick={() => void backup()}>立即备份</button>}>
      {message && <div className="success-note">{message}</div>}
      <div className="table-wrap">
        <table className="table">
          <thead>
            <tr>
              <th>比赛时间</th>
              <th>本帮会</th>
              <th>对手</th>
              <th>记录数</th>
              <th>文件</th>
              <th>规则版本</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {items.map((battle) => (
              <tr key={battle.id}>
                <td>{dateTime(battle.battle_at)}</td>
                <td>{battle.home_guild}</td>
                <td>{battle.opponent_guild}</td>
                <td>{battle.valid_row_count}</td>
                <td>{battle.source_filename}</td>
                <td>{battle.scoring_rule_version}</td>
                <td className="row-actions">
                  <button className="link-btn" onClick={() => onOpen(battle.id)}>查看</button>
                  <button className="link-btn" onClick={() => void reanalyze(battle.id)}>重分析</button>
                  <button className="link-btn danger" onClick={() => void remove(battle.id)}>删除</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Card>
  );
}
