import { ref } from 'vue'
import { get } from '@/api/client'
import { summarizeContainers } from '@/utils/docker'

export function useServerDockerSummary(selectedServerId) {
  const lensState = ref('offline')
  const summary = ref({ running: 0, total: 0 })

  function isCurrentServer(serverId) {
    return Number(selectedServerId.value) === serverId
  }

  function resetSummary() {
    lensState.value = 'offline'
    summary.value = { running: 0, total: 0 }
  }

  async function loadSummary() {
    if (!selectedServerId.value) return

    const serverId = Number(selectedServerId.value)
    lensState.value = 'offline'
    try {
      const containers = await get(`/servers/${serverId}/containers`) || []
      if (!isCurrentServer(serverId)) return
      summary.value = summarizeContainers(containers)
      lensState.value = 'online'
    } catch {
      if (isCurrentServer(serverId)) resetSummary()
    }
  }

  return { lensState, summary, loadSummary, resetSummary }
}
