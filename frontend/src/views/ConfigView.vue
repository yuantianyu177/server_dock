<script setup>
import { ref, onMounted } from 'vue'
import { configApi } from '@/api/config'
import { authApi } from '@/api/auth'
import { useToast } from '@/composables/useToast'

const toast = useToast()
const config = ref({})
const loading = ref(true)
const saving = ref(false)
const testEmailLoading = ref(false)
const activeSection = ref('ports')
const dirty = ref(false)

const pwForm = ref({ oldPassword: '', newPassword: '', confirm: '' })
const pwLoading = ref(false)
const pwError = ref('')
const pwSuccess = ref(false)

const sections = [
  {
    id: 'ports',
    title: 'Port Allocation',
    description: 'Define the range of host ports available for container mapping.',
    keys: [
      { key: 'port_range_start', label: 'Range Start', type: 'number', hint: 'First port in the allocation range' },
      { key: 'port_range_end', label: 'Range End', type: 'number', hint: 'Last port in the allocation range' },
      { key: 'extra_ports_per_container', label: 'Extra Ports', type: 'number', hint: 'Additional ports mapped per container' }
    ]
  },
  {
    id: 'containers',
    title: 'Container Defaults',
    description: 'Default settings applied when creating new containers.',
    keys: [
      { key: 'default_volume_mount_path', label: 'Volume Mount Path', type: 'text', hint: 'e.g. /data' },
      { key: 'docker_extra_args', label: 'Extra Docker Args', type: 'textarea', hint: 'One argument per line, e.g. --gpus all' }
    ]
  },
  {
    id: 'email',
    title: 'Email Notifications',
    description: 'Configure SMTP to send applicants status updates and container credentials.',
    keys: [
      { key: 'email_enabled', label: 'Enable Email', type: 'boolean', hint: 'Send email notifications for applications' },
      { key: 'admin_email', label: 'Admin Email', type: 'text', hint: 'Receives new application notifications' },
      { key: 'smtp_host', label: 'SMTP Host', type: 'text', hint: '' },
      { key: 'smtp_port', label: 'SMTP Port', type: 'number', hint: '' },
      { key: 'smtp_username', label: 'Username', type: 'text', hint: '' },
      { key: 'smtp_password', label: 'Password', type: 'password', hint: '' },
      { key: 'smtp_use_tls', label: 'Use TLS', type: 'boolean', hint: '' }
    ]
  },
  {
    id: 'security',
    title: 'Security',
    description: 'Manage admin account credentials.',
    keys: []
  }
]

const navItems = [
  {
    id: 'ports',
    label: 'Port Allocation',
    icon: '<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="2" width="20" height="8" rx="2"/><rect x="2" y="14" width="20" height="8" rx="2"/><line x1="6" y1="6" x2="6.01" y2="6"/><line x1="6" y1="18" x2="6.01" y2="18"/></svg>'
  },
  {
    id: 'containers',
    label: 'Container Defaults',
    icon: '<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/></svg>'
  },
  {
    id: 'email',
    label: 'Email',
    icon: '<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>'
  },
  {
    id: 'security',
    label: 'Security',
    icon: '<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>'
  }
]

async function loadConfig() {
  loading.value = true
  try {
    config.value = await configApi.all() || {}
  } catch (e) {
    toast.error(e.message)
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
    const keys = sections.flatMap(s => s.keys.map(k => k.key))
    for (const key of keys) {
      const val = config.value[key]
      if (val !== undefined && val !== null) {
        await configApi.update(key, String(val))
      }
    }
    dirty.value = false
    toast.success('All settings saved')
  } catch (e) {
    toast.error(e.message)
  } finally {
    saving.value = false
  }
}

async function testEmail() {
  testEmailLoading.value = true
  try {
    await configApi.testEmail()
    toast.success('Test email sent successfully.')
  } catch (e) {
    toast.error(e.message)
  } finally {
    testEmailLoading.value = false
  }
}

async function changePassword() {
  pwError.value = ''
  pwSuccess.value = false
  if (pwForm.value.newPassword !== pwForm.value.confirm) {
    pwError.value = 'New passwords do not match.'
    return
  }
  pwLoading.value = true
  try {
    await authApi.changePassword(pwForm.value.oldPassword, pwForm.value.newPassword)
    pwForm.value = { oldPassword: '', newPassword: '', confirm: '' }
    pwSuccess.value = true
    setTimeout(() => { pwSuccess.value = false }, 4000)
  } catch (e) {
    pwError.value = e.message
  } finally {
    pwLoading.value = false
  }
}

function scrollToSection(id) {
  activeSection.value = id
  document.getElementById('section-' + id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

onMounted(loadConfig)
</script>

<template>
  <div class="config-page">
    <!-- Page header -->
    <div class="page-header animate-in">
      <div class="page-header-left">
        <h1 class="page-title">Settings</h1>
        <p class="page-subtitle">Manage system configuration and account settings</p>
      </div>
      <div class="header-right">
        <button class="btn btn-primary btn-w-action" @click="saveAll" :disabled="saving || !dirty">
          <span v-if="saving" class="spinner" style="width:13px;height:13px;border-color:rgba(255,255,255,0.3);border-top-color:white" />
          Save Settings
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading-overlay" style="margin-top:80px">
      <span class="spinner" style="width:26px;height:26px" />
    </div>

    <div v-else class="config-body animate-in animate-in-delay-1">
      <!-- Left nav -->
      <nav class="settings-nav">
        <button
          v-for="item in navItems"
          :key="item.id"
          class="settings-nav-item"
          :class="{ active: activeSection === item.id }"
          @click="scrollToSection(item.id)"
        >
          <span class="nav-icon" v-html="item.icon"></span>
          <span class="nav-label">{{ item.label }}</span>
          <span class="nav-active-bar"></span>
        </button>
      </nav>

      <!-- Content -->
      <div class="settings-content">
        <div
          v-for="(section, si) in sections"
          :key="section.id"
          :id="'section-' + section.id"
          class="settings-section"
        >
          <div class="section-header">
            <div class="section-header-inner">
              <span class="section-index">0{{ si + 1 }}</span>
              <div>
                <h2 class="section-title">{{ section.title }}</h2>
                <p class="section-desc">{{ section.description }}</p>
              </div>
            </div>
          </div>

          <div class="settings-panel">
            <!-- Regular config fields -->
            <div
              v-for="(field, i) in section.keys"
              :key="field.key"
              class="field-row"
              :class="{ 'field-row-last': i === section.keys.length - 1 && section.id !== 'email' }"
            >
              <div class="field-meta">
                <div class="field-label">{{ field.label }}</div>
                <div v-if="field.hint" class="field-hint">{{ field.hint }}</div>
              </div>
              <div class="field-control">
                <!-- Boolean toggle -->
                <template v-if="field.type === 'boolean'">
                  <label class="toggle">
                    <input
                      type="checkbox"
                      :checked="config[field.key] === 'true'"
                      @change="e => { config[field.key] = e.target.checked ? 'true' : 'false'; markDirty() }"
                      class="toggle-input"
                    />
                    <span class="toggle-track"><span class="toggle-thumb" /></span>
                  </label>
                </template>
                <!-- Textarea -->
                <template v-else-if="field.type === 'textarea'">
                  <textarea v-model="config[field.key]" class="form-textarea cfg-textarea" @input="markDirty" />
                </template>
                <!-- Text / number / password -->
                <template v-else>
                  <input
                    v-model="config[field.key]"
                    :type="field.type"
                    class="form-input cfg-input"
                    @input="markDirty"
                  />
                </template>
              </div>
            </div>

            <!-- Email test row -->
            <div v-if="section.id === 'email'" class="field-row field-row-last field-row-action">
              <div class="field-meta">
                <div class="field-label">Test Connection</div>
                <div class="field-hint">Send a test email using the current SMTP settings</div>
              </div>
              <div class="field-control">
                <div style="display:flex;flex-direction:column;gap:10px;align-items:flex-start">
                  <button class="btn btn-secondary btn-sm" @click="testEmail" :disabled="testEmailLoading">
                    <span v-if="testEmailLoading" class="spinner" style="width:12px;height:12px" />
                    <svg v-else xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07A19.5 19.5 0 0 1 4.69 13.5a19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 3.62 2.74h3a2 2 0 0 1 2 1.72c.127.96.361 1.903.7 2.81a2 2 0 0 1-.45 2.11L7.91 10.34a16 16 0 0 0 6 6l.87-.87a2 2 0 0 1 2.11-.45c.907.339 1.85.573 2.81.7a2 2 0 0 1 1.72 2.02z"/></svg>
                    Send Test Email
                  </button>
                </div>
              </div>
            </div>

            <!-- Security: password change -->
            <template v-if="section.id === 'security'">
              <Transition name="fade">
                <div v-if="pwError" class="alert alert-error pw-alert">{{ pwError }}</div>
              </Transition>
              <Transition name="fade">
                <div v-if="pwSuccess" class="alert alert-success pw-alert">Password changed successfully.</div>
              </Transition>
              <div class="field-row">
                <div class="field-meta">
                  <div class="field-label">Current Password</div>
                </div>
                <div class="field-control">
                  <input v-model="pwForm.oldPassword" type="password" class="form-input cfg-input" placeholder="••••••••" />
                </div>
              </div>
              <div class="field-row">
                <div class="field-meta">
                  <div class="field-label">New Password</div>
                </div>
                <div class="field-control">
                  <input v-model="pwForm.newPassword" type="password" class="form-input cfg-input" placeholder="••••••••" />
                </div>
              </div>
              <div class="field-row">
                <div class="field-meta">
                  <div class="field-label">Confirm Password</div>
                </div>
                <div class="field-control">
                  <input v-model="pwForm.confirm" type="password" class="form-input cfg-input" placeholder="••••••••" @keyup.enter="changePassword" />
                </div>
              </div>
              <div class="field-row field-row-last field-row-action">
                <div class="field-meta"></div>
                <div class="field-control">
                  <button class="btn btn-primary btn-sm" @click="changePassword" :disabled="pwLoading">
                    <span v-if="pwLoading" class="spinner" style="width:12px;height:12px;border-color:rgba(255,255,255,0.3);border-top-color:white" />
                    Update Password
                  </button>
                </div>
              </div>
            </template>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* -- Layout -- */
.config-page {
  position: relative;
}

.page-header-left {
  display: flex;
  flex-direction: column;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.config-body {
  display: grid;
  grid-template-columns: 188px 1fr;
  gap: 36px;
  align-items: start;
}

/* -- Left nav -- */
.settings-nav {
  position: sticky;
  top: 24px;
  display: flex;
  flex-direction: column;
  gap: 1px;
  background: white;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
  padding: 6px;
}

.settings-nav-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  text-align: left;
  padding: 9px 12px;
  font-size: 13px;
  font-weight: 450;
  color: var(--text-secondary);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.15s;
  background: none;
  border: none;
  overflow: hidden;
}

.settings-nav-item:hover {
  color: var(--text-primary);
  background: var(--cream);
}

.settings-nav-item.active {
  color: var(--accent);
  background: var(--accent-light);
  font-weight: 500;
}

.nav-icon {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  opacity: 0.6;
  transition: opacity 0.15s;
}

.settings-nav-item.active .nav-icon,
.settings-nav-item:hover .nav-icon {
  opacity: 1;
}

.nav-label {
  flex: 1;
}

.nav-active-bar {
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%) scaleY(0);
  width: 3px;
  height: 60%;
  background: var(--accent);
  border-radius: 0 2px 2px 0;
  transition: transform 0.2s cubic-bezier(.34,1.56,.64,1);
}

.settings-nav-item.active .nav-active-bar {
  transform: translateY(-50%) scaleY(1);
}

/* -- Sections -- */
.settings-content {
  display: flex;
  flex-direction: column;
  gap: 44px;
}

.settings-section {
  scroll-margin-top: 16px;
}

.section-header {
  margin-bottom: 14px;
}

.section-header-inner {
  display: flex;
  align-items: flex-start;
  gap: 14px;
}

.section-index {
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  color: var(--accent);
  background: var(--accent-light);
  border: 1px solid rgba(201, 100, 66, 0.2);
  border-radius: var(--radius-sm);
  padding: 3px 7px;
  letter-spacing: 0.04em;
  flex-shrink: 0;
  margin-top: 3px;
}

.section-title {
  font-family: var(--font-serif);
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: 0.01em;
  line-height: 1.2;
}

.section-desc {
  font-size: 13px;
  color: var(--text-muted);
  margin-top: 5px;
  line-height: 1.5;
}

/* -- Panel -- */
.settings-panel {
  background: white;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}

/* -- Field rows -- */
.field-row {
  display: grid;
  grid-template-columns: 210px 1fr;
  gap: 20px;
  padding: 15px 20px;
  border-bottom: 1px solid var(--border-light);
  align-items: center;
  transition: background 0.12s;
}

.field-row:hover {
  background: var(--cream);
}

.field-row:first-child {
  border-top: none;
}

.field-row-last {
  border-bottom: none;
}

.field-meta {
  flex-shrink: 0;
}

.field-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
}

.field-hint {
  font-size: 11.5px;
  color: var(--text-muted);
  margin-top: 3px;
  line-height: 1.45;
}

.field-control {
  display: flex;
  align-items: center;
}

/* -- Inputs -- */
.cfg-input {
  max-width: 300px;
}

.cfg-textarea {
  min-height: 72px;
  font-family: var(--font-mono);
  font-size: 12.5px;
  line-height: 1.6;
}


/* -- Toggle -- */
.toggle {
  display: inline-flex;
  align-items: center;
  cursor: pointer;
}

.toggle-input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-track {
  position: relative;
  width: 38px;
  height: 22px;
  background: var(--cream-border);
  border-radius: 11px;
  transition: background 0.2s;
  display: block;
  border: 1px solid var(--border);
}

.toggle-input:checked + .toggle-track {
  background: var(--accent);
  border-color: var(--accent);
}

.toggle-thumb {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 14px;
  height: 14px;
  background: white;
  border-radius: 50%;
  transition: transform 0.2s cubic-bezier(.34,1.56,.64,1);
  box-shadow: 0 1px 3px rgba(0,0,0,0.18);
  display: block;
}

.toggle-input:checked + .toggle-track .toggle-thumb {
  transform: translateX(16px);
}

/* -- Spinner -- */
.spinner-sm {
  width: 12px;
  height: 12px;
}

/* -- Password alerts -- */
.pw-alert {
  margin: 12px 20px 0;
  border-radius: var(--radius-sm);
}

@media (max-width: 768px) {
  .config-body {
    grid-template-columns: 1fr;
    gap: 20px;
  }

  .settings-nav {
    position: static;
    flex-direction: row;
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
    padding: 4px;
    gap: 0;
  }

  .settings-nav-item {
    white-space: nowrap;
    flex-shrink: 0;
  }

  .nav-active-bar {
    display: none;
  }

  .field-row {
    grid-template-columns: 1fr;
    gap: 8px;
  }
}
</style>
