import type { PrtOption, PrtSeed, PrtSize } from '../_shared/types'

export type { PrtOption }

export interface PrtSegmentedProps {
  // The defining axis. `single` = exactly one active (modelValue is a value);
  // `multiple` = a set (modelValue is an array). This is why it's neither a
  // radio (single only) nor a toggle (binary) nor tabs (navigation): it does
  // single AND multiple selection.
  type?: 'single' | 'multiple'
  // v-model — a value in `single` mode, an array of values in `multiple` mode
  modelValue?: string | number | (string | number)[]
  // the segments
  options: PrtOption[]
  // show only each segment's icon; the label becomes the sr-only accessible
  // name (icons must be present, or segments render empty)
  iconOnly?: boolean
  // group aria-label (prop, never hardcoded)
  label?: string
  // decoration: tints the active segment(s)
  seed?: PrtSeed
  size?: PrtSize
  disabled?: boolean
  // layout-only additions (margin/width) — appearance goes through variants
  class?: string
}
