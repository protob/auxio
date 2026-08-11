import type { Bucket } from '../types'

export const UNGROUPED = ' ungrouped' // groupKeyOf trims, so no real group can collide

export const groupKeyOf = (b: Bucket) => (b.group || '').trim().toLowerCase()

export type GroupedRow =
  | { type: 'header'; key: string; label: string; count: number; open: boolean }
  | { type: 'bucket'; bucket: Bucket }

export interface BucketGroup {
  key: string
  label: string
  buckets: Bucket[]
}

// Groups alphabetical with ungrouped last. The table wants this flattened and
// the sidebar wants it nested, so the ordering rule lives here and neither
// restates it.
export function groupBuckets(
  buckets: Bucket[],
  sort: (a: Bucket, b: Bucket) => number,
): BucketGroup[] {
  const groups = new Map<string, Bucket[]>()
  for (const b of buckets) {
    const k = groupKeyOf(b) || UNGROUPED
    let list = groups.get(k)
    if (!list) {
      list = []
      groups.set(k, list)
    }
    list.push(b)
  }
  return [...groups.keys()]
    .sort((a, z) => (a === UNGROUPED ? 1 : z === UNGROUPED ? -1 : a.localeCompare(z)))
    .map(k => ({
      key: k,
      label: k === UNGROUPED ? 'Ungrouped' : k,
      buckets: groups.get(k)!.slice().sort(sort),
    }))
}

// Flattens to header/bucket rows: a collapsed group contributes only its header.
export function buildGroupRows(
  buckets: Bucket[],
  isOpen: (key: string) => boolean,
  sort: (a: Bucket, b: Bucket) => number,
): GroupedRow[] {
  const rows: GroupedRow[] = []
  for (const g of groupBuckets(buckets, sort)) {
    const open = isOpen(g.key)
    rows.push({ type: 'header', key: g.key, label: g.label, count: g.buckets.length, open })
    if (open) for (const b of g.buckets) rows.push({ type: 'bucket', bucket: b })
  }
  return rows
}
