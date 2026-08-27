import { ref } from 'vue'

const toasts = ref([])
let nextId = 0

export function useToast() {
  function add(message, type = 'error', duration = 4000) {
    const id = nextId++
    toasts.value.push({ id, message, type })
    if (duration > 0) {
      setTimeout(() => remove(id), duration)
    }
  }

  function remove(id) {
    const idx = toasts.value.findIndex(t => t.id === id)
    if (idx !== -1) toasts.value.splice(idx, 1)
  }

  function error(message) { add(message, 'error', 5000) }
  function success(message) { add(message, 'success', 3000) }
  return { toasts, remove, error, success }
}
