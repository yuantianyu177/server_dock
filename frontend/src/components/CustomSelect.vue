<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'

const props = defineProps({
  options: { type: Array, default: () => [] },
  modelValue: { default: '' },
  placeholder: { type: String, default: 'Select…' },
  disabled: { type: Boolean, default: false },
  labelKey: { type: String, default: 'label' },
  valueKey: { type: String, default: 'value' },
  subtitleKey: { type: String, default: '' }
})

const emit = defineEmits(['update:modelValue'])

const open = ref(false)
const dropdownRef = ref(null)

const selected = computed(() => props.options.find(o => {
  const val = typeof o === 'object' ? o[props.valueKey] : o
  return String(val) === String(props.modelValue)
}))

const displayText = computed(() => {
  if (!selected.value) return ''
  return typeof selected.value === 'object' ? selected.value[props.labelKey] : selected.value
})

function select(option) {
  const val = typeof option === 'object' ? option[props.valueKey] : option
  emit('update:modelValue', val)
  open.value = false
}

function toggle() {
  if (!props.disabled) open.value = !open.value
}

function onClickOutside(e) {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target)) {
    open.value = false
  }
}

onMounted(() => document.addEventListener('click', onClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', onClickOutside))
</script>

<template>
  <div class="custom-select" :class="{ disabled }" ref="dropdownRef">
    <button type="button" class="select-trigger" @click="toggle" :class="{ open, placeholder: !selected }" :disabled="disabled">
      <span class="trigger-text">{{ displayText || placeholder }}</span>
      <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="trigger-chevron" :class="{ rotated: open }">
        <path d="M6 9l6 6 6-6"/>
      </svg>
    </button>
    <Transition name="dropdown">
      <div v-if="open" class="select-menu">
        <button
          v-for="(opt, idx) in options"
          :key="idx"
          type="button"
          class="select-item"
          :class="{ active: String(typeof opt === 'object' ? opt[valueKey] : opt) === String(modelValue) }"
          @click="select(opt)"
        >
          <span class="item-dot" :class="{ active: String(typeof opt === 'object' ? opt[valueKey] : opt) === String(modelValue) }" />
          <span class="item-content">
            <span class="item-label">{{ typeof opt === 'object' ? opt[labelKey] : opt }}</span>
            <span v-if="subtitleKey && typeof opt === 'object' && opt[subtitleKey]" class="item-sub">{{ opt[subtitleKey] }}</span>
          </span>
        </button>
        <div v-if="options.length === 0" class="select-empty">No options available</div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.custom-select {
  position: relative;
}

.select-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 9px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: white;
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 400;
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s;
  text-align: left;
}

.select-trigger.placeholder {
  color: var(--text-muted);
}

.select-trigger:hover:not(:disabled) {
  border-color: var(--accent);
}

.select-trigger.open {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(201, 100, 66, 0.12);
}

.select-trigger:disabled {
  background: #F7F3EC;
  color: var(--text-muted);
  cursor: not-allowed;
  border-color: var(--border);
}

.trigger-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.trigger-chevron {
  color: var(--text-muted);
  flex-shrink: 0;
  transition: transform 0.2s;
}

.trigger-chevron.rotated {
  transform: rotate(180deg);
}

.select-menu {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  background: white;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  box-shadow: 0 4px 16px rgba(0,0,0,0.08), 0 1px 3px rgba(0,0,0,0.04);
  padding: 4px;
  z-index: 100;
  max-height: 220px;
  overflow-y: auto;
}

.select-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 8px 10px;
  border: none;
  border-radius: 4px;
  background: none;
  color: var(--text-primary);
  font-size: 13.5px;
  cursor: pointer;
  transition: background 0.1s;
  text-align: left;
}

.select-item:hover {
  background: var(--cream, #FDFBF7);
}

.select-item.active {
  background: rgba(201, 100, 66, 0.08);
}

.item-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--border);
  flex-shrink: 0;
  transition: background 0.15s;
}

.item-dot.active {
  background: var(--accent);
}

.item-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
}

.item-label {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-sub {
  font-size: 11.5px;
  color: var(--text-muted);
}

.select-empty {
  padding: 12px;
  text-align: center;
  color: var(--text-muted);
  font-size: 12.5px;
}

/* Transition */
.dropdown-enter-active {
  transition: opacity 0.12s, transform 0.12s;
}
.dropdown-leave-active {
  transition: opacity 0.08s, transform 0.08s;
}
.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
