import type { PrtSize } from '../_shared/types'

export interface PrtEmptyStateProps {
  // icon class string, e.g. 'i-lucide-inbox' — rendered at muted ink
  icon?: string
  title: string
  description?: string
  size?: PrtSize
  // layout-only additions (margin/width) — appearance goes through variants
  class?: string
}
