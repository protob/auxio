import { ref, shallowRef } from 'vue'
import { defineStore } from 'pinia'
import * as api from '@/api/objects'
import { useStatsStore } from '@/stores/stats'
import type { S3Object } from '@/types'
import { errorMessage } from '@/lib/errors'

export const useObjectsStore = defineStore('objects', () => {
  // shallowRef: a listing is replaced wholesale, never edited in place, so deep
  // reactivity would walk a thousand objects per load for nothing.
  const objects = shallowRef<S3Object[]>([])
  const prefixes = shallowRef<string[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const truncated = ref(false)

  async function load(bucket: string, prefix = '', delimiter = '/') {
    loading.value = true
    error.value = null
    try {
      const result = await api.listObjects(bucket, prefix, delimiter)
      objects.value = result.objects
      prefixes.value = result.common_prefixes
      truncated.value = result.is_truncated
    } catch (e) {
      error.value = errorMessage(e, 'Failed to load objects')
    } finally {
      loading.value = false
    }
  }

  async function remove(bucket: string, keys: string[]) {
    const result = await api.deleteObjects(bucket, keys)
    // Deleting changes the sidebar's totals, and nothing else refreshes them.
    useStatsStore().load()
    return result
  }

  return { objects, prefixes, loading, error, truncated, load, remove }
})
