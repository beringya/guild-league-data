<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from '../components/AppLayout.vue'
import CardPanel from '../components/CardPanel.vue'
import EChart from '../components/EChart.vue'
import EmptyState from '../components/EmptyState.vue'
import { api } from '../api/client'
import type { DimensionScore, ScoredStat } from '../types'
import { compactNumber, percent, score } from '../utils/format'

const route = useRoute()
const missing = ref(false)
const detail = ref<{ player: ScoredStat; dimensions: DimensionScore[]; same_career: any; trend: any[]; score_detail: any } | null>(null)

onMounted(load)

async function load() {
  const battleId = route.params.battleId === 'latest' ? 'latest' : route.params.battleId
  try {
    detail.value = await api.get<{ player: ScoredStat; dimensions: DimensionScore[]; same_career: any; trend: any[]; score_detail: any }>(`/api/battles/${battleId}/players/${route.params.statId}`)
  } catch {
    missing.value = true
  }
}

const radarOption = computed(() => ({
  color: ['#EF6F9F'],
  radar: { indicator: (detail.value?.dimensions || []).map((d) => ({ name: d.label, max: 100 })) },
  series: [{ type: 'radar', areaStyle: { opacity: 0.2 }, data: [{ value: (detail.value?.dimensions || []).map((d) => d.score), name: '本人' }] }]
}))
</script>

<template>
  <AppLayout>
    <EmptyState v-if="missing" title="还没有玩家详情" />
    <template v-else-if="detail">
      <CardPanel>
        <div class="player-head">
          <img class="avatar lg" src="/assets/icons/png/128/user.png" alt="" />
          <div>
            <h2>{{ detail.player.player_name }}</h2>
            <p>{{ detail.player.guild_name }} · {{ detail.player.career }} · {{ detail.player.team_leader }}</p>
          </div>
          <div class="big-score"><label>综合分</label><strong>{{ score(detail.player.composite_score) }}</strong></div>
        </div>
      </CardPanel>
      <div class="hero-row">
        <CardPanel title="六维雷达">
          <EChart :option="radarOption" height="310px" />
        </CardPanel>
        <CardPanel title="核心指标">
          <div class="kpi-grid two">
            <div class="kpi"><label>KDA</label><strong>{{ score(detail.player.kda_ratio) }}</strong></div>
            <div class="kpi"><label>参团率</label><strong>{{ percent(detail.player.participation_rate) }}</strong></div>
            <div class="kpi"><label>玩家伤害</label><strong>{{ compactNumber(detail.player.player_damage) }}</strong></div>
            <div class="kpi"><label>建筑伤害</label><strong>{{ compactNumber(detail.player.building_damage) }}</strong></div>
          </div>
          <p class="mini-note">同职业样本 {{ detail.same_career.sample_size }} 人，本帮平均 {{ score(detail.same_career.home_average) }}，对手平均 {{ score(detail.same_career.opponent_average) }}，百分位 {{ detail.same_career.percentile }}。</p>
        </CardPanel>
      </div>
      <CardPanel title="评分解释">
        <div class="table-wrap">
          <table class="table">
            <thead><tr><th>维度</th><th>原始值</th><th>范围</th><th>维度分</th><th>权重</th><th>贡献</th></tr></thead>
            <tbody>
              <tr v-for="d in detail.dimensions" :key="d.slot">
                <td>{{ d.label }}</td><td>{{ compactNumber(d.value) }}</td><td>{{ compactNumber(d.range_min) }} - {{ compactNumber(d.range_max) }}</td><td>{{ score(d.score) }}</td><td>{{ (d.weight * 100).toFixed(0) }}%</td><td>{{ score(d.contribution) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardPanel>
    </template>
  </AppLayout>
</template>
