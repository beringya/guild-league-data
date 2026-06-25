import { useEffect, useMemo, useState } from "react";
import { api } from "./api/client";
import { Layout, type PageKey } from "./components/Layout";
import { useAuth } from "./context/AuthContext";
import { ForcePasswordPage } from "./pages/SettingsPage";
import { GuildComparisonPage } from "./pages/GuildComparisonPage";
import { HistoryPage } from "./pages/HistoryPage";
import { ImportPage } from "./pages/ImportPage";
import { LoginPage } from "./pages/LoginPage";
import { OverviewPage } from "./pages/OverviewPage";
import { PlayerDetailPage } from "./pages/PlayerDetailPage";
import { RankingPage } from "./pages/RankingPage";
import { SettingsPage } from "./pages/SettingsPage";
import { SquadComparisonPage } from "./pages/SquadComparisonPage";
import { TeamTop3Page } from "./pages/TeamTop3Page";
import type { Battle } from "./types";
import { dateTime } from "./utils";

const titles: Record<PageKey, string> = {
  overview: "首页概览",
  ranking: "个人综合排名",
  player: "个人数据分析",
  teamTop3: "团内 TOP3",
  guildComparison: "对手帮会对比",
  squadComparison: "团队数据对比",
  import: "数据导入",
  history: "历史记录",
  settings: "设置"
};

export function App() {
  const { user, loading } = useAuth();
  const [page, setPage] = useState<PageKey>("overview");
  const [latestBattle, setLatestBattle] = useState<Battle | null>(null);
  const [selectedPlayer, setSelectedPlayer] = useState<number | null>(null);

  const refreshLatest = async () => {
    const response = await api<{ items: Battle[] }>("/battles");
    setLatestBattle(response.items[0] ?? null);
  };

  useEffect(() => {
    if (user) void refreshLatest();
  }, [user]);

  const subtitle = useMemo(() => {
    if (!latestBattle) return "导入联赛 CSV 后开始分析";
    return `比赛时间：${dateTime(latestBattle.battle_at)} · 本帮会：${latestBattle.home_guild} · 对手：${latestBattle.opponent_guild}`;
  }, [latestBattle]);

  if (loading) return <div className="loading-screen">正在连接应用...</div>;
  if (!user) return <LoginPage />;
  if (user.force_password_change) return <ForcePasswordPage />;

  const battleId = latestBattle?.id ?? null;
  const goImport = () => setPage("import");
  const openPlayer = (statId: number) => {
    setSelectedPlayer(statId);
    setPage("player");
  };

  return (
    <Layout page={page} title={titles[page]} subtitle={subtitle} onNavigate={setPage}>
      {page === "overview" && <OverviewPage battleId={battleId} onImport={goImport} onOpenPlayer={openPlayer} />}
      {page === "ranking" && <RankingPage battleId={battleId} onImport={goImport} onOpenPlayer={openPlayer} />}
      {page === "player" && <PlayerDetailPage battleId={battleId} statId={selectedPlayer} onImport={goImport} />}
      {page === "teamTop3" && <TeamTop3Page battleId={battleId} onImport={goImport} />}
      {page === "guildComparison" && <GuildComparisonPage battleId={battleId} onImport={goImport} />}
      {page === "squadComparison" && <SquadComparisonPage battleId={battleId} onImport={goImport} />}
      {page === "import" && (
        <ImportPage
          onImported={(id) => {
            setPage("overview");
            void api<{ battle: Battle }>(`/battles/${id}`).then((response) => setLatestBattle(response.battle));
          }}
        />
      )}
      {page === "history" && (
        <HistoryPage
          onOpen={(id) => {
            void api<{ battle: Battle }>(`/battles/${id}`).then((response) => setLatestBattle(response.battle));
            setPage("overview");
          }}
        />
      )}
      {page === "settings" && <SettingsPage battleId={battleId} />}
    </Layout>
  );
}
