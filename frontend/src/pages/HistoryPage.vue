<script setup lang="ts">
import { onMounted, ref } from 'vue'
import AppLayout from '../components/AppLayout.vue'
import CardPanel from '../components/CardPanel.vue'
import EmptyState from '../components/EmptyState.vue'
import { api } from '../api/client'
import type { BattleSummary } from '../types'
import { dateTime } from '../utils/format'

const battles = ref<BattleSummary[]>([])
const missing = ref(false)

onMounted(load)

async function load() {
  try {
    const res = await api.get<{ items: BattleSummary[] }>('/api/battles')
    battles.value = res.items
    missing.value = false
  } catch {
    missing.value = true
  }
}

async function remove(id: number) {
  await api.del(`/api/battles/${id}`)
  await load()
}

async function reanalyze(id: number) {
  await api.post(`/api/battles/${id}/reanalyze`)
  await load()
}
</script>

<template>
  <AppLayout>
    <EmptyState v-if="missing || battles.length === 0" />
    <CardPanel v-else title="历史记录">
      <div class="table-wrap">
        <table class="table">
          <thead><tr><th>比赛时间</th><th>本帮会</th><th>对手</th><th>记录数</th><th>源文件</th><th>SHA-256</th><th>操作</th></tr></thead>
          <tbody>
            <tr v-for="battle in battles" :key="battle.id">
              <td>{{ dateTime(battle.battle_at) }}</td><td>{{ battle.home_guild }}</td><td>{{ battle.opponent_guild }}</td><td>{{ battle.valid_row_count }}</td><td>{{ battle.source_filename }}</td><td>{{ battle.source_sha256.slice(0, 12) }}...</td>
              <td><button class="btn" @click="reanalyze(battle.id)">重新分析</button><button class="btn danger" @click="remove(battle.id)">删除</button></td>
            </tr>
          </tbody>
        </table>
      </div>
    </CardPanel>
  </AppLayout>
</template>
