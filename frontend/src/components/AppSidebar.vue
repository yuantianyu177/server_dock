<script setup>
import { onBeforeUnmount, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Boxes,
  ClipboardCheck,
  Container,
  Database,
  ExternalLink,
  Layers3,
  LogOut,
  Server,
  Settings,
  X
} from '@lucide/vue'
import { auth } from '@/stores/auth'
import { useApplicationBadge } from '@/composables/useApplicationBadge'
import BrandMark from './BrandMark.vue'

const emit = defineEmits(['close'])
const router = useRouter()
const route = useRoute()
const { pendingCount, refreshPendingCount } = useApplicationBadge()
let pendingTimer = null

const navGroups = [
  {
    label: '基础设施',
    items: [
      { label: '服务器', path: '/servers', icon: Server },
      { label: '容器', path: '/containers', icon: Container },
      { label: '镜像', path: '/images', icon: Layers3 },
      { label: '数据卷', path: '/volumes', icon: Database }
    ]
  },
  {
    label: '工作流',
    items: [
      { label: '申请审批', path: '/applications', icon: ClipboardCheck, pendingBadge: true }
    ]
  },
  {
    label: '系统',
    items: [
      { label: '系统设置', path: '/config', icon: Settings }
    ]
  }
]

function isActive(path) {
  return route.path === path || route.path.startsWith(`${path}/`)
}

function logout() {
  auth.logout()
  router.push('/login')
}

function formatPendingCount(count) {
  return count > 99 ? '99+' : String(count)
}

function handleVisibilityChange() {
  if (document.visibilityState === 'visible') refreshPendingCount()
}

onMounted(() => {
  refreshPendingCount()
  pendingTimer = window.setInterval(() => {
    if (document.visibilityState === 'visible') refreshPendingCount()
  }, 30000)
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onBeforeUnmount(() => {
  window.clearInterval(pendingTimer)
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})
</script>

<template>
  <aside class="sidebar" aria-label="主导航">
    <header class="sidebar-brand">
      <router-link class="brand-link" to="/servers" aria-label="ServerDock 首页" @click="emit('close')">
        <BrandMark :size="32" />
        <span class="brand-name">ServerDock</span>
      </router-link>
      <button class="sidebar-close" type="button" aria-label="关闭导航" @click="emit('close')">
        <X :size="19" aria-hidden="true" />
      </button>
    </header>

    <nav class="sidebar-nav">
      <section v-for="group in navGroups" :key="group.label" class="nav-group">
        <h2 class="nav-group-label">{{ group.label }}</h2>
        <router-link
          v-for="item in group.items"
          :key="item.path"
          :to="item.path"
          class="nav-item"
          :class="{ active: isActive(item.path) }"
          :aria-current="isActive(item.path) ? 'page' : undefined"
          @click="emit('close')"
        >
          <component :is="item.icon" :size="18" :stroke-width="1.8" aria-hidden="true" />
          <span class="nav-item-label">{{ item.label }}</span>
          <span
            v-if="item.pendingBadge && pendingCount > 0"
            class="pending-badge"
            :aria-label="`${pendingCount} 个待处理申请`"
          >
            {{ formatPendingCount(pendingCount) }}
          </span>
        </router-link>
      </section>
    </nav>

    <footer class="sidebar-footer">
      <a class="apply-link" href="/apply" target="_blank" rel="noopener">
        <Boxes :size="17" :stroke-width="1.8" aria-hidden="true" />
        <span>公开申请页</span>
        <ExternalLink :size="13" aria-hidden="true" />
      </a>

      <div class="account-row">
        <span class="account-avatar" aria-hidden="true">
          {{ (auth.user?.username || 'A').slice(0, 1).toUpperCase() }}
        </span>
        <div class="account-copy">
          <span class="account-name">{{ auth.user?.username || '管理员' }}</span>
          <span class="account-role">管理员</span>
        </div>
        <button class="logout-button" type="button" aria-label="退出登录" title="退出登录" @click="logout">
          <LogOut :size="17" :stroke-width="1.8" aria-hidden="true" />
        </button>
      </div>
    </footer>
  </aside>
</template>

<style scoped>
.sidebar {
  width: 232px;
  height: 100%;
  display: flex;
  flex: 0 0 auto;
  flex-direction: column;
  border-right: 1px solid var(--divider);
  background: rgba(255, 255, 255, 0.94);
}

.sidebar-brand {
  min-height: 68px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 14px 12px 18px;
}

.brand-link {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 10px;
  border-radius: 9px;
}

.brand-name {
  font-family: var(--font-display);
  color: var(--ink);
  font-size: 17px;
  font-weight: 720;
  letter-spacing: -0.02em;
}

.sidebar-close {
  width: 34px;
  height: 34px;
  display: none;
  place-items: center;
  border-radius: var(--radius-control);
  background: transparent;
  color: var(--ink-secondary);
  cursor: pointer;
}

.sidebar-nav {
  flex: 1;
  overflow-y: auto;
  padding: 5px 10px 16px;
}

.nav-group + .nav-group {
  margin-top: 21px;
}

.nav-group-label {
  margin: 0 10px 6px;
  color: var(--ink-tertiary);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.09em;
  text-transform: uppercase;
}

.nav-item {
  position: relative;
  min-height: 38px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 10px;
  border-radius: var(--radius-control);
  color: var(--ink-secondary);
  font-size: 13px;
  font-weight: 550;
  transition: background-color 160ms ease, color 160ms ease;
}

.nav-item:hover {
  background: #f0f0f3;
  color: var(--ink);
}

.nav-item.active {
  background: var(--blue-soft);
  color: #0059b5;
}

.nav-item-label {
  min-width: 0;
  flex: 1;
}

.pending-badge {
  min-width: 18px;
  height: 18px;
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  padding: 0 5px;
  border: 0;
  border-radius: 9px;
  background: var(--danger);
  color: #fff;
  font-family: var(--font-mono);
  font-size: 9px;
  font-weight: 750;
  line-height: 1;
}

.nav-item.active::before {
  content: "";
  position: absolute;
  top: 10px;
  bottom: 10px;
  left: -10px;
  width: 3px;
  border-radius: 0 3px 3px 0;
  background: var(--blue);
}

.sidebar-footer {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px 10px 14px;
  border-top: 1px solid var(--divider-subtle);
}

.apply-link {
  min-height: 38px;
  display: grid;
  grid-template-columns: 19px 1fr 14px;
  align-items: center;
  gap: 8px;
  padding: 0 10px;
  border-radius: var(--radius-control);
  color: var(--ink-secondary);
  font-size: 12px;
  font-weight: 550;
}

.apply-link:hover {
  background: #f0f0f3;
  color: var(--ink);
}

.account-row {
  min-height: 48px;
  display: grid;
  grid-template-columns: 30px 1fr 32px;
  align-items: center;
  gap: 9px;
  padding: 4px 5px 4px 8px;
}

.account-avatar {
  width: 30px;
  height: 30px;
  display: grid;
  place-items: center;
  border-radius: 9px;
  background: var(--ink);
  color: #fff;
  font-size: 12px;
  font-weight: 700;
}

.account-copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.account-name {
  overflow: hidden;
  color: var(--ink);
  font-size: 12px;
  font-weight: 650;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.account-role {
  color: var(--ink-tertiary);
  font-size: 10px;
}

.logout-button {
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  border-radius: var(--radius-control);
  background: transparent;
  color: var(--ink-tertiary);
  cursor: pointer;
}

.logout-button:hover {
  background: var(--danger-soft);
  color: var(--danger);
}

@media (max-width: 820px) {
  .sidebar {
    width: min(292px, calc(100vw - 48px));
    border-right: 0;
  }

  .sidebar-close {
    display: grid;
  }
}
</style>
