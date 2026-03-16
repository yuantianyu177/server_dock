<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { serversApi } from '@/api/servers'
import { useToast } from '@/composables/useToast'
import BaseModal from '@/components/BaseModal.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'

const router = useRouter()
const toast = useToast()
const servers = ref([])
const loading = ref(true)

// Modal state
const showModal = ref(false)
const editTarget = ref(null)
const formLoading = ref(false)
const formError = ref('')
const testResult = ref(null)

// Confirm delete
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

async function loadServers() {
  loading.value = true
  try {
    servers.value = await serversApi.list() || []
  } catch (e) {
    servers.value = []
    toast.error(e.message)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editTarget.value = null
  form.value = defaultForm()
  formError.value = ''
  testResult.value = null
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
  formError.value = ''
  testResult.value = null
  showModal.value = true
}

async function saveServer() {
  formLoading.value = true
  formError.value = ''
  try {
    if (editTarget.value) {
      await serversApi.update(editTarget.value.id, form.value)
      toast.success('Server updated successfully')
    } else {
      await serversApi.create(form.value)
      toast.success('Server added successfully')
    }
    showModal.value = false
    loadServers()
  } catch (e) {
    formError.value = e.message
  } finally {
    formLoading.value = false
  }
}

async function testConnection() {
  const f = form.value
  if (!f.hostname || !f.user || !f.auth_type) {
    testResult.value = { ok: false, msg: 'Hostname, user, and auth type are required.' }
    return
  }
  if (!f.credential && !editTarget.value) {
    testResult.value = { ok: false, msg: 'Credential is required.' }
    return
  }
  formLoading.value = true
  testResult.value = null
  try {
    if (f.credential) {
      // Use form data directly
      await serversApi.testConnectionDirect({
        hostname: f.hostname,
        port: f.port || 22,
        user: f.user,
        auth_type: f.auth_type,
        credential: f.credential
      })
    } else {
      // Editing without new credential, test with saved credential
      await serversApi.testConnection(editTarget.value.id)
    }
    testResult.value = { ok: true, msg: 'Connection successful.' }
  } catch (e) {
    testResult.value = { ok: false, msg: e.message }
  } finally {
    formLoading.value = false
  }
}

async function deleteServer() {
  deleteLoading.value = true
  try {
    await serversApi.delete(confirmDelete.value.id)
    confirmDelete.value = null
    toast.success('Server deleted')
    loadServers()
  } catch (e) {
    toast.error(e.message)
  } finally {
    deleteLoading.value = false
  }
}

function openTerminal(serverId) {
  window.open(`/terminal/${serverId}`, '_blank')
}

function formatDate(str) {
  if (!str) return '-'
  return new Date(str).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
}

onMounted(loadServers)
</script>

<template>
  <div>
    <div class="page-header animate-in">
      <div>
        <h1 class="page-title">Servers</h1>
        <p class="page-subtitle">Manage remote servers connected via SSH</p>
      </div>
      <button class="btn btn-primary btn-w-action" @click="openCreate">
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
        </svg>
        Add Server
      </button>
    </div>

    <div class="card animate-in animate-in-delay-1">
      <div v-if="loading" class="loading-overlay">
        <span class="spinner" style="width:24px;height:24px" />
      </div>

      <div v-else-if="servers.length === 0" class="empty-state">
        <div class="empty-state-icon">🖥️</div>
        <div class="empty-state-text">No servers configured.</div>
      </div>

      <div v-else class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Hostname</th>
              <th>Port</th>
              <th>User</th>
              <th>Auth</th>
              <th>Added</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="srv in servers" :key="srv.id">
              <td>
                <button class="server-name-btn" @click="router.push(`/servers/${srv.id}`)">
                  {{ srv.host }}
                </button>
              </td>
              <td><span class="mono">{{ srv.hostname }}</span></td>
              <td><span class="mono">{{ srv.port }}</span></td>
              <td><span class="mono">{{ srv.user }}</span></td>
              <td>
                <span class="badge badge-default">{{ srv.auth_type }}</span>
              </td>
              <td style="color:var(--text-muted)">{{ formatDate(srv.created_at) }}</td>
              <td>
                <div class="row-actions">
                  <button class="btn btn-ghost btn-sm btn-icon" @click="openTerminal(srv.id)" title="Terminal">
                    <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <polyline points="4 17 10 11 4 5"/>
                      <line x1="12" y1="19" x2="20" y2="19"/>
                    </svg>
                  </button>
                  <button class="btn btn-ghost btn-sm" @click="openEdit(srv)">Edit</button>
                  <button class="btn btn-danger btn-sm" @click="confirmDelete = srv">Delete</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create / Edit Modal -->
    <BaseModal
      v-if="showModal"
      :title="editTarget ? 'Edit Server' : 'Add Server'"
      size="md"
      @close="showModal = false"
    >
      <div class="modal-form">
        <div v-if="formError" class="alert alert-error" style="margin-bottom:16px">{{ formError }}</div>
        <div v-if="testResult" class="alert" :class="testResult.ok ? 'alert-success' : 'alert-error'" style="margin-bottom:16px">
          {{ testResult.msg }}
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Display Name *</label>
            <input v-model="form.host" class="form-input" placeholder="GPU Server 01" />
          </div>
          <div class="form-group">
            <label class="form-label">Hostname / IP *</label>
            <input v-model="form.hostname" class="form-input" placeholder="192.168.1.100" />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label">SSH Port</label>
            <input v-model.number="form.port" type="number" class="form-input" />
          </div>
          <div class="form-group">
            <label class="form-label">Username *</label>
            <input v-model="form.user" class="form-input" placeholder="ubuntu" />
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">Auth Type</label>
          <select v-model="form.auth_type" class="form-select">
            <option value="password">Password</option>
            <option value="key">SSH Key</option>
          </select>
        </div>

        <div class="form-group">
          <label class="form-label">
            {{ form.auth_type === 'key' ? 'Private Key' : 'Password' }}
            {{ editTarget ? '(leave blank to keep unchanged)' : '*' }}
          </label>
          <textarea
            v-if="form.auth_type === 'key'"
            v-model="form.credential"
            class="form-textarea"
            placeholder="-----BEGIN OPENSSH PRIVATE KEY-----"
            style="min-height:100px;font-family:var(--font-mono);font-size:12px"
          />
          <input
            v-else
            v-model="form.credential"
            type="password"
            class="form-input"
            placeholder="••••••••"
          />
        </div>

        <div class="form-group">
          <label class="form-label">Description</label>
          <input v-model="form.description" class="form-input" placeholder="Optional notes" />
        </div>
      </div>

      <template #footer>
        <button class="btn btn-ghost" @click="testConnection" :disabled="formLoading">
          Test Connection
        </button>
        <div style="flex:1" />
        <button class="btn btn-secondary" @click="showModal = false" :disabled="formLoading">Cancel</button>
        <button class="btn btn-primary" @click="saveServer" :disabled="formLoading">
          <span v-if="formLoading" class="spinner" style="width:13px;height:13px;border-color:rgba(255,255,255,0.3);border-top-color:white" />
          {{ editTarget ? 'Save Changes' : 'Add Server' }}
        </button>
      </template>
    </BaseModal>

    <!-- Confirm Delete -->
    <ConfirmModal
      v-if="confirmDelete"
      title="Delete Server"
      :message="`Are you sure you want to delete '${confirmDelete.host}'? This action cannot be undone.`"
      confirm-text="Delete"
      confirm-class="btn-danger"
      :loading="deleteLoading"
      @confirm="deleteServer"
      @cancel="confirmDelete = null"
    />
  </div>
</template>

<style scoped>
.server-name-btn {
  font-weight: 600;
  color: var(--accent);
  font-size: 13.5px;
  cursor: pointer;
  background: none;
  border: none;
  padding: 0;
  text-decoration: none;
  transition: opacity 0.15s;
}
.server-name-btn:hover { opacity: 0.7; }

.mono {
  font-family: var(--font-mono);
  font-size: 12.5px;
}

.row-actions {
  display: flex;
  gap: 6px;
  justify-content: flex-end;
}

.row-actions .btn {
  border: none;
  box-shadow: none;
}

.modal-form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

@media (max-width: 768px) {
  .form-row {
    grid-template-columns: 1fr;
  }

  .row-actions {
    flex-wrap: wrap;
  }
}
</style>
