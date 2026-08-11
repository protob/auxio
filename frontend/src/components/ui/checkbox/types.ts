import type { PrtSeed, PrtSize } from '../_shared/types'

// No pill: a pill checkbox mark reads as a radio.
export type PrtCheckboxEdges = 'square' | 'rounded'

export interface PrtCheckboxProps {
  modelValue?: boolean
  // inline label text (D13: prop, not hardcoded); default slot overrides
  label?: string
  id?: string
  seed?: PrtSeed
  size?: PrtSize
  // rounded = --radius-mark (3px) — deliberately tighter than controls
  edges?: PrtCheckboxEdges
  error?: boolean
  disabled?: boolean
  // layout-only additions (margin/width) — appearance goes through variants
  class?: string
}
