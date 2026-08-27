<script setup>
import { TriangleAlert } from '@lucide/vue'
import BaseModal from './BaseModal.vue'

defineProps({
  title: { type: String, default: '确认操作' },
  message: { type: String, required: true },
  detail: { type: String, default: '' },
  confirmText: { type: String, default: '继续操作' },
  confirmClass: { type: String, default: 'btn-danger' },
  loading: Boolean
})

const emit = defineEmits(['confirm', 'cancel'])
</script>

<template>
  <BaseModal :title="title" size="sm" :close-on-backdrop="!loading" @close="emit('cancel')">
    <div class="confirm-content">
      <div class="confirm-icon" aria-hidden="true">
        <TriangleAlert :size="20" />
      </div>
      <div>
        <p class="confirm-message">{{ message }}</p>
        <p v-if="detail" class="confirm-detail">{{ detail }}</p>
      </div>
    </div>
    <template #footer>
      <button class="btn btn-secondary" type="button" :disabled="loading" @click="emit('cancel')">取消</button>
      <button :class="['btn', confirmClass]" type="button" :disabled="loading" @click="emit('confirm')">
        <span v-if="loading" class="spinner" aria-hidden="true" />
        {{ loading ? '正在处理…' : confirmText }}
      </button>
    </template>
  </BaseModal>
</template>

<style scoped>
.confirm-content {
  display: grid;
  grid-template-columns: 36px 1fr;
  gap: 12px;
  align-items: start;
}

.confirm-icon {
  width: 36px;
  height: 36px;
  display: grid;
  place-items: center;
  border-radius: 10px;
  background: var(--danger-soft);
  color: var(--danger);
}

.confirm-message {
  color: var(--ink);
  font-size: 14px;
  font-weight: 600;
  line-height: 1.55;
}

.confirm-detail {
  margin-top: 5px;
  color: var(--ink-secondary);
  font-size: 12px;
  line-height: 1.55;
}

.spinner {
  width: 13px;
  height: 13px;
  color: currentColor;
}
</style>
