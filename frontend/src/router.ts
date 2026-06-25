import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from './stores/auth'
import LoginPage from './pages/LoginPage.vue'
import OverviewPage from './pages/OverviewPage.vue'
import RankingPage from './pages/RankingPage.vue'
import PlayerDetailPage from './pages/PlayerDetailPage.vue'
import TeamTop3Page from './pages/TeamTop3Page.vue'
import GuildComparisonPage from './pages/GuildComparisonPage.vue'
import SquadComparisonPage from './pages/SquadComparisonPage.vue'
import ImportPage from './pages/ImportPage.vue'
import HistoryPage from './pages/HistoryPage.vue'
import SettingsPage from './pages/SettingsPage.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: LoginPage, meta: { public: true } },
    { path: '/', component: OverviewPage },
    { path: '/rankings', component: RankingPage },
    { path: '/players/:battleId/:statId', component: PlayerDetailPage },
    { path: '/team-top3', component: TeamTop3Page },
    { path: '/guild-comparison', component: GuildComparisonPage },
    { path: '/squad-comparison', component: SquadComparisonPage },
    { path: '/import', component: ImportPage },
    { path: '/history', component: HistoryPage },
    { path: '/settings', component: SettingsPage }
  ]
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!auth.user && !auth.loading) {
    await auth.loadMe()
  }
  if (!to.meta.public && !auth.user) {
    return '/login'
  }
  if (to.meta.public && auth.user) {
    return '/'
  }
  return true
})
