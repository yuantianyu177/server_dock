<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'

const props = defineProps({
  servers: { type: Array, default: () => [] },
  modelValue: { default: null }
})

const emit = defineEmits(['update:modelValue'])

const open = ref(false)
const dropdownRef = ref(null)

const selected = computed(() => props.servers.find(s => s.id === props.modelValue))

function select(server) {
  emit('update:modelValue', server.id)
  open.value = false
}

function toggle() {
  open.value = !open.value
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
  <div class="server-dropdown" ref="dropdownRef">
    <button class="dropdown-trigger" @click="toggle" :class="{ open }">
      <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" class="trigger-icon">
        <rect x="2" y="2" width="20" height="8" rx="2"/><rect x="2" y="14" width="20" height="8" rx="2"/>
        <line x1="6" y1="6" x2="6.01" y2="6"/><line x1="6" y1="18" x2="6.01" y2="18"/>
      </svg>
      <span class="trigger-text">{{ selected?.host || 'Select server' }}</span>
      <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="trigger-chevron" :class="{ rotated: open }">
        <path d="M6 9l6 6 6-6"/>
      </svg>
    </button>
    <Transition name="dropdown">
      <div v-if="open" class="dropdown-menu">
        <button
          v-for="s in servers"
          :key="s.id"
          class="dropdown-item"
          :class="{ active: s.id === modelValue }"
          @click="select(s)"
        >
          <span class="item-dot" :class="{ active: s.id === modelValue }" />
          <span class="item-text">{{ s.host }}</span>
          <span v-if="s.hostname" class="item-sub">{{ s.hostname }}</span>
        </button>
        <div v-if="servers.length === 0" class="dropdown-empty">No servers</div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.server-dropdown {
  position: relative;
  z-index: 50;
}

.dropdown-trigger {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 6px 10px 6px 9px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: white;
  color: var(--text-primary);
  font-size: 13px;
  font-weight: 450;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
  min-width: 150px;
}

.dropdown-trigger:hover {
  border-color: var(--accent);
  background: var(--cream, #FDFBF7);
}

.dropdown-trigger.open {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(201, 100, 66, 0.10);
}

.trigger-icon {
  color: var(--text-muted);
  flex-shrink: 0;
}

.trigger-text {
  flex: 1;
  text-align: left;
  overflow: hidden;
  text-overflow: ellipsis;
}

.trigger-chevron {
  color: var(--text-muted);
  flex-shrink: 0;
  transition: transform 0.2s;
}

.trigger-chevron.rotated {
  transform: rotate(180deg);
}

.dropdown-menu {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  min-width: 100%;
  max-width: 280px;
  background: white;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  box-shadow: 0 4px 16px rgba(0,0,0,0.08), 0 1px 3px rgba(0,0,0,0.04);
  padding: 4px;
  z-index: 100;
  max-height: 240px;
  overflow-y: auto;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 7px 10px;
  border: none;
  border-radius: 4px;
  background: none;
  color: var(--text-primary);
  font-size: 13px;
  cursor: pointer;
  transition: background 0.1s;
  text-align: left;
}

.dropdown-item:hover {
  background: var(--cream, #FDFBF7);
}

.dropdown-item.active {
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

.item-text {
  flex: 1;
  font-weight: 450;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-sub {
  font-size: 11.5px;
  color: var(--text-muted);
  font-family: var(--font-mono);
  flex-shrink: 0;
}

.dropdown-empty {
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

@media (max-width: 768px) {
  .dropdown-trigger {
    min-width: unset;
    width: 100%;
  }

  .dropdown-menu {
    left: 0;
    right: 0;
    max-width: none;
  }
}
</style>
