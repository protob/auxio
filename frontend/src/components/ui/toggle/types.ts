import type { PrtSeed, PrtSize } from '../_shared/types'

export interface PrtToggleProps {
  modelValue?: boolean
  // inline label text (prop, not hardcoded); default slot overrides
  label?: string
  id?: string
  seed?: PrtSeed
  size?: PrtSize
  // no `error` by design: a toggle applies immediately — there is no
  // "invalid but applied" state. Validation errors belong to checkboxes
  // (part of a form submit). Noted in the demo prose.
  disabled?: boolean
  // layout-only additions (margin/width) — appearance goes through variants
  class?: string
}
