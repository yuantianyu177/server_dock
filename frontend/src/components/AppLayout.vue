<script setup>
import { ref, onMounted, provide, watch } from 'vue'
import { useRoute } from 'vue-router'
import { auth } from '@/stores/auth'
import AppSidebar from './AppSidebar.vue'

const route = useRoute()
const sidebarOpen = ref(false)

provide('sidebarOpen', sidebarOpen)

function closeSidebar() {
  sidebarOpen.value = false
}

watch(() => route.path, () => {
  sidebarOpen.value = false
})

onMounted(async () => {
  if (auth.isAuthenticated && !auth.user) {
    try { await auth.fetchMe() } catch {}
  }
})
</script>

<template>
  <div class="layout">
    <!-- Mobile header -->
    <header class="mobile-header">
      <button class="mobile-menu-btn" @click="sidebarOpen = true">
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/>
        </svg>
      </button>
      <div class="mobile-brand">
        <div class="mobile-brand-icon">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <rect x="2" y="2" width="20" height="8" rx="2"/>
            <rect x="2" y="14" width="20" height="8" rx="2"/>
            <line x1="6" y1="6" x2="6.01" y2="6"/>
            <line x1="6" y1="18" x2="6.01" y2="18"/>
          </svg>
        </div>
        <span class="mobile-brand-name">ServerDock</span>
      </div>
      <div style="width:36px" />
    </header>

    <!-- Sidebar overlay (mobile) -->
    <Transition name="overlay">
      <div v-if="sidebarOpen" class="sidebar-overlay" @click="closeSidebar" />
    </Transition>

    <AppSidebar :class="{ open: sidebarOpen }" />

    <main class="main">
      <div class="main-inner">
        <router-view v-slot="{ Component }">
          <Transition name="content" mode="out-in">
            <component :is="Component" :key="$route.path" />
          </Transition>
        </router-view>
      </div>
    </main>
  </div>
</template>

<style scoped>
.layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

.main {
  flex: 1;
  overflow-y: auto;
  background: var(--cream);
}

.main-inner {
  max-width: 1200px;
  margin: 0 auto;
  padding: 32px 36px;
}

/* Mobile header - hidden on desktop */
.mobile-header {
  display: none;
}

/* Sidebar overlay - hidden on desktop */
.sidebar-overlay {
  display: none;
}

/* Content transition */
:deep(.content-enter-active) {
  transition: opacity 0.16s ease, transform 0.16s ease;
}
:deep(.content-leave-active) {
  transition: opacity 0.1s ease;
}
:deep(.content-enter-from) {
  opacity: 0;
  transform: translateY(8px);
}
:deep(.content-leave-to) {
  opacity: 0;
}

/* ── Mobile ── */
@media (max-width: 768px) {
  .layout {
    flex-direction: column;
  }

  .mobile-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 52px;
    padding: 0 14px;
    background: var(--sidebar-bg);
    flex-shrink: 0;
    z-index: 100;
  }

  .mobile-menu-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    border-radius: 8px;
    background: none;
    border: none;
    color: var(--sidebar-text);
    cursor: pointer;
    transition: background 0.15s;
  }

  .mobile-menu-btn:hover {
    background: var(--sidebar-hover);
  }

  .mobile-brand {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .mobile-brand-icon {
    width: 26px;
    height: 26px;
    background: var(--accent);
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
  }

  .mobile-brand-name {
    font-family: var(--font-serif);
    font-size: 17px;
    font-weight: 600;
    color: var(--sidebar-text);
    letter-spacing: 0.02em;
  }

  /* Sidebar becomes a drawer */
  :deep(.sidebar) {
    position: fixed;
    left: 0;
    top: 0;
    bottom: 0;
    z-index: 200;
    transform: translateX(-100%);
    transition: transform 0.28s cubic-bezier(0.16, 1, 0.3, 1);
    box-shadow: none;
  }

  :deep(.sidebar.open) {
    transform: translateX(0);
    box-shadow: 4px 0 24px rgba(0,0,0,0.2);
  }

  .sidebar-overlay {
    display: block;
    position: fixed;
    inset: 0;
    background: rgba(0,0,0,0.4);
    z-index: 150;
  }

  .overlay-enter-active {
    transition: opacity 0.2s ease;
  }
  .overlay-leave-active {
    transition: opacity 0.15s ease;
  }
  .overlay-enter-from,
  .overlay-leave-to {
    opacity: 0;
  }

  .main-inner {
    padding: 20px 16px;
  }
}
</style>
