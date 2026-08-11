import type { PrtEdges, PrtSize } from '../_shared/types'

export interface PrtDialogProps {
  // the one v-model. Optional so declarative usage (an external button
  // with command="show-modal" commandfor="<dialog id>") needs no wiring —
  // declarative opens do NOT flip the v-model (v1 limit, by design).
  modelValue?: boolean
  // content width steps — xl/2xl are the advanced-form-dialog widths
  size?: PrtSize | 'xl' | '2xl' | 'full'
  // exact max-width, any CSS length ('68rem', '900px', 'min(90vw, 75rem)').
  // Overrides `size` — rides a style binding, NOT a class (a max-w-[…]
  // utility in `class` would race the size class; trap registry).
  // Still shrinks on small viewports (the panel stays w-full underneath).
  width?: string
  edges?: PrtEdges
  // true (default) → closedby="any" (Esc + outside click); false → "closerequest" (Esc only)
  dismissible?: boolean
  // the X button in the header
  showClose?: boolean
  // D13: sr-only label of the X button
  closeLabel?: string
  // layout-only additions (margin/width) — appearance goes through variants
  class?: string
}
