<script setup lang="ts">
defineProps<{ title?: string }>()
</script>

<template>
  <section class="card">
    <div v-if="title" class="card-title">
      <span>{{ title }}</span>
      <slot name="actions" />
    </div>
    <slot />
  </section>
</template>
