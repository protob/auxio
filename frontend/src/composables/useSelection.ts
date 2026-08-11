import { computed, ref } from 'vue'

// Every caller passes the ids it is currently rendering, never the raw list.
// "Select all" that also picks up rows hidden by a filter is a data-loss trap
// once the next click is Delete.
export function useSelection() {
  const ids = ref<Set<string>>(new Set())

  const count = computed(() => ids.value.size)
  const has = (id: string) => ids.value.has(id)

  // Replacing the Set rather than mutating it, matching toggleAll/set/clear below -
  // consistent semantics, and toggleAll batches multiple changes into one assignment.
  function toggle(id: string) {
    const next = new Set(ids.value)
    next.has(id) ? next.delete(id) : next.add(id)
    ids.value = next
  }

  const allSelected = (visible: string[]) =>
    visible.length > 0 && visible.every(id => ids.value.has(id))

  function toggleAll(visible: string[]) {
    const next = new Set(ids.value)
    if (allSelected(visible)) visible.forEach(id => next.delete(id))
    else visible.forEach(id => next.add(id))
    ids.value = next
  }

  function set(next: string[]) {
    ids.value = new Set(next)
  }

  function clear() {
    ids.value = new Set()
  }

  return { ids, count, has, toggle, allSelected, toggleAll, set, clear }
}
