<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  CircleAlert,
  Pencil,
  Plus,
  RefreshCw,
  Search,
  Server as ServerIcon,
  Terminal,
  Trash2,
  X
} from '@lucide/vue'
import { del, get, post, put } from '@/api/client'
import { useToast } from '@/composables/useToast'
import BaseModal from '@/components/BaseModal.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'
import PageHeader from '@/components/PageHeader.vue'
import StatePanel from '@/components/StatePanel.vue'
import StatusBadge from '@/components/StatusBadge.vue'

const router = useRouter()
const toast = useToast()
const servers = ref([])
const loading = ref(true)
const loadError = ref('')
const searchQuery = ref('')
const connectionStates = ref({})
const checkingServerIds = ref([])
const statusRequestTokens = new Map()
let statusRequestSequence = 0

const showModal = ref(false)
const editTarget = ref(null)
const formLoading = ref(false)
const testLoading = ref(false)

const confirmDelete = ref(null)
const deleteLoading = ref(false)

const defaultForm = () => ({
  host: '',
  hostname: '',
  port: 22,
  user: '',
  auth_type: 'password',
  credential: '',
  description: ''
})

const form = ref(defaultForm())

const filteredServers = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  if (!query) return servers.value
  return servers.value.filter(server =>
    [server.host, server.hostname, server.user, server.description]
      .some(value => String(value || '').toLowerCase().includes(query))
  )
})

async function loadServers() {
  loading.value = true
  loadError.value = ''
  statusRequestTokens.clear()
  checkingServerIds.value = []
  try {
    servers.value = await get('/servers') || []
    connectionStates.value = Object.fromEntries(servers.value.map(server => [server.id, 'offline']))
    refreshAllServerStatuses()
  } catch (error) {
    servers.value = []
    connectionStates.value = {}
    loadError.value = error.message
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editTarget.value = null
  form.value = defaultForm()
  showModal.value = true
}

function openEdit(server) {
  editTarget.value = server
  form.value = {
    host: server.host,
    hostname: server.hostname,
    port: server.port,
    user: server.user,
    auth_type: server.auth_type,
    credential: '',
    description: server.description || ''
  }
  showModal.value = true
}

function closeServerModal() {
  if (formLoading.value || testLoading.value) return
  showModal.value = false
}

function validateForm({ requireCredential = false } = {}) {
  if (!form.value.host.trim()) return '请输入服务器名称。'
  if (!form.value.hostname.trim()) return '请输入主机地址或 IP。'
  if (!form.value.user.trim()) return '请输入 SSH 用户名。'
  if (!Number(form.value.port) || Number(form.value.port) < 1 || Number(form.value.port) > 65535) return 'SSH 端口必须在 1 到 65535 之间。'
  if (requireCredential && !form.value.credential) return '请输入密码或私钥。'
  return ''
}

async function saveServer() {
  const validationError = validateForm({ requireCredential: !editTarget.value })
  if (validationError) {
    toast.warning(validationError)
    return
  }

  formLoading.value = true
  try {
    if (editTarget.value) {
      await put(`/servers/${editTarget.value.id}`, form.value)
      toast.success(`已保存服务器“${form.value.host}”`)
    } else {
      await post('/servers', form.value)
      toast.success(`已添加服务器“${form.value.host}”`)
    }
    showModal.value = false
    await loadServers()
  } catch (error) {
    toast.error(`无法保存服务器：${error.message}`)
  } finally {
    formLoading.value = false
  }
}

async function testConnection() {
  const validationError = validateForm({ requireCredential: !editTarget.value })
  if (validationError) {
    toast.warning(validationError)
    return
  }

  testLoading.value = true
  try {
    if (form.value.credential) {
      await post('/servers/test-direct', {
        hostname: form.value.hostname,
        port: form.value.port || 22,
        user: form.value.user,
        auth_type: form.value.auth_type,
        credential: form.value.credential
      })
    } else {
      await post(`/servers/${editTarget.value.id}/test`)
    }
    toast.success('连接成功，SSH 凭据可用')
  } catch (error) {
    toast.error(`连接失败：${error.message}`)
  } finally {
    testLoading.value = false
  }
}

function setServerChecking(serverId, checking) {
  checkingServerIds.value = checking
    ? [...new Set([...checkingServerIds.value, serverId])]
    : checkingServerIds.value.filter(id => id !== serverId)
}

async function refreshServerStatus(server, { notify = true } = {}) {
  const requestToken = ++statusRequestSequence
  statusRequestTokens.set(server.id, requestToken)
  setServerChecking(server.id, true)
  try {
    await post(`/servers/${server.id}/test`)
    if (statusRequestTokens.get(server.id) !== requestToken) return
    connectionStates.value = { ...connectionStates.value, [server.id]: 'online' }
    if (notify) toast.success(`“${server.host}”当前在线`)
  } catch (error) {
    if (statusRequestTokens.get(server.id) !== requestToken) return
    connectionStates.value = { ...connectionStates.value, [server.id]: 'offline' }
    if (notify) toast.error(`“${server.host}”当前离线：${error.message}`)
  } finally {
    if (statusRequestTokens.get(server.id) === requestToken) {
      statusRequestTokens.delete(server.id)
      setServerChecking(server.id, false)
    }
  }
}

function refreshAllServerStatuses() {
  return Promise.all(servers.value.map(server => refreshServerStatus(server, { notify: false })))
}

async function deleteServer() {
  if (!confirmDelete.value) return
  deleteLoading.value = true
  const serverName = confirmDelete.value.host
  try {
    await del(`/servers/${confirmDelete.value.id}`)
    confirmDelete.value = null
    toast.success(`已删除服务器“${serverName}”`)
    await loadServers()
  } catch (error) {
    toast.error(`无法删除服务器：${error.message}`)
  } finally {
    deleteLoading.value = false
  }
}

function openTerminal(serverId) {
  window.open(`/terminal/${serverId}`, '_blank', 'noopener')
}

function formatDate(value) {
  if (!value) return '—'
  return new Intl.DateTimeFormat('zh-CN', { year: 'numeric', month: 'short', day: 'numeric' }).format(new Date(value))
}

onMounted(loadServers)
</script>

<template>
  <div>
    <PageHeader title="服务器" description="管理用于 SSH 连接和 Docker 工作负载的远程主机。">
      <template #actions>
        <button class="btn btn-primary" type="button" @click="openCreate">
          <Plus :size="16" aria-hidden="true" />
          添加服务器
        </button>
      </template>
    </PageHeader>

    <section class="data-panel responsive-table" aria-label="服务器列表">
      <div class="table-toolbar">
        <div class="search-field">
          <Search :size="16" aria-hidden="true" />
          <label class="sr-only" for="server-search">搜索服务器</label>
          <input id="server-search" v-model="searchQuery" class="form-input" placeholder="搜索名称、地址或用户" />
          <button v-if="searchQuery" class="search-clear" type="button" aria-label="清除搜索" @click="searchQuery = ''">
            <X :size="14" aria-hidden="true" />
          </button>
        </div>
        <div class="toolbar-summary">
          <span class="result-count">{{ filteredServers.length }} 台服务器</span>
          <button class="btn btn-ghost btn-sm" type="button" :disabled="checkingServerIds.length > 0" @click="refreshAllServerStatuses">
            <RefreshCw :size="14" :class="{ spinning: checkingServerIds.length > 0 }" aria-hidden="true" />
            刷新状态
          </button>
        </div>
      </div>

      <div v-if="loading" class="loading-state" role="status">
        <span class="spinner" aria-hidden="true" />
        <span>正在读取服务器…</span>
      </div>

      <StatePanel
        v-else-if="loadError"
        tone="error"
        title="无法读取服务器"
        :description="`${loadError}。请检查 API 服务后重试。`"
      >
        <template #icon><CircleAlert :size="20" /></template>
        <template #actions>
          <button class="btn btn-secondary btn-sm" type="button" @click="loadServers">
            <RefreshCw :size="14" aria-hidden="true" />重新加载
          </button>
        </template>
      </StatePanel>

      <StatePanel
        v-else-if="servers.length === 0"
        title="尚未添加服务器"
        description="添加第一台服务器后，可以管理容器、镜像、数据卷并打开 Web Terminal。"
      >
        <template #icon><ServerIcon :size="20" /></template>
        <template #actions><button class="btn btn-primary btn-sm" type="button" @click="openCreate">添加服务器</button></template>
      </StatePanel>

      <StatePanel
        v-else-if="filteredServers.length === 0"
        title="没有匹配的服务器"
        description="尝试使用服务器名称、主机地址或 SSH 用户搜索。"
      >
        <template #icon><Search :size="20" /></template>
        <template #actions><button class="btn btn-secondary btn-sm" type="button" @click="searchQuery = ''">清除搜索</button></template>
      </StatePanel>

      <div v-else class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>服务器</th>
              <th>IP 地址</th>
              <th>认证</th>
              <th>连接状态</th>
              <th>添加时间</th>
              <th><span class="sr-only">操作</span></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="server in filteredServers" :key="server.id">
              <td data-label="服务器">
                <button
                  class="server-link"
                  type="button"
                  @click="router.push({ path: '/containers', query: { server: server.id } })"
                >
                  {{ server.host }}
                </button>
                <div class="table-secondary">{{ server.description || '未填写说明' }}</div>
              </td>
              <td data-label="IP 地址">
                <div class="mono address">{{ server.hostname }}:{{ server.port }}</div>
                <div class="table-secondary mono">{{ server.user }}</div>
              </td>
              <td data-label="认证">{{ server.auth_type === 'key' ? 'SSH 私钥' : '密码' }}</td>
              <td data-label="连接状态">
                <StatusBadge :status="connectionStates[server.id] || 'offline'" />
              </td>
              <td data-label="添加时间" class="date-cell">{{ formatDate(server.created_at) }}</td>
              <td data-label="">
                <div class="row-actions">
                  <button class="btn btn-ghost btn-sm" type="button" :disabled="checkingServerIds.includes(server.id)" @click="refreshServerStatus(server)">
                    <RefreshCw :size="14" :class="{ spinning: checkingServerIds.includes(server.id) }" aria-hidden="true" />
                    刷新状态
                  </button>
                  <button class="btn btn-ghost btn-sm btn-icon" type="button" title="打开终端" :aria-label="`打开 ${server.host} 的终端`" @click="openTerminal(server.id)">
                    <Terminal :size="15" aria-hidden="true" />
                  </button>
                  <button class="btn btn-ghost btn-sm btn-icon" type="button" title="编辑服务器" :aria-label="`编辑 ${server.host}`" @click="openEdit(server)">
                    <Pencil :size="15" aria-hidden="true" />
                  </button>
                  <button class="btn btn-danger btn-sm btn-icon" type="button" title="删除服务器" :aria-label="`删除 ${server.host}`" @click="confirmDelete = server">
                    <Trash2 :size="15" aria-hidden="true" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <BaseModal
      v-if="showModal"
      :title="editTarget ? '编辑服务器' : '添加服务器'"
      size="md"
      :close-on-backdrop="!formLoading && !testLoading"
      @close="closeServerModal"
    >
      <div class="modal-form">
        <div class="form-grid">
          <div class="form-group">
            <label class="form-label" for="server-name">服务器名称 <span class="required-mark">*</span></label>
            <input id="server-name" v-model="form.host" class="form-input" placeholder="例如：GPU 节点 01" autocomplete="off" required />
          </div>
          <div class="form-group">
            <label class="form-label" for="server-hostname">主机地址或 IP <span class="required-mark">*</span></label>
            <input id="server-hostname" v-model="form.hostname" class="form-input mono" placeholder="192.168.1.100" autocomplete="off" required />
          </div>
          <div class="form-group">
            <label class="form-label" for="server-port">SSH 端口</label>
            <input id="server-port" v-model.number="form.port" type="number" min="1" max="65535" class="form-input mono" />
          </div>
          <div class="form-group">
            <label class="form-label" for="server-user">SSH 用户名 <span class="required-mark">*</span></label>
            <input id="server-user" v-model="form.user" class="form-input mono" placeholder="ubuntu" autocomplete="username" required />
          </div>
        </div>

        <div class="form-group">
          <label class="form-label" for="server-auth">认证方式</label>
          <select id="server-auth" v-model="form.auth_type" class="form-select" required>
            <option value="password">密码</option>
            <option value="key">SSH 私钥</option>
          </select>
        </div>

        <div class="form-group">
          <label class="form-label" for="server-credential">
            {{ form.auth_type === 'key' ? '私钥' : '密码' }}
            <span v-if="!editTarget" class="required-mark">*</span>
          </label>
          <textarea
            v-if="form.auth_type === 'key'"
            id="server-credential"
            v-model="form.credential"
            class="form-textarea mono credential-input"
            placeholder="-----BEGIN OPENSSH PRIVATE KEY-----"
            autocomplete="off"
            :required="!editTarget"
          />
          <input
            v-else
            id="server-credential"
            v-model="form.credential"
            type="password"
            class="form-input"
            placeholder="输入 SSH 密码"
            autocomplete="new-password"
            :required="!editTarget"
          />
          <span v-if="editTarget" class="form-hint">留空将继续使用已保存的凭据。</span>
        </div>

        <div class="form-group">
          <label class="form-label" for="server-description">说明</label>
          <input id="server-description" v-model="form.description" class="form-input" placeholder="例如：实验室公共 GPU 节点" />
        </div>
      </div>

      <template #footer>
        <button class="btn btn-ghost test-button" type="button" :disabled="formLoading || testLoading" @click="testConnection">
          <span v-if="testLoading" class="spinner" aria-hidden="true" />
          <RefreshCw v-else :size="14" aria-hidden="true" />
          {{ testLoading ? '正在测试…' : '测试连接' }}
        </button>
        <span class="modal-spacer" />
        <button class="btn btn-secondary" type="button" :disabled="formLoading || testLoading" @click="closeServerModal">取消</button>
        <button class="btn btn-primary" type="button" :disabled="formLoading || testLoading" @click="saveServer">
          <span v-if="formLoading" class="spinner button-spinner" aria-hidden="true" />
          {{ formLoading ? '正在保存…' : editTarget ? '保存设置' : '添加服务器' }}
        </button>
      </template>
    </BaseModal>

    <ConfirmModal
      v-if="confirmDelete"
      title="删除服务器"
      :message="`确定删除服务器“${confirmDelete.host}”吗？`"
      detail="ServerDock 将删除保存的连接信息；远程主机上的容器和数据不会被删除，但此操作无法撤销。"
      confirm-text="删除服务器"
      :loading="deleteLoading"
      @confirm="deleteServer"
      @cancel="confirmDelete = null"
    />
  </div>
</template>

<style scoped>
.result-count {
  color: var(--ink-secondary);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
}

.toolbar-summary {
  display: flex;
  align-items: center;
  gap: 8px;
}

.server-link {
  padding: 0;
  background: transparent;
  color: #0066cc;
  cursor: pointer;
  font-weight: 650;
  text-align: left;
}

.server-link:hover {
  text-decoration: underline;
  text-underline-offset: 3px;
}

.address {
  font-size: 12px;
}

.date-cell {
  color: var(--ink-secondary);
  font-size: 12px;
  white-space: nowrap;
}

.spinning {
  animation: spin 700ms linear infinite;
}

.modal-form {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

.credential-input {
  min-height: 118px;
  font-size: 12px;
}

.modal-spacer {
  flex: 1;
}

.button-spinner {
  width: 13px;
  height: 13px;
  color: #fff;
}

@media (max-width: 680px) {
  .toolbar-summary {
    justify-content: space-between;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .test-button {
    margin-right: auto;
  }

  .modal-spacer {
    display: none;
  }
}
</style>
