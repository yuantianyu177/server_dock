<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { auth } from '@/stores/auth'
import AppSidebar from './AppSidebar.vue'
import BrandMark from './BrandMark.vue'
import MobileDock from './MobileDock.vue'

const route = useRoute()
const sidebarOpen = ref(false)
const isMobile = ref(false)
const menuButtonRef = ref(null)
let mobileMediaQuery = null

function updateMobileState(event) {
  isMobile.value = event.matches
}

async function openSidebar() {
  sidebarOpen.value = true
  await nextTick()
  document.querySelector('.sidebar.open .sidebar-close')?.focus()
}

async function closeSidebar(restoreFocus = true) {
  sidebarOpen.value = false
  if (restoreFocus) {
    await nextTick()
    menuButtonRef.value?.focus()
  }
}

function handleEscape() {
  if (sidebarOpen.value) closeSidebar()
}

watch(() => route.path, () => closeSidebar(false))

onMounted(async () => {
  mobileMediaQuery = window.matchMedia('(max-width: 820px)')
  updateMobileState(mobileMediaQuery)
  mobileMediaQuery.addEventListener('change', updateMobileState)

  if (auth.isAuthenticated && !auth.user) {
    try {
      await auth.fetchMe()
    } catch {
      // The API client handles expired sessions and redirects to login.
    }
  }
})

onBeforeUnmount(() => {
  mobileMediaQuery?.removeEventListener('change', updateMobileState)
})
</script>

<template>
  <div class="app-layout" :class="{ 'sidebar-is-open': sidebarOpen }" @keydown.esc="handleEscape">
    <a class="skip-link" href="#main-content">跳到主要内容</a>

    <header class="mobile-toolbar">
      <router-link class="mobile-brand" to="/servers" aria-label="ServerDock 首页">
        <BrandMark :size="32" />
      </router-link>
      <div class="mobile-context">
        <span>ServerDock</span>
        <strong>{{ route.meta.title || '基础设施' }}</strong>
      </div>
      <button
        ref="menuButtonRef"
        class="mobile-menu"
        type="button"
        aria-label="打开账户与完整导航"
        :aria-expanded="sidebarOpen"
        @click="openSidebar"
      >
        <span aria-hidden="true">{{ (auth.user?.username || 'A').slice(0, 1).toUpperCase() }}</span>
      </button>
    </header>

    <Transition name="overlay">
      <button
        v-if="sidebarOpen"
        class="sidebar-overlay"
        type="button"
        aria-label="关闭主导航"
        @click="closeSidebar()"
      />
    </Transition>

    <AppSidebar
      :class="{ open: sidebarOpen }"
      :inert="isMobile && !sidebarOpen"
      :aria-hidden="isMobile && !sidebarOpen ? 'true' : undefined"
      @close="closeSidebar"
    />

    <main id="main-content" class="main-content" tabindex="-1">
      <div class="main-inner">
        <router-view v-slot="{ Component }">
          <Transition name="content" mode="out-in">
            <component :is="Component" :key="$route.path" />
          </Transition>
        </router-view>
      </div>
    </main>

    <MobileDock />
  </div>
</template>

<style scoped>
.app-layout {
  height: 100vh;
  height: 100dvh;
  display: flex;
  overflow: hidden;
  background: var(--canvas);
}

.main-content {
  min-width: 0;
  flex: 1;
  overflow-y: auto;
  outline: none;
  scroll-behavior: smooth;
}

.main-inner {
  width: 100%;
  max-width: 1384px;
  margin: 0 auto;
  padding: 30px 32px 48px;
}

.mobile-toolbar,
.sidebar-overlay {
  display: none;
}

.skip-link {
  position: fixed;
  top: 8px;
  left: 8px;
  z-index: 5000;
  padding: 9px 12px;
  border-radius: var(--radius-control);
  background: var(--ink);
  color: #fff;
  transform: translateY(-150%);
  transition: transform 160ms ease;
}

.skip-link:focus {
  transform: translateY(0);
}

.content-enter-active,
.content-leave-active {
  transition: opacity 150ms ease;
}

.content-enter-from,
.content-leave-to {
  opacity: 0;
}

@media (max-width: 820px) {
  .app-layout {
    flex-direction: column;
    background: var(--mobile-canvas);
  }

  .mobile-toolbar {
    position: relative;
    z-index: 80;
    min-height: calc(60px + env(safe-area-inset-top));
    display: grid;
    grid-template-columns: 36px minmax(0, 1fr) 38px;
    align-items: center;
    gap: 10px;
    padding: max(6px, env(safe-area-inset-top)) 12px 6px;
    border-bottom: 1px solid rgba(173, 207, 218, 0.2);
    background: rgba(18, 37, 47, 0.96);
    box-shadow: 0 5px 18px rgba(18, 37, 47, 0.12);
    backdrop-filter: saturate(180%) blur(16px);
    -webkit-backdrop-filter: saturate(180%) blur(16px);
  }

  .mobile-menu {
    width: 38px;
    height: 38px;
    display: grid;
    place-items: center;
    border: 1px solid rgba(173, 207, 218, 0.22);
    border-radius: 13px;
    background: rgba(255, 255, 255, 0.1);
    color: #fff;
    cursor: pointer;
    font-family: var(--font-display);
    font-size: 13px;
    font-weight: 750;
    box-shadow: 0 2px 8px rgba(24, 49, 61, 0.06);
  }

  .mobile-menu:hover {
    background: rgba(255, 255, 255, 0.16);
  }

  .mobile-brand {
    width: 36px;
    height: 36px;
    display: grid;
    align-items: center;
    place-items: center;
    border-radius: 11px;
  }

  .mobile-context {
    min-width: 0;
    display: flex;
    flex-direction: column;
  }

  .mobile-context > span {
    color: #76d2e8;
    font-family: var(--font-mono);
    font-size: 8px;
    font-weight: 700;
    letter-spacing: 0.12em;
    line-height: 1.2;
    text-transform: uppercase;
  }

  .mobile-context > strong {
    overflow: hidden;
    color: #fff;
    font-family: var(--font-display);
    font-size: 18px;
    font-weight: 740;
    letter-spacing: -0.025em;
    line-height: 1.25;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  :deep(.sidebar) {
    position: fixed;
    top: auto;
    right: 0;
    bottom: 0;
    left: 0;
    z-index: 200;
    width: 100%;
    height: auto;
    max-height: min(84dvh, 720px);
    border-radius: 24px 24px 0 0;
    transform: translateY(102%);
    transition: transform 260ms cubic-bezier(0.2, 0, 0, 1);
  }

  :deep(.sidebar.open) {
    transform: translateY(0);
    box-shadow: var(--shadow-popover);
  }

  .sidebar-overlay {
    position: fixed;
    inset: 0;
    z-index: 150;
    display: block;
    background: rgba(12, 27, 35, 0.48);
    backdrop-filter: blur(2px);
    -webkit-backdrop-filter: blur(2px);
    cursor: default;
  }

  .main-inner {
    padding: 14px 12px calc(102px + env(safe-area-inset-bottom));
  }

  .sidebar-is-open .main-content {
    overflow: hidden;
  }

  .overlay-enter-active,
  .overlay-leave-active {
    transition: opacity 180ms ease;
  }

  .overlay-enter-from,
  .overlay-leave-to {
    opacity: 0;
  }
}

@media (max-width: 520px) {
  .main-inner {
    padding-right: 10px;
    padding-left: 10px;
  }
}
</style>
