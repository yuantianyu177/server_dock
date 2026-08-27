import { createRouter, createWebHistory } from 'vue-router'
import { auth } from '@/stores/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/LoginView.vue'),
    meta: { public: true, title: '管理员登录' }
  },
  {
    path: '/apply',
    name: 'Apply',
    component: () => import('@/views/ApplyView.vue'),
    meta: { public: true, title: '容器申请' }
  },
  {
    path: '/',
    component: () => import('@/components/AppLayout.vue'),
    redirect: '/servers',
    children: [
      {
        path: 'servers',
        name: 'Servers',
        component: () => import('@/views/ServersView.vue'),
        meta: { title: '服务器' }
      },
      {
        path: 'containers',
        name: 'Containers',
        component: () => import('@/views/ContainersView.vue'),
        meta: { title: '容器' }
      },
      {
        path: 'images',
        name: 'Images',
        component: () => import('@/views/ImagesView.vue'),
        meta: { title: '镜像' }
      },
      {
        path: 'volumes',
        name: 'Volumes',
        component: () => import('@/views/VolumesView.vue'),
        meta: { title: '数据卷' }
      },
      {
        path: 'applications',
        name: 'Applications',
        component: () => import('@/views/ApplicationsView.vue'),
        meta: { title: '申请审批' }
      },
      {
        path: 'config',
        name: 'Config',
        component: () => import('@/views/ConfigView.vue'),
        meta: { title: '系统设置' }
      }
    ]
  },
  {
    path: '/terminal/:serverId',
    name: 'Terminal',
    component: () => import('@/views/TerminalView.vue'),
    meta: { title: 'Web Terminal' }
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
  const isAuthenticated = auth.isAuthenticated

  if (auth.token && !isAuthenticated) auth.logout()

  if (!to.meta.public && !isAuthenticated) {
    return { name: 'Login', query: { redirect: to.fullPath } }
  }
  if (to.name === 'Login' && isAuthenticated) {
    return { name: 'Servers' }
  }
})

router.afterEach((to) => {
  document.title = `${to.meta.title || '基础设施控制台'} · ServerDock`
})

export default router
