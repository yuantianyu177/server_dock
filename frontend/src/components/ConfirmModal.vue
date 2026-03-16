<script setup>
import BaseModal from './BaseModal.vue'

defineProps({
  title: { type: String, default: 'Confirm' },
  message: String,
  confirmText: { type: String, default: 'Confirm' },
  confirmClass: { type: String, default: 'btn-danger' },
  loading: Boolean
})
const emit = defineEmits(['confirm', 'cancel'])
</script>

<template>
  <BaseModal :title="title" size="sm" @close="emit('cancel')">
    <p class="confirm-message">{{ message }}</p>
    <template #footer>
      <button class="btn btn-secondary" @click="emit('cancel')" :disabled="loading">Cancel</button>
      <button :class="['btn', confirmClass]" @click="emit('confirm')" :disabled="loading">
        <span v-if="loading" class="spinner" />
        {{ confirmText }}
      </button>
    </template>
  </BaseModal>
</template>

<style scoped>
.confirm-message {
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.6;
}
</style>
