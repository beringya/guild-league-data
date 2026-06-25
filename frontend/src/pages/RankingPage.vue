<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import AppLayout from '../components/AppLayout.vue'
import CardPanel from '../components/CardPanel.vue'
import EmptyState from '../components/EmptyState.vue'
import PlayerTable from '../components/PlayerTable.vue'
import { api } from '../api/client'
import type { ScoredStat } from '../types'

const mode = ref<'single' | 'history'>('single')
const side = ref('home')
const career = ref('')
const search = ref('')
const rows = ref<ScoredStat[]>([])
const historyRows = ref<any[]>([])
const missing = ref(false)

onMounted(load)
watch([mode, side, career, search], load)

async function load() {
  try {
    if (mode.value === 'single') {
      const res = await api.get<{ items: ScoredStat[] }>(`/api/battles/latest/rankings?side=${side.value}&career=${encodeURIComponent(career.value)}&search=${encodeURIComponent(search.value)}`)
      rows.value = res.items
    } else {
      const res = await api.get<{ items: any[] }>(`/api/rankings/history?career=${encodeURIComponent(career.value)}&search=${encodeURIComponent(search.value)}&min_matches=1`)
      historyRows.value = res.items
    }
    missing.value = false
  } catch {
    missing.value = true
  }
}
</script>

<template>
  <AppLayout>
    <EmptyState v-if="missing" />
    <CardPanel v-else title="个人综合排名">
      <template #actions>
        <div class="filters">
          <select v-model="mode" class="select"><option value="single">单场排名</option><option value="history">多场总榜</option></select>
          <select v-if="mode === 'single'" v-model="side" class="select"><option value="home">本帮会</option><option value="opponent">对手</option><option value="all">双方合并</option></select>
          <input v-model="career" class="input" placeholder="职业筛选" />
          <input v-model="search" class="input" placeholder="玩家名搜索" />
        </div>
      </template>
      <PlayerTable v-if="mode === 'single'" :items="rows" />
      <div v-else class="table-wrap">
        <table class="table">
          <thead><tr><th>玩家</th><th>帮会</th><th>职业</th><th>参赛</th><th>场均分</th><th>累计贡献</th><th>最高分</th><th>最近分</th></tr></thead>
          <tbody>
            <tr v-for="item in historyRows" :key="`${item.player_name}-${item.career}`">
              <td>{{ item.player_name }}</td><td>{{ item.guild_name }}</td><td>{{ item.career }}</td><td>{{ item.match_count }}</td>
              <td><span class="score">{{ item.average_score.toFixed(2) }}</span></td><td>{{ item.cumulative_score.toFixed(2) }}</td><td>{{ item.best_score.toFixed(2) }}</td><td>{{ item.latest_score.toFixed(2) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </CardPanel>
  </AppLayout>
</template>
