import { computed, shallowRef, type Ref } from 'vue'
import type { SortKey } from '@/lib/sort'

type Comparators<T> = Record<SortKey, (a: T, b: T) => number>

export function useTableSort<T>(items: Ref<T[]>, comparators: Ref<Comparators<T>>) {
  const key = shallowRef<SortKey>('name')
  const dir = shallowRef<'asc' | 'desc'>('asc')

  const sorted = computed(() => {
    const sign = dir.value === 'asc' ? 1 : -1
    return [...items.value].sort((a, b) => sign * comparators.value[key.value](a, b))
  })

  // Clicking the active column flips direction; any other column starts fresh
  // at ascending rather than inheriting the previous column's direction.
  function toggle(next: SortKey) {
    if (key.value === next) dir.value = dir.value === 'asc' ? 'desc' : 'asc'
    else {
      key.value = next
      dir.value = 'asc'
    }
  }

  return { key, dir, sorted, toggle }
}
