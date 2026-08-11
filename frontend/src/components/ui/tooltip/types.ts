import type { PrtSeed } from '../_shared/types'

export type PrtTooltipPlacement = 'top' | 'bottom' | 'left' | 'right'

export interface PrtTooltipProps {
  // tooltip text; #content slot overrides
  content?: string
  // re-tints the tooltip surface (decoration); unset = quiet surface-3
  seed?: PrtSeed
  placement?: PrtTooltipPlacement
  // ms before showing on hover/focus — default 0: tooltips are INSTANT
  // by design (house rule); hide is always immediate
  delay?: number
  disabled?: boolean
  // layout-only additions to the trigger wrapper
  class?: string
}
