<script setup>
import { computed, ref, watch } from 'vue'
import {
  CircleAlert,
  Database,
  Plus,
  RefreshCw,
  Search,
  Server as ServerIcon,
  SquareTerminal,
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

const volumes = ref([])
const volumesLoading = ref(false)
const volumesError = ref('')
const lensState = ref('offline')
const summary = ref({ running: 0, total: 0 })
const searchQuery = ref('')

const createModal = ref(false)
const volumeName = ref('')
const formLoading = ref(false)
const formError = ref('')

const deleteTarget = ref(null)
const deleteLoading = ref(false)
const selectedVolumeNames = ref([])
const batchDeleteConfirm = ref(false)
const batchDeleteLoading = ref(false)

const filteredVolumes = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  if (!query) return volumes.value
  return volumes.value.filter(volume =>
    [volume.Name, volume.name, volume.Driver, volume.driver, volume.Mountpoint, volume.mountpoint]
      .some(value => String(value || '').toLowerCase().includes(query))
  )
})

const visibleVolumeNames = computed(() =>
  filteredVolumes.value.map(volume => volumeValue(volume, 'Name', 'name'))
)
const allVisibleSelected = computed(() =>
  visibleVolumeNames.value.length > 0 && visibleVolumeNames.value.every(name => selectedVolumeNames.value.includes(name))
)
const someVisibleSelected = computed(() =>
  !allVisibleSelected.value && visibleVolumeNames.value.some(name => selectedVolumeNames.value.includes(name))
)

function volumeValue(volume, upper, lower, fallback = '—') {
  return volume?.[upper] || volume?.[lower] || fallback
}

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
  } catch {
    if (Number(selectedServerId.value) !== serverId) return
    summary.value = { running: 0, total: 0 }
    lensState.value = 'offline'
  }
}

async function loadVolumes() {
  if (!selectedServerId.value) return
  const serverId = Number(selectedServerId.value)
  volumesLoading.value = true
  volumesError.value = ''
  try {
    const result = await get(`/servers/${serverId}/volumes`) || []
    if (Number(selectedServerId.value) !== serverId) return
    volumes.value = result
    const availableNames = new Set(result.map(volume => volumeValue(volume, 'Name', 'name')))
    selectedVolumeNames.value = selectedVolumeNames.value.filter(name => availableNames.has(name))
  } catch (error) {
    if (Number(selectedServerId.value) !== serverId) return
    volumes.value = []
    volumesError.value = error.message
  } finally {
    if (Number(selectedServerId.value) === serverId) volumesLoading.value = false
  }
}

function reloadAll() {
  return Promise.all([loadSummary(), loadVolumes()])
}

function openCreateModal() {
  volumeName.value = ''
  formError.value = ''
  createModal.value = true
}

async function createVolume() {
  const name = volumeName.value.trim()
  if (!name) {
    formError.value = '请输入数据卷名称。'
    return
  }
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_.-]*$/.test(name)) {
    formError.value = '名称需以字母或数字开头，只能包含字母、数字、点、连字符和下划线。'
    return
  }

  formLoading.value = true
  formError.value = ''
  try {
    await post(`/servers/${selectedServerId.value}/volumes`, { name })
    createModal.value = false
    toast.success(`已创建数据卷“${name}”`)
    await loadVolumes()
  } catch (error) {
    formError.value = `无法创建数据卷：${error.message}`
  } finally {
    formLoading.value = false
  }
}

async function deleteVolume() {
  if (!deleteTarget.value) return
  const name = volumeValue(deleteTarget.value, 'Name', 'name')
  deleteLoading.value = true
  try {
    await del(`/servers/${selectedServerId.value}/volumes/${encodeURIComponent(name)}`)
    deleteTarget.value = null
    toast.success(`已删除数据卷“${name}”`)
    await loadVolumes()
  } catch (error) {
    toast.error(`无法删除数据卷：${error.message}`)
  } finally {
    deleteLoading.value = false
  }
}

function setVolumeSelected(name, selected) {
  selectedVolumeNames.value = selected
    ? [...new Set([...selectedVolumeNames.value, name])]
    : selectedVolumeNames.value.filter(item => item !== name)
}

function toggleVisibleVolumes(selected) {
  const visibleNames = new Set(visibleVolumeNames.value)
  selectedVolumeNames.value = selected
    ? [...new Set([...selectedVolumeNames.value, ...visibleNames])]
    : selectedVolumeNames.value.filter(name => !visibleNames.has(name))
}

async function deleteSelectedVolumes() {
  const names = [...selectedVolumeNames.value]
  if (names.length === 0) return

  batchDeleteLoading.value = true
  const serverId = selectedServerId.value
  const results = await runSettledBatch(names, name =>
    del(`/servers/${serverId}/volumes/${encodeURIComponent(name)}`)
  )
  const failedNames = names.filter((_, index) => results[index].status === 'rejected')
  const succeededCount = names.length - failedNames.length
  selectedVolumeNames.value = failedNames
  batchDeleteConfirm.value = false

  if (failedNames.length === 0) {
    toast.success(`已批量删除 ${succeededCount} 个数据卷`)
  } else {
    const firstError = results.find(result => result.status === 'rejected')?.reason?.message
    const summary = succeededCount > 0
      ? `已删除 ${succeededCount} 个，${failedNames.length} 个失败`
      : `${failedNames.length} 个数据卷均未能删除`
    toast.error(firstError ? `${summary}：${firstError}` : summary)
  }

  try {
    await loadVolumes()
  } finally {
    batchDeleteLoading.value = false
  }
}

function openTerminal() {
  window.open(`/terminal/${selectedServerId.value}`, '_blank', 'noopener')
}

watch(selectedServerId, id => {
  searchQuery.value = ''
  volumes.value = []
  selectedVolumeNames.value = []
  batchDeleteConfirm.value = false
  if (!id) {
    volumesLoading.value = false
    volumesError.value = ''
    lensState.value = 'offline'
    summary.value = { running: 0, total: 0 }
    return
  }
  reloadAll()
})
</script>

<template>
  <div>
    <PageHeader title="数据卷" description="管理当前服务器上的 Docker 持久化数据，删除前请确认不再被容器使用。" />

    <ServerLens
      v-model="selectedServerId"
      :servers="servers"
      :loading="serversLoading || deleteLoading || batchDeleteLoading"
      :state="lensState"
      :summary="summary"
      :error="volumesError"
      @retry="reloadAll"
    >
      <template #actions>
        <button class="btn btn-secondary" type="button" :disabled="!selectedServerId || lensState === 'offline'" @click="openTerminal">
          <SquareTerminal :size="15" aria-hidden="true" />打开终端
        </button>
        <button class="btn btn-primary" type="button" :disabled="!selectedServerId || lensState === 'offline'" @click="openCreateModal">
          <Plus :size="15" aria-hidden="true" />新建数据卷
        </button>
      </template>
    </ServerLens>

    <section class="data-panel responsive-table" aria-label="数据卷列表">
      <template v-if="!serversLoading && !serversError && servers.length > 0">
        <div class="table-toolbar">
          <div class="table-tools bulk-table-tools">
            <div class="search-field">
              <Search :size="16" aria-hidden="true" />
              <label class="sr-only" for="volume-search">搜索数据卷</label>
              <input id="volume-search" v-model="searchQuery" class="form-input" placeholder="搜索名称、驱动或挂载路径" />
              <button v-if="searchQuery" class="search-clear" type="button" aria-label="清除搜索" @click="searchQuery = ''"><X :size="14" /></button>
            </div>
            <div class="batch-actions" aria-label="数据卷批量操作">
              <span class="batch-count" :data-count="selectedVolumeNames.length">已选 {{ selectedVolumeNames.length }} 项</span>
              <button class="btn btn-danger btn-sm" type="button" :disabled="batchDeleteLoading || deleteLoading || selectedVolumeNames.length === 0" title="批量删除" aria-label="批量删除选中数据卷" @click="batchDeleteConfirm = true">
                <Trash2 :size="14" aria-hidden="true" /><span class="batch-action-label">删除</span>
              </button>
            </div>
          </div>
          <button class="btn btn-ghost btn-sm" type="button" :disabled="volumesLoading || batchDeleteLoading" @click="loadVolumes">
            <RefreshCw :size="14" :class="{ spinning: volumesLoading }" aria-hidden="true" />刷新
          </button>
        </div>

        <div v-if="volumesLoading" class="loading-state" role="status"><span class="spinner" /><span>正在读取数据卷…</span></div>
        <StatePanel v-else-if="volumesError" tone="error" title="无法读取数据卷" :description="`${volumesError}。请检查 SSH 连接和 Docker 服务。`">
          <template #icon><CircleAlert :size="20" /></template>
          <template #actions><button class="btn btn-secondary btn-sm" type="button" @click="reloadAll">重新加载</button></template>
        </StatePanel>
        <StatePanel v-else-if="volumes.length === 0" title="这台服务器还没有数据卷" description="创建数据卷，用于保留容器重建后仍需存在的数据。">
          <template #icon><Database :size="20" /></template>
          <template #actions><button class="btn btn-primary btn-sm" type="button" @click="openCreateModal">新建数据卷</button></template>
        </StatePanel>
        <StatePanel v-else-if="filteredVolumes.length === 0" title="没有匹配的数据卷" description="尝试搜索数据卷名称、驱动或挂载路径。">
          <template #icon><Search :size="20" /></template>
          <template #actions><button class="btn btn-secondary btn-sm" type="button" @click="searchQuery = ''">清除搜索</button></template>
        </StatePanel>
        <div v-else class="table-wrap">
          <table>
            <thead>
              <tr>
                <th class="selection-cell">
                  <input
                    class="row-checkbox"
                    type="checkbox"
                    :checked="allVisibleSelected"
                    :indeterminate="someVisibleSelected"
                    :disabled="batchDeleteLoading || deleteLoading"
                    aria-label="选择当前筛选结果中的全部数据卷"
                    @change="toggleVisibleVolumes($event.target.checked)"
                  />
                </th>
                <th>名称</th><th>驱动</th><th>挂载路径</th><th><span class="sr-only">操作</span></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="volume in filteredVolumes" :key="volumeValue(volume, 'Name', 'name')" :class="{ 'is-selected': selectedVolumeNames.includes(volumeValue(volume, 'Name', 'name')) }">
                <td class="selection-cell" data-label="选择">
                  <input
                    class="row-checkbox"
                    type="checkbox"
                    :checked="selectedVolumeNames.includes(volumeValue(volume, 'Name', 'name'))"
                    :disabled="batchDeleteLoading || deleteLoading"
                    :aria-label="`选择数据卷 ${volumeValue(volume, 'Name', 'name')}`"
                    @change="setVolumeSelected(volumeValue(volume, 'Name', 'name'), $event.target.checked)"
                  />
                </td>
                <td data-label="名称"><span class="table-primary mono">{{ volumeValue(volume, 'Name', 'name') }}</span></td>
                <td data-label="驱动"><span class="mono secondary-value">{{ volumeValue(volume, 'Driver', 'driver', 'local') }}</span></td>
                <td data-label="挂载路径"><span class="mono mountpoint">{{ volumeValue(volume, 'Mountpoint', 'mountpoint') }}</span></td>
                <td data-label="">
                  <div class="row-actions">
                    <button class="btn btn-danger btn-sm" type="button" :disabled="batchDeleteLoading || deleteLoading" @click="deleteTarget = volume">
                      <Trash2 :size="14" aria-hidden="true" />删除数据卷
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>

      <StatePanel v-else-if="!serversLoading && serversError" tone="error" title="无法读取服务器" :description="`${serversError}。请检查 API 服务后重试。`">
        <template #icon><CircleAlert :size="20" /></template>
        <template #actions><button class="btn btn-secondary btn-sm" type="button" @click="loadServers">重新加载</button></template>
      </StatePanel>
      <StatePanel v-else-if="!serversLoading" title="请先添加服务器" description="数据卷由具体服务器上的 Docker 服务管理。">
        <template #icon><ServerIcon :size="20" /></template>
        <template #actions><router-link class="btn btn-primary btn-sm" to="/servers">前往服务器</router-link></template>
      </StatePanel>
      <div v-else class="loading-state" role="status"><span class="spinner" /><span>正在读取服务器…</span></div>
    </section>

    <BaseModal v-if="createModal" title="新建数据卷" size="sm" :close-on-backdrop="!formLoading" @close="!formLoading && (createModal = false)">
      <div class="modal-form">
        <div v-if="formError" class="alert alert-error" role="alert"><CircleAlert :size="17" />{{ formError }}</div>
        <div class="form-group">
          <label class="form-label" for="volume-name">数据卷名称 <span class="required-mark">*</span></label>
          <input id="volume-name" v-model="volumeName" class="form-input mono" placeholder="例如：training-data" required @keyup.enter="createVolume" />
          <span class="form-hint">名称可包含字母、数字、点、连字符和下划线。</span>
        </div>
      </div>
      <template #footer>
        <button class="btn btn-secondary" type="button" :disabled="formLoading" @click="createModal = false">取消</button>
        <button class="btn btn-primary" type="button" :disabled="formLoading || !volumeName.trim()" @click="createVolume">
          <span v-if="formLoading" class="spinner button-spinner" />{{ formLoading ? '正在创建…' : '创建数据卷' }}
        </button>
      </template>
    </BaseModal>

    <ConfirmModal
      v-if="deleteTarget"
      title="删除数据卷"
      :message="`确定永久删除数据卷“${volumeValue(deleteTarget, 'Name', 'name')}”吗？`"
      detail="数据卷中的全部持久化数据将永久丢失；如果仍被容器使用，Docker 会拒绝删除。此操作无法撤销。"
      confirm-text="删除数据卷"
      :loading="deleteLoading"
      @confirm="deleteVolume"
      @cancel="deleteTarget = null"
    />

    <ConfirmModal
      v-if="batchDeleteConfirm"
      title="批量删除数据卷"
      :message="`确定永久删除选中的 ${selectedVolumeNames.length} 个数据卷吗？`"
      detail="这些数据卷中的全部持久化数据将永久丢失；仍被容器使用的数据卷会删除失败并保持选中，便于重试。此操作无法撤销。"
      confirm-text="批量删除"
      :loading="batchDeleteLoading"
      @confirm="deleteSelectedVolumes"
      @cancel="batchDeleteConfirm = false"
    />
  </div>
</template>

<style scoped>
.secondary-value {
  color: var(--ink-secondary);
  font-size: 12px;
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

.mountpoint {
  display: inline-block;
  max-width: 520px;
  overflow: hidden;
  color: var(--ink-secondary);
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.spinning {
  animation: spin 700ms linear infinite;
}

.modal-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
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
