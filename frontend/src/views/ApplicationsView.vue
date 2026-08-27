<script setup>
import { ref, computed, onMounted } from 'vue'
import { get, post } from '@/api/client'
import { useToast } from '@/composables/useToast'
import StatusBadge from '@/components/StatusBadge.vue'
import BaseModal from '@/components/BaseModal.vue'

const toast = useToast()
const apps = ref([])
const loading = ref(true)
const activeFilter = ref('all')

const detailModal = ref(null)
const actionModal = ref(null)  // { app, action: 'approve' | 'reject' }
const adminNotes = ref('')
const actionLoading = ref(false)
const actionError = ref('')

const filters = ['all', 'pending', 'approved', 'rejected']

const filtered = computed(() => {
  if (activeFilter.value === 'all') return apps.value
  return apps.value.filter(a => a.status === activeFilter.value)
})

const counts = computed(() => {
  const c = { all: apps.value.length }
  for (const f of ['pending', 'approved', 'rejected']) {
    c[f] = apps.value.filter(a => a.status === f).length
  }
  return c
})

async function loadApps() {
  loading.value = true
  try {
    apps.value = await get('/applications') || []
  } catch (e) {
    toast.error(e.message)
  } finally {
    loading.value = false
  }
}

function openAction(app, action) {
  actionModal.value = { app, action }
  adminNotes.value = ''
  actionError.value = ''
}

async function doAction() {
  actionLoading.value = true
  actionError.value = ''
  try {
    await post(`/applications/${actionModal.value.app.id}/action`, {
      action: actionModal.value.action,
      admin_notes: adminNotes.value
    })
    const actionName = actionModal.value.action === 'approve' ? 'approved' : 'rejected'
    toast.success(`Application ${actionName} successfully`)
    actionModal.value = null
    loadApps()
  } catch (e) {
    actionError.value = e.message
  } finally {
    actionLoading.value = false
  }
}

function formatDate(str) {
  if (!str) return '-'
  return new Date(str).toLocaleString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

onMounted(loadApps)
</script>

<template>
  <div>
    <div class="page-header animate-in">
      <div>
        <h1 class="page-title">Applications</h1>
        <p class="page-subtitle">Container requests from users</p>
      </div>
    </div>

    <!-- Filter Tabs -->
    <div class="tabs animate-in animate-in-delay-1" style="margin-bottom:20px">
      <button
        v-for="f in filters"
        :key="f"
        class="tab-btn"
        :class="{ active: activeFilter === f }"
        @click="activeFilter = f"
      >
        {{ f.charAt(0).toUpperCase() + f.slice(1) }}
        <span class="tab-count">{{ counts[f] }}</span>
      </button>
    </div>

    <div class="card animate-in animate-in-delay-2">
      <div v-if="loading" class="loading-overlay">
        <span class="spinner" style="width:24px;height:24px" />
      </div>

      <div v-else-if="filtered.length === 0" class="empty-state">
        <div class="empty-state-icon">📋</div>
        <div class="empty-state-text">
          {{ activeFilter === 'all' ? 'No applications yet.' : `No ${activeFilter} applications.` }}
        </div>
      </div>

      <div v-else class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>#</th>
              <th>Applicant</th>
              <th>Server</th>
              <th>Image</th>
              <th>Status</th>
              <th>Submitted</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="app in filtered" :key="app.id">
              <td style="color:var(--text-muted);font-size:12px">#{{ app.id }}</td>
              <td>
                <div style="font-weight:500;font-size:13.5px">{{ app.applicant_name }}</div>
                <div style="font-size:12px;color:var(--text-muted)">{{ app.applicant_email }}</div>
              </td>
              <td style="font-size:13px">{{ app.server_host || '-' }}</td>
              <td style="font-size:13px">{{ app.image_name || '-' }}</td>
              <td><StatusBadge :status="app.status" /></td>
              <td style="color:var(--text-muted);font-size:12.5px">{{ formatDate(app.created_at) }}</td>
              <td>
                <div class="row-actions">
                  <button class="btn btn-ghost btn-sm" @click="detailModal = app">View</button>
                  <template v-if="app.status === 'pending'">
                    <button class="btn btn-success btn-sm" @click="openAction(app, 'approve')">Approve</button>
                    <button class="btn btn-danger btn-sm" @click="openAction(app, 'reject')">Reject</button>
                  </template>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Detail Modal -->
    <BaseModal v-if="detailModal" :title="`Application #${detailModal.id}`" size="md" @close="detailModal = null">
      <div class="detail-grid">
        <div class="detail-row">
          <span class="detail-label">Applicant</span>
          <span class="detail-value">{{ detailModal.applicant_name }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">Email</span>
          <span class="detail-value">{{ detailModal.applicant_email }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">Server</span>
          <span class="detail-value">{{ detailModal.server_host || '-' }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">Image</span>
          <span class="detail-value">{{ detailModal.image_name || '-' }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">Status</span>
          <span class="detail-value"><StatusBadge :status="detailModal.status" /></span>
        </div>
        <div class="detail-row">
          <span class="detail-label">Submitted</span>
          <span class="detail-value">{{ formatDate(detailModal.created_at) }}</span>
        </div>
        <div v-if="detailModal.admin_notes" class="detail-row">
          <span class="detail-label">Admin Notes</span>
          <span class="detail-value">{{ detailModal.admin_notes }}</span>
        </div>
      </div>
      <template #footer>
        <template v-if="detailModal.status === 'pending'">
          <button class="btn btn-danger btn-sm" @click="openAction(detailModal, 'reject'); detailModal = null">Reject</button>
          <button class="btn btn-success btn-sm" @click="openAction(detailModal, 'approve'); detailModal = null">Approve</button>
        </template>
        <button v-else class="btn btn-ghost btn-sm" @click="detailModal = null">Close</button>
      </template>
    </BaseModal>

    <!-- Approve / Reject Modal -->
    <BaseModal
      v-if="actionModal"
      :title="actionModal.action === 'approve' ? 'Approve Application' : 'Reject Application'"
      size="sm"
      @close="actionModal = null"
    >
      <div style="display:flex;flex-direction:column;gap:14px">
        <p style="font-size:13.5px;color:var(--text-secondary);line-height:1.6">
          <template v-if="actionModal.action === 'approve'">
            Approving will automatically create a Docker container for
            <strong>{{ actionModal.app.applicant_name }}</strong> and send them connection details via email.
          </template>
          <template v-else>
            Rejecting will notify <strong>{{ actionModal.app.applicant_name }}</strong> via email.
          </template>
        </p>
        <div v-if="actionError" class="alert alert-error">{{ actionError }}</div>
        <div class="form-group">
          <label class="form-label">Admin Notes (optional)</label>
          <textarea v-model="adminNotes" class="form-textarea" placeholder="Optional message to the applicant…" />
        </div>
      </div>
      <template #footer>
        <button class="btn btn-secondary" @click="actionModal = null" :disabled="actionLoading">Cancel</button>
        <button
          class="btn"
          :class="actionModal.action === 'approve' ? 'btn-success' : 'btn-danger'"
          @click="doAction"
          :disabled="actionLoading"
        >
          <span v-if="actionLoading" class="spinner" style="width:13px;height:13px;border-color:rgba(255,255,255,0.3);border-top-color:white" />
          {{ actionModal.action === 'approve' ? 'Approve & Create Container' : 'Reject' }}
        </button>
      </template>
    </BaseModal>
  </div>
</template>

<style scoped>
.tab-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 9px;
  background: var(--cream-dark);
  color: var(--text-muted);
  font-size: 11px;
  font-weight: 600;
  margin-left: 6px;
}

.tab-btn.active .tab-count {
  background: var(--accent-light);
  color: var(--accent);
}

.row-actions {
  display: flex;
  gap: 4px;
  justify-content: flex-end;
}

.row-actions .btn,
:deep(.modal-footer) .btn {
  border: none;
  box-shadow: none;
}

.detail-grid {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.detail-row {
  display: flex;
  padding: 11px 0;
  border-bottom: 1px solid var(--border-light);
  gap: 16px;
  align-items: baseline;
}

.detail-row:last-child {
  border-bottom: none;
}

.detail-label {
  width: 110px;
  flex-shrink: 0;
  font-size: 12.5px;
  font-weight: 500;
  color: var(--text-muted);
}

.detail-value {
  font-size: 13.5px;
  color: var(--text-primary);
  line-height: 1.5;
}

@media (max-width: 768px) {
  .row-actions {
    flex-wrap: wrap;
  }
}
</style>
