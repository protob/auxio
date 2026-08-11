import { computed, ref, watch, type Ref } from 'vue'

export function usePagination<T>(items: Ref<T[]>, pageSize: number) {
  const page = ref(1)

  const totalPages = computed(() => Math.max(1, Math.ceil(items.value.length / pageSize)))
  const pageItems = computed(() => items.value.slice((page.value - 1) * pageSize, page.value * pageSize))
  const rangeStart = computed(() => (items.value.length === 0 ? 0 : (page.value - 1) * pageSize + 1))
  const rangeEnd = computed(() => Math.min(page.value * pageSize, items.value.length))

  // Deleting the last bucket on the last page would otherwise leave the view
  // parked past the end, showing nothing.
  watch(totalPages, (tp) => { if (page.value > tp) page.value = tp })

  return { page, totalPages, pageItems, rangeStart, rangeEnd }
}
