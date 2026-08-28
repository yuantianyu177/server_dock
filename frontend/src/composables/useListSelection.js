import { computed, ref } from 'vue'

export function useListSelection(visibleItems) {
  const selectedItems = ref([])

  const allVisibleSelected = computed(() =>
    visibleItems.value.length > 0 && visibleItems.value.every(item => selectedItems.value.includes(item))
  )
  const someVisibleSelected = computed(() =>
    !allVisibleSelected.value && visibleItems.value.some(item => selectedItems.value.includes(item))
  )

  function setItemSelected(item, selected) {
    selectedItems.value = selected
      ? [...new Set([...selectedItems.value, item])]
      : selectedItems.value.filter(selectedItem => selectedItem !== item)
  }

  function toggleVisibleItems(selected) {
    const visibleSet = new Set(visibleItems.value)
    selectedItems.value = selected
      ? [...new Set([...selectedItems.value, ...visibleSet])]
      : selectedItems.value.filter(item => !visibleSet.has(item))
  }

  function retainAvailableItems(availableItems) {
    const availableSet = new Set(availableItems)
    selectedItems.value = selectedItems.value.filter(item => availableSet.has(item))
  }

  function clearSelection() {
    selectedItems.value = []
  }

  return {
    selectedItems,
    allVisibleSelected,
    someVisibleSelected,
    setItemSelected,
    toggleVisibleItems,
    retainAvailableItems,
    clearSelection
  }
}
