import { BarChart3, History, Home, LogOut, Settings, Shield, Swords, Upload, User, Users } from "lucide-react";
import type { ReactNode } from "react";
import { useAuth } from "../context/AuthContext";

export type PageKey =
  | "overview"
  | "ranking"
  | "player"
  | "teamTop3"
  | "guildComparison"
  | "squadComparison"
  | "import"
  | "history"
  | "settings";

const nav = [
  { key: "overview", label: "首页概览", icon: Home },
  { key: "ranking", label: "个人排名", icon: BarChart3 },
  { key: "teamTop3", label: "团内 TOP3", icon: Shield },
  { key: "guildComparison", label: "对手帮会对比", icon: Swords },
  { key: "squadComparison", label: "团队数据对比", icon: Users },
  { key: "player", label: "个人数据分析", icon: User },
  { key: "import", label: "数据导入", icon: Upload },
  { key: "history", label: "历史记录", icon: History },
  { key: "settings", label: "设置", icon: Settings }
] as const;

export function Layout({
  page,
  title,
  subtitle,
  onNavigate,
  children
}: {
  page: PageKey;
  title: string;
  subtitle: string;
  onNavigate: (page: PageKey) => void;
  children: ReactNode;
}) {
  const { logout } = useAuth();
  return (
    <div className="app-shell">
      <aside className="sidebar">
        <button className="brand" onClick={() => onNavigate("overview")}>
          <img src="/assets/brand/logo.svg" alt="" />
          <span>逆水寒联赛分析</span>
        </button>
        <nav className="nav">
          {nav.map((item) => {
            const NavIcon = item.icon;
            return (
              <button key={item.key} className={page === item.key ? "active" : ""} onClick={() => onNavigate(item.key)}>
                <NavIcon size={20} />
                <span>{item.label}</span>
              </button>
            );
          })}
        </nav>
        <div className="side-bottom">
          <img src="/assets/brand/mascot_bunny.svg" alt="" />
          <span>v1.0.0 · 管理员模式</span>
        </div>
      </aside>
      <section className="workspace">
        <header className="topbar">
          <div>
            <h1>{title}</h1>
            <p>{subtitle}</p>
          </div>
          <div className="top-actions">
            <button className="btn" onClick={() => onNavigate("history")}>
              <History size={18} />
              历史记录
            </button>
            <button className="btn primary" onClick={() => onNavigate("import")}>
              <Upload size={18} />
              导入数据
            </button>
            <button className="btn" onClick={() => void logout()}>
              <LogOut size={18} />
              退出
            </button>
          </div>
        </header>
        <main className="main">{children}</main>
      </section>
    </div>
  );
}
