<script setup>
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

async function handleLogin() {
  if (!username.value || !password.value) {
    error.value = 'Please enter username and password.'
    return
  }
  loading.value = true
  error.value = ''
  try {
    await auth.login(username.value, password.value)
    const redirect = route.query.redirect || '/servers'
    router.push(redirect)
  } catch (e) {
    error.value = e.message || 'Login failed. Please check your credentials.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <!-- Background grid -->
    <div class="login-bg" aria-hidden="true">
      <div class="bg-blob bg-blob-1" />
      <div class="bg-blob bg-blob-2" />
    </div>

    <div class="login-card animate-in">
      <!-- Brand -->
      <div class="login-brand">
        <div class="login-brand-icon">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <rect x="2" y="2" width="20" height="8" rx="2"/>
            <rect x="2" y="14" width="20" height="8" rx="2"/>
            <line x1="6" y1="6" x2="6.01" y2="6"/>
            <line x1="6" y1="18" x2="6.01" y2="18"/>
          </svg>
        </div>
        <h1 class="login-brand-name">ServerDock</h1>
      </div>

      <div class="login-heading">
        <h2 class="login-title">Welcome back</h2>
        <p class="login-subtitle">Sign in to the admin panel</p>
      </div>

      <form class="login-form" @submit.prevent="handleLogin">
        <Transition name="fade">
          <div v-if="error" class="alert alert-error">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="flex-shrink:0;margin-top:1px">
              <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
            </svg>
            {{ error }}
          </div>
        </Transition>

        <div class="form-group">
          <label class="form-label">Username</label>
          <input
            v-model="username"
            type="text"
            class="form-input"
            placeholder="admin"
            autocomplete="username"
            autofocus
          />
        </div>

        <div class="form-group">
          <label class="form-label">Password</label>
          <input
            v-model="password"
            type="password"
            class="form-input"
            placeholder="••••••••"
            autocomplete="current-password"
            @keyup.enter="handleLogin"
          />
        </div>

        <button type="submit" class="btn btn-primary login-btn" :disabled="loading">
          <span v-if="loading" class="spinner" style="width:14px;height:14px;border-color:rgba(255,255,255,0.3);border-top-color:white" />
          {{ loading ? 'Signing in…' : 'Sign in' }}
        </button>
      </form>

      <div class="login-footer">
        <a href="/apply" class="login-apply-link">
          Apply for a container
          <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="9 18 15 12 9 6"/>
          </svg>
        </a>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  background: var(--cream);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  position: relative;
  overflow: hidden;
}

.login-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.bg-blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
}

.bg-blob-1 {
  width: 500px;
  height: 500px;
  background: radial-gradient(circle, rgba(201, 100, 66, 0.07) 0%, transparent 70%);
  top: -150px;
  right: -100px;
}

.bg-blob-2 {
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, rgba(201, 100, 66, 0.05) 0%, transparent 70%);
  bottom: -100px;
  left: -80px;
}

.login-card {
  width: 100%;
  max-width: 368px;
  background: white;
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 36px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.08), 0 1px 4px rgba(0,0,0,0.04);
  position: relative;
  z-index: 1;
}

.login-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 28px;
}

.login-brand-icon {
  width: 36px;
  height: 36px;
  background: var(--accent);
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  box-shadow: 0 2px 8px rgba(201, 100, 66, 0.3);
}

.login-brand-name {
  font-family: var(--font-serif);
  font-size: 22px;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: 0.02em;
}

.login-heading {
  margin-bottom: 24px;
}

.login-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1.2;
}

.login-subtitle {
  font-size: 13.5px;
  color: var(--text-muted);
  margin-top: 4px;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.login-btn {
  width: 100%;
  justify-content: center;
  padding: 10px;
  font-size: 14.5px;
  margin-top: 4px;
  box-shadow: 0 1px 4px rgba(201, 100, 66, 0.25);
}

.login-footer {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid var(--border-light);
  text-align: center;
}

.login-apply-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--text-muted);
  text-decoration: none;
  transition: color 0.15s;
}

.login-apply-link:hover {
  color: var(--accent);
}
</style>
