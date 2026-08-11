import type { PrtEdges, PrtSize } from '../_shared/types'

export interface PrtPaginationProps {
  // v-model:page — 1-based current page
  page: number
  totalPages: number
  // pages shown on each side of the current page
  siblingCount?: number
  // pages pinned at each end
  boundaryCount?: number
  // prev/next chevrons
  showArrows?: boolean
  // first/last double chevrons — quieter default
  showEdges?: boolean
  size?: PrtSize
  edges?: PrtEdges
  disabled?: boolean
  // layout-only additions (margin/width) — appearance goes through variants
  class?: string
  // every aria string is a prop with an English default
  label?: string
  prevLabel?: string
  nextLabel?: string
  firstLabel?: string
  lastLabel?: string
  // aria-label prefix for page cells, e.g. 'Page' → 'Page 4'
  pageLabel?: string
}
