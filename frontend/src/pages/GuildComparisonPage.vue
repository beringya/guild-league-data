<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import AppLayout from '../components/AppLayout.vue'
import CardPanel from '../components/CardPanel.vue'
import EChart from '../components/EChart.vue'
import EmptyState from '../components/EmptyState.vue'
import { api } from '../api/client'
import type { BattleSummary, Insight, Totals } from '../types'
import { compactNumber } from '../utils/format'

const data = ref<{ battle: BattleSummary; totals: Record<string, Totals>; per_capita: Record<string, Totals>; careers: any[]; insights: Insight[] } | null>(null)
const missing = ref(false)

onMounted(async () => {
  try {
    data.value = await api.get<{ battle: BattleSummary; totals: Record<string, Totals>; per_capita: Record<string, Totals>; careers: any[]; insights: Insight[] }>('/api/battles/latest/guild-comparison')
  } catch {
    missing.value = true
  }
})

const chartOption = computed(() => {
  const battle = data.value?.battle
  const totals = data.value?.totals || {}
  const home = battle ? totals[battle.home_guild] : undefined
  const opponent = battle ? totals[battle.opponent_guild] : undefined
  return {
    color: ['#EF6F9F', '#70A8E7'],
    tooltip: {},
    legend: { bottom: 0 },
    radar: { indicator: ['击败', '助攻', '玩家伤害', '建筑伤害', '治疗', '控制'].map((name) => ({ name, max: 100 })) },
    series: [{
      type: 'radar',
      data: [
        { name: battle?.home_guild || '本帮', value: [home?.kills, home?.assists, home?.player_damage, home?.building_damage, home?.healing, home?.control].map(Number) },
        { name: battle?.opponent_guild || '对手', value: [opponent?.kills, opponent?.assists, opponent?.player_damage, opponent?.building_damage, opponent?.healing, opponent?.control].map(Number) }
      ]
    }]
  }
})
</script>

<template>
  <AppLayout>
    <EmptyState v-if="missing" />
    <template v-else-if="data">
      <div class="compare-grid">
        <CardPanel title="双方总量对比">
          <div class="table-wrap">
            <table class="table compare-table">
              <thead><tr><th>指标</th><th>{{ data.battle.home_guild }}</th><th>{{ data.battle.opponent_guild }}</th></tr></thead>
              <tbody>
                <tr v-for="key in ['kills','assists','player_damage','building_damage','healing','damage_taken','control','composite_avg']" :key="key">
                  <td>{{ key }}</td>
                  <td>{{ compactNumber((data.totals[data.battle.home_guild] as any)[key]) }}</td>
                  <td>{{ compactNumber((data.totals[data.battle.opponent_guild] as any)[key]) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </CardPanel>
        <CardPanel title="规则化结论">
          <div v-for="item in data.insights" :key="item.metric" class="insight">
            <div class="mark" :class="item.delta >= 0 ? 'good' : 'bad'">{{ item.delta >= 0 ? '+' : '!' }}</div>
            <div><strong>{{ item.label }}</strong><p>{{ item.message }}</p></div>
          </div>
        </CardPanel>
      </div>
      <CardPanel title="职业人均对比">
        <div class="table-wrap">
          <table class="table">
            <thead><tr><th>职业</th><th>本帮人数</th><th>对手人数</th><th>本帮均分</th><th>对手均分</th></tr></thead>
            <tbody><tr v-for="career in data.careers" :key="career.career"><td>{{ career.career }}</td><td>{{ career.home_count }}</td><td>{{ career.opponent_count }}</td><td>{{ career.home_avg_score }}</td><td>{{ career.opponent_avg_score }}</td></tr></tbody>
          </table>
        </div>
      </CardPanel>
      <CardPanel title="结构雷达">
        <EChart :option="chartOption" height="340px" />
      </CardPanel>
    </template>
  </AppLayout>
</template>
