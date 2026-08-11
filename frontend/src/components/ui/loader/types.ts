import type { PrtSeed, PrtSize } from '../_shared/types'

// spinner is the inline default; dots/squares are the alternatives
export type PrtLoaderVariant = 'spinner' | 'dots' | 'squares'

export interface PrtLoaderProps {
  variant?: PrtLoaderVariant
  seed?: PrtSeed
  size?: PrtSize
  // layout-only additions (margin/placement)
  class?: string
}
