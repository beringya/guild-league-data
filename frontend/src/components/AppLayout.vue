<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const nav = [
  { path: '/', label: '首页概览', icon: 'home' },
  { path: '/rankings', label: '个人排名', icon: 'ranking' },
  { path: '/team-top3', label: '团内 TOP3', icon: 'team_top3' },
  { path: '/guild-comparison', label: '对手帮会对比', icon: 'opponent' },
  { path: '/squad-comparison', label: '团队数据对比', icon: 'squads' },
  { path: '/players/latest/1', label: '个人数据分析', icon: 'user' },
  { path: '/import', label: '数据导入', icon: 'import' },
  { path: '/history', label: '历史记录', icon: 'history' },
  { path: '/settings', label: '设置', icon: 'settings' }
]

const title = computed(() => nav.find((item) => item.path === route.path)?.label || route.meta.title || '联赛分析')

async function logout() {
  await auth.logout()
  router.push('/login')
}
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="brand">
        <img src="/assets/brand/logo.svg" alt="" />
        <span>逆水寒联赛分析</span>
      </div>
      <nav class="nav">
        <router-link v-for="item in nav" :key="item.path" :to="item.path" :class="{ active: route.path === item.path }">
          <img :src="`/assets/icons/svg/${item.icon}.svg`" alt="" />
          <span>{{ item.label }}</span>
        </router-link>
      </nav>
      <div class="side-bottom">
        <img src="/assets/brand/mascot_bunny.svg" alt="" />
        <div>v1.0.0 · 管理员模式</div>
      </div>
    </aside>
    <section class="workspace">
      <header class="topbar">
        <div>
          <div class="page-title">{{ title }}</div>
          <div class="subtitle">真实 API 数据流 · 当前账号：{{ auth.user?.username || 'admin' }}</div>
        </div>
        <div class="top-actions">
          <router-link class="btn primary" to="/import"><img src="/assets/icons/svg/import.svg" alt="" />导入数据</router-link>
          <button class="btn" @click="logout"><img src="/assets/icons/svg/logout.svg" alt="" />退出</button>
        </div>
      </header>
      <main class="main">
        <slot />
      </main>
    </section>
  </div>
</template>
