<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { del, get, post } from '@/api/client'
import { useToast } from '@/composables/useToast'
import BaseModal from '@/components/BaseModal.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'

const route = useRoute()
const router = useRouter()
const toast = useToast()

const servers = ref([])
const selectedServerId = ref(null)
const serversLoading = ref(true)

const dbImages = ref([])
const dbImagesLoading = ref(false)
const imageModal = ref(false)
const imageForm = ref({ name: '', image_id: '' })
const imageFormLoading = ref(false)
const imageFormError = ref('')

const remoteImages = ref([])
const remoteLoading = ref(false)
const pullModal = ref(false)
const pullForm = ref({ image: '', tag: 'latest' })
const pullLoading = ref(false)
const confirmDeleteImage = ref(null)
const deleteTarget = ref(null) // 'db' | 'remote'

async function loadServers() {
  serversLoading.value = true
  try {
    servers.value = await get('/servers') || []
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

async function loadDbImages() {
  if (!selectedServerId.value) return
  dbImagesLoading.value = true
  try {
    dbImages.value = await get('/images', { server_id: selectedServerId.value }) || []
  } catch (e) {
    toast.error(e.message)
  } finally {
    dbImagesLoading.value = false
  }
}

async function loadRemoteImages() {
  if (!selectedServerId.value) return
  remoteLoading.value = true
  try {
    remoteImages.value = await get(`/servers/${selectedServerId.value}/images`) || []
  } catch (e) {
    toast.error(e.message)
  } finally {
    remoteLoading.value = false
  }
}

function formatRemoteImageAddress(img) {
  if (!img) return ''
  return img.tag && img.tag !== '<none>' ? `${img.repository}:${img.tag}` : img.repository
}

async function saveDbImage() {
  imageFormLoading.value = true
  imageFormError.value = ''
  try {
    const selectedRemoteImage = remoteImages.value.find((img) => img.image_id === imageForm.value.image_id)
    if (!selectedRemoteImage) {
      throw new Error('Please select an image from the server.')
    }
    await post('/images', {
      server_id: selectedServerId.value,
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

function confirmDelete(img, type) {
  confirmDeleteImage.value = img
  deleteTarget.value = type
}

async function doDelete() {
  try {
    if (deleteTarget.value === 'db') {
      await del(`/images/${confirmDeleteImage.value.id}`)
      toast.success('Image removed')
      loadDbImages()
    } else {
      await del(`/servers/${selectedServerId.value}/images/${encodeURIComponent(confirmDeleteImage.value.image_id)}`)
      toast.success('Image deleted from server')
      loadRemoteImages()
    }
  } catch (e) {
    toast.error(e.message)
  } finally {
    confirmDeleteImage.value = null
    deleteTarget.value = null
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
    await post(`/servers/${selectedServerId.value}/images/pull`, { image: tag ? `${image}:${tag}` : image })
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

function formatSize(size) {
  return size || '-'
}

watch(selectedServerId, (id) => {
  if (id) router.replace({ query: { server: id } })
  loadDbImages()
  loadRemoteImages()
})
onMounted(loadServers)
</script>

<template>
  <div>
    <div class="page-header animate-in">
      <div>
        <h1 class="page-title">Images</h1>
        <p class="page-subtitle">Manage Docker images and registry configurations</p>
      </div>
      <div class="header-actions">
        <select v-if="servers.length > 0" v-model.number="selectedServerId" class="form-select server-filter">
          <option v-for="server in servers" :key="server.id" :value="server.id">{{ server.host }}</option>
        </select>
      </div>
    </div>

    <div v-if="!serversLoading && servers.length === 0" class="card animate-in">
      <div class="empty-state">
        <div class="empty-state-icon">🖥️</div>
        <div class="empty-state-text">No servers configured. Add a server first.</div>
      </div>
    </div>
    <template v-else-if="!serversLoading">
      <!-- Available Images -->
      <div style="margin-bottom:24px" class="animate-in animate-in-delay-1">
        <div class="section-toolbar">
          <div>
            <div class="section-title">Available Images</div>
            <div class="section-subtitle">Images shown to users in the application form</div>
          </div>
          <button class="btn btn-primary btn-w-action image-action-btn" @click="imageModal = true" :disabled="!selectedServerId">
            <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            Add Image
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
                      <button class="btn btn-danger btn-sm image-action-btn" @click="confirmDelete(img, 'db')">Delete</button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Remote Images -->
      <div class="animate-in animate-in-delay-2">
        <div class="section-toolbar">
          <div>
            <div class="section-title">Remote Images</div>
            <div class="section-subtitle">Images present on the server (docker images)</div>
          </div>
          <div style="display:flex;gap:8px">
            <button class="btn btn-primary btn-w-action image-action-btn" @click="pullModal = true" :disabled="!selectedServerId">Pull Image</button>
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
                      <button class="btn btn-danger btn-sm image-action-btn" @click="confirmDelete(img, 'remote')">Delete</button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </template>

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
        <button class="btn btn-primary" @click="saveDbImage" :disabled="imageFormLoading || remoteLoading || remoteImages.length === 0">
          <span v-if="imageFormLoading" class="spinner" style="width:13px;height:13px;border-color:rgba(255,255,255,0.3);border-top-color:white" />
          Add Image
        </button>
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

    <!-- Confirm Delete -->
    <ConfirmModal
      v-if="confirmDeleteImage"
      title="Delete Image"
      :message="deleteTarget === 'db'
        ? `Delete '${confirmDeleteImage.name}' from available images?`
        : `Delete image '${confirmDeleteImage.repository}:${confirmDeleteImage.tag}' from the server?`"
      confirm-text="Delete"
      confirm-class="btn-danger"
      @confirm="doDelete"
      @cancel="confirmDeleteImage = null"
    />
  </div>
</template>

<style scoped>
.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  gap: 12px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.section-subtitle {
  font-size: 12.5px;
  color: var(--text-muted);
  margin-top: 2px;
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

.image-action-btn {
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

  .section-toolbar {
    flex-wrap: wrap;
  }
}
</style>
