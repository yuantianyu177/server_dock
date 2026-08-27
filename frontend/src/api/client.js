export const AUTH_UNAUTHORIZED_EVENT = 'serverdock:auth-unauthorized'

async function request(path, { method = 'GET', params, data } = {}) {
  const url = new URL(`/api${path}`, window.location.origin)
  for (const [key, value] of Object.entries(params || {})) {
    if (value !== undefined && value !== null) url.searchParams.set(key, value)
  }

  const headers = {}
  const token = localStorage.getItem('token')
  if (token) headers.Authorization = `Bearer ${token}`
  if (data !== undefined) headers['Content-Type'] = 'application/json'

  let response
  try {
    response = await fetch(url, {
      method,
      headers,
      body: data === undefined ? undefined : JSON.stringify(data),
      signal: AbortSignal.timeout(60000)
    })
  } catch (error) {
    if (error.name === 'TimeoutError') throw new Error('服务响应超时，请稍后重试')
    throw new Error('无法连接 ServerDock API，请检查服务是否正在运行')
  }

  if (response.status === 401) {
    localStorage.removeItem('token')
    window.dispatchEvent(new Event(AUTH_UNAUTHORIZED_EVENT))
  }

  const text = await response.text()
  let body = null
  if (text) {
    try { body = JSON.parse(text) } catch { body = text }
  }
  if (!response.ok) {
    throw new Error(localizeError(body?.error || response.statusText || '请求失败'))
  }
  return body
}

function localizeError(message) {
  const text = String(message || '')
  const exact = {
    'invalid credentials': '用户名或密码错误',
    'invalid token': '登录已过期，请重新登录',
    'invalid or expired token': '登录已过期，请重新登录',
    'authorization header required': '登录已过期，请重新登录',
    'invalid authorization format': '登录已过期，请重新登录',
    'server not found': '未找到该服务器',
    'image not found': '未找到该镜像',
    'application not found': '未找到该申请',
    'application is not pending': '该申请已处理，不能重复审批',
    'old password is incorrect': '当前密码不正确',
    'admin_email not configured': '尚未配置管理员邮箱',
    'cannot delete: image is registered as available in the system': '该镜像仍登记为可申请镜像，请先取消登记',
    'invalid container name': '容器名称格式无效',
    'image is required': '请选择镜像',
    'failed to change password': '密码更新失败'
  }
  if (exact[text]) return exact[text]
  if (text.toLowerCase().includes('volume is in use')) return '该数据卷正在被容器使用，请先解除容器挂载或删除相关容器后重试'
  if (text.startsWith('connection failed:')) return `连接失败：${text.slice('connection failed:'.length).trim()}`
  if (text.startsWith('not enough available ports:')) return `可用端口不足：${text.slice('not enough available ports:'.length).trim()}`
  if (text.startsWith('container provisioning failed:')) return `容器创建失败：${text.slice('container provisioning failed:'.length).trim()}`
  if (text.startsWith('failed to send test email:')) return `测试邮件发送失败：${text.slice('failed to send test email:'.length).trim()}`
  if (text.startsWith('invalid request:')) return `请求内容无效：${text.slice('invalid request:'.length).trim()}`
  return text
}

export const get = (path, params) => request(path, { params })
export const post = (path, data) => request(path, { method: 'POST', data })
export const put = (path, data) => request(path, { method: 'PUT', data })
export const del = (path) => request(path, { method: 'DELETE' })
