<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import {
  BellOff,
  Check,
  CircleAlert,
  CircleCheck,
  CircleX,
  ClipboardCheck,
  Copy,
  Eye,
  RefreshCw,
  Search,
  SquareTerminal,
  TriangleAlert,
  X
} from '@lucide/vue'
import { get, post } from '@/api/client'
import { useApplicationBadge } from '@/composables/useApplicationBadge'
import { useToast } from '@/composables/useToast'
import BaseModal from '@/components/BaseModal.vue'
import PageHeader from '@/components/PageHeader.vue'
import StatePanel from '@/components/StatePanel.vue'
import StatusBadge from '@/components/StatusBadge.vue'

const toast = useToast()
const { setPendingCount } = useApplicationBadge()
const applications = ref([])
const loading = ref(true)
const loadError = ref('')
const activeFilter = ref('all')
const searchQuery = ref('')

const detailModal = ref(null)
const processingAction = ref(null)
const approvalModal = ref(null)
const approvalCopied = ref(false)
let syncTimer = null
let syncing = false

const applicationSyncInterval = 5000

const filters = [
  { id: 'all', label: '全部' },
  { id: 'pending', label: '待处理' },
  { id: 'approved', label: '已批准' },
  { id: 'rejected', label: '已拒绝' },
  { id: 'ignored', label: '已忽略' }
]

const actionDefinitions = {
  approve: {
    loadingLabel: '批准中…',
    errorLabel: '无法批准申请'
  },
  reject: {
    loadingLabel: '拒绝中…',
    errorLabel: '无法拒绝申请'
  },
  ignore: {
    loadingLabel: '忽略中…',
    errorLabel: '无法忽略申请'
  }
}

const counts = computed(() => ({
  all: applications.value.length,
  pending: applications.value.filter(application => application.status === 'pending').length,
  approved: applications.value.filter(application => application.status === 'approved').length,
  rejected: applications.value.filter(application => application.status === 'rejected').length,
  ignored: applications.value.filter(application => application.status === 'ignored').length
}))

const filteredApplications = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  return applications.value.filter(application => {
    const matchesStatus = activeFilter.value === 'all' || application.status === activeFilter.value
    const matchesQuery = !query || [
      application.applicant_name,
      application.applicant_email,
      application.server_host,
      application.image_name
    ].some(value => String(value || '').toLowerCase().includes(query))
    return matchesStatus && matchesQuery
  })
})

const approvalCopyText = computed(() => {
  if (!approvalModal.value) return ''
  return [
    '连接信息',
    `服务器：${approvalModal.value.server}`,
    `用户：${approvalModal.value.user}`,
    `密码：${approvalModal.value.password}`,
    `SSH 端口：${approvalModal.value.ssh_port}`,
    `额外端口：${approvalModal.value.extra_ports}`,
    '',
    approvalModal.value.ssh_command
  ].join('\n')
})

function applyApplications(nextApplications) {
  applications.value = nextApplications
  loadError.value = ''
  setPendingCount(nextApplications)
  if (detailModal.value) {
    detailModal.value = nextApplications.find(application => application.id === detailModal.value.id) || null
  }
}

async function loadApplications() {
  loading.value = true
  loadError.value = ''
  try {
    applyApplications(await get('/applications') || [])
  } catch (error) {
    applications.value = []
    loadError.value = error.message
  } finally {
    loading.value = false
  }
}

async function syncApplications() {
  if (document.visibilityState !== 'visible' || loading.value || processingAction.value || syncing) return
  syncing = true
  try {
    applyApplications(await get('/applications') || [])
  } catch {
    // Keep the last successful view during background synchronization failures.
  } finally {
    syncing = false
  }
}

function isProcessing(application, action) {
  return processingAction.value?.applicationId === application.id && processingAction.value.action === action
}

async function submitAction(application, action) {
  if (processingAction.value) return
  processingAction.value = { applicationId: application.id, action }
  try {
    const response = await post(`/applications/${application.id}/action`, { action })
    if (detailModal.value?.id === application.id) detailModal.value = null
    if (action === 'approve' && response?.connection_info) {
      approvalModal.value = response.connection_info
      approvalCopied.value = false
    } else if (action === 'approve') {
      toast.warning('申请已批准，但接口未返回连接信息，请立即检查容器创建结果')
    } else if (action === 'ignore') {
      toast.success(`已忽略 ${application.applicant_name} 的申请，未发送邮件`)
    } else {
      toast.success(`已拒绝 ${application.applicant_name} 的申请`)
    }
    await loadApplications()
  } catch (error) {
    const suffix = action === 'approve' ? '。请检查服务器、镜像和端口配置。' : ''
    toast.error(`${actionDefinitions[action].errorLabel}：${error.message}${suffix}`)
  } finally {
    processingAction.value = null
  }
}

function closeApprovalModal() {
  approvalModal.value = null
  approvalCopied.value = false
}

function copyWithFallback(text) {
  const previousFocus = document.activeElement
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.setAttribute('readonly', '')
  textarea.style.position = 'fixed'
  textarea.style.opacity = '0'
  document.body.appendChild(textarea)
  textarea.select()
  const copied = document.execCommand('copy')
  textarea.remove()
  previousFocus?.focus?.()
  if (!copied) throw new Error('浏览器未允许复制')
}

async function copyApprovalInfo() {
  try {
    if (navigator.clipboard?.writeText && window.isSecureContext) {
      await navigator.clipboard.writeText(approvalCopyText.value)
    } else {
      copyWithFallback(approvalCopyText.value)
    }
    approvalCopied.value = true
    toast.success('连接信息已复制到剪贴板')
  } catch (error) {
    toast.error(`无法复制连接信息：${error.message}`)
  }
}

function formatDate(value, detail = false) {
  if (!value) return '—'
  return new Intl.DateTimeFormat('zh-CN', detail
    ? { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }
    : { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }
  ).format(new Date(value))
}

onMounted(() => {
  loadApplications()
  syncTimer = window.setInterval(syncApplications, applicationSyncInterval)
  document.addEventListener('visibilitychange', syncApplications)
})

onBeforeUnmount(() => {
  window.clearInterval(syncTimer)
  document.removeEventListener('visibilitychange', syncApplications)
})
</script>

<template>
  <div>
    <PageHeader title="申请审批" description="审核容器申请；批准会创建容器，忽略则静默归档，邮件审批结果会自动同步。">
      <template #actions>
        <button class="btn btn-secondary" type="button" :disabled="loading" @click="loadApplications">
          <RefreshCw :size="15" :class="{ spinning: loading }" aria-hidden="true" />刷新申请
        </button>
      </template>
    </PageHeader>

    <div class="tabs application-tabs" aria-label="按申请状态筛选">
      <button
        v-for="filter in filters"
        :key="filter.id"
        class="tab-btn"
        :class="{ active: activeFilter === filter.id }"
        type="button"
        :aria-pressed="activeFilter === filter.id"
        @click="activeFilter = filter.id"
      >
        {{ filter.label }}
        <span class="tab-count">{{ counts[filter.id] }}</span>
      </button>
    </div>

    <section class="data-panel responsive-table" aria-label="容器申请列表">
      <div class="table-toolbar">
        <div class="search-field">
          <Search :size="16" aria-hidden="true" />
          <label class="sr-only" for="application-search">搜索申请</label>
          <input id="application-search" v-model="searchQuery" class="form-input" placeholder="搜索申请人、邮箱、服务器或镜像" />
          <button v-if="searchQuery" class="search-clear" type="button" aria-label="清除搜索" @click="searchQuery = ''"><X :size="14" /></button>
        </div>
        <span class="result-count">显示 {{ filteredApplications.length }} 条</span>
      </div>

      <div v-if="loading" class="loading-state" role="status"><span class="spinner" /><span>正在读取申请…</span></div>
      <StatePanel v-else-if="loadError" tone="error" title="无法读取申请" :description="`${loadError}。请检查 API 服务后重试。`">
        <template #icon><CircleAlert :size="20" /></template>
        <template #actions><button class="btn btn-secondary btn-sm" type="button" @click="loadApplications">重新加载</button></template>
      </StatePanel>
      <StatePanel v-else-if="applications.length === 0" title="还没有容器申请" description="新申请提交后会出现在这里，按提交时间倒序排列。">
        <template #icon><ClipboardCheck :size="20" /></template>
        <template #actions><a class="btn btn-secondary btn-sm" href="/apply" target="_blank" rel="noopener">打开公开申请页</a></template>
      </StatePanel>
      <StatePanel v-else-if="filteredApplications.length === 0" title="没有匹配的申请" description="调整状态筛选或搜索词后重试。">
        <template #icon><Search :size="20" /></template>
        <template #actions><button class="btn btn-secondary btn-sm" type="button" @click="activeFilter = 'all'; searchQuery = ''">清除筛选</button></template>
      </StatePanel>
      <div v-else class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>申请人</th>
              <th>服务器</th>
              <th>镜像</th>
              <th>状态</th>
              <th>提交时间</th>
              <th><span class="sr-only">操作</span></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="application in filteredApplications" :key="application.id">
              <td data-label="申请人">
                <div class="table-primary">{{ application.applicant_name }}</div>
                <div class="table-secondary">{{ application.applicant_email }}</div>
              </td>
              <td data-label="服务器">{{ application.server_host || '—' }}</td>
              <td data-label="镜像">{{ application.image_name || '—' }}</td>
              <td data-label="状态"><StatusBadge :status="application.status" /></td>
              <td data-label="提交时间" class="date-cell">{{ formatDate(application.created_at) }}</td>
              <td data-label="">
                <div class="row-actions">
                  <button class="btn btn-ghost btn-sm" type="button" @click="detailModal = application">
                    <Eye :size="14" aria-hidden="true" />查看
                  </button>
                  <template v-if="application.status === 'pending'">
                    <button class="btn btn-secondary btn-sm" type="button" :disabled="processingAction !== null" @click="submitAction(application, 'ignore')">
                      <span v-if="isProcessing(application, 'ignore')" class="spinner inline-action-spinner" />
                      <BellOff v-else :size="14" aria-hidden="true" />
                      {{ isProcessing(application, 'ignore') ? actionDefinitions.ignore.loadingLabel : '忽略' }}
                    </button>
                    <button class="btn btn-danger btn-sm" type="button" :disabled="processingAction !== null" @click="submitAction(application, 'reject')">
                      <span v-if="isProcessing(application, 'reject')" class="spinner inline-action-spinner" />
                      <X v-else :size="14" aria-hidden="true" />
                      {{ isProcessing(application, 'reject') ? actionDefinitions.reject.loadingLabel : '拒绝' }}
                    </button>
                    <button class="btn btn-success btn-sm" type="button" :disabled="processingAction !== null" @click="submitAction(application, 'approve')">
                      <span v-if="isProcessing(application, 'approve')" class="spinner inline-action-spinner" />
                      <Check v-else :size="14" aria-hidden="true" />
                      {{ isProcessing(application, 'approve') ? actionDefinitions.approve.loadingLabel : '批准' }}
                    </button>
                  </template>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <BaseModal v-if="detailModal" title="申请详情" size="md" @close="detailModal = null">
      <dl class="detail-list">
        <div class="detail-row"><dt>申请人</dt><dd>{{ detailModal.applicant_name }}</dd></div>
        <div class="detail-row"><dt>邮箱</dt><dd>{{ detailModal.applicant_email }}</dd></div>
        <div class="detail-row"><dt>服务器</dt><dd>{{ detailModal.server_host || '—' }}</dd></div>
        <div class="detail-row"><dt>镜像</dt><dd>{{ detailModal.image_name || '—' }}</dd></div>
        <div class="detail-row"><dt>状态</dt><dd><StatusBadge :status="detailModal.status" /></dd></div>
        <div class="detail-row"><dt>提交时间</dt><dd>{{ formatDate(detailModal.created_at, true) }}</dd></div>
      </dl>
      <template #footer>
        <template v-if="detailModal.status === 'pending'">
          <button class="btn btn-secondary" type="button" :disabled="processingAction !== null" @click="submitAction(detailModal, 'ignore')">
            <span v-if="isProcessing(detailModal, 'ignore')" class="spinner inline-action-spinner" />
            <BellOff v-else :size="15" aria-hidden="true" />
            {{ isProcessing(detailModal, 'ignore') ? actionDefinitions.ignore.loadingLabel : '忽略申请' }}
          </button>
          <button class="btn btn-danger" type="button" :disabled="processingAction !== null" @click="submitAction(detailModal, 'reject')">
            <span v-if="isProcessing(detailModal, 'reject')" class="spinner inline-action-spinner" />
            <CircleX v-else :size="15" aria-hidden="true" />
            {{ isProcessing(detailModal, 'reject') ? actionDefinitions.reject.loadingLabel : '拒绝申请' }}
          </button>
          <button class="btn btn-primary" type="button" :disabled="processingAction !== null" @click="submitAction(detailModal, 'approve')">
            <span v-if="isProcessing(detailModal, 'approve')" class="spinner inline-action-spinner" />
            <CircleCheck v-else :size="15" aria-hidden="true" />
            {{ isProcessing(detailModal, 'approve') ? actionDefinitions.approve.loadingLabel : '批准并创建容器' }}
          </button>
        </template>
        <button v-else class="btn btn-secondary" type="button" @click="detailModal = null">关闭</button>
      </template>
    </BaseModal>

    <BaseModal
      v-if="approvalModal"
      title="申请已批准"
      size="md"
      :close-on-backdrop="false"
      @close="closeApprovalModal"
    >
      <div class="approval-result">
        <div class="approval-outcome">
          <span class="approval-outcome-icon"><CircleCheck :size="22" aria-hidden="true" /></span>
          <div>
            <strong>容器已创建</strong>
            <p>以下信息用于连接新容器。</p>
          </div>
        </div>

        <div class="alert alert-warning one-time-notice" role="note">
          <TriangleAlert :size="17" aria-hidden="true" />
          <span>这是一次性连接信息，关闭后无法从审批记录中再次查看本次密码，请先复制并妥善保管。</span>
        </div>

        <section class="connection-card" aria-labelledby="connection-title">
          <header class="connection-card-header">
            <h3 id="connection-title"><SquareTerminal :size="15" aria-hidden="true" />连接信息</h3>
          </header>
          <dl class="connection-details">
            <div><dt>服务器</dt><dd><code>{{ approvalModal.server }}</code></dd></div>
            <div><dt>用户</dt><dd><code>{{ approvalModal.user }}</code></dd></div>
            <div><dt>密码</dt><dd><code>{{ approvalModal.password }}</code></dd></div>
            <div><dt>SSH 端口</dt><dd><code>{{ approvalModal.ssh_port }}</code></dd></div>
            <div><dt>额外端口</dt><dd><code>{{ approvalModal.extra_ports }}</code></dd></div>
          </dl>
          <div class="connection-command">
            <pre>{{ approvalModal.ssh_command }}</pre>
          </div>
        </section>
      </div>
      <template #footer>
        <button class="btn btn-secondary" type="button" @click="closeApprovalModal">关闭</button>
        <button class="btn btn-primary" type="button" @click="copyApprovalInfo">
          <Check v-if="approvalCopied" :size="15" aria-hidden="true" />
          <Copy v-else :size="15" aria-hidden="true" />
          <span aria-live="polite">{{ approvalCopied ? '已复制' : '复制连接信息' }}</span>
        </button>
      </template>
    </BaseModal>
  </div>
</template>

<style scoped>
.application-tabs {
  margin-bottom: 14px;
}

.result-count,
.date-cell {
  color: var(--ink-secondary);
  font-size: 12px;
  white-space: nowrap;
}

.spinning {
  animation: spin 700ms linear infinite;
}

.detail-list {
  margin: 0;
}

.detail-row {
  display: grid;
  grid-template-columns: 116px minmax(0, 1fr);
  gap: 16px;
  padding: 12px 0;
  border-bottom: 1px solid var(--divider-subtle);
}

.detail-row:first-child {
  padding-top: 0;
}

.detail-row:last-child {
  padding-bottom: 0;
  border-bottom: 0;
}

.detail-row dt {
  color: var(--ink-secondary);
  font-size: 12px;
  font-weight: 600;
}

.detail-row dd {
  min-width: 0;
  margin: 0;
  color: var(--ink);
  font-size: 13px;
  overflow-wrap: anywhere;
}

.inline-action-spinner {
  width: 13px;
  height: 13px;
  color: currentColor;
}

.approval-result {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.approval-outcome {
  display: grid;
  grid-template-columns: 42px minmax(0, 1fr);
  align-items: start;
  gap: 12px;
}

.approval-outcome-icon {
  width: 42px;
  height: 42px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  background: var(--success-soft);
  color: var(--success);
}

.approval-outcome strong {
  display: block;
  margin-top: 1px;
  color: var(--ink);
  font-size: 15px;
}

.approval-outcome p {
  margin-top: 4px;
  color: var(--ink-secondary);
  font-size: 13px;
  line-height: 1.55;
}

.one-time-notice {
  margin: 0;
}

.connection-card {
  overflow: hidden;
  border: 1px solid var(--divider);
  border-radius: 12px;
  background: var(--surface);
}

.connection-card-header {
  min-height: 42px;
  display: flex;
  align-items: center;
  padding: 9px 13px;
  border-bottom: 1px solid var(--divider-subtle);
  background: var(--surface-subtle);
}

.connection-card-header h3 {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  margin: 0;
  color: var(--ink);
  font-size: 12px;
  font-weight: 700;
}

.connection-card-header svg {
  color: var(--blue);
}

.connection-details {
  margin: 0;
}

.connection-details > div {
  min-height: 42px;
  display: grid;
  grid-template-columns: 92px minmax(0, 1fr);
  align-items: center;
  gap: 12px;
  padding: 9px 13px;
  border-bottom: 1px solid var(--divider-subtle);
}

.connection-details dt {
  color: var(--ink-tertiary);
  font-size: 11px;
  font-weight: 600;
}

.connection-details dd {
  min-width: 0;
  margin: 0;
  overflow-wrap: anywhere;
}

.connection-details code {
  color: var(--ink);
  font-size: 12px;
  font-weight: 600;
}

.connection-command {
  padding: 12px 13px 14px;
}

.connection-command pre {
  margin: 0;
  padding: 12px 13px;
  overflow-x: auto;
  border-radius: 9px;
  background: var(--lens);
  color: #f5f5f7;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  font-size: 12px;
  line-height: 1.6;
}

@media (max-width: 520px) {
  .detail-row {
    grid-template-columns: 92px minmax(0, 1fr);
  }

  .approval-outcome {
    grid-template-columns: 36px minmax(0, 1fr);
  }

  .approval-outcome-icon {
    width: 36px;
    height: 36px;
  }
}
</style>
