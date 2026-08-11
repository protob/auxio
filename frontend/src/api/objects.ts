import { request } from './client'
import { DeleteObjectsResult, ListObjectsResponse } from './schemas'
import { OBJECT_MAX_KEYS } from '@/lib/constants'

export function listObjects(bucket: string, prefix = '', delimiter = '/', maxKeys = OBJECT_MAX_KEYS) {
  return request(`/buckets/${encodeURIComponent(bucket)}/objects`, ListObjectsResponse, {
    query: { prefix, delimiter, max_keys: maxKeys },
  })
}

export function deleteObjects(bucket: string, keys: string[]) {
  return request(`/buckets/${encodeURIComponent(bucket)}/objects`, DeleteObjectsResult, {
    method: 'DELETE',
    body: { keys },
  })
}
