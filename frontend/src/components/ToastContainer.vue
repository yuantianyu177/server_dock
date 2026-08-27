<script setup>
import { useToast } from '@/composables/useToast'

const { toasts, remove } = useToast()

function iconFor(type) {
  if (type === 'success') return '<svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>'
  return '<svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>'
}
</script>

<template>
  <Teleport to="body">
    <div class="toast-stack" aria-live="polite">
      <TransitionGroup name="toast">
        <div
          v-for="toast in toasts"
          :key="toast.id"
          class="toast-item"
          :class="`toast-${toast.type}`"
          @click="remove(toast.id)"
        >
          <span class="toast-icon" v-html="iconFor(toast.type)" />
          <span class="toast-msg">{{ toast.message }}</span>
          <button class="toast-close" @click.stop="remove(toast.id)">
            <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-stack {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 8px;
  pointer-events: none;
  max-width: calc(100vw - 40px);
}

.toast-item {
  pointer-events: auto;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border-radius: 10px;
  font-size: 13.5px;
  font-weight: 450;
  line-height: 1.4;
  cursor: pointer;
  box-shadow: 0 4px 20px rgba(0,0,0,0.12), 0 1px 4px rgba(0,0,0,0.06);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  max-width: 420px;
  min-width: 200px;
  word-break: break-word;
}

.toast-error {
  background: rgba(255, 252, 250, 0.97);
  color: var(--danger);
  border: 1px solid var(--danger-border);
}

.toast-success {
  background: rgba(252, 255, 252, 0.97);
  color: var(--success);
  border: 1px solid var(--success-border);
}

.toast-icon {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.toast-msg {
  flex: 1;
}

.toast-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 4px;
  border: none;
  background: none;
  cursor: pointer;
  opacity: 0.5;
  flex-shrink: 0;
  color: inherit;
  transition: opacity 0.15s;
}

.toast-close:hover {
  opacity: 1;
}

/* Transition */
.toast-enter-active {
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}
.toast-leave-active {
  transition: all 0.18s ease;
}
.toast-enter-from {
  opacity: 0;
  transform: translateX(30px) scale(0.96);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(20px) scale(0.96);
}
.toast-move {
  transition: transform 0.25s ease;
}

@media (max-width: 640px) {
  .toast-stack {
    top: auto;
    bottom: 20px;
    right: 12px;
    left: 12px;
    max-width: none;
    align-items: stretch;
  }

  .toast-item {
    max-width: none;
    width: 100%;
  }

  .toast-enter-from {
    transform: translateY(20px) scale(0.96);
  }
  .toast-leave-to {
    transform: translateY(10px) scale(0.96);
  }
}
</style>
