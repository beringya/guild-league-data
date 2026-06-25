<script setup lang="ts">
import { onMounted, ref } from 'vue'
import AppLayout from '../components/AppLayout.vue'
import CardPanel from '../components/CardPanel.vue'
import { api } from '../api/client'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const settings = ref<Record<string, any>>({})
const rules = ref<any[]>([])
const ranges = ref<any[]>([])
const passwordForm = ref({ current: '', next: '' })
const message = ref('')

onMounted(load)

async function load() {
  settings.value = await api.get<Record<string, any>>('/api/settings')
  rules.value = (await api.get<{ items: any[] }>('/api/scoring-rules')).items
  ranges.value = (await api.get<{ items: any[] }>('/api/scoring-ranges')).items
}

async function saveSettings() {
  await api.put('/api/settings', settings.value)
  message.value = '设置已保存。'
}

async function changePassword() {
  await auth.changePassword(passwordForm.value.current, passwordForm.value.next)
  passwordForm.value = { current: '', next: '' }
  message.value = '密码已修改，其他会话已撤销。'
}

async function backup() {
  const res = await api.post<{ path: string }>('/api/backups')
  message.value = `备份已创建：${res.path}`
}
</script>

<template>
  <AppLayout>
    <div class="settings-grid">
      <CardPanel title="通用设置">
        <div class="setting-row"><div><h4>默认本帮会</h4><p>导入确认时会记住最近选择。</p></div><input v-model="settings.default_home_guild" class="input" /></div>
        <div class="setting-row"><div><h4>多场总榜最低场次</h4><p>默认值为 3，页面可临时覆盖。</p></div><input v-model.number="settings.aggregate_min_matches" class="input" type="number" /></div>
        <div class="setting-row"><div><h4>优势判断阈值</h4><p>0.05 表示差异超过 5% 记为优势或不足。</p></div><input v-model.number="settings.comparison_threshold" class="input" type="number" step="0.01" /></div>
        <button class="btn primary" @click="saveSettings">保存设置</button>
      </CardPanel>
      <CardPanel title="管理员安全">
        <input v-model="passwordForm.current" class="input full" type="password" placeholder="当前密码" />
        <input v-model="passwordForm.next" class="input full" type="password" placeholder="新密码，至少 10 位" />
        <button class="btn primary" @click="changePassword">修改密码</button>
      </CardPanel>
      <CardPanel title="评分规则版本">
        <div class="table-wrap">
          <table class="table"><thead><tr><th>版本</th><th>名称</th><th>状态</th><th>激活</th></tr></thead><tbody><tr v-for="rule in rules" :key="rule.version"><td>{{ rule.version }}</td><td>{{ rule.name }}</td><td>{{ rule.status }}</td><td>{{ rule.is_active ? '是' : '否' }}</td></tr></tbody></table>
        </div>
      </CardPanel>
      <CardPanel title="范围版本与备份">
        <div class="table-wrap">
          <table class="table"><thead><tr><th>版本</th><th>来源</th><th>冻结</th><th>激活</th></tr></thead><tbody><tr v-for="range in ranges" :key="range.version"><td>{{ range.version }}</td><td>{{ range.source_method }}</td><td>{{ range.is_frozen ? '是' : '否' }}</td><td>{{ range.is_active ? '是' : '否' }}</td></tr></tbody></table>
        </div>
        <button class="btn" @click="backup">立即备份</button>
      </CardPanel>
    </div>
    <p v-if="message" class="toast">{{ message }}</p>
  </AppLayout>
</template>
