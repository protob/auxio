import type { S3Object } from '../types'
import { objectName } from './objectPath'

export const SORT_KEYS = ['name', 'size', 'modified'] as const
export type SortKey = (typeof SORT_KEYS)[number]

// A Record keyed by SortKey rather than a switch: adding a column to SORT_KEYS
// is then a compile error here, not a silently-unsorted header.
// Takes the prefix because the name column sorts on what the table shows.
export function objectComparators(prefix: string): Record<SortKey, (a: S3Object, b: S3Object) => number> {
  return {
    name: (a, b) => objectName(a.key, prefix).localeCompare(objectName(b.key, prefix)),
    size: (a, b) => a.size - b.size,
    modified: (a, b) => new Date(a.last_modified).getTime() - new Date(b.last_modified).getTime(),
  }
}

// Shared by the bucket table and the sidebar so the ordering rule lives in
// exactly one place.
export function byPinnedThenName(
  a: { pinned: boolean; name: string },
  b: { pinned: boolean; name: string },
) {
  if (a.pinned !== b.pinned) return a.pinned ? -1 : 1
  return a.name.localeCompare(b.name)
}
