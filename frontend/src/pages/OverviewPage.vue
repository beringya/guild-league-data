<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import AppLayout from '../components/AppLayout.vue'
import CardPanel from '../components/CardPanel.vue'
import EChart from '../components/EChart.vue'
import EmptyState from '../components/EmptyState.vue'
import PlayerTable from '../components/PlayerTable.vue'
import { api } from '../api/client'
import type { EChartsOption } from 'echarts'
import type { BattleSummary, Insight, ScoredStat, Totals } from '../types'
import { compactNumber, dateTime } from '../utils/format'

const loading = ref(true)
const missing = ref(false)
const data = ref<{
  battle: BattleSummary
  totals: Record<string, Totals>
  advantages: Insight[]
  weaknesses: Insight[]
  top_home: ScoredStat[]
  top_opponent: ScoredStat[]
} | null>(null)

onMounted(load)

async function load() {
  loading.value = true
  try {
    data.value = await api.get<{
      battle: BattleSummary
      totals: Record<string, Totals>
      advantages: Insight[]
      weaknesses: Insight[]
      top_home: ScoredStat[]
      top_opponent: ScoredStat[]
    }>('/api/battles/latest/overview')
    missing.value = false
  } catch {
    missing.value = true
  } finally {
    loading.value = false
  }
}

const metricCards = computed(() => {
  if (!data.value) return []
  const home = data.value.totals[data.value.battle.home_guild] || {}
  const opponent = data.value.totals[data.value.battle.opponent_guild] || {}
  return [
    ['击败', home.kills, opponent.kills, 'kill'],
    ['助攻', home.assists, opponent.assists, 'assist'],
    ['对玩家伤害', home.player_damage, opponent.player_damage, 'pvp_damage'],
    ['对建筑伤害', home.building_damage, opponent.building_damage, 'building_damage'],
    ['治疗值', home.healing, opponent.healing, 'heal'],
    ['承受伤害', home.damage_taken, opponent.damage_taken, 'damage_taken'],
    ['平均综合分', home.composite_avg, opponent.composite_avg, 'score']
  ]
})

const barOption = computed<EChartsOption>(() => ({
  color: ['#EF6F9F', '#70A8E7'],
  tooltip: {},
  legend: { bottom: 0 },
  grid: { left: 38, right: 16, top: 24, bottom: 44 },
  xAxis: { type: 'category', data: ['击败', '助攻', '玩家伤害', '建筑伤害', '治疗', '控制'] },
  yAxis: { type: 'value' },
  series: [
    { name: data.value?.battle.home_guild || '本帮', type: 'bar', data: metricCards.value.slice(0, 6).map((m) => m[1]) },
    { name: data.value?.battle.opponent_guild || '对手', type: 'bar', data: metricCards.value.slice(0, 6).map((m) => m[2]) }
  ]
}))
</script>

<template>
  <AppLayout>
    <EmptyState v-if="!loading && missing" />
    <template v-else-if="data">
      <div class="summary-strip">
        <CardPanel class="guild-score">
          <div>
            <div class="label">本帮会 · {{ data.battle.home_guild }}</div>
            <strong>{{ data.totals[data.battle.home_guild]?.member_count || 0 }} 人</strong>
          </div>
          <div class="vs">VS</div>
          <div>
            <div class="label">对手 · {{ data.battle.opponent_guild }}</div>
            <strong class="blue">{{ data.totals[data.battle.opponent_guild]?.member_count || 0 }} 人</strong>
          </div>
        </CardPanel>
        <CardPanel v-for="metric in metricCards" :key="String(metric[0])" class="metric">
          <div class="metric-icon"><img :src="`/assets/icons/svg/${metric[3]}.svg`" alt="" /></div>
          <div>
            <div class="label">{{ metric[0] }}</div>
            <div class="value">{{ compactNumber(metric[1] as number) }}</div>
            <div class="delta">对手 {{ compactNumber(metric[2] as number) }}</div>
          </div>
        </CardPanel>
      </div>
      <div class="hero-row">
        <CardPanel title="本帮会 vs 对手帮会 · 核心指标">
          <EChart :option="barOption" height="280px" />
        </CardPanel>
        <CardPanel title="自动分析结论">
          <div v-for="item in [...data.advantages, ...data.weaknesses].slice(0, 5)" :key="item.metric" class="insight">
            <div class="mark" :class="item.delta >= 0 ? 'good' : 'bad'">{{ item.delta >= 0 ? '+' : '!' }}</div>
            <div>
              <strong>{{ item.label }}</strong>
              <p>{{ item.message }}</p>
            </div>
          </div>
        </CardPanel>
      </div>
      <div class="content-row">
        <CardPanel>
          <template #actions>
            <span class="badge">{{ dateTime(data.battle.battle_at) }}</span>
          </template>
          <PlayerTable :items="data.top_home" />
        </CardPanel>
        <CardPanel title="对手 TOP3">
          <div class="top-list">
            <div v-for="(item, idx) in data.top_opponent" :key="item.row_number" class="top-item">
              <div class="medal" :class="`m${idx + 1}`">{{ idx + 1 }}</div>
              <div>
                <div class="name">{{ item.player_name }}</div>
                <div class="meta">{{ item.career }} · {{ item.team_leader }}</div>
              </div>
              <div class="score">{{ item.composite_score.toFixed(2) }}</div>
            </div>
          </div>
        </CardPanel>
      </div>
    </template>
  </AppLayout>
</template>
