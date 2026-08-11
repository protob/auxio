import { z } from 'zod'

// The Go structs are the other side of this contract. Every array is
// `.nullable().transform(v => v ?? [])` because nil slices marshal to null.
// huma's extra `$schema` key is a non-issue: zod strips unknown keys.

export const Bucket = z.object({
  name: z.string(),
  created_at: z.string(),
  region: z.string(),
  public: z.boolean(),
  group: z.string(),
  pinned: z.boolean(),
  object_count: z.number(),
  total_size: z.number(),
})

// PATCH returns metadata only, no object stats (Go UpdateBucketOutput), so this
// must not be typed as a full Bucket.
export const BucketMeta = Bucket.omit({ object_count: true, total_size: true })

export const S3Object = z.object({
  key: z.string(),
  size: z.number(),
  content_type: z.string(),
  last_modified: z.string(),
  etag: z.string(),
})

export const ListObjectsResponse = z.object({
  objects: z.array(S3Object).nullable().transform(v => v ?? []),
  common_prefixes: z.array(z.string()).nullable().transform(v => v ?? []),
  is_truncated: z.boolean(),
  next_continuation_token: z.string().optional().default(''),
})

export const Stats = z.object({
  buckets: z.number(),
  objects: z.number(),
  total_size: z.number(),
  total_size_human: z.string(),
  uploads_in_progress: z.number(),
  cache_size: z.number(),
  cache_size_human: z.string(),
})

export const DeleteObjectsResult = z.object({
  deleted: z.array(z.string()).nullable().transform(v => v ?? []),
  errors: z.array(z.object({ key: z.string(), error: z.string() })).nullable().transform(v => v ?? []),
})

export const LoginResponse = z.object({
  token: z.string(),
})

// The one request-side schema. Approximates the server's bucket-name rules (no
// dots) so a typo is caught before the POST; zod v4's Standard Schema is what
// PrtForm validates against, and displayError shows the first issue per path.
export const BucketCreate = z.object({
  name: z.string()
    .min(1, 'Name is required')
    .regex(/^[a-z0-9]/, 'Must start with a lowercase letter or digit')
    .regex(/^[a-z0-9-]+$/, 'Lowercase letters, digits, and hyphens only')
    .min(3, 'Minimum 3 characters')
    .max(63, 'Maximum 63 characters'),
  group: z.string(),
  public: z.boolean(),
})

export type Bucket = z.infer<typeof Bucket>
export type S3Object = z.infer<typeof S3Object>
export type ListObjectsResponse = z.infer<typeof ListObjectsResponse>
export type Stats = z.infer<typeof Stats>
export type DeleteObjectsResult = z.infer<typeof DeleteObjectsResult>
