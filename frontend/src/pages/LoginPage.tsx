import { Lock, User } from "lucide-react";
import { FormEvent, useState } from "react";
import { useAuth } from "../context/AuthContext";

export function LoginPage() {
  const { login } = useAuth();
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    setError("");
    try {
      await login(username, password);
    } catch (err) {
      setError(err instanceof Error ? err.message : "登录失败");
    }
  };

  return (
    <div className="login-page">
      <div className="login-shell">
        <section className="login-art">
          <div className="logo-line">
            <img src="/assets/brand/logo.svg" alt="" />
            <span>逆水寒联赛分析</span>
          </div>
          <h1>帮会联赛数据分析平台</h1>
          <p>导入联赛 CSV 后，自动生成个人综合排名、团内 TOP3、双方横向对比和职业六维分析。</p>
          <img className="mascot" src="/assets/brand/mascot_bunny.svg" alt="" />
        </section>
        <form className="login-card" onSubmit={submit}>
          <h2>管理员登录</h2>
          <p>首次部署账号为 admin，随机密码只会在首次启动日志中显示。</p>
          <label className="field">
            <span>管理员账号</span>
            <div className="fieldbox">
              <User size={20} />
              <input value={username} onChange={(event) => setUsername(event.target.value)} />
            </div>
          </label>
          <label className="field">
            <span>密码</span>
            <div className="fieldbox">
              <Lock size={20} />
              <input type="password" value={password} onChange={(event) => setPassword(event.target.value)} />
            </div>
          </label>
          {error && <div className="form-error">{error}</div>}
          <button className="login-button">登录</button>
          <div className="login-tip">系统不开放注册和访客入口。首次登录后会要求修改随机初始密码。</div>
        </form>
      </div>
    </div>
  );
}
