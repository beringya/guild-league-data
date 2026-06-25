<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import AppLayout from '../components/AppLayout.vue'
import CardPanel from '../components/CardPanel.vue'
import EmptyState from '../components/EmptyState.vue'
import { api } from '../api/client'
import { compactNumber, score } from '../utils/format'

const side = ref('')
const data = ref<{ squads: any[] } | null>(null)
const missing = ref(false)

onMounted(load)
watch(side, load)

async function load() {
  try {
    data.value = await api.get<{ squads: any[] }>(`/api/battles/latest/squad-comparison?side=${side.value}`)
    missing.value = false
  } catch {
    missing.value = true
  }
}
</script>

<template>
  <AppLayout>
    <EmptyState v-if="missing" />
    <CardPanel v-else-if="data" title="团队数据对比">
      <template #actions>
        <select v-model="side" class="select"><option value="">双方全部</option><option value="home">仅本帮</option><option value="opponent">仅对手</option></select>
      </template>
      <div class="table-wrap">
        <table class="table">
          <thead><tr><th>帮会</th><th>所在团长</th><th>人数</th><th>击败</th><th>助攻</th><th>玩家伤害</th><th>建筑伤害</th><th>平均综合分</th><th>TOP3</th></tr></thead>
          <tbody>
            <tr v-for="squad in data.squads" :key="`${squad.guild_name}-${squad.team_leader}`">
              <td>{{ squad.guild_name }}</td><td>{{ squad.team_leader }}</td><td>{{ squad.member_count }}</td>
              <td>{{ squad.totals.kills }}</td><td>{{ squad.totals.assists }}</td><td>{{ compactNumber(squad.totals.player_damage) }}</td><td>{{ compactNumber(squad.totals.building_damage) }}</td><td>{{ score(squad.totals.composite_avg) }}</td>
              <td>{{ (squad.top3 || []).map((p: any) => p.player_name).join('、') }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </CardPanel>
  </AppLayout>
</template>
