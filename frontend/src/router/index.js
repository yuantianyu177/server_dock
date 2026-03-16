import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/LoginView.vue'),
    meta: { public: true }
  },
  {
    path: '/apply',
    name: 'Apply',
    component: () => import('@/views/ApplyView.vue'),
    meta: { public: true }
  },
  {
    path: '/',
    component: () => import('@/components/AppLayout.vue'),
    redirect: '/servers',
    children: [
      {
        path: 'servers',
        name: 'Servers',
        component: () => import('@/views/ServersView.vue')
      },
      {
        path: 'servers/:id',
        name: 'ServerDetail',
        component: () => import('@/views/ServerDetailView.vue')
      },
      {
        path: 'containers',
        name: 'Containers',
        component: () => import('@/views/ContainersView.vue')
      },
      {
        path: 'images',
        name: 'Images',
        component: () => import('@/views/ImagesView.vue')
      },
      {
        path: 'volumes',
        name: 'Volumes',
        component: () => import('@/views/VolumesView.vue')
      },
      {
        path: 'applications',
        name: 'Applications',
        component: () => import('@/views/ApplicationsView.vue')
      },
      {
        path: 'config',
        name: 'Config',
        component: () => import('@/views/ConfigView.vue')
      }
    ]
  },
  {
    path: '/terminal/:serverId',
    name: 'Terminal',
    component: () => import('@/views/TerminalView.vue')
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.isAuthenticated) {
    return { name: 'Login', query: { redirect: to.fullPath } }
  }
  if (to.name === 'Login' && auth.isAuthenticated) {
    return { name: 'Servers' }
  }
})

export default router
