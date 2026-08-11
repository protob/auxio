import { ref } from 'vue'
import { defineStore } from 'pinia'
import { getStats } from '@/api/stats'
import type { Stats } from '@/types'

export const useStatsStore = defineStore('stats', () => {
  const stats = ref<Stats | null>(null)

  // Counts go stale on every upload and delete, so callers refresh after a
  // mutation rather than relying on the one fetch at mount. A failed refresh
  // keeps the last good numbers - blanking the sidebar is worse than stale.
  async function load() {
    stats.value = await getStats().catch(() => stats.value)
  }

  return { stats, load }
})
