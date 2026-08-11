import type { PrtEdges, PrtSeed, PrtSize } from '../_shared/types'

export type PrtChipVariant = 'solid' | 'outline' | 'ghost'

export interface PrtChipProps {
  variant?: PrtChipVariant
  seed?: PrtSeed
  size?: PrtSize
  // default 'rounded' (control radius — technical); 'pill' is the opt-in exception (reads well inside tag/select input boxes), never the default
  edges?: PrtEdges
  // leading icon class string, e.g. 'i-lucide-tag'
  icon?: string
  // renders the × button; clicking emits 'dismiss' (the parent owns the list)
  dismissible?: boolean
  disabled?: boolean
  // layout-only additions (margin/placement) — appearance goes through variants
  class?: string
}
