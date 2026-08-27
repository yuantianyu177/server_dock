<script setup>
defineProps({
  title: { type: String, required: true },
  description: { type: String, default: '' }
})
</script>

<template>
  <header class="page-header">
    <div class="page-heading">
      <h1 class="page-title">{{ title }}</h1>
      <p v-if="description" class="page-subtitle">{{ description }}</p>
    </div>
    <div v-if="$slots.actions" class="page-actions">
      <slot name="actions" />
    </div>
  </header>
</template>
