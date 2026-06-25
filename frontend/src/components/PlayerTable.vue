<script setup lang="ts">
import type { ScoredStat } from '../types'
import { compactNumber, percent, score } from '../utils/format'

defineProps<{ items: ScoredStat[] }>()
</script>

<template>
  <div class="table-wrap">
    <table class="table">
      <thead>
        <tr>
          <th>排名</th>
          <th>玩家</th>
          <th>帮会</th>
          <th>职业</th>
          <th>所在团</th>
          <th>击败</th>
          <th>助攻</th>
          <th>KDA</th>
          <th>参团率</th>
          <th>玩家伤害</th>
          <th>建筑伤害</th>
          <th>综合分</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in items" :key="`${item.guild_name}-${item.player_name}-${item.row_number}`">
          <td><span class="rank">{{ item.guild_rank }}</span></td>
          <td>
            <router-link class="player-link" :to="`/players/latest/${item.row_number}`">{{ item.player_name }}</router-link>
          </td>
          <td>{{ item.guild_name }}</td>
          <td><span class="badge purple">{{ item.career }}</span></td>
          <td>{{ item.team_leader }}</td>
          <td>{{ item.kills }}</td>
          <td>{{ item.assists }}</td>
          <td>{{ score(item.kda_ratio) }}</td>
          <td>{{ percent(item.participation_rate) }}</td>
          <td>{{ compactNumber(item.player_damage) }}</td>
          <td>{{ compactNumber(item.building_damage) }}</td>
          <td><span class="score">{{ score(item.composite_score) }}</span></td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
