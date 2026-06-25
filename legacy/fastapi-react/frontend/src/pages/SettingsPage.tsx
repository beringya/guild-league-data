import { FormEvent, useEffect, useState } from "react";
import { api, postJson, putJson } from "../api/client";
import { Card } from "../components/Card";
import { useAuth } from "../context/AuthContext";

export function SettingsPage({ battleId }: { battleId?: number | null }) {
  const { changePassword, user } = useAuth();
  const [settings, setSettings] = useState<Record<string, any>>({});
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [message, setMessage] = useState("");
  const [rules, setRules] = useState<any[]>([]);
  const [ranges, setRanges] = useState<any[]>([]);
  const [rangeDraft, setRangeDraft] = useState<any | null>(null);

  useEffect(() => {
    void api<{ settings: Record<string, any> }>("/settings").then((response) => setSettings(response.settings));
    void api<{ items: any[] }>("/scoring-rules").then((response) => setRules(response.items));
    void api<{ items: any[] }>("/scoring-ranges").then((response) => setRanges(response.items));
  }, []);

  const saveSettings = async () => {
    const response = await putJson<{ settings: Record<string, any> }>("/settings", { values: settings });
    setSettings(response.settings);
    setMessage("设置已保存");
  };

  const submitPassword = async (event: FormEvent) => {
    event.preventDefault();
    await changePassword(oldPassword, newPassword);
    setOldPassword("");
    setNewPassword("");
    setMessage("密码已修改，其他会话已失效");
  };

  const suggestRanges = async () => {
    if (!battleId) {
      setMessage("需要先导入或选择一场比赛");
      return;
    }
    const response = await postJson<any>("/scoring-rules/range-suggestions", { battle_id: battleId });
    setRangeDraft(response);
    setMessage("已生成范围建议，确认后可发布为冻结版本");
  };

  const publishRanges = async () => {
    if (!rangeDraft) return;
    await postJson("/scoring-ranges", {
      name: "手动发布职业范围",
      config: rangeDraft.config,
      source_battle_id: battleId,
      sample_summary: rangeDraft.sample_summary
    });
    const response = await api<{ items: any[] }>("/scoring-ranges");
    setRanges(response.items);
    setRangeDraft(null);
    setMessage("职业范围版本已发布");
  };

  return (
    <div className="settings-grid">
      {user?.force_password_change && <div className="force-banner">首次登录需要修改随机初始密码。</div>}
      {message && <div className="success-note span-2">{message}</div>}
      <Card title="通用设置">
        <div className="setting-row">
          <div>
            <h4>默认本帮会</h4>
            <p>导入预览会自动预选该帮会。</p>
          </div>
          <input className="input" value={settings.default_home_guild || ""} onChange={(event) => setSettings({ ...settings, default_home_guild: event.target.value })} />
        </div>
        <div className="setting-row">
          <div>
            <h4>优势阈值</h4>
            <p>默认 5%，用于规则化优势/不足判断。</p>
          </div>
          <input
            className="input"
            type="number"
            min="0"
            step="0.01"
            value={settings.advantage_threshold ?? 0.05}
            onChange={(event) => setSettings({ ...settings, advantage_threshold: Number(event.target.value) })}
          />
        </div>
        <div className="setting-row">
          <div>
            <h4>多场榜最低场次</h4>
            <p>默认 3 场，页面可临时筛选。</p>
          </div>
          <input
            className="input"
            type="number"
            min="1"
            value={settings.multi_match_min_matches ?? 3}
            onChange={(event) => setSettings({ ...settings, multi_match_min_matches: Number(event.target.value) })}
          />
        </div>
        <button className="btn primary" onClick={() => void saveSettings()}>保存设置</button>
      </Card>
      <Card title="管理员密码">
        <form onSubmit={submitPassword} className="password-form">
          <input className="input" type="password" placeholder="旧密码" value={oldPassword} onChange={(event) => setOldPassword(event.target.value)} />
          <input className="input" type="password" placeholder="新密码，至少 10 位" value={newPassword} onChange={(event) => setNewPassword(event.target.value)} />
          <button className="btn primary">修改密码</button>
        </form>
      </Card>
      <Card title="职业评分规则" className="span-2">
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>版本</th>
                <th>名称</th>
                <th>状态</th>
                <th>启用</th>
                <th>职业数</th>
              </tr>
            </thead>
            <tbody>
              {rules.map((rule) => (
                <tr key={rule.version}>
                  <td>{rule.version}</td>
                  <td>{rule.name}</td>
                  <td>{rule.status}</td>
                  <td>{rule.is_active ? "是" : "否"}</td>
                  <td>{Object.keys(rule.config.career_profiles || {}).length}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>
      <Card title="职业范围版本" className="span-2">
        <div className="filters">
          <button className="btn" onClick={() => void suggestRanges()}>按当前比赛生成范围建议</button>
          <button className="btn primary" disabled={!rangeDraft} onClick={() => void publishRanges()}>发布建议范围</button>
        </div>
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>版本</th>
                <th>名称</th>
                <th>来源</th>
                <th>冻结</th>
                <th>启用</th>
              </tr>
            </thead>
            <tbody>
              {ranges.map((range) => (
                <tr key={range.version}>
                  <td>{range.version}</td>
                  <td>{range.name}</td>
                  <td>{range.source_method}</td>
                  <td>{range.is_frozen ? "是" : "否"}</td>
                  <td>{range.is_active ? "是" : "否"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  );
}

export function ForcePasswordPage() {
  return (
    <div className="force-password-page">
      <SettingsPage />
    </div>
  );
}
