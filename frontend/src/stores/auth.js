import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(null)

  const isAuthenticated = computed(() => !!token.value)

  function setToken(t) {
    token.value = t
    localStorage.setItem('token', t)
  }

  async function login(username, password) {
    const res = await authApi.login(username, password)
    setToken(res.token)
    user.value = { username }
    return res
  }

  async function fetchMe() {
    const res = await authApi.me()
    user.value = res
    return res
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
  }

  return { token, user, isAuthenticated, login, logout, fetchMe }
})
