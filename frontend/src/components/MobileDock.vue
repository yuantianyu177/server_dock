<script setup>
import { useRoute } from 'vue-router'
import {
  ClipboardCheck,
  Container,
  Database,
  Layers3,
  Server,
  Settings
} from '@lucide/vue'
import { useApplicationBadge } from '@/composables/useApplicationBadge'

const route = useRoute()
const { pendingCount } = useApplicationBadge()

const items = [
  { label: '服务器', path: '/servers', icon: Server },
  { label: '容器', path: '/containers', icon: Container },
  { label: '镜像', path: '/images', icon: Layers3 },
  { label: '数据卷', path: '/volumes', icon: Database },
  { label: '审批', path: '/applications', icon: ClipboardCheck, badge: true },
  { label: '设置', path: '/config', icon: Settings }
]

function isActive(path) {
  return route.path === path || route.path.startsWith(`${path}/`)
}

function formatPendingCount(count) {
  return count > 99 ? '99+' : String(count)
}
</script>

<template>
  <nav class="mobile-dock" aria-label="移动端主导航">
    <router-link
      v-for="item in items"
      :key="item.path"
      :to="item.path"
      class="dock-item"
      :class="{ active: isActive(item.path) }"
      :aria-current="isActive(item.path) ? 'page' : undefined"
    >
      <span class="dock-icon">
        <component :is="item.icon" :size="19" :stroke-width="1.85" aria-hidden="true" />
        <span
          v-if="item.badge && pendingCount > 0"
          class="dock-badge"
          :aria-label="`${pendingCount} 个待处理申请`"
        >
          {{ formatPendingCount(pendingCount) }}
        </span>
      </span>
      <span class="dock-label">{{ item.label }}</span>
    </router-link>
  </nav>
</template>

<style scoped>
.mobile-dock {
  display: none;
}

@media (max-width: 820px) {
  .mobile-dock {
    position: fixed;
    right: 0;
    bottom: 0;
    left: 0;
    z-index: 100;
    height: calc(68px + env(safe-area-inset-bottom));
    display: grid;
    grid-template-columns: repeat(6, minmax(0, 1fr));
    gap: 2px;
    padding: 6px 8px max(6px, env(safe-area-inset-bottom));
    border: 1px solid rgba(173, 207, 218, 0.18);
    border-width: 1px 0 0;
    border-radius: 21px 21px 0 0;
    background: rgba(18, 37, 47, 0.96);
    box-shadow: 0 -8px 28px rgba(18, 37, 47, 0.18);
    backdrop-filter: blur(18px) saturate(140%);
    -webkit-backdrop-filter: blur(18px) saturate(140%);
  }

  .dock-item {
    position: relative;
    min-width: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-direction: column;
    gap: 2px;
    border-radius: 15px;
    color: #92a7b2;
    touch-action: manipulation;
    transition: background-color 160ms ease, color 160ms ease, transform 120ms ease;
  }

  .dock-item::before {
    content: "";
    position: absolute;
    top: 3px;
    width: 14px;
    height: 2px;
    border-radius: 2px;
    background: #76d2e8;
    opacity: 0;
    transform: scaleX(0.4);
    transition: opacity 160ms ease, transform 180ms ease;
  }

  .dock-item:active {
    transform: scale(0.96);
  }

  .dock-item.active {
    background: rgba(118, 210, 232, 0.1);
    color: #f7fbfc;
  }

  .dock-item.active::before {
    opacity: 1;
    transform: scaleX(1);
  }

  .dock-icon {
    position: relative;
    display: grid;
    place-items: center;
    line-height: 0;
  }

  .dock-item.active .dock-icon {
    color: #76d2e8;
  }

  .dock-label {
    max-width: 100%;
    overflow: hidden;
    font-size: 9px;
    font-weight: 650;
    line-height: 1.15;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .dock-badge {
    position: absolute;
    top: -8px;
    right: -11px;
    min-width: 16px;
    height: 16px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0 4px;
    border: 2px solid #12252f;
    border-radius: 8px;
    background: #e36345;
    color: #fff;
    font-family: var(--font-mono);
    font-size: 8px;
    font-weight: 750;
    line-height: 1;
  }
}

@media (max-width: 350px) {
  .mobile-dock {
    padding-right: 5px;
    padding-left: 5px;
  }

  .dock-label {
    font-size: 8px;
  }
}
</style>
