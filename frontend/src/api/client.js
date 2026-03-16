import axios from 'axios'
import router from '@/router'

function unwrapResponseData(payload) {
  if (payload && typeof payload === 'object' && Object.prototype.hasOwnProperty.call(payload, 'data')) {
    return payload.data
  }
  return payload
}

const client = axios.create({
  baseURL: '/api',
  timeout: 60000
})

// Attach JWT token to every request
client.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Handle 401 globally
client.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      router.push('/login')
    }
    // Normalize error message
    const message = err.response?.data?.message || err.response?.data?.error || err.message || 'Unknown error'
    return Promise.reject(new Error(message))
  }
)

// Unwrap the standard { success, data } envelope
export async function get(url, params) {
  const res = await client.get(url, { params })
  return unwrapResponseData(res.data)
}

export async function post(url, data) {
  const res = await client.post(url, data)
  return unwrapResponseData(res.data)
}

export async function put(url, data) {
  const res = await client.put(url, data)
  return unwrapResponseData(res.data)
}

export async function del(url) {
  const res = await client.delete(url)
  return unwrapResponseData(res.data)
}

export default client
