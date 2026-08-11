// API response types are derived from the zod schemas in api/schemas.ts -
// the runtime contract and the static types cannot drift apart.
export type { Bucket, S3Object, ListObjectsResponse, Stats } from './api/schemas'

export interface UploadProgress {
  file: File
  key: string
  progress: number
  status: 'pending' | 'uploading' | 'complete' | 'error'
  error?: string
}
