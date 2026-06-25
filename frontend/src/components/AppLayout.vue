<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../api/client'
import { useAuthStore } from '../stores/auth'
import type { UpdateInfo } from '../types'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const versionInfo = ref<UpdateInfo | null>(null)
const updateOpen = ref(false)
const checkingUpdate = ref(false)
const copiedCommand = ref(false)

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
const currentVersion = computed(() => versionInfo.value?.current_version || '1.0.0')
const latestVersion = computed(() => versionInfo.value?.latest_version || currentVersion.value)
const currentVersionLabel = computed(() => formatVersionLabel(currentVersion.value))
const latestVersionLabel = computed(() => formatVersionLabel(latestVersion.value))
const updateCommand = computed(() => versionInfo.value?.install_command || 'cd /opt/nsh-guild-analytics && sudo bash scripts/install.sh && sudo systemctl restart nsh-guild-analytics')

function formatVersionLabel(version: string) {
  const value = version.trim()
  return value.toLowerCase().startsWith('v') ? value : `v${value}`
}

async function checkUpdate() {
  checkingUpdate.value = true
  try {
    const firstCheck = versionInfo.value === null
    const info = await api.get<UpdateInfo>('/api/system/version')
    versionInfo.value = info
    if (info.update_available || firstCheck) {
      updateOpen.value = Boolean(info.update_available)
    }
  } catch {
    const fallbackVersion = versionInfo.value?.current_version || '1.0.0'
    versionInfo.value = { current_version: fallbackVersion, latest_version: fallbackVersion, update_available: false, channel: 'stable', source: 'local', checked_at: new Date().toISOString(), error: 'check_failed' }
  } finally {
    checkingUpdate.value = false
  }
}

async function copyUpdateCommand() {
  if (!navigator.clipboard) return
  try {
    await navigator.clipboard.writeText(updateCommand.value)
    copiedCommand.value = true
    window.setTimeout(() => { copiedCommand.value = false }, 1800)
  } catch {
    copiedCommand.value = false
  }
}

async function logout() {
  await auth.logout()
  router.push('/login')
}

onMounted(checkUpdate)
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="brand">
        <img src="/assets/brand/logo.svg" alt="" />
        <div class="brand-stack">
          <span>逆水寒联赛分析</span>
          <button class="version-pill" type="button" :class="{ update: versionInfo?.update_available }" @click="updateOpen = !updateOpen">
            {{ currentVersionLabel }}
            <b v-if="versionInfo?.update_available">有更新</b>
          </button>
        </div>
      </div>
      <div v-if="updateOpen" class="update-panel">
        <div class="update-head">
          <strong v-if="versionInfo?.update_available">发现 {{ latestVersionLabel }}</strong>
          <strong v-else>当前 {{ currentVersionLabel }}</strong>
          <button type="button" @click="updateOpen = false">×</button>
        </div>
        <p v-if="versionInfo?.update_available">下载新版本后，在服务器重启服务即可完成更新。</p>
        <p v-else-if="versionInfo?.error === 'update_check_not_configured'">尚未配置更新检查地址。</p>
        <p v-else-if="versionInfo?.error">更新检查暂不可用：{{ versionInfo.error }}</p>
        <p v-else>{{ checkingUpdate ? '正在检查更新...' : '已是当前配置源的最新版本。' }}</p>
        <div v-if="versionInfo?.notes" class="update-notes">{{ versionInfo.notes }}</div>
        <div class="update-actions">
          <a v-if="versionInfo?.download_url" class="btn primary" :href="versionInfo.download_url" target="_blank" rel="noreferrer">下载更新</a>
          <a v-if="versionInfo?.release_url" class="btn" :href="versionInfo.release_url" target="_blank" rel="noreferrer">发布页</a>
          <button class="btn" type="button" @click="checkUpdate" :disabled="checkingUpdate">重新检查</button>
        </div>
        <button class="update-command" type="button" @click="copyUpdateCommand">{{ copiedCommand ? '已复制' : updateCommand }}</button>
      </div>
      <nav class="nav">
        <router-link v-for="item in nav" :key="item.path" :to="item.path" :class="{ active: route.path === item.path }">
          <img :src="`/assets/icons/svg/${item.icon}.svg`" alt="" />
          <span>{{ item.label }}</span>
        </router-link>
      </nav>
      <div class="side-bottom">
        <img src="/assets/brand/mascot_bunny.svg" alt="" />
        <div>管理员模式</div>
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
