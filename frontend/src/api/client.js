async function request(path, { method = 'GET', params, data } = {}) {
  const url = new URL(`/api${path}`, window.location.origin)
  for (const [key, value] of Object.entries(params || {})) {
    if (value !== undefined && value !== null) url.searchParams.set(key, value)
  }

  const headers = {}
  const token = localStorage.getItem('token')
  if (token) headers.Authorization = `Bearer ${token}`
  if (data !== undefined) headers['Content-Type'] = 'application/json'

  const response = await fetch(url, {
    method,
    headers,
    body: data === undefined ? undefined : JSON.stringify(data),
    signal: AbortSignal.timeout(60000)
  })

  if (response.status === 401) {
    localStorage.removeItem('token')
    if (location.pathname !== '/login') location.assign('/login')
  }

  const text = await response.text()
  let body = null
  if (text) {
    try { body = JSON.parse(text) } catch { body = text }
  }
  if (!response.ok) {
    throw new Error(body?.error || response.statusText || 'Request failed')
  }
  return body
}

export const get = (path, params) => request(path, { params })
export const post = (path, data) => request(path, { method: 'POST', data })
export const put = (path, data) => request(path, { method: 'PUT', data })
export const del = (path) => request(path, { method: 'DELETE' })
