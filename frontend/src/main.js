import { createApp } from 'vue'
import App from './App.vue'
import { AUTH_UNAUTHORIZED_EVENT } from './api/client'
import router from './router'
import { auth } from './stores/auth'
import './style.css'

let redirectingToLogin = false

window.addEventListener(AUTH_UNAUTHORIZED_EVENT, () => {
  const currentRoute = router.currentRoute.value
  const redirect = currentRoute.fullPath

  auth.logout()

  if (
    redirectingToLogin ||
    currentRoute.name === 'Login' ||
    currentRoute.meta.public ||
    currentRoute.matched.length === 0
  ) return

  redirectingToLogin = true
  const finishRedirect = () => { redirectingToLogin = false }
  void router.replace({ name: 'Login', query: { redirect } }).then(finishRedirect, finishRedirect)
})

const app = createApp(App)
app.use(router)
app.mount('#app')
