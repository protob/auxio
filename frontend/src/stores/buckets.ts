import { computed, ref, shallowRef } from 'vue'
import { defineStore } from 'pinia'
import * as api from '@/api/buckets'
import { useStatsStore } from '@/stores/stats'
import type { BucketPatch } from '@/api/buckets'
import type { Bucket } from '@/types'
import { errorMessage } from '@/lib/errors'

export const useBucketsStore = defineStore('buckets', () => {
  // shallowRef: every mutation reloads the whole list, so nothing edits a
  // bucket in place.
  const buckets = shallowRef<Bucket[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const groupNames = computed(() =>
    [...new Set(buckets.value.map(b => b.group).filter(Boolean))].sort(),
  )

  // Creating and deleting move the sidebar's bucket count, and nothing else
  // refetches it.
  function refreshStats() {
    useStatsStore().load()
  }

  async function load() {
    loading.value = true
    error.value = null
    try {
      buckets.value = await api.listBuckets()
    } catch (e) {
      error.value = errorMessage(e, 'Failed to load buckets')
    } finally {
      loading.value = false
    }
  }

  async function create(name: string, isPublic = false, group = '') {
    await api.createBucket(name, isPublic, group)
    await load()
    refreshStats()
  }

  async function remove(name: string) {
    await api.deleteBucket(name)
    await load()
    refreshStats()
  }

  async function update(name: string, patch: BucketPatch) {
    await api.updateBucket(name, patch)
    await load()
  }

  // Returns the names that failed. Partial success is the normal case for both
  // of these - a bulk delete hits 409 on whichever buckets are non-empty - so
  // neither aborts the rest of the batch on the first rejection.
  async function updateMany(names: string[], patch: BucketPatch): Promise<string[]> {
    const results = await Promise.allSettled(names.map(n => api.updateBucket(n, patch)))
    await load()
    return names.filter((_, i) => results[i]?.status === 'rejected')
  }

  async function removeMany(names: string[]): Promise<string[]> {
    const results = await Promise.allSettled(names.map(n => api.deleteBucket(n)))
    await load()
    refreshStats()
    return names.filter((_, i) => results[i]?.status === 'rejected')
  }

  return { buckets, loading, error, groupNames, load, create, remove, update, updateMany, removeMany }
})
