<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { serversApi } from '@/api/servers'
import { containersApi } from '@/api/containers'
import { useToast } from '@/composables/useToast'
import BaseModal from '@/components/BaseModal.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'
import ServerSelect from '@/components/ServerSelect.vue'

const route = useRoute()
const router = useRouter()
const toast = useToast()

const servers = ref([])
const selectedServerId = ref(null)
const volumes = ref([])
const serversLoading = ref(true)
const volumesLoading = ref(false)

const createModal = ref(false)
const volumeName = ref('')
const formLoading = ref(false)
const formError = ref('')

const deleteTarget = ref(null)
const deleteLoading = ref(false)

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

async function loadVolumes() {
  if (!selectedServerId.value) return
  volumesLoading.value = true
  try {
    volumes.value = await containersApi.listVolumes(selectedServerId.value) || []
  } catch (e) {
    toast.error(e.message)
  } finally {
    volumesLoading.value = false
  }
}

function openCreateModal() {
  formError.value = ''
  volumeName.value = ''
  createModal.value = true
}

async function createVolume() {
  formLoading.value = true
  formError.value = ''
  try {
    await containersApi.createVolume(selectedServerId.value, volumeName.value)
    createModal.value = false
    volumeName.value = ''
    toast.success('Volume created successfully')
    loadVolumes()
  } catch (e) {
    formError.value = e.message
  } finally {
    formLoading.value = false
  }
}

async function confirmDelete(vol) {
  deleteTarget.value = vol
}

async function doDelete() {
  deleteLoading.value = true
  try {
    await containersApi.deleteVolume(selectedServerId.value, deleteTarget.value.Name || deleteTarget.value.name)
    deleteTarget.value = null
    toast.success('Volume deleted')
    loadVolumes()
  } catch (e) {
    toast.error(e.message)
    deleteTarget.value = null
  } finally {
    deleteLoading.value = false
  }
}

watch(selectedServerId, (id) => {
  if (id) router.replace({ query: { server: id } })
  loadVolumes()
})

onMounted(async () => {
  await loadServers()
  if (selectedServerId.value) {
    loadVolumes()
  }
})
</script>

<template>
  <div>
    <div class="page-header animate-in">
      <div>
        <h1 class="page-title">Volumes</h1>
        <p class="page-subtitle">Manage Docker volumes across all servers</p>
      </div>
      <div class="header-actions">
        <ServerSelect
          v-if="servers.length > 0"
          v-model="selectedServerId"
          :servers="servers"
        />
        <button class="btn btn-primary btn-w-action" @click="openCreateModal" :disabled="!selectedServerId">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          Create Volume
        </button>
      </div>
    </div>

    <div class="card animate-in animate-in-delay-1">
      <div v-if="!serversLoading && servers.length === 0" class="empty-state">
        <div class="empty-state-icon">🖥️</div>
        <div class="empty-state-text">No servers configured. Add a server first.</div>
      </div>
      <template v-else-if="!serversLoading">
        <div v-if="volumesLoading" class="loading-overlay">
          <span class="spinner" style="width:24px;height:24px" />
        </div>
        <div v-else-if="volumes.length === 0" class="empty-state">
          <div class="empty-state-icon">💾</div>
          <div class="empty-state-text">No volumes on this server.</div>
        </div>
        <div v-else class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>Driver</th>
                <th>Mountpoint</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="v in volumes" :key="v.Name || v.name">
                <td><span class="mono" style="font-weight:500">{{ v.Name || v.name }}</span></td>
                <td><span class="mono" style="font-size:12px;color:var(--text-secondary)">{{ v.Driver || v.driver || 'local' }}</span></td>
                <td><span class="mono" style="font-size:11.5px;color:var(--text-muted)">{{ v.Mountpoint || v.mountpoint || '-' }}</span></td>
                <td>
                  <div class="row-actions">
                    <button class="btn btn-danger btn-sm" @click="confirmDelete(v)">Delete</button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>

    <!-- Create Volume Modal -->
    <BaseModal v-if="createModal" title="Create Volume" size="md" @close="createModal = false">
      <div style="display:flex;flex-direction:column;gap:14px">
        <div v-if="formError" class="alert alert-error">{{ formError }}</div>
        <div class="form-group">
          <label class="form-label">Volume Name *</label>
          <input v-model="volumeName" class="form-input" placeholder="my-volume" @keyup.enter="createVolume" />
          <span class="form-hint">Alphanumeric, hyphens and underscores only</span>
        </div>
      </div>
      <template #footer>
        <button class="btn btn-secondary" @click="createModal = false">Cancel</button>
        <button class="btn btn-primary" @click="createVolume" :disabled="formLoading || !volumeName.trim()">
          <span v-if="formLoading" class="spinner" style="width:13px;height:13px;border-color:rgba(255,255,255,0.3);border-top-color:white" />
          Create
        </button>
      </template>
    </BaseModal>

    <!-- Confirm Delete -->
    <ConfirmModal
      v-if="deleteTarget"
      title="Delete Volume"
      :message="`Permanently delete volume '${deleteTarget.Name || deleteTarget.name}'? This cannot be undone.`"
      confirm-text="Delete"
      confirm-class="btn-danger"
      :loading="deleteLoading"
      @confirm="doDelete"
      @cancel="deleteTarget = null"
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

@media (max-width: 768px) {
  .header-actions {
    flex-wrap: wrap;
  }

  .row-actions {
    flex-wrap: wrap;
  }
}
</style>
