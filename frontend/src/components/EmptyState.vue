<script setup lang="ts">
defineProps<{ title?: string; message?: string }>()
</script>

<template>
  <div class="empty-state">
    <img src="/assets/icons/svg/import.svg" alt="" />
    <h3>{{ title || '还没有可分析的数据' }}</h3>
    <p>{{ message || '导入一份帮会联赛 CSV 后，这里会显示排名、对比和六维分析。' }}</p>
    <router-link class="btn primary" to="/import">导入 CSV</router-link>
  </div>
</template>
