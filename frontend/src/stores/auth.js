import { reactive } from 'vue'
import { get, post } from '@/api/client'

const state = reactive({
  token: localStorage.getItem('token') || '',
  user: null
})

function hasUnexpiredToken(token) {
  if (!token) return false

  try {
    const payload = token.split('.')[1]
    if (!payload) return false

    const base64 = payload.replace(/-/g, '+').replace(/_/g, '/')
    const padded = base64.padEnd(Math.ceil(base64.length / 4) * 4, '=')
    const claims = JSON.parse(atob(padded))
    return Number.isFinite(claims.exp) && claims.exp * 1000 > Date.now()
  } catch {
    return false
  }
}

async function login(username, password) {
  const response = await post('/auth/login', { username, password })
  state.token = response.token
  localStorage.setItem('token', response.token)
  state.user = { username }
  return response
}

async function fetchMe() {
  state.user = await get('/auth/me')
  return state.user
}

function logout() {
  state.token = ''
  state.user = null
  localStorage.removeItem('token')
}

export const auth = {
  get token() { return state.token },
  get user() { return state.user },
  get isAuthenticated() { return hasUnexpiredToken(state.token) },
  login,
  logout,
  fetchMe
}
