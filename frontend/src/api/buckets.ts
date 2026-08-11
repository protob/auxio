import { z } from 'zod'
import { request } from './client'
import { Bucket, BucketMeta } from './schemas'

export interface BucketPatch {
  public?: boolean
  group?: string
  pinned?: boolean
}

export function listBuckets() {
  return request('/buckets', z.array(Bucket))
}

// Returns nothing the UI uses - validated as "we don't care", not cast.
export function createBucket(name: string, isPublic = false, group = '') {
  return request('/buckets', z.unknown(), {
    method: 'POST',
    body: { name, public: isPublic, group },
  })
}

export function deleteBucket(name: string) {
  return request(`/buckets/${encodeURIComponent(name)}`, z.unknown(), { method: 'DELETE' })
}

export function updateBucket(name: string, patch: BucketPatch) {
  return request(`/buckets/${encodeURIComponent(name)}`, BucketMeta, {
    method: 'PATCH',
    body: patch,
  })
}
