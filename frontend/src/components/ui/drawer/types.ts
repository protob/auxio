import type { PrtSize } from '../_shared/types'

export type PrtDrawerPlacement = 'left' | 'right' | 'bottom'

export interface PrtDrawerProps {
  // the one v-model
  modelValue?: boolean
  placement?: PrtDrawerPlacement
  // side placements: width steps; bottom: height steps — xl/2xl are the working-panel sizes
  size?: PrtSize | 'xl' | '2xl' | 'full'
  // exact width for SIDE placements, any CSS length ('40rem', '700px').
  // Overrides `size`; ignored on placement="bottom". Style-bound, NOT a
  // class — a w-[…] utility in `class` would race the size class
  // (trap registry). Still capped by max-w-full on small viewports.
  width?: string
  // true (default) → closedby="any" (Esc + outside click); false → "closerequest" (Esc only)
  dismissible?: boolean
  // the X button in the header
  showClose?: boolean
  // D13: sr-only label of the X button
  closeLabel?: string
  // layout-only additions (margin/width) — appearance goes through variants
  class?: string
}
