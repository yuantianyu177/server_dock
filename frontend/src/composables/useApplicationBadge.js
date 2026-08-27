import { ref } from 'vue'
import { get } from '@/api/client'

const pendingCount = ref(0)
let pendingRequest = null

function setPendingCount(applicationsOrCount) {
  if (Array.isArray(applicationsOrCount)) {
    pendingCount.value = applicationsOrCount.filter(application => application.status === 'pending').length
    return
  }

  const count = Number(applicationsOrCount)
  pendingCount.value = Number.isFinite(count) ? Math.max(0, count) : 0
}

async function refreshPendingCount() {
  if (pendingRequest) return pendingRequest

  pendingRequest = get('/applications')
    .then(applications => {
      setPendingCount(applications || [])
      return pendingCount.value
    })
    .catch(() => pendingCount.value)
    .finally(() => {
      pendingRequest = null
    })

  return pendingRequest
}

export function useApplicationBadge() {
  return { pendingCount, refreshPendingCount, setPendingCount }
}
