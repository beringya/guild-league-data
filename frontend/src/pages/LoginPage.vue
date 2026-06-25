<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const username = ref('admin')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(username.value, password.value)
    router.push('/')
  } catch {
    error.value = '账号或密码错误，请检查首次启动日志中的随机密码。'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-body">
    <section class="login-shell">
      <div class="login-art">
        <div class="logo-line">
          <img src="/assets/brand/logo.svg" alt="" />
          <span>逆水寒联赛分析</span>
        </div>
        <h1>帮会联赛数据后台</h1>
        <p>导入 CSV 后查看职业评分、六维雷达、团内 TOP3、对手帮会对比和多场历史总榜。</p>
        <img class="mascot" src="/assets/brand/mascot_bunny.svg" alt="" />
      </div>
      <form class="login-card" @submit.prevent="submit">
        <h2>管理员登录</h2>
        <p>首次部署账号为 admin，随机密码只在首次容器日志中显示一次。</p>
        <label class="field">
          <span>账号</span>
          <input v-model="username" class="input" autocomplete="username" />
        </label>
        <label class="field">
          <span>密码</span>
          <input v-model="password" class="input" type="password" autocomplete="current-password" />
        </label>
        <button class="login-button" :disabled="loading">{{ loading ? '登录中...' : '进入后台' }}</button>
        <div v-if="error" class="login-tip error">{{ error }}</div>
        <div class="login-tip">登录后如果账号处于首次初始化状态，系统会要求立即修改密码。</div>
      </form>
    </section>
  </div>
</template>
