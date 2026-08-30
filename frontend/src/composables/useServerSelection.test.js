import assert from 'node:assert/strict'
import test from 'node:test'
import { createSSRApp, h, nextTick } from 'vue'
import { renderToString } from '@vue/server-renderer'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createServer } from 'vite'

test('keeps the selected server when switching management pages', async (t) => {
  const originalWindow = globalThis.window
  const originalLocalStorage = Object.getOwnPropertyDescriptor(globalThis, 'localStorage')
  const originalFetch = globalThis.fetch

  globalThis.window = { location: { origin: 'http://localhost' } }
  Object.defineProperty(globalThis, 'localStorage', {
    configurable: true,
    value: {
      getItem: () => null,
      removeItem: () => {}
    }
  })
  globalThis.fetch = async () => new Response(JSON.stringify([
    { id: 1, host: 'Server 1' },
    { id: 2, host: 'Server 2' }
  ]), {
    status: 200,
    headers: { 'Content-Type': 'application/json' }
  })

  t.after(() => {
    if (originalWindow === undefined) delete globalThis.window
    else globalThis.window = originalWindow
    if (originalLocalStorage === undefined) delete globalThis.localStorage
    else Object.defineProperty(globalThis, 'localStorage', originalLocalStorage)
    globalThis.fetch = originalFetch
  })

  const vite = await createServer({
    root: process.cwd(),
    server: { middlewareMode: true },
    appType: 'custom'
  })
  t.after(() => vite.close())

  const { useServerSelection } = await vite.ssrLoadModule('/src/composables/useServerSelection.js')
  const EmptyView = { render: () => h('div') }
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/containers', component: EmptyView },
      { path: '/images', component: EmptyView }
    ]
  })

  await router.push('/containers')
  await router.isReady()

  let containersSelection
  let imagesSelection
  const app = createSSRApp({
    setup() {
      containersSelection = useServerSelection()
      imagesSelection = useServerSelection()
      return () => h('div')
    }
  })
  app.use(router)
  await renderToString(app)

  await containersSelection.loadServers()
  containersSelection.selectedServerId.value = 2
  await nextTick()
  await new Promise(resolve => setTimeout(resolve, 0))

  await router.push('/images')
  await imagesSelection.loadServers()

  assert.equal(imagesSelection.selectedServerId.value, 2)
})
