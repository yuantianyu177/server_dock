<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref, useId } from 'vue'
import { X } from '@lucide/vue'

const props = defineProps({
  title: { type: String, required: true },
  size: { type: String, default: 'md' },
  closeOnBackdrop: { type: Boolean, default: true }
})

const emit = defineEmits(['close'])
const modalRef = ref(null)
const titleId = `modal-title-${useId()}`
let previousFocus = null
let previousOverflow = ''

const focusableSelector = [
  'a[href]',
  'button:not([disabled])',
  'textarea:not([disabled])',
  'input:not([disabled])',
  'select:not([disabled])',
  '[tabindex]:not([tabindex="-1"])'
].join(',')

function close() {
  emit('close')
}

function handleBackdrop() {
  if (props.closeOnBackdrop) close()
}

function handleKeydown(event) {
  if (event.key === 'Escape') {
    event.preventDefault()
    close()
    return
  }
  if (event.key !== 'Tab' || !modalRef.value) return

  const focusable = [...modalRef.value.querySelectorAll(focusableSelector)]
  if (focusable.length === 0) {
    event.preventDefault()
    modalRef.value.focus()
    return
  }
  const first = focusable[0]
  const last = focusable[focusable.length - 1]
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault()
    last.focus()
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault()
    first.focus()
  }
}

onMounted(async () => {
  previousFocus = document.activeElement
  previousOverflow = document.body.style.overflow
  document.body.style.overflow = 'hidden'
  document.addEventListener('keydown', handleKeydown)
  await nextTick()
  const firstFocusable = modalRef.value?.querySelector(focusableSelector)
  ;(firstFocusable || modalRef.value)?.focus()
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleKeydown)
  document.body.style.overflow = previousOverflow
  previousFocus?.focus?.()
})
</script>

<template>
  <Teleport to="body">
    <Transition name="modal-fade" appear>
      <div class="modal-backdrop" @mousedown.self="handleBackdrop">
        <section
          ref="modalRef"
          class="modal-box"
          :class="`modal-${size}`"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="titleId"
          tabindex="-1"
        >
          <header class="modal-header">
            <h2 :id="titleId" class="modal-title">{{ title }}</h2>
            <button class="modal-close" type="button" aria-label="关闭对话框" @click="close">
              <X :size="18" aria-hidden="true" />
            </button>
          </header>
          <div class="modal-body">
            <slot />
          </div>
          <footer v-if="$slots.footer" class="modal-footer">
            <slot name="footer" />
          </footer>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: rgba(0, 0, 0, 0.38);
}

.modal-box {
  width: min(100%, 520px);
  max-height: calc(100vh - 48px);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: var(--radius-modal);
  outline: none;
  background: var(--surface);
  box-shadow: var(--shadow-popover);
}

.modal-sm {
  max-width: 420px;
}

.modal-md {
  max-width: 560px;
}

.modal-lg {
  max-width: 780px;
}

.modal-header {
  min-height: 58px;
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 18px 12px 22px;
  border-bottom: 1px solid var(--divider-subtle);
}

.modal-title {
  color: var(--ink);
  font-size: 17px;
  font-weight: 700;
  letter-spacing: -0.01em;
}

.modal-close {
  width: 32px;
  height: 32px;
  display: grid;
  flex: 0 0 auto;
  place-items: center;
  border-radius: var(--radius-control);
  background: transparent;
  color: var(--ink-secondary);
  cursor: pointer;
  transition: background-color 160ms ease, color 160ms ease;
}

.modal-close:hover {
  background: #eeeeF1;
  color: var(--ink);
}

.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 20px 22px;
}

.modal-footer {
  min-height: 60px;
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  padding: 11px 22px;
  border-top: 1px solid var(--divider-subtle);
  background: #fafafa;
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 180ms ease;
}

.modal-fade-enter-active .modal-box,
.modal-fade-leave-active .modal-box {
  transition: opacity 180ms ease, transform 220ms cubic-bezier(0.2, 0, 0, 1);
}

.modal-fade-enter-from,
.modal-fade-leave-to,
.modal-fade-enter-from .modal-box,
.modal-fade-leave-to .modal-box {
  opacity: 0;
}

.modal-fade-enter-from .modal-box,
.modal-fade-leave-to .modal-box {
  transform: translateY(8px) scale(0.985);
}

@media (max-width: 640px) {
  .modal-backdrop {
    align-items: flex-end;
    padding: 0;
  }

  .modal-box,
  .modal-sm,
  .modal-md,
  .modal-lg {
    position: relative;
    width: 100%;
    max-width: none;
    max-height: min(92dvh, 820px);
    border-radius: 24px 24px 0 0;
  }

  .modal-box::before {
    content: "";
    position: absolute;
    top: 8px;
    left: 50%;
    z-index: 1;
    width: 38px;
    height: 4px;
    border-radius: 4px;
    background: #d7e0e4;
    transform: translateX(-50%);
  }

  .modal-header {
    min-height: 64px;
    padding-top: 20px;
    padding-right: 14px;
    padding-left: 18px;
  }

  .modal-body {
    padding: 18px;
  }

  .modal-footer {
    padding: 12px 18px max(12px, env(safe-area-inset-bottom));
    flex-wrap: wrap;
  }

  .modal-footer :deep(.btn) {
    min-width: 112px;
    flex: 1 1 auto;
  }
}
</style>
