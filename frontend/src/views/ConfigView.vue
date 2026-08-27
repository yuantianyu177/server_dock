<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import {
  Box,
  CircleAlert,
  CircleCheck,
  KeyRound,
  Mail,
  Network,
  RefreshCw,
  Save,
  Send,
  ShieldCheck
} from '@lucide/vue'
import { get, post, put } from '@/api/client'
import { useToast } from '@/composables/useToast'
import PageHeader from '@/components/PageHeader.vue'
import StatePanel from '@/components/StatePanel.vue'

const toast = useToast()
const config = ref({})
const loading = ref(true)
const loadError = ref('')
const saving = ref(false)
const testEmailLoading = ref(false)
const activeSection = ref('ports')
const dirty = ref(false)

const passwordForm = ref({ oldPassword: '', newPassword: '', confirm: '' })
const passwordLoading = ref(false)
const passwordError = ref('')
const passwordSuccess = ref(false)

const sections = [
  {
    id: 'ports',
    title: '端口分配',
    description: '定义新容器可使用的宿主机端口范围。',
    icon: Network,
    fields: [
      { key: 'port_range_start', label: '起始端口', type: 'number', hint: '自动分配范围中的第一个端口' },
      { key: 'port_range_end', label: '结束端口', type: 'number', hint: '自动分配范围中的最后一个端口' },
      { key: 'extra_ports_per_container', label: '额外端口数', type: 'number', hint: '除 SSH 端口外，为每个容器额外映射的端口数量' }
    ]
  },
  {
    id: 'containers',
    title: '容器默认值',
    description: '创建容器时自动应用的存储路径和 Docker 参数。',
    icon: Box,
    fields: [
      { key: 'default_volume_mount_path', label: '数据卷挂载路径', type: 'text', hint: '例如 /data' },
      { key: 'docker_extra_args', label: '额外 Docker 参数', type: 'textarea', hint: '例如 --gpus all；保存时会合并空白字符' }
    ]
  },
  {
    id: 'email',
    title: '邮件通知',
    description: '向管理员发送带快捷审批按钮的新申请通知，并向申请人发送审批结果和连接信息。',
    icon: Mail,
    fields: [
      { key: 'email_enabled', label: '启用邮件通知', type: 'boolean', hint: '关闭后不发送申请与审批邮件' },
      { key: 'admin_email', label: '管理员邮箱', type: 'email', hint: '接收新容器申请通知' },
      { key: 'public_url', label: '公开访问地址', type: 'url', hint: '邮件审批按钮的访问地址，例如 https://serverdock.example.com' },
      { key: 'smtp_host', label: 'SMTP 主机', type: 'text', hint: '例如 smtp.example.com' },
      { key: 'smtp_port', label: 'SMTP 端口', type: 'number', hint: '常用端口为 465 或 587' },
      { key: 'smtp_username', label: 'SMTP 用户名', type: 'text', hint: '用于登录 SMTP 服务' },
      { key: 'smtp_password', label: 'SMTP 密码', type: 'password', hint: '用于登录 SMTP 服务' },
      { key: 'smtp_use_tls', label: '使用 TLS', type: 'boolean', hint: '通过加密连接发送邮件' }
    ]
  },
  {
    id: 'security',
    title: '账户安全',
    description: '更新当前管理员账户的登录密码。',
    icon: ShieldCheck,
    fields: []
  }
]

const configKeys = sections.flatMap(section => section.fields.map(field => field.key))

async function loadConfig() {
  loading.value = true
  loadError.value = ''
  try {
    config.value = await get('/config') || {}
    dirty.value = false
  } catch (error) {
    config.value = {}
    loadError.value = error.message
  } finally {
    loading.value = false
  }
}

function markDirty() {
  dirty.value = true
}

async function saveAll() {
  saving.value = true
  try {
    for (const key of configKeys) {
      if (config.value[key] !== undefined && config.value[key] !== null) {
        await put(`/config/${key}`, { value: String(config.value[key]) })
      }
    }
    dirty.value = false
    toast.success('已保存系统设置')
  } catch (error) {
    toast.error(`无法保存系统设置：${error.message}`)
  } finally {
    saving.value = false
  }
}

async function testEmail() {
  testEmailLoading.value = true
  try {
    await post('/config/test-email')
    toast.success('测试邮件已发送，请检查收件箱')
  } catch (error) {
    toast.error(`无法发送测试邮件：${error.message}`)
  } finally {
    testEmailLoading.value = false
  }
}

async function changePassword() {
  passwordError.value = ''
  passwordSuccess.value = false
  if (!passwordForm.value.oldPassword) {
    passwordError.value = '请输入当前密码。'
    return
  }
  if (passwordForm.value.newPassword.length < 6) {
    passwordError.value = '新密码至少需要 6 个字符。'
    return
  }
  if (passwordForm.value.newPassword !== passwordForm.value.confirm) {
    passwordError.value = '两次输入的新密码不一致。'
    return
  }

  passwordLoading.value = true
  try {
    await post('/auth/change-password', {
      old_password: passwordForm.value.oldPassword,
      new_password: passwordForm.value.newPassword
    })
    passwordForm.value = { oldPassword: '', newPassword: '', confirm: '' }
    passwordSuccess.value = true
    toast.success('管理员密码已更新')
  } catch (error) {
    passwordError.value = `无法更新密码：${error.message}`
  } finally {
    passwordLoading.value = false
  }
}

function scrollToSection(id) {
  activeSection.value = id
  const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  document.getElementById(`settings-${id}`)?.scrollIntoView({
    behavior: reducedMotion ? 'auto' : 'smooth',
    block: 'start'
  })
}

function beforeUnload(event) {
  if (!dirty.value) return
  event.preventDefault()
  event.returnValue = ''
}

onMounted(() => {
  loadConfig()
  window.addEventListener('beforeunload', beforeUnload)
})

onBeforeUnmount(() => window.removeEventListener('beforeunload', beforeUnload))
</script>

<template>
  <div class="config-page">
    <PageHeader title="系统设置" description="配置端口、容器默认值、邮件通知和管理员账户。">
      <template #actions>
        <span v-if="dirty" class="unsaved-indicator" role="status"><span />有未保存更改</span>
        <button class="btn btn-primary" type="button" :disabled="saving || !dirty" @click="saveAll">
          <span v-if="saving" class="spinner button-spinner" aria-hidden="true" />
          <Save v-else :size="15" aria-hidden="true" />
          {{ saving ? '正在保存…' : '保存设置' }}
        </button>
      </template>
    </PageHeader>

    <div v-if="loading" class="config-loading" role="status"><span class="spinner" /><span>正在读取系统设置…</span></div>
    <div v-else-if="loadError" class="surface">
      <StatePanel tone="error" title="无法读取系统设置" :description="`${loadError}。请检查 API 服务后重试。`">
        <template #icon><CircleAlert :size="20" /></template>
        <template #actions><button class="btn btn-secondary btn-sm" type="button" @click="loadConfig"><RefreshCw :size="14" />重新加载</button></template>
      </StatePanel>
    </div>

    <div v-else class="config-layout">
      <nav class="settings-nav" aria-label="设置类别">
        <button
          v-for="section in sections"
          :key="section.id"
          class="settings-nav-item"
          :class="{ active: activeSection === section.id }"
          :aria-pressed="activeSection === section.id"
          type="button"
          @click="scrollToSection(section.id)"
        >
          <component :is="section.icon" :size="17" :stroke-width="1.8" aria-hidden="true" />
          <span>{{ section.title }}</span>
        </button>
      </nav>

      <div class="settings-content">
        <section
          v-for="section in sections"
          :id="`settings-${section.id}`"
          :key="section.id"
          class="settings-section"
          :aria-labelledby="`settings-title-${section.id}`"
        >
          <header class="settings-heading">
            <div class="settings-heading-icon" aria-hidden="true"><component :is="section.icon" :size="18" :stroke-width="1.8" /></div>
            <div>
              <h2 :id="`settings-title-${section.id}`">{{ section.title }}</h2>
              <p>{{ section.description }}</p>
            </div>
          </header>

          <div class="settings-panel">
            <div v-for="field in section.fields" :key="field.key" class="setting-row">
              <div class="setting-copy">
                <label class="setting-label" :for="`config-${field.key}`">{{ field.label }}</label>
                <p v-if="field.hint" class="setting-hint">{{ field.hint }}</p>
              </div>
              <div class="setting-control">
                <label v-if="field.type === 'boolean'" class="switch" :aria-label="field.label">
                  <input
                    :id="`config-${field.key}`"
                    type="checkbox"
                    :checked="config[field.key] === 'true'"
                    @change="event => { config[field.key] = event.target.checked ? 'true' : 'false'; markDirty() }"
                  />
                  <span class="switch-track" />
                </label>
                <textarea
                  v-else-if="field.type === 'textarea'"
                  :id="`config-${field.key}`"
                  v-model="config[field.key]"
                  class="form-textarea config-textarea mono"
                  @input="markDirty"
                />
                <input
                  v-else
                  :id="`config-${field.key}`"
                  v-model="config[field.key]"
                  :type="field.type"
                  class="form-input config-input"
                  :class="{ mono: field.type === 'number' || field.key === 'smtp_host' }"
                  :autocomplete="field.type === 'password' ? 'new-password' : 'off'"
                  @input="markDirty"
                />
              </div>
            </div>

            <div v-if="section.id === 'email'" class="setting-row setting-action-row">
              <div class="setting-copy">
                <span class="setting-label">测试邮件</span>
                <p class="setting-hint">使用已保存的 SMTP 设置向管理员邮箱发送测试邮件。</p>
              </div>
              <div class="setting-control test-control">
                <button class="btn btn-secondary btn-sm" type="button" :disabled="testEmailLoading || dirty" @click="testEmail">
                  <span v-if="testEmailLoading" class="spinner small-spinner" />
                  <Send v-else :size="14" aria-hidden="true" />
                  {{ testEmailLoading ? '正在发送…' : '发送测试邮件' }}
                </button>
                <span v-if="dirty" class="test-hint">请先保存上方更改。</span>
              </div>
            </div>

            <template v-if="section.id === 'security'">
              <div v-if="passwordError" class="password-alert alert alert-error" role="alert"><CircleAlert :size="17" />{{ passwordError }}</div>
              <div v-if="passwordSuccess" class="password-alert alert alert-success" role="status"><CircleCheck :size="17" />密码已更新。</div>

              <div class="setting-row">
                <div class="setting-copy"><label class="setting-label" for="current-password">当前密码</label></div>
                <div class="setting-control"><input id="current-password" v-model="passwordForm.oldPassword" type="password" class="form-input config-input" autocomplete="current-password" /></div>
              </div>
              <div class="setting-row">
                <div class="setting-copy"><label class="setting-label" for="new-password">新密码</label><p class="setting-hint">至少 6 个字符。</p></div>
                <div class="setting-control"><input id="new-password" v-model="passwordForm.newPassword" type="password" class="form-input config-input" autocomplete="new-password" /></div>
              </div>
              <div class="setting-row">
                <div class="setting-copy"><label class="setting-label" for="confirm-password">再次输入新密码</label></div>
                <div class="setting-control"><input id="confirm-password" v-model="passwordForm.confirm" type="password" class="form-input config-input" autocomplete="new-password" @keyup.enter="changePassword" /></div>
              </div>
              <div class="setting-row setting-action-row">
                <div class="setting-copy" />
                <div class="setting-control">
                  <button class="btn btn-secondary" type="button" :disabled="passwordLoading" @click="changePassword">
                    <span v-if="passwordLoading" class="spinner small-spinner" />
                    <KeyRound v-else :size="15" aria-hidden="true" />
                    {{ passwordLoading ? '正在更新…' : '更新密码' }}
                  </button>
                </div>
              </div>
            </template>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<style scoped>
.config-page {
  position: relative;
}

.unsaved-indicator {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--warning);
  font-size: 12px;
  font-weight: 600;
}

.unsaved-indicator > span {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--warning);
}

.config-loading {
  min-height: 360px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: var(--ink-secondary);
}

.config-layout {
  display: grid;
  grid-template-columns: 190px minmax(0, 1fr);
  gap: 34px;
  align-items: start;
}

.settings-nav {
  position: sticky;
  top: 24px;
  display: flex;
  flex-direction: column;
  gap: 3px;
  padding-right: 15px;
  border-right: 1px solid var(--divider);
}

.settings-nav-item {
  min-height: 38px;
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 0 10px;
  border-radius: var(--radius-control);
  background: transparent;
  color: var(--ink-secondary);
  cursor: pointer;
  font-size: 13px;
  font-weight: 550;
  text-align: left;
}

.settings-nav-item:hover {
  background: #eeeef1;
  color: var(--ink);
}

.settings-nav-item.active {
  background: var(--blue-soft);
  color: #0059b5;
}

.settings-content {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 38px;
}

.settings-section {
  scroll-margin-top: 20px;
}

.settings-heading {
  display: grid;
  grid-template-columns: 36px 1fr;
  align-items: start;
  gap: 11px;
  margin-bottom: 12px;
}

.settings-heading-icon {
  width: 36px;
  height: 36px;
  display: grid;
  place-items: center;
  border: 1px solid var(--divider);
  border-radius: 10px;
  background: var(--surface);
  color: var(--ink-secondary);
}

.settings-heading h2 {
  color: var(--ink);
  font-size: 17px;
  font-weight: 700;
  letter-spacing: -0.01em;
}

.settings-heading p {
  margin-top: 2px;
  color: var(--ink-secondary);
  font-size: 12px;
}

.settings-panel {
  overflow: hidden;
  border: 1px solid var(--divider);
  border-radius: var(--radius-panel);
  background: var(--surface);
}

.setting-row {
  min-height: 70px;
  display: grid;
  grid-template-columns: minmax(190px, 0.72fr) minmax(260px, 1fr);
  align-items: center;
  gap: 24px;
  padding: 13px 16px;
  border-bottom: 1px solid var(--divider-subtle);
}

.setting-row:last-child {
  border-bottom: 0;
}

.setting-copy {
  min-width: 0;
}

.setting-label {
  color: var(--ink);
  font-size: 13px;
  font-weight: 600;
}

.setting-hint {
  max-width: 390px;
  margin-top: 3px;
  color: var(--ink-secondary);
  font-size: 11px;
  line-height: 1.45;
}

.setting-control {
  min-width: 0;
  display: flex;
  align-items: center;
}

.config-input,
.config-textarea {
  max-width: 390px;
}

.config-textarea {
  min-height: 76px;
  font-size: 12px;
}

.setting-action-row {
  background: #fafafa;
}

.test-control {
  flex-wrap: wrap;
  gap: 9px;
}

.test-hint {
  color: var(--warning);
  font-size: 11px;
}

.password-alert {
  margin: 14px 16px 0;
}

.button-spinner,
.small-spinner {
  width: 13px;
  height: 13px;
  color: currentColor;
}

.button-spinner {
  color: #fff;
}

@media (max-width: 920px) {
  .config-layout {
    grid-template-columns: 1fr;
    gap: 22px;
  }

  .settings-nav {
    position: static;
    flex-direction: row;
    gap: 4px;
    overflow-x: auto;
    padding: 0 0 9px;
    border-right: 0;
    border-bottom: 1px solid var(--divider);
    scrollbar-width: none;
  }

  .settings-nav::-webkit-scrollbar {
    display: none;
  }

  .settings-nav-item {
    flex: 0 0 auto;
    white-space: nowrap;
  }
}

@media (max-width: 680px) {
  .setting-row {
    grid-template-columns: 1fr;
    gap: 9px;
    padding: 14px 12px;
  }

  .config-input,
  .config-textarea {
    max-width: none;
  }

  .setting-action-row .setting-copy:empty {
    display: none;
  }

  .unsaved-indicator {
    width: 100%;
  }
}
</style>
