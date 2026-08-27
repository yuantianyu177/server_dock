<script setup>
import { computed, ref, watch } from 'vue'
import {
  CircleAlert,
  Download,
  ImagePlus,
  Layers3,
  RefreshCw,
  Search,
  Server as ServerIcon,
  Trash2,
  X
} from '@lucide/vue'
import { del, get, post } from '@/api/client'
import { useToast } from '@/composables/useToast'
import { useServerSelection } from '@/composables/useServerSelection'
import { runSettledBatch } from '@/utils/batch'
import BaseModal from '@/components/BaseModal.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'
import PageHeader from '@/components/PageHeader.vue'
import ServerLens from '@/components/ServerLens.vue'
import StatePanel from '@/components/StatePanel.vue'

const toast = useToast()
const {
  servers,
  selectedServerId,
  serversLoading,
  serversError,
  loadServers
} = useServerSelection()

const dbImages = ref([])
const dbImagesLoading = ref(false)
const dbImagesError = ref('')
const dbSearch = ref('')

const remoteImages = ref([])
const remoteLoading = ref(false)
const remoteError = ref('')
const remoteSearch = ref('')

const lensState = ref('offline')
const summary = ref({ running: 0, total: 0 })

const imageModal = ref(false)
const imageForm = ref({ name: '', image_id: '' })
const imageFormLoading = ref(false)
const imageFormError = ref('')

const pullModal = ref(false)
const pullForm = ref({ image: '', tag: 'latest' })
const pullLoading = ref(false)
const pullError = ref('')

const confirmDeleteImage = ref(null)
const deleteTarget = ref(null)
const deleteLoading = ref(false)
const selectedDbImageIds = ref([])
const selectedRemoteImageIds = ref([])
const batchDeleteTarget = ref('')
const batchDeleteLoading = ref(false)

const imageActionLoading = computed(() => deleteLoading.value || batchDeleteLoading.value)

const filteredDbImages = computed(() => {
  const query = dbSearch.value.trim().toLowerCase()
  if (!query) return dbImages.value
  return dbImages.value.filter(image =>
    [image.name, image.image_address].some(value => String(value || '').toLowerCase().includes(query))
  )
})

const filteredRemoteImages = computed(() => {
  const query = remoteSearch.value.trim().toLowerCase()
  if (!query) return remoteImages.value
  return remoteImages.value.filter(image =>
    [image.repository, image.tag, image.image_id].some(value => String(value || '').toLowerCase().includes(query))
  )
})

const visibleDbImageIds = computed(() => filteredDbImages.value.map(image => image.id))
const allVisibleDbSelected = computed(() =>
  visibleDbImageIds.value.length > 0 && visibleDbImageIds.value.every(id => selectedDbImageIds.value.includes(id))
)
const someVisibleDbSelected = computed(() =>
  !allVisibleDbSelected.value && visibleDbImageIds.value.some(id => selectedDbImageIds.value.includes(id))
)
const visibleRemoteImageIds = computed(() =>
  [...new Set(filteredRemoteImages.value.map(image => image.image_id))]
)
const allVisibleRemoteSelected = computed(() =>
  visibleRemoteImageIds.value.length > 0 && visibleRemoteImageIds.value.every(id => selectedRemoteImageIds.value.includes(id))
)
const someVisibleRemoteSelected = computed(() =>
  !allVisibleRemoteSelected.value && visibleRemoteImageIds.value.some(id => selectedRemoteImageIds.value.includes(id))
)

async function loadSummary() {
  if (!selectedServerId.value) return
  const serverId = Number(selectedServerId.value)
  lensState.value = 'offline'
  try {
    const containers = await get(`/servers/${serverId}/containers`) || []
    if (Number(selectedServerId.value) !== serverId) return
    summary.value = {
      total: containers.length,
      running: containers.filter(container => container.status?.toLowerCase().startsWith('up')).length
    }
    lensState.value = 'online'
  } catch (error) {
    if (Number(selectedServerId.value) !== serverId) return
    summary.value = { running: 0, total: 0 }
    lensState.value = 'offline'
  }
}

async function loadDbImages() {
  if (!selectedServerId.value) return
  const serverId = Number(selectedServerId.value)
  dbImagesLoading.value = true
  dbImagesError.value = ''
  try {
    const result = await get('/images', { server_id: serverId }) || []
    if (Number(selectedServerId.value) !== serverId) return
    dbImages.value = result
    const availableIds = new Set(result.map(image => image.id))
    selectedDbImageIds.value = selectedDbImageIds.value.filter(id => availableIds.has(id))
  } catch (error) {
    if (Number(selectedServerId.value) !== serverId) return
    dbImages.value = []
    dbImagesError.value = error.message
  } finally {
    if (Number(selectedServerId.value) === serverId) dbImagesLoading.value = false
  }
}

async function loadRemoteImages() {
  if (!selectedServerId.value) return
  const serverId = Number(selectedServerId.value)
  remoteLoading.value = true
  remoteError.value = ''
  try {
    const result = await get(`/servers/${serverId}/images`) || []
    if (Number(selectedServerId.value) !== serverId) return
    remoteImages.value = result
    const availableIds = new Set(result.map(image => image.image_id))
    selectedRemoteImageIds.value = selectedRemoteImageIds.value.filter(id => availableIds.has(id))
  } catch (error) {
    if (Number(selectedServerId.value) !== serverId) return
    remoteImages.value = []
    remoteError.value = error.message
  } finally {
    if (Number(selectedServerId.value) === serverId) remoteLoading.value = false
  }
}

function reloadAll() {
  return Promise.all([loadSummary(), loadDbImages(), loadRemoteImages()])
}

function formatRemoteImageAddress(image) {
  if (!image) return ''
  return image.tag && image.tag !== '<none>' ? `${image.repository}:${image.tag}` : image.repository
}

function openImageModal() {
  imageForm.value = { name: '', image_id: '' }
  imageFormError.value = ''
  imageModal.value = true
  if (!remoteImages.value.length && !remoteLoading.value) loadRemoteImages()
}

async function saveDbImage() {
  if (!imageForm.value.name.trim()) {
    imageFormError.value = '请输入展示名称。'
    return
  }
  const selectedImage = remoteImages.value.find(image => image.image_id === imageForm.value.image_id)
  if (!selectedImage) {
    imageFormError.value = '请选择服务器上已有的镜像。'
    return
  }

  imageFormLoading.value = true
  imageFormError.value = ''
  try {
    await post('/images', {
      server_id: selectedServerId.value,
      name: imageForm.value.name.trim(),
      image_id: selectedImage.image_id,
      image_address: formatRemoteImageAddress(selectedImage)
    })
    imageModal.value = false
    toast.success(`已登记可申请镜像“${imageForm.value.name.trim()}”`)
    await loadDbImages()
  } catch (error) {
    imageFormError.value = `无法登记镜像：${error.message}`
  } finally {
    imageFormLoading.value = false
  }
}

function requestDelete(image, target) {
  confirmDeleteImage.value = image
  deleteTarget.value = target
}

async function deleteImage() {
  if (!confirmDeleteImage.value) return
  deleteLoading.value = true
  try {
    if (deleteTarget.value === 'db') {
      await del(`/images/${confirmDeleteImage.value.id}`)
      toast.success(`已取消登记“${confirmDeleteImage.value.name}”`)
      await loadDbImages()
    } else {
      await del(`/servers/${selectedServerId.value}/images/${encodeURIComponent(confirmDeleteImage.value.image_id)}`)
      toast.success(`已从服务器删除镜像“${formatRemoteImageAddress(confirmDeleteImage.value)}”`)
      await loadRemoteImages()
    }
    confirmDeleteImage.value = null
    deleteTarget.value = null
  } catch (error) {
    toast.error(`无法删除镜像：${error.message}`)
  } finally {
    deleteLoading.value = false
  }
}

function setDbImageSelected(id, selected) {
  selectedDbImageIds.value = selected
    ? [...new Set([...selectedDbImageIds.value, id])]
    : selectedDbImageIds.value.filter(item => item !== id)
}

function toggleVisibleDbImages(selected) {
  const visibleIds = new Set(visibleDbImageIds.value)
  selectedDbImageIds.value = selected
    ? [...new Set([...selectedDbImageIds.value, ...visibleIds])]
    : selectedDbImageIds.value.filter(id => !visibleIds.has(id))
}

function setRemoteImageSelected(id, selected) {
  selectedRemoteImageIds.value = selected
    ? [...new Set([...selectedRemoteImageIds.value, id])]
    : selectedRemoteImageIds.value.filter(item => item !== id)
}

function toggleVisibleRemoteImages(selected) {
  const visibleIds = new Set(visibleRemoteImageIds.value)
  selectedRemoteImageIds.value = selected
    ? [...new Set([...selectedRemoteImageIds.value, ...visibleIds])]
    : selectedRemoteImageIds.value.filter(id => !visibleIds.has(id))
}

function requestBatchDelete(target) {
  const selectedIds = target === 'db' ? selectedDbImageIds.value : selectedRemoteImageIds.value
  if (selectedIds.length === 0 || imageActionLoading.value) return
  batchDeleteTarget.value = target
}

async function deleteSelectedImages() {
  const target = batchDeleteTarget.value
  const ids = [...(target === 'db' ? selectedDbImageIds.value : selectedRemoteImageIds.value)]
  if (!target || ids.length === 0) return

  batchDeleteLoading.value = true
  const serverId = selectedServerId.value
  const results = await runSettledBatch(ids, id => target === 'db'
    ? del(`/images/${id}`)
    : del(`/servers/${serverId}/images/${encodeURIComponent(id)}`)
  )
  const failedIds = ids.filter((_, index) => results[index].status === 'rejected')
  const succeededCount = ids.length - failedIds.length
  const action = target === 'db' ? '取消登记' : '删除'

  if (target === 'db') selectedDbImageIds.value = failedIds
  else selectedRemoteImageIds.value = failedIds
  batchDeleteTarget.value = ''

  if (failedIds.length === 0) {
    toast.success(`已批量${action} ${succeededCount} 个镜像`)
  } else {
    const firstError = results.find(result => result.status === 'rejected')?.reason?.message
    const summary = succeededCount > 0
      ? `已${action} ${succeededCount} 个，${failedIds.length} 个失败`
      : `${failedIds.length} 个镜像均未能${action}`
    toast.error(firstError ? `${summary}：${firstError}` : summary)
  }

  try {
    if (target === 'db') await loadDbImages()
    else await loadRemoteImages()
  } finally {
    batchDeleteLoading.value = false
  }
}

function openPullModal() {
  pullForm.value = { image: '', tag: 'latest' }
  pullError.value = ''
  pullModal.value = true
}

async function pullImage() {
  const image = pullForm.value.image.trim()
  const tag = pullForm.value.tag.trim()
  if (!image) {
    pullError.value = '请输入镜像名称或仓库地址。'
    return
  }

  pullLoading.value = true
  pullError.value = ''
  try {
    const address = tag ? `${image}:${tag}` : image
    await post(`/servers/${selectedServerId.value}/images/pull`, { image: address })
    pullModal.value = false
    toast.success(`已拉取镜像“${address}”`)
    await loadRemoteImages()
  } catch (error) {
    pullError.value = `无法拉取镜像：${error.message}`
  } finally {
    pullLoading.value = false
  }
}

function formatDate(value) {
  if (!value) return '—'
  return new Intl.DateTimeFormat('zh-CN', { year: 'numeric', month: 'short', day: 'numeric' }).format(new Date(value))
}

watch(selectedServerId, id => {
  dbSearch.value = ''
  remoteSearch.value = ''
  dbImages.value = []
  remoteImages.value = []
  selectedDbImageIds.value = []
  selectedRemoteImageIds.value = []
  batchDeleteTarget.value = ''
  if (!id) {
    dbImagesLoading.value = false
    remoteLoading.value = false
    dbImagesError.value = ''
    remoteError.value = ''
    lensState.value = 'offline'
    summary.value = { running: 0, total: 0 }
    return
  }
  reloadAll()
})
</script>

<template>
  <div>
    <PageHeader title="镜像" description="区分服务器本地镜像与允许用户申请的镜像，避免误删仍在使用的环境。" />

    <ServerLens
      v-model="selectedServerId"
      :servers="servers"
      :loading="serversLoading || imageActionLoading"
      :state="lensState"
      :summary="summary"
      :error="remoteError"
      @retry="reloadAll"
    >
      <template #actions>
        <button class="btn btn-secondary" type="button" :disabled="!selectedServerId || lensState === 'offline'" @click="openImageModal">
          <ImagePlus :size="15" aria-hidden="true" />登记镜像
        </button>
        <button class="btn btn-primary" type="button" :disabled="!selectedServerId || lensState === 'offline'" @click="openPullModal">
          <Download :size="15" aria-hidden="true" />拉取镜像
        </button>
      </template>
    </ServerLens>

    <template v-if="!serversLoading && !serversError && servers.length > 0">
      <section class="section-block" aria-labelledby="available-images-title">
        <div class="section-heading">
          <div>
            <h2 id="available-images-title" class="section-title">可申请镜像</h2>
            <p class="section-subtitle">这里登记的镜像会出现在公开容器申请表中。</p>
          </div>
          <button class="btn btn-secondary btn-sm" type="button" :disabled="dbImagesLoading || imageActionLoading" @click="loadDbImages">
            <RefreshCw :size="14" :class="{ spinning: dbImagesLoading }" aria-hidden="true" />刷新
          </button>
        </div>

        <div class="data-panel responsive-table">
          <div class="table-toolbar">
            <div class="table-tools bulk-table-tools">
              <div class="search-field">
                <Search :size="16" aria-hidden="true" />
                <label class="sr-only" for="available-image-search">搜索可申请镜像</label>
                <input id="available-image-search" v-model="dbSearch" class="form-input" placeholder="搜索名称或镜像地址" />
                <button v-if="dbSearch" class="search-clear" type="button" aria-label="清除搜索" @click="dbSearch = ''"><X :size="14" /></button>
              </div>
              <div class="batch-actions" aria-label="可申请镜像批量操作">
                <span class="batch-count" :data-count="selectedDbImageIds.length">已选 {{ selectedDbImageIds.length }} 项</span>
                <button class="btn btn-danger btn-sm" type="button" :disabled="imageActionLoading || selectedDbImageIds.length === 0" title="批量取消登记" aria-label="批量取消登记选中镜像" @click="requestBatchDelete('db')">
                  <Trash2 :size="14" aria-hidden="true" /><span class="batch-action-label">取消登记</span>
                </button>
              </div>
            </div>
            <span class="result-count">{{ filteredDbImages.length }} 个镜像</span>
          </div>

          <div v-if="dbImagesLoading" class="loading-state" role="status"><span class="spinner" /><span>正在读取可申请镜像…</span></div>
          <StatePanel v-else-if="dbImagesError" tone="error" title="无法读取可申请镜像" :description="`${dbImagesError}。请重试。`">
            <template #icon><CircleAlert :size="20" /></template>
            <template #actions><button class="btn btn-secondary btn-sm" type="button" @click="loadDbImages">重新加载</button></template>
          </StatePanel>
          <StatePanel v-else-if="dbImages.length === 0" title="还没有可申请镜像" description="从这台服务器已有的 Docker 镜像中登记一个环境。">
            <template #icon><ImagePlus :size="20" /></template>
            <template #actions><button class="btn btn-primary btn-sm" type="button" @click="openImageModal">登记镜像</button></template>
          </StatePanel>
          <StatePanel v-else-if="filteredDbImages.length === 0" title="没有匹配的镜像" description="尝试搜索展示名称或完整镜像地址。">
            <template #icon><Search :size="20" /></template>
            <template #actions><button class="btn btn-secondary btn-sm" type="button" @click="dbSearch = ''">清除搜索</button></template>
          </StatePanel>
          <div v-else class="table-wrap">
            <table>
              <thead>
                <tr>
                  <th class="selection-cell">
                    <input
                      class="row-checkbox"
                      type="checkbox"
                      :checked="allVisibleDbSelected"
                      :indeterminate="someVisibleDbSelected"
                      :disabled="imageActionLoading"
                      aria-label="选择当前筛选结果中的全部可申请镜像"
                      @change="toggleVisibleDbImages($event.target.checked)"
                    />
                  </th>
                  <th>展示名称</th><th>镜像地址</th><th>登记时间</th><th><span class="sr-only">操作</span></th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="image in filteredDbImages" :key="image.id" :class="{ 'is-selected': selectedDbImageIds.includes(image.id) }">
                  <td class="selection-cell" data-label="选择">
                    <input
                      class="row-checkbox"
                      type="checkbox"
                      :checked="selectedDbImageIds.includes(image.id)"
                      :disabled="imageActionLoading"
                      :aria-label="`选择可申请镜像 ${image.name}`"
                      @change="setDbImageSelected(image.id, $event.target.checked)"
                    />
                  </td>
                  <td data-label="展示名称"><span class="table-primary">{{ image.name }}</span></td>
                  <td data-label="镜像地址"><span class="mono image-address">{{ image.image_address }}</span></td>
                  <td data-label="登记时间" class="date-cell">{{ formatDate(image.created_at) }}</td>
                  <td data-label="">
                    <div class="row-actions">
                      <button class="btn btn-danger btn-sm" type="button" :disabled="imageActionLoading" @click="requestDelete(image, 'db')">
                        <Trash2 :size="14" aria-hidden="true" />取消登记
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>

      <section class="section-block" aria-labelledby="server-images-title">
        <div class="section-heading">
          <div>
            <h2 id="server-images-title" class="section-title">服务器镜像</h2>
            <p class="section-subtitle">Docker 当前保存在这台服务器上的全部镜像。</p>
          </div>
          <button class="btn btn-secondary btn-sm" type="button" :disabled="remoteLoading || imageActionLoading" @click="loadRemoteImages">
            <RefreshCw :size="14" :class="{ spinning: remoteLoading }" aria-hidden="true" />刷新
          </button>
        </div>

        <div class="data-panel responsive-table">
          <div class="table-toolbar">
            <div class="table-tools bulk-table-tools">
              <div class="search-field">
                <Search :size="16" aria-hidden="true" />
                <label class="sr-only" for="remote-image-search">搜索服务器镜像</label>
                <input id="remote-image-search" v-model="remoteSearch" class="form-input" placeholder="搜索仓库、标签或镜像 ID" />
                <button v-if="remoteSearch" class="search-clear" type="button" aria-label="清除搜索" @click="remoteSearch = ''"><X :size="14" /></button>
              </div>
              <div class="batch-actions" aria-label="服务器镜像批量操作">
                <span class="batch-count" :data-count="selectedRemoteImageIds.length">已选 {{ selectedRemoteImageIds.length }} 项</span>
                <button class="btn btn-danger btn-sm" type="button" :disabled="imageActionLoading || selectedRemoteImageIds.length === 0" title="批量删除" aria-label="批量删除选中服务器镜像" @click="requestBatchDelete('remote')">
                  <Trash2 :size="14" aria-hidden="true" /><span class="batch-action-label">删除</span>
                </button>
              </div>
            </div>
            <span class="result-count">{{ filteredRemoteImages.length }} 个镜像</span>
          </div>

          <div v-if="remoteLoading" class="loading-state" role="status"><span class="spinner" /><span>正在读取服务器镜像…</span></div>
          <StatePanel v-else-if="remoteError" tone="error" title="无法读取服务器镜像" :description="`${remoteError}。请检查 SSH 连接和 Docker 服务。`">
            <template #icon><CircleAlert :size="20" /></template>
            <template #actions><button class="btn btn-secondary btn-sm" type="button" @click="loadRemoteImages">重新加载</button></template>
          </StatePanel>
          <StatePanel v-else-if="remoteImages.length === 0" title="服务器上没有镜像" description="从镜像仓库拉取一个镜像后，可将其登记为申请环境。">
            <template #icon><Layers3 :size="20" /></template>
            <template #actions><button class="btn btn-primary btn-sm" type="button" @click="openPullModal">拉取镜像</button></template>
          </StatePanel>
          <StatePanel v-else-if="filteredRemoteImages.length === 0" title="没有匹配的镜像" description="调整仓库、标签或镜像 ID 搜索词。">
            <template #icon><Search :size="20" /></template>
            <template #actions><button class="btn btn-secondary btn-sm" type="button" @click="remoteSearch = ''">清除搜索</button></template>
          </StatePanel>
          <div v-else class="table-wrap">
            <table>
              <thead>
                <tr>
                  <th class="selection-cell">
                    <input
                      class="row-checkbox"
                      type="checkbox"
                      :checked="allVisibleRemoteSelected"
                      :indeterminate="someVisibleRemoteSelected"
                      :disabled="imageActionLoading"
                      aria-label="选择当前筛选结果中的全部服务器镜像"
                      @change="toggleVisibleRemoteImages($event.target.checked)"
                    />
                  </th>
                  <th>仓库</th><th>标签</th><th>镜像 ID</th><th>大小</th><th>创建时间</th><th><span class="sr-only">操作</span></th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="image in filteredRemoteImages" :key="`${image.image_id}-${image.repository}-${image.tag}`" :class="{ 'is-selected': selectedRemoteImageIds.includes(image.image_id) }">
                  <td class="selection-cell" data-label="选择">
                    <input
                      class="row-checkbox"
                      type="checkbox"
                      :checked="selectedRemoteImageIds.includes(image.image_id)"
                      :disabled="imageActionLoading"
                      :aria-label="`选择服务器镜像 ${formatRemoteImageAddress(image)}`"
                      @change="setRemoteImageSelected(image.image_id, $event.target.checked)"
                    />
                  </td>
                  <td data-label="仓库"><span class="mono image-address">{{ image.repository }}</span></td>
                  <td data-label="标签"><span class="mono">{{ image.tag }}</span></td>
                  <td data-label="镜像 ID"><span class="mono image-id">{{ image.image_id }}</span></td>
                  <td data-label="大小" class="date-cell">{{ image.size || '—' }}</td>
                  <td data-label="创建时间" class="date-cell">{{ image.created || '—' }}</td>
                  <td data-label="">
                    <div class="row-actions">
                      <button class="btn btn-danger btn-sm btn-icon" type="button" :disabled="imageActionLoading" title="从服务器删除镜像" :aria-label="`删除镜像 ${formatRemoteImageAddress(image)}`" @click="requestDelete(image, 'remote')">
                        <Trash2 :size="15" aria-hidden="true" />
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>
    </template>

    <section v-else class="data-panel">
      <StatePanel v-if="!serversLoading && serversError" tone="error" title="无法读取服务器" :description="`${serversError}。请检查 API 服务后重试。`">
        <template #icon><CircleAlert :size="20" /></template>
        <template #actions><button class="btn btn-secondary btn-sm" type="button" @click="loadServers">重新加载</button></template>
      </StatePanel>
      <StatePanel v-else-if="!serversLoading" title="请先添加服务器" description="镜像由具体服务器上的 Docker 服务管理。">
        <template #icon><ServerIcon :size="20" /></template>
        <template #actions><router-link class="btn btn-primary btn-sm" to="/servers">前往服务器</router-link></template>
      </StatePanel>
      <div v-else class="loading-state" role="status"><span class="spinner" /><span>正在读取服务器…</span></div>
    </section>

    <BaseModal v-if="imageModal" title="登记可申请镜像" size="md" :close-on-backdrop="!imageFormLoading" @close="!imageFormLoading && (imageModal = false)">
      <div class="modal-form">
        <div v-if="imageFormError" class="alert alert-error" role="alert"><CircleAlert :size="17" />{{ imageFormError }}</div>
        <div class="form-group">
          <label class="form-label" for="image-display-name">展示名称 <span class="required-mark">*</span></label>
          <input id="image-display-name" v-model="imageForm.name" class="form-input" placeholder="例如：Ubuntu 22.04 + CUDA" required />
          <span class="form-hint">该名称会显示在公开容器申请表中。</span>
        </div>
        <div class="form-group">
          <label class="form-label" for="image-address">服务器镜像 <span class="required-mark">*</span></label>
          <select id="image-address" v-model="imageForm.image_id" class="form-select" :disabled="remoteLoading || remoteImages.length === 0" required>
            <option value="" disabled>{{ remoteLoading ? '正在读取镜像…' : remoteImages.length ? '选择服务器镜像' : '服务器上没有镜像' }}</option>
            <option v-for="image in remoteImages" :key="`${image.image_id}-${image.repository}-${image.tag}`" :value="image.image_id">
              {{ formatRemoteImageAddress(image) }}
            </option>
          </select>
          <span v-if="!remoteLoading && remoteImages.length === 0" class="form-error">请先将镜像拉取到当前服务器。</span>
        </div>
      </div>
      <template #footer>
        <button class="btn btn-secondary" type="button" :disabled="imageFormLoading" @click="imageModal = false">取消</button>
        <button class="btn btn-primary" type="button" :disabled="imageFormLoading || remoteLoading || remoteImages.length === 0" @click="saveDbImage">
          <span v-if="imageFormLoading" class="spinner button-spinner" />{{ imageFormLoading ? '正在登记…' : '登记镜像' }}
        </button>
      </template>
    </BaseModal>

    <BaseModal v-if="pullModal" title="拉取镜像" size="sm" :close-on-backdrop="!pullLoading" @close="!pullLoading && (pullModal = false)">
      <div class="modal-form">
        <div v-if="pullError" class="alert alert-error" role="alert"><CircleAlert :size="17" />{{ pullError }}</div>
        <div class="form-group">
          <label class="form-label" for="pull-image-name">镜像名称或仓库地址 <span class="required-mark">*</span></label>
          <input id="pull-image-name" v-model="pullForm.image" class="form-input mono" placeholder="ubuntu 或 ghcr.io/org/image" required />
        </div>
        <div class="form-group">
          <label class="form-label" for="pull-image-tag">标签</label>
          <input id="pull-image-tag" v-model="pullForm.tag" class="form-input mono" placeholder="latest" />
        </div>
        <p class="pull-note">拉取耗时取决于镜像大小和服务器网络。关闭此对话框不会加快操作。</p>
      </div>
      <template #footer>
        <button class="btn btn-secondary" type="button" :disabled="pullLoading" @click="pullModal = false">取消</button>
        <button class="btn btn-primary" type="button" :disabled="pullLoading" @click="pullImage">
          <span v-if="pullLoading" class="spinner button-spinner" />{{ pullLoading ? '正在拉取…' : '拉取镜像' }}
        </button>
      </template>
    </BaseModal>

    <ConfirmModal
      v-if="confirmDeleteImage"
      title="删除镜像"
      :message="deleteTarget === 'db'
        ? `取消登记可申请镜像“${confirmDeleteImage.name}”吗？`
        : `从当前服务器删除镜像“${formatRemoteImageAddress(confirmDeleteImage)}”吗？`"
      :detail="deleteTarget === 'db'
        ? '该镜像将不再出现在公开申请表中，但服务器上的 Docker 镜像仍会保留。'
        : '服务器将执行 docker rmi；已登记为可申请的镜像必须先取消登记。此操作不会删除正在运行的容器。'"
      :confirm-text="deleteTarget === 'db' ? '取消登记' : '删除镜像'"
      :loading="deleteLoading"
      @confirm="deleteImage"
      @cancel="confirmDeleteImage = null; deleteTarget = null"
    />

    <ConfirmModal
      v-if="batchDeleteTarget"
      :title="batchDeleteTarget === 'db' ? '批量取消登记' : '批量删除服务器镜像'"
      :message="batchDeleteTarget === 'db'
        ? `确定取消登记选中的 ${selectedDbImageIds.length} 个可申请镜像吗？`
        : `确定从当前服务器删除选中的 ${selectedRemoteImageIds.length} 个镜像吗？`"
      :detail="batchDeleteTarget === 'db'
        ? '这些镜像将不再出现在公开申请表中，但服务器上的 Docker 镜像仍会保留。'
        : '服务器将逐个执行 docker rmi；已登记或正在被容器使用的镜像会删除失败并保持选中，便于重试。'"
      :confirm-text="batchDeleteTarget === 'db' ? '批量取消登记' : '批量删除'"
      :loading="batchDeleteLoading"
      @confirm="deleteSelectedImages"
      @cancel="batchDeleteTarget = ''"
    />
  </div>
</template>

<style scoped>
.result-count,
.date-cell {
  color: var(--ink-secondary);
  font-size: 12px;
  white-space: nowrap;
}

.bulk-table-tools {
  min-width: 0;
  flex: 1;
  flex-wrap: nowrap;
  overflow-x: auto;
  scrollbar-width: none;
}

.bulk-table-tools::-webkit-scrollbar {
  display: none;
}

.bulk-table-tools .search-field {
  flex: 1 0 210px;
}

.batch-actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 5px;
}

.batch-count {
  margin: 0 3px 0 5px;
  color: var(--ink-secondary);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.image-address {
  color: var(--ink-secondary);
  font-size: 12px;
}

.image-id {
  color: var(--ink-tertiary);
  font-size: 11px;
}

.spinning {
  animation: spin 700ms linear infinite;
}

.modal-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.pull-note {
  color: var(--ink-secondary);
  font-size: 12px;
  line-height: 1.55;
}

.button-spinner {
  width: 13px;
  height: 13px;
  color: #fff;
}

@media (max-width: 680px) {
  .bulk-table-tools .search-field {
    width: auto;
    min-width: 160px;
  }

  .batch-count {
    min-width: 28px;
    margin: 0;
    font-size: 0;
    text-align: center;
  }

  .batch-count::after {
    content: attr(data-count) " 项";
    font-size: 11px;
  }

  .batch-actions .btn {
    width: 30px;
    padding: 0;
  }

  .batch-action-label {
    display: none;
  }
}
</style>
