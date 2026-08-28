<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import {
  ArrowLeft,
  CircleCheck,
  Mail,
  Send,
  UserRound
} from '@lucide/vue'
import { get, post } from '@/api/client'
import { useToast } from '@/composables/useToast'
import BrandMark from '@/components/BrandMark.vue'

const toast = useToast()
const servers = ref([])
const images = ref([])
const serversLoading = ref(false)
const imagesLoading = ref(false)
const serversError = ref('')
const imagesError = ref('')

const form = ref({
  applicant_name: '',
  applicant_email: '',
  server_id: '',
  image_id: ''
})

const submitting = ref(false)
const submittedApplication = ref(null)

const hasServers = computed(() => servers.value.length > 0)
const hasImages = computed(() => images.value.length > 0)
const selectedServer = computed(() => servers.value.find(server => Number(server.id) === Number(form.value.server_id)))

const canSubmit = computed(() =>
  !submitting.value &&
  !imagesLoading.value &&
  !!form.value.applicant_name.trim() &&
  !!form.value.applicant_email.trim() &&
  !!form.value.server_id &&
  !!form.value.image_id
)

async function loadServers() {
  serversLoading.value = true
  serversError.value = ''
  try {
    servers.value = await get('/applications/public/servers') || []
    if (!servers.value.some(server => Number(server.id) === Number(form.value.server_id))) {
      form.value.server_id = ''
      form.value.image_id = ''
      images.value = []
    }
  } catch (error) {
    servers.value = []
    images.value = []
    serversError.value = `无法读取可申请服务器：${error.message}`
    toast.error(`${serversError.value}。请刷新页面后重试`)
  } finally {
    serversLoading.value = false
  }
}

async function loadImages(serverId) {
  form.value.image_id = ''
  imagesError.value = ''
  if (!serverId) {
    images.value = []
    return
  }

  imagesLoading.value = true
  try {
    images.value = await get(`/applications/public/server/${serverId}/images`) || []
  } catch (error) {
    images.value = []
    imagesError.value = `无法读取该服务器的镜像：${error.message}`
    toast.error(`${imagesError.value}。请重新选择服务器后重试`)
  } finally {
    imagesLoading.value = false
  }
}

function validateForm() {
  if (!form.value.applicant_name.trim()) return '请输入姓名。'
  if (!form.value.applicant_email.trim()) return '请输入邮箱地址。'
  if (!/^\S+@\S+\.\S+$/.test(form.value.applicant_email.trim())) return '请输入有效的邮箱地址。'
  if (!form.value.server_id) return '请选择服务器。'
  if (!form.value.image_id) return '请选择镜像。'
  return ''
}

async function submitApplication() {
  const validationError = validateForm()
  if (validationError) {
    toast.warning(validationError)
    return
  }

  submitting.value = true
  try {
    submittedApplication.value = await post('/applications/public/apply', {
      applicant_name: form.value.applicant_name.trim(),
      applicant_email: form.value.applicant_email.trim(),
      server_id: Number(form.value.server_id),
      image_id: Number(form.value.image_id)
    })
  } catch (error) {
    toast.error(`无法提交申请：${error.message}。请检查填写内容后重试`)
  } finally {
    submitting.value = false
  }
}

function resetApplication() {
  form.value = { applicant_name: '', applicant_email: '', server_id: '', image_id: '' }
  images.value = []
  submittedApplication.value = null
}

function formatServerOption(server) {
  const parts = [server.host]
  if (server.description?.trim()) parts.push(server.description.trim())
  parts.push(server.load_available
    ? `${server.running_containers || 0} 个容器正在运行`
    : '容器运行数暂不可用')
  return parts.join(' — ')
}

watch(() => form.value.server_id, loadImages)
onMounted(loadServers)
</script>

<template>
  <div class="apply-page">
    <header class="public-header">
      <router-link class="public-brand" to="/apply" aria-label="ServerDock 容器申请页">
        <BrandMark :size="34" />
        <span>ServerDock</span>
      </router-link>
      <router-link class="login-link" to="/login"><ArrowLeft :size="15" aria-hidden="true" />管理员登录</router-link>
    </header>

    <main class="apply-main">
      <aside class="process-panel" aria-labelledby="process-title">
        <p class="process-kicker">容器申请</p>
        <h1 id="process-title">申请一个容器</h1>
        <p class="process-description">选择服务器和管理员提供的镜像。审批通过后，连接地址、端口和密码会发送到你的邮箱。</p>

        <ol class="process-list">
          <li>
            <span class="step-number">1</span>
            <div><strong>填写申请</strong><span>选择服务器和镜像</span></div>
          </li>
          <li>
            <span class="step-number">2</span>
            <div><strong>管理员审核</strong><span>确认资源并创建容器</span></div>
          </li>
          <li>
            <span class="step-number">3</span>
            <div><strong>接收连接信息</strong><span>结果发送至申请邮箱</span></div>
          </li>
        </ol>
      </aside>

      <section class="application-panel" aria-labelledby="application-title">
        <template v-if="submittedApplication">
          <div class="success-state">
            <div class="success-heading">
              <div class="success-icon"><CircleCheck :size="24" aria-hidden="true" /></div>
              <div>
                <p class="success-kicker">提交成功</p>
                <h2 id="application-title">等待管理员审核</h2>
              </div>
            </div>

            <dl class="submission-summary" aria-label="本次申请内容">
              <div>
                <dt>服务器</dt>
                <dd>{{ submittedApplication.server_host || selectedServer?.host || '—' }}</dd>
              </div>
              <div>
                <dt>镜像</dt>
                <dd>{{ submittedApplication.image_name || '—' }}</dd>
              </div>
            </dl>

            <div class="success-note">
              <Mail :size="17" aria-hidden="true" />
              <span>管理员审核并创建容器后，连接信息会发送到 <strong>{{ submittedApplication.applicant_email || form.applicant_email }}</strong></span>
            </div>

            <button class="btn btn-primary" type="button" @click="resetApplication">再申请一个容器</button>
          </div>
        </template>

        <template v-else>
          <header class="application-heading">
            <h2 id="application-title">容器申请表</h2>
          </header>

          <form class="application-form" @submit.prevent="submitApplication">
            <fieldset>
              <legend><UserRound :size="15" aria-hidden="true" />申请人信息</legend>
              <div class="form-group">
                <label class="form-label" for="applicant-name">姓名 <span class="required-mark">*</span></label>
                <input id="applicant-name" v-model="form.applicant_name" class="form-input" autocomplete="name" placeholder="请输入真实姓名" :disabled="submitting" required />
              </div>
              <div class="form-group">
                <label class="form-label" for="applicant-email">邮箱 <span class="required-mark">*</span></label>
                <input id="applicant-email" v-model="form.applicant_email" type="email" class="form-input" autocomplete="email" placeholder="name@example.com" :disabled="submitting" required />
                <span class="form-hint">审批结果和容器连接信息会发送到此邮箱。</span>
              </div>
            </fieldset>

            <fieldset>
              <legend class="sr-only">容器环境</legend>

              <div class="form-group">
                <label class="form-label" for="application-server">服务器 <span class="required-mark">*</span></label>
                <select id="application-server" v-model.number="form.server_id" class="form-select" :disabled="submitting || serversLoading || !hasServers" required>
                  <option value="" disabled>{{ serversLoading ? '正在读取服务器…' : hasServers ? '选择服务器' : '当前没有可申请服务器' }}</option>
                  <option v-for="serverItem in servers" :key="serverItem.id" :value="serverItem.id">
                    {{ formatServerOption(serverItem) }}
                  </option>
                </select>
                <span v-if="!serversLoading && !serversError && !hasServers" class="form-error">管理员尚未配置可申请服务器。</span>
              </div>

              <div class="form-group">
                <label class="form-label" for="application-image">镜像 <span class="required-mark">*</span></label>
                <select id="application-image" v-model.number="form.image_id" class="form-select" :disabled="submitting || !form.server_id || imagesLoading || !hasImages" required>
                  <option value="" disabled>
                    {{ !form.server_id ? '请先选择服务器' : imagesLoading ? '正在读取镜像…' : hasImages ? '选择镜像' : '该服务器没有可申请镜像' }}
                  </option>
                  <option v-for="image in images" :key="image.id" :value="image.id">{{ image.name }}</option>
                </select>
                <span v-if="!imagesError && form.server_id && !imagesLoading && !hasImages" class="form-error">请联系管理员为该服务器登记可申请镜像。</span>
              </div>
            </fieldset>

            <button class="btn btn-primary submit-button" type="submit" :disabled="!canSubmit">
              <span v-if="submitting" class="spinner submit-spinner" aria-hidden="true" />
              <Send v-else :size="16" aria-hidden="true" />
              {{ submitting ? '正在提交…' : '提交容器申请' }}
            </button>
          </form>
        </template>
      </section>
    </main>
  </div>
</template>

<style scoped>
.apply-page {
  min-height: 100vh;
  min-height: 100dvh;
  padding: 0 42px;
  background: var(--canvas);
}

.public-header {
  width: 100%;
  max-width: 1120px;
  min-height: 76px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 0 auto;
  border-bottom: 1px solid var(--divider);
}

.public-brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-family: var(--font-display);
  color: var(--ink);
  font-size: 18px;
  font-weight: 720;
  letter-spacing: -0.02em;
}

.login-link {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  color: #0066cc;
  font-size: 13px;
  font-weight: 600;
}

.login-link:hover {
  text-decoration: underline;
  text-underline-offset: 3px;
}

.apply-main {
  width: 100%;
  max-width: 1000px;
  display: grid;
  grid-template-columns: minmax(280px, 0.72fr) minmax(440px, 1fr);
  align-items: start;
  gap: 76px;
  margin: 0 auto;
  padding: 58px 0 72px;
}

.process-panel {
  position: sticky;
  top: 42px;
  padding-top: 12px;
}

.process-kicker,
.success-kicker {
  margin-bottom: 10px;
  color: #0066cc;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.09em;
  text-transform: uppercase;
}

.process-panel h1 {
  max-width: 380px;
  color: var(--ink);
  font-family: var(--font-display);
  font-size: 36px;
  font-weight: 730;
  letter-spacing: -0.04em;
  line-height: 1.08;
}

.process-description {
  max-width: 390px;
  margin-top: 15px;
  color: var(--ink-secondary);
  font-size: 14px;
  line-height: 1.65;
}

.process-list {
  display: flex;
  flex-direction: column;
  gap: 0;
  margin: 38px 0 0;
  padding: 0;
  list-style: none;
}

.process-list li {
  position: relative;
  display: grid;
  grid-template-columns: 30px 1fr;
  align-items: start;
  gap: 11px;
  min-height: 63px;
}

.process-list li:not(:last-child)::after {
  content: "";
  position: absolute;
  top: 30px;
  bottom: 0;
  left: 14px;
  width: 1px;
  background: var(--divider);
}

.step-number {
  width: 30px;
  height: 30px;
  display: grid;
  z-index: 1;
  place-items: center;
  border: 1px solid var(--divider);
  border-radius: 9px;
  background: var(--surface);
  color: var(--ink-secondary);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 700;
}

.process-list strong,
.process-list div > span {
  display: block;
}

.process-list strong {
  margin-top: 2px;
  color: var(--ink);
  font-size: 13px;
}

.process-list div > span {
  margin-top: 2px;
  color: var(--ink-secondary);
  font-size: 11px;
}

.application-panel {
  overflow: hidden;
  border: 1px solid var(--divider);
  border-radius: var(--radius-modal);
  background: var(--surface);
}

.application-heading {
  padding: 26px 28px 20px;
  border-bottom: 1px solid var(--divider-subtle);
}

.application-heading h2,
.success-state h2 {
  color: var(--ink);
  font-size: 21px;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.application-form {
  display: flex;
  flex-direction: column;
  gap: 24px;
  padding: 24px 28px 28px;
}

fieldset {
  display: flex;
  flex-direction: column;
  gap: 15px;
  margin: 0;
  padding: 0;
  border: 0;
}

legend {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  margin-bottom: 14px;
  padding: 0;
  color: var(--ink);
  font-size: 13px;
  font-weight: 700;
}

.submit-button {
  width: 100%;
  min-height: 42px;
}

.submit-spinner {
  width: 14px;
  height: 14px;
  color: #fff;
}

.success-state {
  display: flex;
  flex-direction: column;
  padding: 38px 36px 34px;
}

.success-heading {
  display: grid;
  grid-template-columns: 52px 1fr;
  align-items: center;
  gap: 15px;
}

.success-icon {
  width: 52px;
  height: 52px;
  display: grid;
  place-items: center;
  border: 1px solid #b9dec8;
  border-radius: 14px;
  background: var(--success-soft);
  color: var(--success);
}

.success-heading .success-kicker {
  margin-bottom: 4px;
  color: var(--success);
}

.submission-summary {
  margin: 23px 0 0;
  border-top: 1px solid var(--divider-subtle);
  border-bottom: 1px solid var(--divider-subtle);
}

.submission-summary > div {
  min-height: 48px;
  display: grid;
  grid-template-columns: 78px minmax(0, 1fr);
  align-items: center;
  gap: 12px;
}

.submission-summary > div + div {
  border-top: 1px solid var(--divider-subtle);
}

.submission-summary dt {
  color: var(--ink-secondary);
  font-size: 12px;
}

.submission-summary dd {
  overflow-wrap: anywhere;
  color: var(--ink);
  font-size: 12px;
  font-weight: 600;
}

.success-note {
  display: grid;
  grid-template-columns: 20px 1fr;
  align-items: center;
  gap: 10px;
  margin-top: 22px;
  padding: 13px 14px;
  border-radius: var(--radius-control);
  background: #f6f6f8;
  color: var(--ink-secondary);
}

.success-note > svg {
  color: var(--ink-tertiary);
}

.success-note > span {
  min-width: 0;
  overflow-wrap: anywhere;
  color: var(--ink-secondary);
  font-size: 12px;
  line-height: 1.5;
}

.success-note strong {
  color: var(--ink);
  font-weight: 650;
}

.success-state .btn {
  align-self: center;
  margin-top: 24px;
}

@media (max-width: 880px) {
  .apply-page {
    padding: 0 24px;
  }

  .apply-main {
    grid-template-columns: 1fr;
    gap: 34px;
    max-width: 560px;
    padding: 42px 0 60px;
  }

  .process-panel {
    position: static;
    padding-top: 0;
  }

  .process-panel h1 {
    font-size: 32px;
  }

  .process-list {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
    margin-top: 28px;
  }

  .process-list li {
    grid-template-columns: 30px 1fr;
    min-height: 0;
  }

  .process-list li::after {
    display: none;
  }
}

@media (max-width: 580px) {
  .apply-page {
    padding: 0 12px;
    background: var(--mobile-canvas);
  }

  .public-header {
    width: auto;
    min-height: calc(62px + env(safe-area-inset-top));
    margin-right: -12px;
    margin-left: -12px;
    padding: env(safe-area-inset-top) 12px 0;
    border-bottom-color: rgba(173, 207, 218, 0.2);
    background: var(--mobile-dock);
  }

  .public-brand {
    gap: 8px;
    color: #fff;
    font-size: 16px;
  }

  .login-link {
    color: #76d2e8;
    font-size: 12px;
  }

  .apply-main {
    gap: 12px;
    padding: 18px 0 36px;
  }

  .process-panel {
    padding: 20px;
    overflow: hidden;
    border-radius: 20px;
    background: var(--mobile-dock);
    box-shadow: 0 10px 28px rgba(18, 37, 47, 0.14);
  }

  .process-kicker {
    margin-bottom: 6px;
    color: #76d2e8;
    font-family: var(--font-mono);
    font-size: 9px;
  }

  .process-panel h1 {
    color: #fff;
    font-size: 27px;
  }

  .process-description {
    margin-top: 9px;
    color: #b4c4cb;
    font-size: 12px;
    line-height: 1.55;
  }

  .process-list {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 6px;
    margin-top: 17px;
    padding-top: 15px;
    border-top: 1px solid rgba(255, 255, 255, 0.11);
  }

  .process-list li {
    display: flex;
    align-items: flex-start;
    flex-direction: column;
    gap: 7px;
  }

  .step-number {
    width: 25px;
    height: 25px;
    border-color: rgba(118, 210, 232, 0.26);
    border-radius: 8px;
    background: rgba(118, 210, 232, 0.09);
    color: #76d2e8;
    font-size: 9px;
  }

  .process-list strong {
    margin-top: 0;
    color: #f5fafb;
    font-size: 10px;
    line-height: 1.35;
  }

  .process-list div > span {
    display: none;
  }

  .application-panel {
    border-color: rgba(171, 189, 198, 0.66);
    border-radius: 20px;
    box-shadow: 0 6px 22px rgba(26, 56, 69, 0.07);
  }

  .application-heading,
  .application-form {
    padding-right: 20px;
    padding-left: 20px;
  }

  .success-state {
    padding: 30px 20px 26px;
  }
}
</style>
