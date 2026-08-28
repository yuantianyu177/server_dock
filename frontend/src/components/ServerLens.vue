<script setup>
import { computed } from 'vue'
import { RefreshCw, Server, Box, Network } from '@lucide/vue'

const props = defineProps({
  servers: { type: Array, default: () => [] },
  modelValue: { type: [Number, String], default: null },
  loading: Boolean,
  state: { type: String, default: 'offline' },
  summary: { type: Object, default: () => ({ running: 0, total: 0 }) },
  error: { type: String, default: '' }
})

const emit = defineEmits(['update:modelValue', 'retry'])

const selected = computed(() => props.servers.find(server => Number(server.id) === Number(props.modelValue)))

const stateText = computed(() => ({
  online: '在线',
  offline: '离线'
}[props.state] || '离线'))

function selectServer(event) {
  emit('update:modelValue', Number(event.target.value))
}
</script>

<template>
  <section class="server-lens" aria-labelledby="server-lens-title">
    <div class="lens-identity">
      <div class="lens-orbit" :class="`is-${state}`" aria-hidden="true">
        <span class="lens-orbit-ring" />
        <span class="lens-orbit-dot" />
      </div>
      <div class="lens-selector">
        <label id="server-lens-title" class="lens-kicker" for="server-lens-select">Server Lens · 当前服务器</label>
        <select
          id="server-lens-select"
          class="lens-select"
          :value="modelValue ?? ''"
          :disabled="loading || servers.length === 0"
          @change="selectServer"
        >
          <option v-if="servers.length === 0" value="">尚未添加服务器</option>
          <option v-for="serverItem in servers" :key="serverItem.id" :value="serverItem.id">
            {{ serverItem.host }}
          </option>
        </select>
      </div>
    </div>

    <div v-if="selected" class="lens-readings" aria-live="polite">
      <div class="lens-reading">
        <Network :size="15" aria-hidden="true" />
        <span class="lens-reading-label">SSH 通道</span>
        <span class="lens-reading-value" :class="`is-${state}`">
          <span class="lens-status-dot" />{{ stateText }}
        </span>
      </div>
      <div class="lens-reading">
        <Server :size="15" aria-hidden="true" />
        <span class="lens-reading-label">Docker</span>
        <span class="lens-reading-value" :class="`is-${state}`">
          <span class="lens-status-dot" />{{ stateText }}
        </span>
      </div>
      <div class="lens-reading">
        <Box :size="15" aria-hidden="true" />
        <span class="lens-reading-label">容器</span>
        <span class="lens-reading-value mono">
          <template v-if="state === 'online'">{{ summary.running }}/{{ summary.total }} 运行</template>
          <template v-else>—</template>
        </span>
      </div>
      <button
        v-if="state === 'offline'"
        class="lens-retry"
        type="button"
        :title="error || '重新检测服务器'"
        @click="emit('retry')"
      >
        <RefreshCw :size="14" aria-hidden="true" />
        重新检测
      </button>
    </div>

    <div v-if="$slots.actions" class="lens-actions">
      <slot name="actions" />
    </div>
  </section>
</template>

<style scoped>
.server-lens {
  position: relative;
  min-height: 82px;
  display: grid;
  grid-template-columns: minmax(230px, 1fr) auto auto;
  align-items: center;
  gap: 22px;
  margin-bottom: 18px;
  padding: 13px 14px;
  overflow: hidden;
  border-radius: var(--radius-panel);
  background: var(--lens);
  color: #fff;
}

.lens-identity {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 12px;
}

.lens-orbit {
  position: relative;
  width: 42px;
  height: 42px;
  display: grid;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 50%;
}

.lens-orbit-ring {
  width: 24px;
  height: 24px;
  border: 1px solid rgba(255, 255, 255, 0.28);
  border-radius: 50%;
}

.lens-orbit-dot {
  position: absolute;
  width: 8px;
  height: 8px;
  border: 2px solid var(--lens);
  border-radius: 50%;
  background: var(--ink-tertiary);
  box-sizing: content-box;
}

.lens-orbit.is-online .lens-orbit-dot,
.lens-reading-value.is-online .lens-status-dot {
  background: #4dcc77;
}

.lens-orbit.is-offline .lens-orbit-dot,
.lens-reading-value.is-offline .lens-status-dot {
  background: #ff665c;
}

.lens-selector {
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.lens-kicker {
  margin-bottom: 2px;
  color: var(--lens-muted);
  font-size: 10px;
  font-weight: 650;
  letter-spacing: 0.075em;
  text-transform: uppercase;
}

.lens-select {
  min-width: 0;
  max-width: 190px;
  padding: 0 20px 0 0;
  overflow: hidden;
  border: 0;
  outline: 0;
  background: transparent;
  color: #fff;
  font-size: 16px;
  font-weight: 650;
  text-overflow: ellipsis;
  cursor: pointer;
}

.lens-select option {
  background: var(--lens);
  color: #fff;
}

.lens-readings {
  display: flex;
  align-items: stretch;
}

.lens-reading {
  min-width: 108px;
  display: grid;
  grid-template-columns: 17px auto;
  grid-template-rows: auto auto;
  column-gap: 6px;
  padding: 3px 15px;
  border-left: 1px solid rgba(255, 255, 255, 0.12);
}

.lens-reading > svg {
  grid-row: 1 / 3;
  align-self: center;
  color: #c7c7cc;
}

.lens-reading-label {
  color: var(--lens-muted);
  font-size: 10px;
}

.lens-reading-value {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  color: #f5f5f7;
  font-size: 11px;
  font-weight: 600;
}

.lens-status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--ink-tertiary);
}

.lens-retry {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  margin-left: 8px;
  padding: 7px 8px;
  border-radius: 7px;
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
  cursor: pointer;
  font-size: 11px;
}

.lens-retry:hover {
  background: rgba(255, 255, 255, 0.14);
}

.lens-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}

.lens-actions :deep(.btn-secondary),
.lens-actions :deep(.btn-ghost) {
  border-color: rgba(255, 255, 255, 0.18);
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
}

.lens-actions :deep(.btn-secondary:hover:not(:disabled)),
.lens-actions :deep(.btn-ghost:hover:not(:disabled)) {
  border-color: rgba(255, 255, 255, 0.3);
  background: rgba(255, 255, 255, 0.14);
}

@media (max-width: 1120px) {
  .server-lens {
    grid-template-columns: minmax(230px, 1fr) auto;
  }

  .lens-readings {
    grid-column: 1 / -1;
    grid-row: 2;
    padding-top: 10px;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
  }

  .lens-reading:first-child {
    padding-left: 0;
    border-left: 0;
  }
}

@media (max-width: 680px) {
  .server-lens {
    grid-template-columns: 1fr;
    gap: 10px;
    min-height: 0;
    margin-bottom: 12px;
    padding: 12px;
    border: 1px solid rgba(136, 203, 219, 0.12);
    border-radius: 18px;
    background: var(--mobile-dock);
    box-shadow: 0 8px 24px rgba(18, 37, 47, 0.13);
  }

  .lens-identity {
    gap: 10px;
  }

  .lens-orbit {
    width: 34px;
    height: 34px;
  }

  .lens-orbit-ring {
    width: 19px;
    height: 19px;
  }

  .lens-orbit-dot {
    width: 7px;
    height: 7px;
  }

  .lens-kicker {
    font-family: var(--font-mono);
    font-size: 8px;
  }

  .lens-select {
    max-width: min(68vw, 260px);
    font-size: 15px;
  }

  .lens-actions {
    justify-content: stretch;
    gap: 7px;
  }

  .lens-actions :deep(.btn) {
    flex: 1;
    min-width: 0;
    min-height: 38px;
    padding-right: 9px;
    padding-left: 9px;
    font-size: 12px;
  }

  .lens-readings {
    grid-column: auto;
    grid-row: auto;
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    overflow: visible;
    padding-top: 10px;
  }

  .lens-reading {
    min-width: 0;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    flex-direction: column;
    gap: 1px;
    padding: 0 10px;
  }

  .lens-reading:first-child {
    padding-left: 2px;
  }

  .lens-reading > svg {
    display: none;
  }

  .lens-reading-label {
    font-family: var(--font-mono);
    font-size: 8px;
    letter-spacing: 0.035em;
  }

  .lens-reading-value {
    min-width: 0;
    overflow: hidden;
    font-size: 10px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .lens-retry {
    margin-left: 7px;
    padding: 6px 7px;
    font-size: 10px;
  }
}
</style>
