<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '../components/AppLayout.vue'
import CardPanel from '../components/CardPanel.vue'
import { api } from '../api/client'
import type { ImportPreview } from '../types'

const router = useRouter()
const preview = ref<ImportPreview | null>(null)
const selectedGuild = ref('')
const loading = ref(false)
const message = ref('')

async function onFile(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return
  loading.value = true
  message.value = ''
  try {
    preview.value = await api.uploadPreview(file)
    selectedGuild.value = preview.value.guilds[0]?.name || ''
  } catch {
    message.value = '解析失败，请检查 CSV 文件。'
  } finally {
    loading.value = false
  }
}

async function confirm() {
  if (!preview.value) return
  loading.value = true
  try {
    const res = await api.post<{ battle_id: number }>('/api/battles/import/confirm', {
      token: preview.value.token,
      home_guild: selectedGuild.value,
      battle_at: preview.value.inferred_battle_at
    })
    router.push(`/history?battle=${res.battle_id}`)
  } catch {
    message.value = '确认入库失败，请检查校验错误。'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AppLayout>
    <div class="steps">
      <span class="step active"><b>1</b>上传 CSV</span><i class="line" /><span class="step" :class="{ active: preview }"><b>2</b>预览校验</span><i class="line" /><span class="step"><b>3</b>确认入库</span>
    </div>
    <CardPanel title="数据导入">
      <div class="dropzone">
        <img src="/assets/icons/svg/import.svg" alt="" />
        <h3>{{ loading ? '处理中...' : '选择或拖入帮会联赛 CSV' }}</h3>
        <p>支持 UTF-8 BOM、UTF-8、GB18030、重复表头清理和数字逗号清理。</p>
        <input type="file" accept=".csv,text/csv" @change="onFile" />
      </div>
      <p v-if="message" class="error-text">{{ message }}</p>
    </CardPanel>
    <CardPanel v-if="preview" title="预览结果">
      <template #actions>
        <button class="btn primary" :disabled="Boolean(preview.errors.length) || loading" @click="confirm">确认入库</button>
      </template>
      <div class="summary-strip compact">
        <div v-for="guild in preview.guilds" :key="guild.name" class="kpi">
          <label>{{ guild.name }}</label><strong>{{ guild.member_count }} 人</strong>
        </div>
        <label class="kpi"><span>选择本帮会</span><select v-model="selectedGuild" class="select"><option v-for="guild in preview.guilds" :key="guild.name" :value="guild.name">{{ guild.name }}</option></select></label>
      </div>
      <div v-if="preview.errors.length" class="message-list error-text">
        <div v-for="err in preview.errors" :key="`${err.code}-${err.row_number}`">{{ err.row_number ? `第 ${err.row_number} 行：` : '' }}{{ err.message }}</div>
      </div>
      <div v-if="preview.warnings.length" class="message-list">
        <div v-for="warn in preview.warnings" :key="`${warn.code}-${warn.row_number}`">{{ warn.row_number ? `第 ${warn.row_number} 行：` : '' }}{{ warn.message }}</div>
      </div>
      <div class="table-wrap">
        <table class="table">
          <thead><tr><th>帮会</th><th>玩家</th><th>职业</th><th>所在团长</th><th>击败</th><th>助攻</th><th>玩家伤害</th><th>建筑伤害</th></tr></thead>
          <tbody><tr v-for="row in preview.preview_rows" :key="`${row.guild_name}-${row.player_name}`"><td>{{ row.guild_name }}</td><td>{{ row.player_name }}</td><td>{{ row.career }}</td><td>{{ row.team_leader }}</td><td>{{ row.kills }}</td><td>{{ row.assists }}</td><td>{{ row.player_damage }}</td><td>{{ row.building_damage }}</td></tr></tbody>
        </table>
      </div>
    </CardPanel>
  </AppLayout>
</template>
