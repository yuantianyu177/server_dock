<script setup>
import { CircleAlert, CircleCheck, Info, TriangleAlert, X } from '@lucide/vue'
import { useToast } from '@/composables/useToast'

const { toasts, remove } = useToast()
const icons = { success: CircleCheck, error: CircleAlert, warning: TriangleAlert, info: Info }
</script>

<template>
  <Teleport to="body">
    <div class="toast-stack" aria-live="polite" aria-atomic="false">
      <TransitionGroup name="toast">
        <div
          v-for="toast in toasts"
          :key="toast.id"
          class="toast-item"
          :class="`toast-${toast.type}`"
          :role="toast.type === 'error' ? 'alert' : 'status'"
        >
          <component :is="icons[toast.type] || Info" :size="18" class="toast-icon" aria-hidden="true" />
          <span class="toast-message">{{ toast.message }}</span>
          <button class="toast-close" type="button" aria-label="关闭通知" @click="remove(toast.id)">
            <X :size="15" aria-hidden="true" />
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-stack {
  position: fixed;
  top: 18px;
  right: 18px;
  z-index: 2000;
  width: min(390px, calc(100vw - 36px));
  display: flex;
  flex-direction: column;
  gap: 8px;
  pointer-events: none;
}

.toast-item {
  min-width: 0;
  min-height: 48px;
  display: grid;
  grid-template-columns: 20px minmax(0, 1fr) 28px;
  align-items: start;
  gap: 9px;
  padding: 9px 9px 9px 13px;
  border: 1px solid var(--divider);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: var(--shadow-popover);
  color: var(--ink);
  pointer-events: auto;
}

.toast-icon {
  margin-top: 1px;
  color: var(--blue);
}

.toast-success .toast-icon { color: var(--success); }
.toast-error .toast-icon { color: var(--danger); }
.toast-warning .toast-icon { color: var(--warning); }

.toast-message {
  min-width: 0;
  overflow-wrap: anywhere;
  word-break: break-word;
  font-size: 13px;
  line-height: 1.45;
}

.toast-close {
  width: 28px;
  height: 28px;
  display: grid;
  align-self: start;
  place-items: center;
  border-radius: 7px;
  background: transparent;
  color: var(--ink-secondary);
  cursor: pointer;
}

.toast-close:hover {
  background: #eeeeF1;
  color: var(--ink);
}

.toast-enter-active,
.toast-leave-active,
.toast-move {
  transition: opacity 180ms ease, transform 220ms cubic-bezier(0.2, 0, 0, 1);
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

@media (max-width: 640px) {
  .toast-stack {
    top: auto;
    right: 12px;
    bottom: max(12px, env(safe-area-inset-bottom));
    left: 12px;
    width: auto;
  }
}
</style>
