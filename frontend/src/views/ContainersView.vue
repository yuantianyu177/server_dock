<script setup>
import { computed, ref, watch } from 'vue'
import {
  Box,
  CircleAlert,
  CircleStop,
  Play,
  Plus,
  RefreshCw,
  RotateCw,
  ScrollText,
  Search,
  Server as ServerIcon,
  SquareTerminal,
  Trash2,
  X
} from '@lucide/vue'
import { get, post } from '@/api/client'
import { useListSelection } from '@/composables/useListSelection'
import { useToast } from '@/composables/useToast'
import { useServerSelection } from '@/composables/useServerSelection'
import { runSettledBatch, summarizeBatchResults } from '@/utils/batch'
import { formatIPv4Ports, isContainerRunning, summarizeContainers } from '@/utils/docker'
import BaseModal from '@/components/BaseModal.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'
import PageHeader from '@/components/PageHeader.vue'
import ServerLens from '@/components/ServerLens.vue'
import StatePanel from '@/components/StatePanel.vue'
import StatusBadge from '@/components/StatusBadge.vue'

const toast = useToast()
const {
  servers,
  selectedServerId,
  serversLoading,
  serversError,
  loadServers
} = useServerSelection()

const containers = ref([])
const containersLoading = ref(false)
const containersError = ref('')
const lensState = ref('offline')
const searchQuery = ref('')
const statusFilter = ref('all')

const logsModal = ref(null)
const logsContent = ref('')
const logsLoading = ref(false)

const availableImages = ref([])
const availableImagesLoading = ref(false)
const createModal = ref(false)
const containerForm = ref({ name: '', image: '' })
const formLoading = ref(false)

const confirmAction = ref(null)
const actionLoading = ref(false)
const directActionKey = ref('')
const batchConfirmAction = ref('')

const actionLabels = {
  start: '启动',
  stop: '停止',
  restart: '重启',
  delete: '删除'
}

const summary = computed(() => summarizeContainers(containers.value))

const filteredContainers = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  return containers.value.filter(container => {
    const running = isContainerRunning(container)
    const matchesStatus = statusFilter.value === 'all' || (statusFilter.value === 'running' ? running : !running)
    const matchesQuery = !query || [container.name, container.image, container.ports, container.status]
      .some(value => String(value || '').toLowerCase().includes(query))
    return matchesStatus && matchesQuery
  })
})

const visibleContainerNames = computed(() => filteredContainers.value.map(container => container.name))
const {
  selectedItems: selectedContainerNames,
  allVisibleSelected,
  someVisibleSelected,
  setItemSelected: setContainerSelected,
  toggleVisibleItems: toggleVisibleContainers,
  retainAvailableItems: retainAvailableContainerNames,
  clearSelection: clearContainerSelection
} = useListSelection(visibleContainerNames)

async function loadContainers() {
  if (!selectedServerId.value) {
    containers.value = []
    containersError.value = ''
    lensState.value = 'offline'
    return
  }

  const serverId = Number(selectedServerId.value)
  containersLoading.value = true
  containersError.value = ''
  lensState.value = 'offline'
  try {
    const result = await get(`/servers/${serverId}/containers`) || []
    if (Number(selectedServerId.value) !== serverId) return
    containers.value = result
    retainAvailableContainerNames(result.map(container => container.name))
    lensState.value = 'online'
  } catch (error) {
    if (Number(selectedServerId.value) !== serverId) return
    containers.value = []
    containersError.value = error.message
    lensState.value = 'offline'
  } finally {
    if (Number(selectedServerId.value) === serverId) containersLoading.value = false
  }
}

async function loadAvailableImages() {
  if (!selectedServerId.value) return
  const serverId = Number(selectedServerId.value)
  availableImagesLoading.value = true
  try {
    const result = await get('/images', { server_id: serverId }) || []
    if (Number(selectedServerId.value) !== serverId) return
    availableImages.value = result
  } catch (error) {
    if (Number(selectedServerId.value) !== serverId) return
    availableImages.value = []
    toast.error(`无法读取可用镜像：${error.message}`)
  } finally {
    if (Number(selectedServerId.value) === serverId) availableImagesLoading.value = false
  }
}

function openCreateModal() {
  containerForm.value = { name: '', image: '' }
  createModal.value = true
  loadAvailableImages()
}

function requestContainerAction(container, action) {
  if (action === 'delete' || action === 'stop') {
    confirmAction.value = { container, action }
    return
  }
  doContainerAction(container, action)
}

async function doContainerAction(container, action) {
  const actionKey = `${container.name}:${action}`
  directActionKey.value = actionKey
  actionLoading.value = true
  try {
    await post(`/servers/${selectedServerId.value}/containers/${encodeURIComponent(container.name)}/action`, { action })
    toast.success(`已${actionLabels[action]}容器“${container.name}”`)
    confirmAction.value = null
    await loadContainers()
  } catch (error) {
    toast.error(`无法${actionLabels[action]}容器：${error.message}`)
  } finally {
    actionLoading.value = false
    directActionKey.value = ''
  }
}

function requestBatchContainerAction(action) {
  if (selectedContainerNames.value.length === 0 || actionLoading.value) return
  if (action === 'stop' || action === 'delete') {
    batchConfirmAction.value = action
    return
  }
  doBatchContainerAction(action)
}

async function doBatchContainerAction(action = batchConfirmAction.value) {
  const names = [...selectedContainerNames.value]
  if (!action || names.length === 0) return

  actionLoading.value = true
  const serverId = selectedServerId.value
  const results = await runSettledBatch(names, name =>
    post(`/servers/${serverId}/containers/${encodeURIComponent(name)}/action`, { action })
  )
  const {
    failedItems: failedNames,
    succeededCount,
    firstError
  } = summarizeBatchResults(names, results)
  selectedContainerNames.value = failedNames
  batchConfirmAction.value = ''

  if (failedNames.length === 0) {
    toast.success(`已批量${actionLabels[action]} ${succeededCount} 个容器`)
  } else {
    const summary = succeededCount > 0
      ? `已${actionLabels[action]} ${succeededCount} 个，${failedNames.length} 个失败`
      : `${failedNames.length} 个容器均未能${actionLabels[action]}`
    toast.error(firstError ? `${summary}：${firstError}` : summary)
  }

  try {
    await loadContainers()
  } finally {
    actionLoading.value = false
  }
}

async function showLogs(container) {
  logsModal.value = container
  logsContent.value = ''
  logsLoading.value = true
  try {
    const response = await get(`/servers/${selectedServerId.value}/containers/${encodeURIComponent(container.name)}/logs`, { tail: 200 })
    logsContent.value = response?.logs || response || '当前没有日志输出。'
  } catch (error) {
    logsContent.value = `无法读取日志：${error.message}`
  } finally {
    logsLoading.value = false
  }
}

async function createContainer() {
  const name = containerForm.value.name.trim()
  if (!name) {
    toast.warning('请输入容器名称')
    return
  }
  if (!/^[a-zA-Z0-9_-]+$/.test(name)) {
    toast.warning('容器名称只能包含字母、数字、连字符和下划线')
    return
  }
  if (!containerForm.value.image) {
    toast.warning('请选择镜像')
    return
  }

  formLoading.value = true
  try {
    await post(`/servers/${selectedServerId.value}/containers`, { ...containerForm.value, name })
    createModal.value = false
    toast.success(`已创建容器“${name}”`)
    await loadContainers()
  } catch (error) {
    toast.error(`无法创建容器：${error.message}`)
  } finally {
    formLoading.value = false
  }
}

function openTerminal(containerName = '') {
  const query = containerName ? `?container=${encodeURIComponent(containerName)}` : ''
  window.open(`/terminal/${selectedServerId.value}${query}`, '_blank', 'noopener')
}

function actionIsLoading(container, action) {
  return directActionKey.value === `${container.name}:${action}`
}

watch(selectedServerId, () => {
  searchQuery.value = ''
  statusFilter.value = 'all'
  clearContainerSelection()
  batchConfirmAction.value = ''
  loadContainers()
})
</script>

<template>
  <div>
    <PageHeader title="容器" description="查看运行状态、端口映射，并控制当前服务器上的 Docker 容器。" />

    <ServerLens
      v-model="selectedServerId"
      :servers="servers"
      :loading="serversLoading || actionLoading"
      :state="lensState"
      :summary="summary"
      :error="containersError"
      @retry="loadContainers"
    >
      <template #actions>
        <button class="btn btn-secondary" type="button" :disabled="!selectedServerId || lensState === 'offline'" @click="openTerminal()">
          <SquareTerminal :size="15" aria-hidden="true" />打开终端
        </button>
        <button class="btn btn-primary" type="button" :disabled="!selectedServerId || lensState === 'offline'" @click="openCreateModal">
          <Plus :size="15" aria-hidden="true" />新建容器
        </button>
      </template>
    </ServerLens>

    <section class="data-panel responsive-table" aria-label="容器列表">
      <template v-if="!serversLoading && !serversError && servers.length > 0">
        <div class="table-toolbar">
          <div class="table-tools bulk-table-tools">
            <div class="search-field">
              <Search :size="16" aria-hidden="true" />
              <label class="sr-only" for="container-search">搜索容器</label>
              <input id="container-search" v-model="searchQuery" class="form-input" placeholder="搜索名称、镜像或端口" />
              <button v-if="searchQuery" class="search-clear" type="button" aria-label="清除搜索" @click="searchQuery = ''">
                <X :size="14" aria-hidden="true" />
              </button>
            </div>
            <label class="sr-only" for="container-status-filter">按状态筛选</label>
            <select id="container-status-filter" v-model="statusFilter" class="form-select compact-select">
              <option value="all">全部状态</option>
              <option value="running">运行中</option>
              <option value="stopped">未运行</option>
            </select>
            <div class="batch-actions" aria-label="容器批量操作">
              <span class="batch-count" :data-count="selectedContainerNames.length">已选 {{ selectedContainerNames.length }} 项</span>
              <button class="btn btn-success btn-sm" type="button" :disabled="actionLoading || selectedContainerNames.length === 0" title="批量启动" aria-label="批量启动选中容器" @click="requestBatchContainerAction('start')">
                <Play :size="14" aria-hidden="true" /><span class="batch-action-label">启动</span>
              </button>
              <button class="btn btn-warning btn-sm" type="button" :disabled="actionLoading || selectedContainerNames.length === 0" title="批量停止" aria-label="批量停止选中容器" @click="requestBatchContainerAction('stop')">
                <CircleStop :size="14" aria-hidden="true" /><span class="batch-action-label">停止</span>
              </button>
              <button class="btn btn-secondary btn-sm" type="button" :disabled="actionLoading || selectedContainerNames.length === 0" title="批量重启" aria-label="批量重启选中容器" @click="requestBatchContainerAction('restart')">
                <RotateCw :size="14" aria-hidden="true" /><span class="batch-action-label">重启</span>
              </button>
              <button class="btn btn-danger btn-sm" type="button" :disabled="actionLoading || selectedContainerNames.length === 0" title="批量删除" aria-label="批量删除选中容器" @click="requestBatchContainerAction('delete')">
                <Trash2 :size="14" aria-hidden="true" /><span class="batch-action-label">删除</span>
              </button>
            </div>
          </div>
          <button class="btn btn-ghost btn-sm" type="button" :disabled="containersLoading" @click="loadContainers">
            <RefreshCw :size="14" :class="{ spinning: containersLoading }" aria-hidden="true" />刷新
          </button>
        </div>

        <div v-if="containersLoading" class="loading-state" role="status">
          <span class="spinner" aria-hidden="true" />
          <span>正在读取容器…</span>
        </div>

        <StatePanel
          v-else-if="containersError"
          tone="error"
          title="无法连接 Docker"
          :description="`${containersError}。请检查服务器连接和 Docker 服务后重试。`"
        >
          <template #icon><CircleAlert :size="20" /></template>
          <template #actions><button class="btn btn-secondary btn-sm" type="button" @click="loadContainers">重新加载</button></template>
        </StatePanel>

        <StatePanel
          v-else-if="containers.length === 0"
          title="这台服务器还没有容器"
          description="选择一个已登记的镜像，创建第一个容器。"
        >
          <template #icon><Box :size="20" /></template>
          <template #actions><button class="btn btn-primary btn-sm" type="button" @click="openCreateModal">新建容器</button></template>
        </StatePanel>

        <StatePanel
          v-else-if="filteredContainers.length === 0"
          title="没有匹配的容器"
          description="调整搜索词或状态筛选后重试。"
        >
          <template #icon><Search :size="20" /></template>
          <template #actions><button class="btn btn-secondary btn-sm" type="button" @click="searchQuery = ''; statusFilter = 'all'">清除筛选</button></template>
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
                    :disabled="actionLoading"
                    aria-label="选择当前筛选结果中的全部容器"
                    @change="toggleVisibleContainers($event.target.checked)"
                  />
                </th>
                <th>容器</th>
                <th>镜像</th>
                <th>状态</th>
                <th>端口映射</th>
                <th><span class="sr-only">操作</span></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="container in filteredContainers" :key="container.id || container.name" :class="{ 'is-selected': selectedContainerNames.includes(container.name) }">
                <td class="selection-cell" data-label="选择">
                  <input
                    class="row-checkbox"
                    type="checkbox"
                    :checked="selectedContainerNames.includes(container.name)"
                    :disabled="actionLoading"
                    :aria-label="`选择容器 ${container.name}`"
                    @change="setContainerSelected(container.name, $event.target.checked)"
                  />
                </td>
                <td data-label="容器"><span class="table-primary mono">{{ container.name }}</span></td>
                <td data-label="镜像"><span class="image-address mono">{{ container.image }}</span></td>
                <td data-label="状态">
                  <StatusBadge :status="container.status" />
                  <div class="table-secondary status-detail">{{ container.status }}</div>
                </td>
                <td data-label="端口映射"><span class="ports mono">{{ formatIPv4Ports(container.ports) }}</span></td>
                <td data-label="">
                  <div class="row-actions">
                    <button class="btn btn-ghost btn-sm btn-icon" type="button" title="打开容器终端" :aria-label="`打开 ${container.name} 的终端`" @click="openTerminal(container.name)">
                      <SquareTerminal :size="15" aria-hidden="true" />
                    </button>
                    <button class="btn btn-ghost btn-sm" type="button" @click="showLogs(container)">
                      <ScrollText :size="14" aria-hidden="true" />日志
                    </button>
                    <button
                      v-if="!isContainerRunning(container)"
                      class="btn btn-success btn-sm"
                      type="button"
                      :disabled="actionLoading"
                      @click="requestContainerAction(container, 'start')"
                    >
                      <span v-if="actionIsLoading(container, 'start')" class="spinner action-spinner" aria-hidden="true" />
                      <Play v-else :size="14" aria-hidden="true" />启动
                    </button>
                    <button v-else class="btn btn-warning btn-sm" type="button" :disabled="actionLoading" @click="requestContainerAction(container, 'stop')">
                      <CircleStop :size="14" aria-hidden="true" />停止
                    </button>
                    <button class="btn btn-ghost btn-sm btn-icon" type="button" :disabled="actionLoading" title="重启容器" :aria-label="`重启 ${container.name}`" @click="requestContainerAction(container, 'restart')">
                      <span v-if="actionIsLoading(container, 'restart')" class="spinner action-spinner" aria-hidden="true" />
                      <RotateCw v-else :size="15" aria-hidden="true" />
                    </button>
                    <button class="btn btn-danger btn-sm btn-icon" type="button" :disabled="actionLoading" title="删除容器" :aria-label="`删除 ${container.name}`" @click="requestContainerAction(container, 'delete')">
                      <Trash2 :size="15" aria-hidden="true" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>

      <StatePanel
        v-else-if="!serversLoading && serversError"
        tone="error"
        title="无法读取服务器"
        :description="`${serversError}。请检查 API 服务后重试。`"
      >
        <template #icon><CircleAlert :size="20" /></template>
        <template #actions><button class="btn btn-secondary btn-sm" type="button" @click="loadServers">重新加载</button></template>
      </StatePanel>

      <StatePanel
        v-else-if="!serversLoading"
        title="请先添加服务器"
        description="容器需要运行在已配置 SSH 连接的服务器上。"
      >
        <template #icon><ServerIcon :size="20" /></template>
        <template #actions><router-link class="btn btn-primary btn-sm" to="/servers">前往服务器</router-link></template>
      </StatePanel>

      <div v-else class="loading-state" role="status"><span class="spinner" aria-hidden="true" /><span>正在读取服务器…</span></div>
    </section>

    <BaseModal v-if="createModal" title="新建容器" size="md" :close-on-backdrop="!formLoading" @close="!formLoading && (createModal = false)">
      <div class="modal-form">
        <div class="form-group">
          <label class="form-label" for="container-name">容器名称 <span class="required-mark">*</span></label>
          <input id="container-name" v-model="containerForm.name" class="form-input mono" placeholder="例如：training-zhangsan" autocomplete="off" required />
          <span class="form-hint">可使用字母、数字、连字符和下划线。</span>
        </div>
        <div class="form-group">
          <label class="form-label" for="container-image">镜像 <span class="required-mark">*</span></label>
          <select id="container-image" v-model="containerForm.image" class="form-select" :disabled="availableImagesLoading || availableImages.length === 0" required>
            <option value="" disabled>
              {{ availableImagesLoading ? '正在读取镜像…' : availableImages.length ? '选择镜像' : '当前服务器没有可用镜像' }}
            </option>
            <option v-for="image in availableImages" :key="image.id" :value="image.image_address">
              {{ image.name }} — {{ image.image_address }}
            </option>
          </select>
          <span v-if="!availableImagesLoading && availableImages.length === 0" class="form-error">请先在镜像页登记一个可申请镜像。</span>
        </div>
      </div>
      <template #footer>
        <button class="btn btn-secondary" type="button" :disabled="formLoading" @click="createModal = false">取消</button>
        <button class="btn btn-primary" type="button" :disabled="formLoading || availableImagesLoading || availableImages.length === 0" @click="createContainer">
          <span v-if="formLoading" class="spinner button-spinner" aria-hidden="true" />
          {{ formLoading ? '正在创建…' : '创建容器' }}
        </button>
      </template>
    </BaseModal>

    <BaseModal v-if="logsModal" :title="`容器日志 · ${logsModal.name}`" size="lg" @close="logsModal = null">
      <div v-if="logsLoading" class="logs-loading" role="status"><span class="spinner" aria-hidden="true" />正在读取最近 200 行日志…</div>
      <pre v-else class="logs-output" tabindex="0">{{ logsContent }}</pre>
    </BaseModal>

    <ConfirmModal
      v-if="confirmAction"
      :title="confirmAction.action === 'delete' ? '删除容器' : '停止容器'"
      :message="`${confirmAction.action === 'delete' ? '确定删除' : '确定停止'}容器“${confirmAction.container.name}”吗？`"
      :detail="confirmAction.action === 'delete'
        ? '容器将被强制删除；未写入数据卷的数据会永久丢失。'
        : '容器中的运行进程会终止，已挂载数据卷中的数据不受影响。'"
      :confirm-text="confirmAction.action === 'delete' ? '删除容器' : '停止容器'"
      :confirm-class="confirmAction.action === 'delete' ? 'btn-danger' : 'btn-warning'"
      :loading="actionLoading"
      @confirm="doContainerAction(confirmAction.container, confirmAction.action)"
      @cancel="confirmAction = null"
    />

    <ConfirmModal
      v-if="batchConfirmAction"
      :title="batchConfirmAction === 'delete' ? '批量删除容器' : '批量停止容器'"
      :message="`确定${batchConfirmAction === 'delete' ? '删除' : '停止'}选中的 ${selectedContainerNames.length} 个容器吗？`"
      :detail="batchConfirmAction === 'delete'
        ? '选中的容器将被强制删除；未写入数据卷的数据会永久丢失。执行失败的容器会保留选中，便于重试。'
        : '选中容器内的运行进程会终止，已挂载数据卷中的数据不受影响。'"
      :confirm-text="batchConfirmAction === 'delete' ? '批量删除' : '批量停止'"
      :confirm-class="batchConfirmAction === 'delete' ? 'btn-danger' : 'btn-warning'"
      :loading="actionLoading"
      @confirm="doBatchContainerAction()"
      @cancel="batchConfirmAction = ''"
    />
  </div>
</template>

<style scoped>
.image-address {
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

.table-primary {
  white-space: nowrap;
}

.ports {
  color: var(--ink-secondary);
  font-size: 11px;
  line-height: 1.5;
}

.status-detail {
  max-width: 210px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.spinning {
  animation: spin 700ms linear infinite;
}

.action-spinner,
.button-spinner {
  width: 13px;
  height: 13px;
  color: currentColor;
}

.button-spinner {
  color: #fff;
}

.modal-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.logs-loading {
  min-height: 260px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: var(--ink-secondary);
  font-size: 13px;
}

.logs-output {
  min-height: 300px;
  max-height: min(62vh, 560px);
  margin: 0;
  padding: 16px;
  overflow: auto;
  border: 1px solid #2e3036;
  border-radius: var(--radius-control);
  outline: none;
  background: var(--terminal);
  color: #d7d7dc;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}

@media (max-width: 680px) {
  .bulk-table-tools {
    width: 100%;
    display: grid;
    grid-template-columns: minmax(0, 1fr) 104px;
    gap: 8px;
    overflow: visible;
  }

  .bulk-table-tools .search-field {
    width: 100%;
    min-width: 0;
    flex: none;
  }

  .bulk-table-tools .compact-select {
    width: 104px;
    min-width: 104px;
    flex: none;
  }

  .batch-actions {
    grid-column: 1 / -1;
    justify-content: flex-end;
    padding-top: 8px;
    border-top: 1px solid var(--divider-subtle);
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
    width: 34px;
    padding: 0;
  }

  .batch-action-label {
    display: none;
  }
}
</style>
