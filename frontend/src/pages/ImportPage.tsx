import { ChangeEvent, useState } from "react";
import { upload } from "../api/client";
import { Card } from "../components/Card";

type Preview = {
  filename: string;
  sha256: string;
  valid_player_rows: number;
  repeated_header_rows_removed: number;
  guilds: Record<string, number>;
  teams: Record<string, Record<string, number>>;
  careers: string[];
  detected_battle_time: string | null;
  suggested_home_guild: string | null;
  warnings: { message: string }[];
  errors: { message: string }[];
  can_confirm: boolean;
};

export function ImportPage({ onImported }: { onImported: (battleId: number) => void }) {
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<Preview | null>(null);
  const [homeGuild, setHomeGuild] = useState("");
  const [error, setError] = useState("");

  const chooseFile = async (event: ChangeEvent<HTMLInputElement>) => {
    const selected = event.target.files?.[0] ?? null;
    setFile(selected);
    setPreview(null);
    setError("");
    if (!selected) return;
    const data = new FormData();
    data.append("file", selected);
    try {
      const response = await upload<Preview>("/battles/import/preview", data);
      setPreview(response);
      setHomeGuild(response.suggested_home_guild || Object.keys(response.guilds)[0] || "");
    } catch (err) {
      setError(err instanceof Error ? err.message : "预览失败");
    }
  };

  const confirm = async () => {
    if (!file) return;
    const data = new FormData();
    data.append("file", file);
    data.append("home_guild", homeGuild);
    if (preview?.detected_battle_time) data.append("battle_at", preview.detected_battle_time);
    try {
      const response = await upload<{ battle: { id: number } }>("/battles/import/confirm", data);
      onImported(response.battle.id);
    } catch (err) {
      setError(err instanceof Error ? err.message : "导入失败");
    }
  };

  return (
    <div className="import-layout">
      <Card title="数据导入" hint="CSV 预览、校验、本帮会选择">
        <label className="dropzone">
          <input type="file" accept=".csv,text/csv" onChange={chooseFile} />
          <img src="/assets/icons/svg/import.svg" alt="" />
          <h3>选择或拖入联赛 CSV</h3>
          <p>系统会自动检测编码、清理重复表头并识别双方帮会。</p>
        </label>
        {error && <div className="form-error">{error}</div>}
      </Card>
      <Card title="预览结果">
        {!preview ? (
          <p className="muted">选择 CSV 后显示导入预览。</p>
        ) : (
          <div className="preview">
            <div className="kpi-grid">
              <div className="kpi"><span>有效玩家</span><strong>{preview.valid_player_rows}</strong></div>
              <div className="kpi"><span>重复表头</span><strong>{preview.repeated_header_rows_removed}</strong></div>
              <div className="kpi"><span>职业数</span><strong>{preview.careers.length}</strong></div>
              <div className="kpi"><span>SHA-256</span><strong>{preview.sha256.slice(0, 10)}...</strong></div>
            </div>
            <div className="guild-choice">
              <span>选择本帮会</span>
              {Object.entries(preview.guilds).map(([guild, count]) => (
                <button className={homeGuild === guild ? "seg active" : "seg"} key={guild} onClick={() => setHomeGuild(guild)}>
                  {guild} · {count} 人
                </button>
              ))}
            </div>
            <div className="issue-list">
              {preview.errors.map((item) => <div className="issue error" key={item.message}>{item.message}</div>)}
              {preview.warnings.map((item) => <div className="issue warning" key={item.message}>{item.message}</div>)}
            </div>
            <button className="btn primary" disabled={!preview.can_confirm || !homeGuild} onClick={() => void confirm()}>
              确认导入并分析
            </button>
          </div>
        )}
      </Card>
    </div>
  );
}
