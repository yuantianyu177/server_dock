<script setup>
import { computed } from 'vue'

const props = defineProps({
  status: String
})

function prettifyStatusLabel(status) {
  return String(status || '')
    .trim()
    .replace(/[_-]+/g, ' ')
    .replace(/\s+/g, ' ')
    .replace(/\b\w/g, (char) => char.toUpperCase())
}

const config = computed(() => {
  const s = (props.status || '').toLowerCase()
  if (s.includes('error') || s.includes('failed')) {
    return { class: 'badge-danger', label: 'Error' }
  }
  if (['running', 'up'].some(k => s.includes(k))) {
    return { class: 'badge-success', label: 'Running' }
  }
  if (s.includes('approved')) {
    return { class: 'badge-success', label: 'Approved' }
  }
  if (s.includes('active')) {
    return { class: 'badge-success', label: 'Active' }
  }
  if (s.includes('restarting')) {
    return { class: 'badge-warning', label: 'Restarting' }
  }
  if (s.includes('starting')) {
    return { class: 'badge-warning', label: 'Starting' }
  }
  if (s.includes('paused')) {
    return { class: 'badge-warning', label: 'Paused' }
  }
  if (s.includes('pending')) {
    return { class: 'badge-warning', label: 'Pending' }
  }
  if (s.includes('created')) {
    return { class: 'badge-info', label: 'Created' }
  }
  if (['stopped', 'exited', 'dead'].some(k => s.includes(k))) {
    return { class: 'badge-danger', label: 'Stopped' }
  }
  if (s.includes('rejected')) {
    return { class: 'badge-danger', label: 'Rejected' }
  }
  return { class: 'badge-default', label: prettifyStatusLabel(props.status) || 'Unknown' }
})
</script>

<template>
  <span class="badge" :class="config.class" :title="status || config.label">
    <span class="badge-dot" />
    {{ config.label }}
  </span>
</template>
