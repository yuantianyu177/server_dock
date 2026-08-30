import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { get } from '@/api/client'

let rememberedServerId = null

export function useServerSelection() {
  const route = useRoute()
  const router = useRouter()
  const servers = ref([])
  const currentServerId = ref(null)
  const selectedServerId = computed({
    get: () => currentServerId.value,
    set: (id) => {
      currentServerId.value = id
      if (id) rememberedServerId = id
    }
  })
  const serversLoading = ref(true)
  const serversError = ref('')

  const selectedServer = computed(() =>
    servers.value.find(server => Number(server.id) === Number(selectedServerId.value)) || null
  )

  async function loadServers() {
    serversLoading.value = true
    serversError.value = ''
    try {
      servers.value = await get('/servers') || []
      const queryId = Number(route.query.server)
      const queryServer = servers.value.find(server => Number(server.id) === queryId)
      const rememberedServer = servers.value.find(server => Number(server.id) === Number(rememberedServerId))
      selectedServerId.value = queryServer?.id || rememberedServer?.id || servers.value[0]?.id || null
    } catch (error) {
      servers.value = []
      selectedServerId.value = null
      serversError.value = error.message
    } finally {
      serversLoading.value = false
    }
  }

  watch(selectedServerId, (id) => {
    if (!id) return
    if (Number(route.query.server) === Number(id)) return
    router.replace({ query: { ...route.query, server: id } })
  })

  watch(() => route.query.server, (value) => {
    const id = Number(value)
    if (id && servers.value.some(server => Number(server.id) === id)) {
      selectedServerId.value = id
    }
  })

  onMounted(loadServers)

  return {
    servers,
    selectedServerId,
    selectedServer,
    serversLoading,
    serversError,
    loadServers
  }
}
