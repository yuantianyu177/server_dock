<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { serversApi } from '@/api/servers'
import { imagesApi } from '@/api/images'
import { containersApi } from '@/api/containers'
import { useToast } from '@/composables/useToast'
import StatusBadge from '@/components/StatusBadge.vue'
import BaseModal from '@/components/BaseModal.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'
import ServerSelect from '@/components/ServerSelect.vue'
import { formatIPv4Ports } from '@/utils/docker'

const route = useRoute()
const router = useRouter()
const toast = useToast()

const servers = ref([])
const selectedServerId = ref(null)
const containers = ref([])
const serversLoading = ref(true)
const containersLoading = ref(false)

const logsModal = ref(null)
const logsContent = ref('')
const logsLoading = ref(false)

const availableImages = ref([])
const availableImagesLoading = ref(false)
const createModal = ref(false)
const containerForm = ref({ name: '', image: '' })
const formLoading = ref(false)
const formError = ref('')

const confirmAction = ref(null)
const actionLoading = ref(false)

async function loadServers() {
  serversLoading.value = true
  try {
    servers.value = await serversApi.list() || []
    const queryId = route.query.server ? Number(route.query.server) : null
    if (queryId && servers.value.find(s => s.id === queryId)) {
      selectedServerId.value = queryId
    } else if (servers.value.length > 0) {
      selectedServerId.value = servers.value[0].id
    }
  } catch (e) {
    toast.error(e.message)
  } finally {
    serversLoading.value = false
  }
}

async function loadContainers() {
  if (!selectedServerId.value) return
  containersLoading.value = true
  try {
    containers.value = await containersApi.list(selectedServerId.value) || []
  } catch (e) {
    toast.error(e.message)
  } finally {
    containersLoading.value = false
  }
}

async function loadAvailableImages() {
  if (!selectedServerId.value) {
    availableImages.value = []
    return
  }
  availableImagesLoading.value = true
  try {
    availableImages.value = await imagesApi.list(selectedServerId.value) || []
  } catch (e) {
    availableImages.value = []
    formError.value = e.message
  } finally {
    availableImagesLoading.value = false
  }
}

async function openCreateModal() {
  formError.value = ''
  containerForm.value = { name: '', image: '' }
  await loadAvailableImages()
  createModal.value = true
}

async function containerAction(container, action) {
  if (['delete', 'stop'].includes(action)) {
    confirmAction.value = { container, action }
    return
  }
  await doContainerAction(container, action)
}

async function doContainerAction(container, action) {
  actionLoading.value = true
  try {
    await containersApi.action(selectedServerId.value, container.name, action)
    toast.success(`Container ${action}ed successfully`)
    loadContainers()
  } catch (e) {
    toast.error(e.message)
  } finally {
    actionLoading.value = false
    confirmAction.value = null
  }
}

async function showLogs(container) {
  logsModal.value = container
  logsContent.value = ''
  logsLoading.value = true
  try {
    const res = await containersApi.logs(selectedServerId.value, container.name)
    logsContent.value = res?.logs || res || '(no logs)'
  } catch (e) {
    logsContent.value = e.message
  } finally {
    logsLoading.value = false
  }
}

async function createContainer() {
  formLoading.value = true
  formError.value = ''
  try {
    await containersApi.create(selectedServerId.value, containerForm.value)
    createModal.value = false
    containerForm.value = { name: '', image: '' }
    toast.success('Container created successfully')
    loadContainers()
  } catch (e) {
    formError.value = e.message
  } finally {
    formLoading.value = false
  }
}

function openTerminal(containerName) {
  const query = containerName ? `?container=${containerName}` : ''
  window.open(`/terminal/${selectedServerId.value}${query}`, '_blank')
}

watch(selectedServerId, (id) => {
  if (id) router.replace({ query: { server: id } })
  loadContainers()
  loadAvailableImages()
})
onMounted(async () => {
  await loadServers()
  if (selectedServerId.value) {
    loadContainers()
    loadAvailableImages()
  }
})
</script>

<template>
  <div>
    <div class="page-header animate-in">
      <div>
        <h1 class="page-title">Containers</h1>
        <p class="page-subtitle">Manage Docker containers across all servers</p>
      </div>
      <div class="header-actions">
        <ServerSelect
          v-if="servers.length > 0"
          v-model="selectedServerId"
          :servers="servers"
        />
        <button class="btn btn-primary btn-w-action" @click="openCreateModal" :disabled="!selectedServerId">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          Add Container
        </button>
      </div>
    </div>

    <div class="card animate-in animate-in-delay-1">
      <div v-if="!serversLoading && servers.length === 0" class="empty-state">
        <div class="empty-state-icon">🖥️</div>
        <div class="empty-state-text">No servers configured. Add a server first.</div>
      </div>
      <template v-else-if="!serversLoading">
        <div v-if="containersLoading" class="loading-overlay">
          <span class="spinner" style="width:24px;height:24px" />
        </div>
        <div v-else-if="containers.length === 0" class="empty-state">
          <div class="empty-state-icon">📦</div>
          <div class="empty-state-text">No containers on this server.</div>
        </div>
        <div v-else class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>Image</th>
                <th>Status</th>
                <th>Ports</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="c in containers" :key="c.id || c.name">
                <td><span class="mono" style="font-weight:500">{{ c.name }}</span></td>
                <td><span class="mono" style="font-size:12px;color:var(--text-secondary)">{{ c.image }}</span></td>
                <td><StatusBadge :status="c.status" /></td>
                <td><span class="mono" style="font-size:11.5px;color:var(--text-muted)">{{ formatIPv4Ports(c.ports) }}</span></td>
                <td>
                  <div class="row-actions">
                    <button class="btn btn-ghost btn-sm btn-icon" @click="openTerminal(c.name)" title="Terminal">
                      <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg>
                    </button>
                    <button class="btn btn-ghost btn-sm" @click="showLogs(c)">Logs</button>
                    <button v-if="!c.status?.toLowerCase().startsWith('up')" class="btn btn-success btn-sm" @click="containerAction(c, 'start')">Start</button>
                    <button v-else class="btn btn-warning btn-sm" @click="containerAction(c, 'stop')">Stop</button>
                    <button class="btn btn-ghost btn-sm" @click="containerAction(c, 'restart')">Restart</button>
                    <button class="btn btn-danger btn-sm" @click="containerAction(c, 'delete')">Delete</button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>

    <!-- Create Container Modal -->
    <BaseModal v-if="createModal" title="Create Container" size="md" @close="createModal = false">
      <div style="display:flex;flex-direction:column;gap:14px">
        <div v-if="formError" class="alert alert-error">{{ formError }}</div>
        <div class="form-group">
          <label class="form-label">Container Name *</label>
          <input v-model="containerForm.name" class="form-input" placeholder="my-container" />
          <span class="form-hint">Alphanumeric, hyphens and underscores only</span>
        </div>
        <div class="form-group">
          <label class="form-label">Image *</label>
          <select v-model="containerForm.image" class="form-select" :disabled="availableImagesLoading || availableImages.length === 0">
            <option value="" disabled>
              {{ availableImagesLoading ? 'Loading images…' : availableImages.length === 0 ? 'No images available for this server' : 'Select an image…' }}
            </option>
            <option v-for="img in availableImages" :key="img.id" :value="img.image_address">
              {{ img.name }} — {{ img.image_address }}
            </option>
          </select>
          <span v-if="!availableImagesLoading && availableImages.length === 0" class="form-hint" style="color:var(--warning)">
            Add an available image for this server before creating a container.
          </span>
        </div>
      </div>
      <template #footer>
        <button class="btn btn-secondary" @click="createModal = false">Cancel</button>
        <button class="btn btn-primary" @click="createContainer" :disabled="formLoading || availableImagesLoading || availableImages.length === 0">
          <span v-if="formLoading" class="spinner" style="width:13px;height:13px;border-color:rgba(255,255,255,0.3);border-top-color:white" />
          Create
        </button>
      </template>
    </BaseModal>

    <!-- Logs Modal -->
    <BaseModal v-if="logsModal" :title="`Logs — ${logsModal.name}`" size="lg" @close="logsModal = null">
      <div v-if="logsLoading" class="loading-overlay"><span class="spinner" /></div>
      <pre v-else class="logs-pre">{{ logsContent }}</pre>
    </BaseModal>

    <!-- Confirm Action -->
    <ConfirmModal
      v-if="confirmAction"
      :title="`${confirmAction.action.charAt(0).toUpperCase() + confirmAction.action.slice(1)} Container`"
      :message="`${confirmAction.action === 'delete' ? 'Permanently delete' : 'Stop'} container '${confirmAction.container.name}'?`"
      :confirm-text="confirmAction.action.charAt(0).toUpperCase() + confirmAction.action.slice(1)"
      :confirm-class="confirmAction.action === 'delete' ? 'btn-danger' : 'btn-warning'"
      :loading="actionLoading"
      @confirm="doContainerAction(confirmAction.container, confirmAction.action)"
      @cancel="confirmAction = null"
    />
  </div>
</template>

<style scoped>
.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.mono {
  font-family: var(--font-mono);
  font-size: 12.5px;
}

.row-actions {
  display: flex;
  gap: 4px;
  justify-content: flex-end;
  align-items: center;
}

.row-actions .btn {
  border: none;
  box-shadow: none;
}

.logs-pre {
  font-family: var(--font-mono);
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  color: var(--text-primary);
  background: var(--cream);
  padding: 16px;
  border-radius: var(--radius-sm);
  max-height: 500px;
  overflow-y: auto;
}

@media (max-width: 768px) {
  .header-actions {
    flex-wrap: wrap;
  }

  .row-actions {
    flex-wrap: wrap;
  }
}
</style>
