<script setup>
import { computed } from 'vue'

const props = defineProps({
  status: { type: String, default: '' }
})

const config = computed(() => {
  const status = props.status.trim().toLowerCase()
  if (status.includes('error') || status.includes('failed')) return { tone: 'danger', label: '错误' }
  if (status.includes('restarting')) return { tone: 'warning', label: '重启中' }
  if (status.includes('starting')) return { tone: 'warning', label: '启动中' }
  if (status.includes('approved')) return { tone: 'success', label: '已批准' }
  if (status.includes('rejected')) return { tone: 'danger', label: '已拒绝' }
  if (status.includes('ignored')) return { tone: 'default', label: '已忽略' }
  if (status.includes('pending')) return { tone: 'warning', label: '待处理' }
  if (status.includes('paused')) return { tone: 'warning', label: '已暂停' }
  if (status.includes('created')) return { tone: 'info', label: '已创建' }
  if (status.includes('检测中') || status.includes('testing') || status.includes('checking')) return { tone: 'warning', label: '检测中' }
  if (status === 'online') return { tone: 'success', label: '在线' }
  if (status === 'offline') return { tone: 'default', label: '离线' }
  if (status.includes('disconnected')) return { tone: 'default', label: '离线' }
  if (status.includes('connected')) return { tone: 'success', label: '已连接' }
  if (['running', 'up', 'active'].some(value => status.includes(value))) return { tone: 'success', label: '运行中' }
  if (['stopped', 'exited', 'dead'].some(value => status.includes(value))) return { tone: 'danger', label: '已停止' }
  return { tone: 'default', label: props.status || '未知' }
})
</script>

<template>
  <span class="status-badge" :class="`status-${config.tone}`" role="status" :title="status || config.label">
    <span class="status-dot" aria-hidden="true" />
    {{ config.label }}
  </span>
</template>

<style scoped>
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--offline);
  font-size: 12px;
  font-weight: 600;
  line-height: 1.4;
  white-space: nowrap;
}

.status-dot {
  width: 7px;
  height: 7px;
  flex: 0 0 auto;
  border-radius: 50%;
  background: currentColor;
}

.status-success { color: var(--success); }
.status-warning { color: var(--warning); }
.status-danger { color: var(--danger); }
.status-info { color: #0066cc; }
.status-default { color: var(--offline); }
</style>
