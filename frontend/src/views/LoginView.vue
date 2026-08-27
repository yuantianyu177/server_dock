<script setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ArrowRight,
  Boxes,
  Container,
  Eye,
  EyeOff,
  LockKeyhole,
  Server,
  SquareTerminal
} from '@lucide/vue'
import { auth } from '@/stores/auth'
import { useToast } from '@/composables/useToast'
import BrandMark from '@/components/BrandMark.vue'

const router = useRouter()
const route = useRoute()
const toast = useToast()
const username = ref('')
const password = ref('')
const passwordVisible = ref(false)
const loading = ref(false)

async function handleLogin() {
  if (!username.value.trim() || !password.value) {
    toast.warning('请输入用户名和密码')
    return
  }

  loading.value = true
  try {
    await auth.login(username.value.trim(), password.value)
    const redirect = typeof route.query.redirect === 'string' && route.query.redirect.startsWith('/')
      ? route.query.redirect
      : '/servers'
    await router.push(redirect)
  } catch (loginError) {
    toast.error(`无法登录：${loginError.message || '请检查用户名和密码'}`)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <header class="public-header">
      <router-link class="public-brand" to="/login" aria-label="ServerDock 登录页">
        <BrandMark :size="34" />
        <span>ServerDock</span>
      </router-link>
      <router-link class="application-link" to="/apply">
        申请容器
        <ArrowRight :size="15" aria-hidden="true" />
      </router-link>
    </header>

    <main class="login-main">
      <section class="login-context" aria-labelledby="login-context-title">
        <p class="context-kicker">基础设施控制台</p>
        <h1 id="login-context-title">进入 ServerDock</h1>
        <p class="context-description">登录后管理服务器、Docker 资源、Web Terminal 和容器申请。</p>

        <div class="context-flow" aria-label="ServerDock 管理范围">
          <div class="flow-item">
            <span class="flow-icon"><Server :size="18" :stroke-width="1.7" /></span>
            <div><strong>服务器连接</strong><span>SSH 主机与凭据</span></div>
          </div>
          <div class="flow-line" aria-hidden="true" />
          <div class="flow-item">
            <span class="flow-icon"><Container :size="18" :stroke-width="1.7" /></span>
            <div><strong>Docker 资源</strong><span>容器、镜像与数据卷</span></div>
          </div>
          <div class="flow-line" aria-hidden="true" />
          <div class="flow-item">
            <span class="flow-icon"><SquareTerminal :size="18" :stroke-width="1.7" /></span>
            <div><strong>远程操作</strong><span>Web Terminal 与审批</span></div>
          </div>
        </div>
      </section>

      <section class="login-panel" aria-labelledby="login-title">
        <div class="panel-icon" aria-hidden="true"><LockKeyhole :size="20" /></div>
        <h2 id="login-title">管理员登录</h2>
        <p class="login-description">使用 ServerDock 管理员账户继续。</p>

        <form class="login-form" @submit.prevent="handleLogin">
          <div class="form-group">
            <label class="form-label" for="login-username">用户名</label>
            <input
              id="login-username"
              v-model="username"
              class="form-input"
              type="text"
              autocomplete="username"
              autofocus
              placeholder="请输入用户名"
              :disabled="loading"
              required
            />
          </div>

          <div class="form-group">
            <label class="form-label" for="login-password">密码</label>
            <div class="password-field">
              <input
                id="login-password"
                v-model="password"
                class="form-input"
                :type="passwordVisible ? 'text' : 'password'"
                autocomplete="current-password"
                placeholder="请输入密码"
                :disabled="loading"
                required
              />
              <button
                class="password-toggle"
                type="button"
                :aria-label="passwordVisible ? '隐藏密码' : '显示密码'"
                :aria-pressed="passwordVisible"
                @click="passwordVisible = !passwordVisible"
              >
                <Eye v-if="passwordVisible" :size="17" aria-hidden="true" />
                <EyeOff v-else :size="17" aria-hidden="true" />
              </button>
            </div>
          </div>

          <button class="btn btn-primary login-button" type="submit" :disabled="loading">
            <span v-if="loading" class="spinner login-spinner" aria-hidden="true" />
            {{ loading ? '正在登录…' : '登录管理控制台' }}
          </button>
        </form>

        <div class="panel-footer">
          <Boxes :size="15" aria-hidden="true" />
          <span>需要容器环境？</span>
          <router-link to="/apply">提交申请</router-link>
        </div>
      </section>
    </main>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  padding: 0 42px;
  background: var(--canvas);
}

.public-header {
  width: 100%;
  max-width: 1120px;
  min-height: 76px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 0 auto;
  border-bottom: 1px solid var(--divider);
}

.public-brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-family: var(--font-display);
  color: var(--ink);
  font-size: 18px;
  font-weight: 720;
  letter-spacing: -0.02em;
}

.application-link {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  color: #0066cc;
  font-size: 13px;
  font-weight: 600;
}

.application-link:hover {
  text-decoration: underline;
  text-underline-offset: 3px;
}

.login-main {
  width: 100%;
  max-width: 1000px;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 410px;
  align-items: center;
  gap: 96px;
  margin: auto;
  padding: 62px 0 84px;
}

.context-kicker {
  margin-bottom: 12px;
  color: #0066cc;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.09em;
  text-transform: uppercase;
}

.login-context h1 {
  max-width: 430px;
  color: var(--ink);
  font-family: var(--font-display);
  font-size: clamp(34px, 5vw, 48px);
  font-weight: 730;
  letter-spacing: -0.045em;
  line-height: 1.03;
}

.context-description {
  max-width: 430px;
  margin-top: 16px;
  color: var(--ink-secondary);
  font-size: 15px;
  line-height: 1.65;
}

.context-flow {
  margin-top: 42px;
}

.flow-item {
  display: grid;
  grid-template-columns: 38px 1fr;
  align-items: center;
  gap: 11px;
}

.flow-icon {
  width: 38px;
  height: 38px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--divider);
  border-radius: 10px;
  background: var(--surface);
  color: var(--ink-secondary);
  line-height: 0;
}

.flow-icon :deep(svg) {
  display: block;
  flex: 0 0 auto;
}

.flow-item div > strong,
.flow-item div > span {
  display: block;
}

.flow-item strong {
  color: var(--ink);
  font-size: 13px;
}

.flow-item div > span {
  margin-top: 1px;
  color: var(--ink-secondary);
  font-size: 11px;
}

.flow-line {
  width: 1px;
  height: 14px;
  margin-left: 19px;
  background: var(--divider);
}

.login-panel {
  padding: 32px;
  border: 1px solid var(--divider);
  border-radius: var(--radius-modal);
  background: var(--surface);
}

.panel-icon {
  width: 40px;
  height: 40px;
  display: grid;
  place-items: center;
  margin-bottom: 22px;
  border-radius: 11px;
  background: var(--ink);
  color: #fff;
}

.login-panel h2 {
  color: var(--ink);
  font-size: 22px;
  font-weight: 700;
  letter-spacing: -0.025em;
}

.login-description {
  margin-top: 4px;
  color: var(--ink-secondary);
  font-size: 13px;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-top: 26px;
}

.password-field {
  position: relative;
}

.password-field .form-input {
  padding-right: 42px;
}

.password-toggle {
  position: absolute;
  top: 50%;
  right: 6px;
  width: 30px;
  height: 30px;
  display: grid;
  place-items: center;
  border-radius: 7px;
  background: transparent;
  color: var(--ink-tertiary);
  cursor: pointer;
  transform: translateY(-50%);
}

.password-toggle:hover {
  background: #eeeef1;
  color: var(--ink);
}

.login-button {
  width: 100%;
  min-height: 42px;
  margin-top: 3px;
}

.login-spinner {
  width: 14px;
  height: 14px;
  color: #fff;
}

.panel-footer {
  display: flex;
  align-items: center;
  gap: 5px;
  margin-top: 24px;
  padding-top: 18px;
  border-top: 1px solid var(--divider-subtle);
  color: var(--ink-secondary);
  font-size: 12px;
}

.panel-footer a {
  margin-left: 2px;
  color: #0066cc;
  font-weight: 600;
}

@media (max-width: 860px) {
  .login-page {
    padding: 0 24px;
  }

  .login-main {
    grid-template-columns: 1fr;
    gap: 40px;
    max-width: 460px;
    padding: 44px 0 64px;
  }

  .login-context h1 {
    font-size: 34px;
  }

  .context-flow {
    display: none;
  }
}

@media (max-width: 520px) {
  .login-page {
    padding: 0 16px;
  }

  .public-header {
    min-height: 64px;
  }

  .login-main {
    display: block;
    padding: 32px 0 42px;
  }

  .login-context {
    margin-bottom: 28px;
  }

  .login-context h1 {
    font-size: 30px;
  }

  .context-description {
    font-size: 14px;
  }

  .login-panel {
    padding: 24px 20px;
  }
}
</style>
