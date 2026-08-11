import type { PrtEdges, PrtSeed, PrtSize } from '../_shared/types'

export type PrtBtnVariant = 'solid' | 'outline' | 'ghost'

export interface PrtBtnProps {
  variant?: PrtBtnVariant
  // Palette pad 1–10 from the seed rack — one value, the whole style set
  // (bg/ink/hover/wash) derives from it. Unset = quiet surface button,
  // which also derives its tint from an ancestor's `data-seed` via the cascade.
  seed?: PrtSeed
  size?: PrtSize
  edges?: PrtEdges
  disabled?: boolean
  loading?: boolean
  fullWidth?: boolean
  // polymorphic root: 'button' (default), 'a', ...
  tag?: string
  // layout-only additions (margin/width) — appearance goes through variants
  class?: string
}
