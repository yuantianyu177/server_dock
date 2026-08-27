<script setup>
import { nextTick, onMounted, ref, watch } from 'vue'
import { Menu } from '@lucide/vue'
import { useRoute } from 'vue-router'
import { auth } from '@/stores/auth'
import AppSidebar from './AppSidebar.vue'
import BrandMark from './BrandMark.vue'

const route = useRoute()
const sidebarOpen = ref(false)
const menuButtonRef = ref(null)

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
  if (auth.isAuthenticated && !auth.user) {
    try {
      await auth.fetchMe()
    } catch {
      // The API client handles expired sessions and redirects to login.
    }
  }
})
</script>

<template>
  <div class="app-layout" @keydown.esc="handleEscape">
    <a class="skip-link" href="#main-content">跳到主要内容</a>

    <header class="mobile-toolbar">
      <button
        ref="menuButtonRef"
        class="mobile-menu"
        type="button"
        aria-label="打开主导航"
        :aria-expanded="sidebarOpen"
        @click="openSidebar"
      >
        <Menu :size="20" aria-hidden="true" />
      </button>
      <router-link class="mobile-brand" to="/servers" aria-label="ServerDock 首页">
        <BrandMark :size="28" />
        <span>ServerDock</span>
      </router-link>
      <span class="toolbar-spacer" aria-hidden="true" />
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

    <AppSidebar :class="{ open: sidebarOpen }" @close="closeSidebar()" />

    <main id="main-content" class="main-content" tabindex="-1">
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
  }

  .mobile-toolbar {
    position: relative;
    z-index: 80;
    min-height: 52px;
    display: grid;
    grid-template-columns: 36px 1fr 36px;
    align-items: center;
    padding: 0 10px;
    border-bottom: 1px solid rgba(210, 210, 215, 0.78);
    background: rgba(255, 255, 255, 0.86);
    backdrop-filter: saturate(180%) blur(16px);
    -webkit-backdrop-filter: saturate(180%) blur(16px);
  }

  .mobile-menu {
    width: 36px;
    height: 36px;
    display: grid;
    place-items: center;
    border-radius: var(--radius-control);
    background: transparent;
    color: var(--ink);
    cursor: pointer;
  }

  .mobile-menu:hover {
    background: #eeeef1;
  }

  .mobile-brand {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    font-family: var(--font-display);
    font-size: 15px;
    font-weight: 700;
    letter-spacing: -0.015em;
  }

  :deep(.sidebar) {
    position: fixed;
    top: 0;
    bottom: 0;
    left: 0;
    z-index: 200;
    transform: translateX(-102%);
    transition: transform 240ms cubic-bezier(0.2, 0, 0, 1);
  }

  :deep(.sidebar.open) {
    transform: translateX(0);
    box-shadow: var(--shadow-popover);
  }

  .sidebar-overlay {
    position: fixed;
    inset: 0;
    z-index: 150;
    display: block;
    background: rgba(0, 0, 0, 0.32);
    cursor: default;
  }

  .main-inner {
    padding: 22px 18px 40px;
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
    padding: 18px 12px 32px;
  }
}
</style>
