<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import AppLayout from '../components/AppLayout.vue'
import CardPanel from '../components/CardPanel.vue'
import EmptyState from '../components/EmptyState.vue'
import { api } from '../api/client'
import type { ScoredStat } from '../types'
import { score } from '../utils/format'

const side = ref('home')
const items = ref<Record<string, ScoredStat[]>>({})
const missing = ref(false)

onMounted(load)
watch(side, load)

async function load() {
  try {
    const res = await api.get<{ items: Record<string, ScoredStat[]> }>(`/api/battles/latest/team-top3?side=${side.value}`)
    items.value = res.items
    missing.value = false
  } catch {
    missing.value = true
  }
}
</script>

<template>
  <AppLayout>
    <EmptyState v-if="missing" />
    <CardPanel v-else title="团内 TOP3">
      <template #actions>
        <select v-model="side" class="select"><option value="home">本帮会</option><option value="opponent">对手帮会</option><option value="">双方全部</option></select>
      </template>
      <div class="team-grid">
        <section v-for="(players, team) in items" :key="team" class="team-card">
          <div class="team-head">
            <strong>{{ team }}</strong>
            <span class="badge">{{ players.length }} 人上榜</span>
          </div>
          <div class="top-list">
            <div v-for="(item, idx) in players" :key="item.row_number" class="top-item">
              <div class="medal" :class="`m${idx + 1}`">{{ idx + 1 }}</div>
              <div>
                <div class="name">{{ item.player_name }}</div>
                <div class="meta">{{ item.career }} · {{ item.guild_name }}</div>
              </div>
              <div class="score">{{ score(item.composite_score) }}</div>
            </div>
          </div>
        </section>
      </div>
    </CardPanel>
  </AppLayout>
</template>
