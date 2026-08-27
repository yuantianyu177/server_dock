import { reactive } from 'vue'
import { get, post } from '@/api/client'

const state = reactive({
  token: localStorage.getItem('token') || '',
  user: null
})

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
  get isAuthenticated() { return Boolean(state.token) },
  login,
  logout,
  fetchMe
}
