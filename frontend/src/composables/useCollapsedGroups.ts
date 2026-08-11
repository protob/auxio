import { useLocalStorage } from '@vueuse/core'

// Stored as an array rather than a Set so the JSON on disk stays readable and
// survives a hand edit.
export function useCollapsedGroups(storageKey: string) {
  const collapsed = useLocalStorage<string[]>(storageKey, [])

  function isCollapsed(key: string) {
    return collapsed.value.includes(key)
  }

  function toggle(key: string) {
    collapsed.value = isCollapsed(key)
      ? collapsed.value.filter(k => k !== key)
      : [...collapsed.value, key]
  }

  function setAll(keys: string[]) {
    collapsed.value = keys
  }

  return { collapsed, isCollapsed, toggle, setAll }
}
