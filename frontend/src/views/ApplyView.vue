<script setup>
import { computed, ref, onMounted, watch } from 'vue'
import { applicationsApi } from '@/api/applications'
import CustomSelect from '@/components/CustomSelect.vue'

const servers = ref([])
const images = ref([])
const serversLoading = ref(false)
const imagesLoading = ref(false)

const form = ref({
  applicant_name: '',
  applicant_email: '',
  server_id: '',
  image_id: ''
})

const submitting = ref(false)
const error = ref('')
const success = ref(false)

const hasServers = computed(() => servers.value.length > 0)
const hasImages = computed(() => images.value.length > 0)

const serverOptions = computed(() => servers.value.map(s => ({
  value: s.id,
  label: s.host,
  subtitle: (s.description ? s.description + ' · ' : '') + (s.container_count !== undefined ? `${s.container_count} running` : '')
})))

const imageOptions = computed(() => images.value.map(i => ({
  value: i.id,
  label: i.name
})))
const canSubmit = computed(() =>
  !submitting.value &&
  !imagesLoading.value &&
  hasServers.value &&
  !!form.value.server_id &&
  !!form.value.image_id
)

async function loadServers() {
  serversLoading.value = true
  error.value = ''
  try {
    servers.value = await applicationsApi.publicServers() || []
    if (!servers.value.find((srv) => srv.id === Number(form.value.server_id))) {
      form.value.server_id = ''
      form.value.image_id = ''
      images.value = []
    }
  } catch (e) {
    servers.value = []
    images.value = []
  } finally {
    serversLoading.value = false
  }
}

async function loadImages(serverId) {
  form.value.image_id = ''
  if (!serverId) {
    images.value = []
    imagesLoading.value = false
    return
  }
  imagesLoading.value = true
  try {
    images.value = await applicationsApi.publicImages(serverId) || []
  } catch (e) {
    images.value = []
  } finally {
    imagesLoading.value = false
  }
}

watch(() => form.value.server_id, (id) => {
  loadImages(id)
})

async function submit() {
  if (!form.value.applicant_name || !form.value.applicant_email || !form.value.server_id || !form.value.image_id) {
    error.value = 'Please fill in all required fields.'
    return
  }
  submitting.value = true
  error.value = ''
  try {
    await applicationsApi.apply({
      applicant_name: form.value.applicant_name,
      applicant_email: form.value.applicant_email,
      server_id: Number(form.value.server_id),
      image_id: Number(form.value.image_id)
    })
    success.value = true
  } catch (e) {
    error.value = e.message || 'Submission failed. Please try again.'
  } finally {
    submitting.value = false
  }
}

onMounted(loadServers)
</script>

<template>
  <div class="apply-page">
    <div class="apply-bg" aria-hidden="true">
      <div class="bg-blob bg-blob-1" />
      <div class="bg-blob bg-blob-2" />
    </div>

    <div class="apply-wrap">
      <!-- Page header (outside card) -->
      <div class="apply-top animate-in">
        <div class="apply-brand">
          <div class="apply-brand-icon">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
              <rect x="2" y="2" width="20" height="8" rx="2"/>
              <rect x="2" y="14" width="20" height="8" rx="2"/>
              <line x1="6" y1="6" x2="6.01" y2="6"/>
              <line x1="6" y1="18" x2="6.01" y2="18"/>
            </svg>
          </div>
          <span class="apply-brand-name">ServerDock</span>
        </div>
      </div>

      <!-- Card -->
      <div class="apply-card animate-in animate-in-delay-1">
        <div class="apply-card-header">
          <h2 class="apply-title">Request a Container</h2>
          <p class="apply-desc">Submit a request for access to a GPU server environment. You'll receive SSH connection details via email once approved.</p>
        </div>

        <!-- Success state -->
        <div v-if="success" class="apply-success">
          <div class="success-icon-wrap">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="20 6 9 17 4 12"/>
            </svg>
          </div>
          <h3 class="success-title">Application Submitted</h3>
          <p class="success-desc">Your request has been submitted. An administrator will review it and you'll receive an email notification with the result.</p>
          <button class="btn btn-secondary" style="margin-top:24px" @click="success = false; form = { applicant_name: '', applicant_email: '', server_id: '', image_id: '' }">
            Submit another request
          </button>
        </div>

        <!-- Form -->
        <form v-else class="apply-form" @submit.prevent="submit">
          <div v-if="error" class="alert alert-error">{{ error }}</div>

          <div class="form-section">
            <div class="form-section-label">Your information</div>
            <div class="form-group">
              <label class="form-label">Full Name <span class="required">*</span></label>
              <input v-model="form.applicant_name" class="form-input" placeholder="Zhang San" />
            </div>
            <div class="form-group">
              <label class="form-label">Email Address <span class="required">*</span></label>
              <input v-model="form.applicant_email" type="email" class="form-input" placeholder="you@example.com" />
              <span class="form-hint">Connection details will be sent to this address</span>
            </div>
          </div>

          <div class="form-section">
            <div class="form-section-label">Container configuration</div>
            <div class="form-group">
              <label class="form-label">Server <span class="required">*</span></label>
              <CustomSelect
                v-model="form.server_id"
                :options="serverOptions"
                :placeholder="serversLoading ? 'Loading servers…' : hasServers ? 'Select a server…' : 'No servers available'"
                :disabled="serversLoading || !hasServers"
                label-key="label"
                value-key="value"
                subtitle-key="subtitle"
              />
              <span v-if="!serversLoading && !hasServers" class="form-hint" style="color:var(--warning)">
                No servers are currently available for new requests.
              </span>
            </div>
            <div class="form-group">
              <label class="form-label">Image <span class="required">*</span></label>
              <CustomSelect
                v-model="form.image_id"
                :options="imageOptions"
                :placeholder="!hasServers ? 'No environments available' : !form.server_id ? 'Select a server first' : imagesLoading ? 'Loading…' : hasImages ? 'Select an image…' : 'No images available'"
                :disabled="!form.server_id || imagesLoading || !hasImages"
                label-key="label"
                value-key="value"
              />
              <span v-if="form.server_id && !hasImages && !imagesLoading" class="form-hint" style="color:var(--warning)">
                No images available for this server.
              </span>
            </div>
          </div>

          <button type="submit" class="btn btn-primary apply-submit" :disabled="!canSubmit">
            <span v-if="submitting" class="spinner" style="width:14px;height:14px;border-color:rgba(255,255,255,0.3);border-top-color:white" />
            {{ submitting ? 'Submitting…' : 'Submit Request' }}
          </button>
        </form>

        <div class="apply-card-footer">
          Already an admin?
          <a href="/login">Sign in</a>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.apply-page {
  min-height: 100vh;
  background: var(--cream);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 48px 24px 64px;
  position: relative;
  overflow: hidden;
}

.apply-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.bg-blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
}

.bg-blob-1 {
  width: 600px;
  height: 600px;
  background: radial-gradient(circle, rgba(201, 100, 66, 0.07) 0%, transparent 70%);
  top: -200px;
  right: -150px;
}

.bg-blob-2 {
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, rgba(201, 100, 66, 0.05) 0%, transparent 70%);
  bottom: -100px;
  left: -100px;
}

.apply-wrap {
  width: 100%;
  max-width: 560px;
  position: relative;
  z-index: 1;
}

.apply-top {
  margin-bottom: 24px;
}

.apply-brand {
  display: flex;
  align-items: center;
  gap: 9px;
}

.apply-brand-icon {
  width: 30px;
  height: 30px;
  background: var(--accent);
  border-radius: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  box-shadow: 0 2px 8px rgba(201, 100, 66, 0.28);
}

.apply-brand-name {
  font-family: var(--font-serif);
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: 0.02em;
}

.apply-card {
  background: white;
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: 0 4px 24px rgba(0,0,0,0.07), 0 1px 4px rgba(0,0,0,0.04);
}

.apply-card-header {
  padding: 28px 32px 24px;
  border-bottom: 1px solid var(--border-light);
}

.apply-title {
  font-size: 19px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 6px;
}

.apply-desc {
  font-size: 13.5px;
  color: var(--text-secondary);
  line-height: 1.6;
}

.apply-form {
  padding: 24px 32px 28px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.form-section-label {
  font-size: 11.5px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  color: var(--text-muted);
}

.required {
  color: var(--accent);
  margin-left: 1px;
}

.apply-submit {
  width: 100%;
  justify-content: center;
  padding: 11px;
  font-size: 14.5px;
  box-shadow: 0 1px 4px rgba(201, 100, 66, 0.25);
}

.apply-card-footer {
  padding: 16px 32px;
  border-top: 1px solid var(--border-light);
  text-align: center;
  font-size: 13px;
  color: var(--text-muted);
  background: var(--cream);
}

.apply-card-footer a {
  color: var(--accent);
  margin-left: 4px;
  font-weight: 500;
}

.apply-card-footer a:hover {
  opacity: 0.8;
}

/* Success state */
.apply-success {
  padding: 52px 32px;
  text-align: center;
}

.success-icon-wrap {
  width: 52px;
  height: 52px;
  background: var(--success-bg);
  color: var(--success);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 20px;
  border: 2px solid var(--success-border);
}

.success-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 10px;
}

.success-desc {
  font-size: 13.5px;
  color: var(--text-secondary);
  line-height: 1.65;
  max-width: 380px;
  margin: 0 auto;
}

@media (max-width: 768px) {
  .apply-page {
    padding: 24px 16px 48px;
  }

  .apply-card-header {
    padding: 20px 20px 18px;
  }

  .apply-form {
    padding: 18px 20px 22px;
  }

  .apply-card-footer {
    padding: 14px 20px;
  }

  .apply-success {
    padding: 36px 20px;
  }
}
</style>
