<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { serversApi } from '@/api/servers'
import { imagesApi } from '@/api/images'
import { containersApi } from '@/api/containers'
import { useToast } from '@/composables/useToast'
import StatusBadge from '@/components/StatusBadge.vue'
import BaseModal from '@/components/BaseModal.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'
import { formatIPv4Ports } from '@/utils/docker'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const serverId = computed(() => Number(route.params.id))

const server = ref(null)
const activeTab = ref('containers')
const loading = ref(true)

// Containers
const containers = ref([])
const containersLoading = ref(false)
const logsModal = ref(null)
const logsContent = ref('')
const logsLoading = ref(false)
const createContainerModal = ref(false)
const containerForm = ref({ name: '', image: '' })
const containerFormLoading = ref(false)
const containerFormError = ref('')
const confirmAction = ref(null)
const actionLoading = ref(false)

// DB Images
const dbImages = ref([])
const dbImagesLoading = ref(false)
const imageModal = ref(false)
const imageForm = ref({ name: '', image_id: '' })
const imageFormLoading = ref(false)
const imageFormError = ref('')

// Remote Images
const remoteImages = ref([])
const remoteLoading = ref(false)
const pullModal = ref(false)
const pullForm = ref({ image: '', tag: 'latest' })
const pullLoading = ref(false)
const confirmDeleteImage = ref(null)

// Volumes
const volumes = ref([])
const volumesLoading = ref(false)
const newVolumeName = ref('')
const volumeCreateLoading = ref(false)
const confirmDeleteVolume = ref(null)

async function loadServer() {
  loading.value = true
  try {
    server.value = await serversApi.get(serverId.value)
  } catch (e) {
    server.value = null
    toast.error(e.message)
  } finally {
    loading.value = false
  }
}

async function loadContainers() {
  containersLoading.value = true
  try {
    containers.value = await containersApi.list(serverId.value)
  } catch (e) {
    toast.error(e.message)
  } finally {
    containersLoading.value = false
  }
}

async function loadDbImages() {
  dbImagesLoading.value = true
  try {
    dbImages.value = await imagesApi.list(serverId.value)
  } catch (e) { /* ignore */ } finally {
    dbImagesLoading.value = false
  }
}

async function loadRemoteImages() {
  remoteLoading.value = true
  try {
    remoteImages.value = await imagesApi.listRemote(serverId.value)
  } catch (e) { /* ignore */ } finally {
    remoteLoading.value = false
  }
}

async function loadVolumes() {
  volumesLoading.value = true
  try {
    volumes.value = await containersApi.listVolumes(serverId.value)
  } catch (e) { /* ignore */ } finally {
    volumesLoading.value = false
  }
}

function formatRemoteImageAddress(img) {
  if (!img) return ''
  return img.tag && img.tag !== '<none>' ? `${img.repository}:${img.tag}` : img.repository
}

function switchTab(tab) {
  activeTab.value = tab
  if (tab === 'containers') loadContainers()
  if (tab === 'images') { loadDbImages(); loadRemoteImages() }
  if (tab === 'volumes') loadVolumes()
}

async function openCreateContainerModal() {
  containerFormError.value = ''
  containerForm.value = { name: '', image: '' }
  await loadDbImages()
  createContainerModal.value = true
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
    await containersApi.action(serverId.value, container.name, action)
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
    const res = await containersApi.logs(serverId.value, container.name)
    logsContent.value = res?.logs || res || '(no logs)'
  } catch (e) {
    logsContent.value = e.message
  } finally {
    logsLoading.value = false
  }
}

async function createContainer() {
  containerFormLoading.value = true
  containerFormError.value = ''
  try {
    const payload = {
      name: containerForm.value.name,
      image: containerForm.value.image
    }
    await containersApi.create(serverId.value, payload)
    createContainerModal.value = false
    containerForm.value = { name: '', image: '' }
    toast.success('Container created successfully')
    loadContainers()
  } catch (e) {
    containerFormError.value = e.message
  } finally {
    containerFormLoading.value = false
  }
}

async function saveDbImage() {
  imageFormLoading.value = true
  imageFormError.value = ''
  try {
    const selectedRemoteImage = remoteImages.value.find((img) => img.image_id === imageForm.value.image_id)
    if (!selectedRemoteImage) {
      throw new Error('Please select an image from the server.')
    }
    await imagesApi.create({
      server_id: serverId.value,
      name: imageForm.value.name,
      image_id: selectedRemoteImage.image_id,
      image_address: formatRemoteImageAddress(selectedRemoteImage)
    })
    imageModal.value = false
    imageForm.value = { name: '', image_id: '' }
    toast.success('Image added successfully')
    loadDbImages()
  } catch (e) {
    imageFormError.value = e.message
  } finally {
    imageFormLoading.value = false
  }
}

async function deleteDbImage(img) {
  try {
    await imagesApi.delete(img.id)
    toast.success('Image removed')
    loadDbImages()
  } catch (e) {
    toast.error(e.message)
  }
}

async function pullImage() {
  const image = pullForm.value.image.trim()
  const tag = pullForm.value.tag.trim()
  if (!image) {
    toast.error('Image name is required.')
    return
  }
  pullLoading.value = true
  try {
    await imagesApi.pull(serverId.value, image, tag)
    pullModal.value = false
    pullForm.value = { image: '', tag: 'latest' }
    toast.success('Image pulled successfully')
    loadRemoteImages()
  } catch (e) {
    toast.error(e.message)
  } finally {
    pullLoading.value = false
  }
}

async function deleteRemoteImage() {
  try {
    await imagesApi.deleteRemote(serverId.value, confirmDeleteImage.value.image_id)
    confirmDeleteImage.value = null
    toast.success('Image deleted from server')
    loadRemoteImages()
  } catch (e) {
    toast.error(e.message)
  }
}

async function createVolume() {
  if (!newVolumeName.value.trim()) return
  volumeCreateLoading.value = true
  try {
    await containersApi.createVolume(serverId.value, newVolumeName.value.trim())
    newVolumeName.value = ''
    toast.success('Volume created successfully')
    loadVolumes()
  } catch (e) {
    toast.error(e.message)
  } finally {
    volumeCreateLoading.value = false
  }
}

async function deleteVolume() {
  try {
    await containersApi.deleteVolume(serverId.value, confirmDeleteVolume.value)
    confirmDeleteVolume.value = null
    toast.success('Volume deleted')
    loadVolumes()
  } catch (e) {
    toast.error(e.message)
  }
}

function openTerminal(containerName) {
  const query = containerName ? `?container=${containerName}` : ''
  window.open(`/terminal/${serverId.value}${query}`, '_blank')
}

function formatSize(size) {
  return size || '-'
}

onMounted(async () => {
  await loadServer()
  if (server.value) {
    loadContainers()
    loadDbImages()
  }
})
</script>

<template>
  <div>
    <!-- Breadcrumb -->
    <div class="breadcrumb animate-in">
      <button class="btn-ghost" style="padding:0;font-size:13px;color:var(--text-muted)" @click="router.push('/servers')">
        Servers
      </button>
      <span style="color:var(--text-muted)">/</span>
      <span style="font-size:13px;color:var(--text-secondary)">{{ server?.host || '…' }}</span>
    </div>

    <div v-if="loading" class="loading-overlay" style="margin-top:40px">
      <span class="spinner" style="width:28px;height:28px" />
    </div>

    <template v-else-if="server">
      <!-- Header -->
      <div class="page-header animate-in animate-in-delay-1" style="margin-top:8px">
        <div>
          <h1 class="page-title">{{ server.host }}</h1>
          <p class="page-subtitle" style="font-family:var(--font-mono);font-size:12.5px">
            {{ server.user }}@{{ server.hostname }}:{{ server.port }}
          </p>
        </div>
        <button class="btn btn-secondary" @click="openTerminal(null)">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/>
          </svg>
          Open Terminal
        </button>
      </div>

      <!-- Tabs -->
      <div class="tabs" style="margin-bottom:20px">
        <button class="tab-btn" :class="{ active: activeTab === 'containers' }" @click="switchTab('containers')">Containers</button>
        <button class="tab-btn" :class="{ active: activeTab === 'images' }" @click="switchTab('images')">Images</button>
        <button class="tab-btn" :class="{ active: activeTab === 'volumes' }" @click="switchTab('volumes')">Volumes</button>
      </div>

      <!-- Containers Tab -->
      <div v-if="activeTab === 'containers'">
        <div class="tab-toolbar">
          <span style="font-size:13px;color:var(--text-muted)">{{ containers.length }} container{{ containers.length !== 1 ? 's' : '' }}</span>
          <div style="display:flex;gap:8px">
            <button class="btn btn-primary btn-w-action" @click="openCreateContainerModal">
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
              Add Container
            </button>
          </div>
        </div>

        <div class="card">
          <div v-if="containersLoading" class="loading-overlay">
            <span class="spinner" style="width:22px;height:22px" />
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
                      <button class="btn btn-ghost btn-sm" @click="openTerminal(c.name)" title="Terminal">
                        <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg>
                      </button>
                      <button class="btn btn-ghost btn-sm" @click="showLogs(c)" title="Logs">Logs</button>
                      <button v-if="!c.status?.toLowerCase().includes('running')" class="btn btn-success btn-sm" @click="containerAction(c, 'start')">Start</button>
                      <button v-else class="btn btn-warning btn-sm" @click="containerAction(c, 'stop')">Stop</button>
                      <button class="btn btn-ghost btn-sm" @click="containerAction(c, 'restart')">Restart</button>
                      <button class="btn btn-danger btn-sm" @click="containerAction(c, 'delete')">Delete</button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Images Tab -->
      <div v-if="activeTab === 'images'">
        <!-- DB Images (Available) -->
        <div style="margin-bottom:24px">
          <div class="tab-toolbar">
            <div>
              <div style="font-size:14px;font-weight:600;color:var(--text-primary)">Available Images</div>
              <div style="font-size:12.5px;color:var(--text-muted);margin-top:2px">Images shown to users in the application form</div>
            </div>
            <button class="btn btn-primary btn-w-action image-action-btn" @click="imageModal = true">
              <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
              Add
            </button>
          </div>
          <div class="card">
            <div v-if="dbImagesLoading" class="loading-overlay"><span class="spinner" /></div>
            <div v-else-if="dbImages.length === 0" class="empty-state" style="padding:24px">
              <div class="empty-state-text">No available images configured.</div>
            </div>
            <div v-else class="table-wrap">
              <table>
                <thead><tr><th>Name</th><th>Image Address</th><th>Added</th><th></th></tr></thead>
                <tbody>
                  <tr v-for="img in dbImages" :key="img.id">
                    <td style="font-weight:500">{{ img.name }}</td>
                    <td><span class="mono" style="font-size:12px">{{ img.image_address }}</span></td>
                    <td style="color:var(--text-muted);font-size:12.5px">{{ new Date(img.created_at).toLocaleDateString() }}</td>
                    <td>
                      <div class="row-actions">
                        <button class="btn btn-danger btn-sm image-action-btn" @click="deleteDbImage(img)">Delete</button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- Remote Images -->
        <div>
          <div class="tab-toolbar">
            <div>
              <div style="font-size:14px;font-weight:600;color:var(--text-primary)">Remote Images</div>
              <div style="font-size:12.5px;color:var(--text-muted);margin-top:2px">Images present on the server (docker images)</div>
            </div>
            <div style="display:flex;gap:8px">
              <button class="btn btn-primary btn-w-action image-action-btn" @click="pullModal = true">Pull Image</button>
            </div>
          </div>
          <div class="card">
            <div v-if="remoteLoading" class="loading-overlay"><span class="spinner" /></div>
            <div v-else-if="remoteImages.length === 0" class="empty-state" style="padding:24px">
              <div class="empty-state-text">No images found on the server.</div>
            </div>
            <div v-else class="table-wrap">
              <table>
                <thead><tr><th>Repository</th><th>Tag</th><th>Image ID</th><th>Size</th><th></th></tr></thead>
                <tbody>
                  <tr v-for="img in remoteImages" :key="img.image_id">
                    <td><span class="mono" style="font-size:12px">{{ img.repository }}</span></td>
                    <td><span class="mono" style="font-size:12px">{{ img.tag }}</span></td>
                    <td><span class="mono" style="font-size:11px;color:var(--text-muted)">{{ img.image_id?.slice(7, 19) }}</span></td>
                    <td style="color:var(--text-muted);font-size:12.5px">{{ formatSize(img.size) }}</td>
                    <td>
                      <div class="row-actions">
                        <button class="btn btn-danger btn-sm image-action-btn" @click="confirmDeleteImage = img">Delete</button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      <!-- Volumes Tab -->
      <div v-if="activeTab === 'volumes'">
        <div class="tab-toolbar">
          <span style="font-size:13px;color:var(--text-muted)">{{ volumes.length }} volume{{ volumes.length !== 1 ? 's' : '' }}</span>
          <div style="display:flex;gap:8px;align-items:center">
            <input v-model="newVolumeName" class="form-input" style="width:200px;padding:6px 10px" placeholder="volume-name" />
            <button class="btn btn-primary btn-sm" @click="createVolume" :disabled="volumeCreateLoading">Create</button>
          </div>
        </div>
        <div class="card">
          <div v-if="volumesLoading" class="loading-overlay"><span class="spinner" /></div>
          <div v-else-if="volumes.length === 0" class="empty-state">
            <div class="empty-state-icon">💾</div>
            <div class="empty-state-text">No volumes on this server.</div>
          </div>
          <div v-else class="table-wrap">
            <table>
              <thead><tr><th>Name</th><th>Driver</th><th>Mountpoint</th><th></th></tr></thead>
              <tbody>
                <tr v-for="vol in volumes" :key="vol.name">
                  <td><span class="mono" style="font-weight:500">{{ vol.name }}</span></td>
                  <td><span class="mono" style="font-size:12px;color:var(--text-muted)">{{ vol.driver || 'local' }}</span></td>
                  <td><span class="mono" style="font-size:11.5px;color:var(--text-muted)">{{ vol.mountpoint }}</span></td>
                  <td><button class="btn btn-danger btn-sm" @click="confirmDeleteVolume = vol.name">Delete</button></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </template>

    <div v-else class="empty-state" style="margin-top:40px">
      <div class="empty-state-icon">🖥️</div>
      <div class="empty-state-text">Server not found.</div>
    </div>

    <!-- Create Container Modal -->
    <BaseModal v-if="createContainerModal" title="Create Container" size="md" @close="createContainerModal = false">
      <div style="display:flex;flex-direction:column;gap:14px">
        <div v-if="containerFormError" class="alert alert-error">{{ containerFormError }}</div>
        <div class="form-group">
          <label class="form-label">Container Name *</label>
          <input v-model="containerForm.name" class="form-input" placeholder="my-container" />
          <span class="form-hint">Alphanumeric, hyphens and underscores only</span>
        </div>
        <div class="form-group">
          <label class="form-label">Image *</label>
          <select v-model="containerForm.image" class="form-select" :disabled="dbImagesLoading || dbImages.length === 0">
            <option value="" disabled>
              {{ dbImagesLoading ? 'Loading images…' : dbImages.length === 0 ? 'No images available for this server' : 'Select an image…' }}
            </option>
            <option v-for="img in dbImages" :key="img.id" :value="img.image_address">
              {{ img.name }} — {{ img.image_address }}
            </option>
          </select>
          <span v-if="!dbImagesLoading && dbImages.length === 0" class="form-hint" style="color:var(--warning)">
            Add an available image for this server before creating a container.
          </span>
        </div>
      </div>
      <template #footer>
        <button class="btn btn-secondary" @click="createContainerModal = false">Cancel</button>
        <button class="btn btn-primary" @click="createContainer" :disabled="containerFormLoading || dbImagesLoading || dbImages.length === 0">
          <span v-if="containerFormLoading" class="spinner" style="width:13px;height:13px;border-color:rgba(255,255,255,0.3);border-top-color:white" />
          Create
        </button>
      </template>
    </BaseModal>

    <!-- Logs Modal -->
    <BaseModal v-if="logsModal" :title="`Logs — ${logsModal.name}`" size="lg" @close="logsModal = null">
      <div v-if="logsLoading" class="loading-overlay"><span class="spinner" /></div>
      <pre v-else class="logs-pre">{{ logsContent }}</pre>
    </BaseModal>

    <!-- Add DB Image Modal -->
    <BaseModal v-if="imageModal" title="Add Available Image" size="md" @close="imageModal = false">
      <div style="display:flex;flex-direction:column;gap:14px">
        <div v-if="imageFormError" class="alert alert-error">{{ imageFormError }}</div>
        <div class="form-group">
          <label class="form-label">Display Name *</label>
          <input v-model="imageForm.name" class="form-input" placeholder="Ubuntu 22.04 + CUDA" />
        </div>
        <div class="form-group">
          <label class="form-label">Image Address *</label>
          <select v-model="imageForm.image_id" class="form-select" :disabled="remoteLoading || remoteImages.length === 0">
            <option value="" disabled>
              {{ remoteLoading ? 'Loading images…' : remoteImages.length === 0 ? 'No images found on this server' : 'Select an image…' }}
            </option>
            <option v-for="img in remoteImages" :key="img.image_id" :value="img.image_id">
              {{ formatRemoteImageAddress(img) }}
            </option>
          </select>
          <span v-if="!remoteLoading && remoteImages.length === 0" class="form-hint" style="color:var(--warning)">
            Pull an image to this server before adding it to the available images list.
          </span>
        </div>
      </div>
      <template #footer>
        <button class="btn btn-secondary" @click="imageModal = false">Cancel</button>
        <button class="btn btn-primary" @click="saveDbImage" :disabled="imageFormLoading || remoteLoading || remoteImages.length === 0">Add Image</button>
      </template>
    </BaseModal>

    <!-- Pull Image Modal -->
    <BaseModal v-if="pullModal" title="Pull Image" size="sm" @close="pullModal = false">
      <div style="display:flex;flex-direction:column;gap:14px">
        <div class="form-group">
          <label class="form-label">Image Name</label>
          <input v-model="pullForm.image" class="form-input" placeholder="ubuntu" />
        </div>
        <div class="form-group">
          <label class="form-label">Tag</label>
          <input v-model="pullForm.tag" class="form-input" placeholder="22.04" />
        </div>
      </div>
      <template #footer>
        <button class="btn btn-secondary" @click="pullModal = false">Cancel</button>
        <button class="btn btn-primary" @click="pullImage" :disabled="pullLoading">
          <span v-if="pullLoading" class="spinner" style="width:13px;height:13px;border-color:rgba(255,255,255,0.3);border-top-color:white" />
          Pull
        </button>
      </template>
    </BaseModal>

    <!-- Confirm Container Action -->
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

    <!-- Confirm Delete Remote Image -->
    <ConfirmModal
      v-if="confirmDeleteImage"
      title="Delete Remote Image"
      :message="`Delete image '${confirmDeleteImage.repository}:${confirmDeleteImage.tag}' from the server?`"
      confirm-text="Delete"
      confirm-class="btn-danger"
      @confirm="deleteRemoteImage"
      @cancel="confirmDeleteImage = null"
    />

    <!-- Confirm Delete Volume -->
    <ConfirmModal
      v-if="confirmDeleteVolume"
      title="Delete Volume"
      :message="`Delete volume '${confirmDeleteVolume}'? All data stored in this volume will be lost.`"
      confirm-text="Delete"
      confirm-class="btn-danger"
      @confirm="deleteVolume"
      @cancel="confirmDeleteVolume = null"
    />
  </div>
</template>

<style scoped>
.breadcrumb {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
  font-size: 13px;
}

.tab-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  gap: 12px;
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

.image-action-btn {
  border: none;
  box-shadow: none;
}

.mono {
  font-family: var(--font-mono);
  font-size: 12.5px;
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
  .tab-toolbar {
    flex-wrap: wrap;
  }

  .row-actions {
    flex-wrap: wrap;
  }
}
</style>
