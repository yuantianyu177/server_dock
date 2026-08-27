<script setup>
defineProps({
  title: { type: String, required: true },
  description: { type: String, default: '' },
  tone: { type: String, default: 'empty' }
})
</script>

<template>
  <div :class="tone === 'error' ? 'error-state' : 'empty-state'" :role="tone === 'error' ? 'alert' : 'status'">
    <div class="state-icon" aria-hidden="true">
      <slot name="icon" />
    </div>
    <div>
      <div class="state-title">{{ title }}</div>
      <p v-if="description" class="state-description">{{ description }}</p>
    </div>
    <div v-if="$slots.actions">
      <slot name="actions" />
    </div>
  </div>
</template>
