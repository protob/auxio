import type { PrtEdges, PrtSize } from '../_shared/types'

export type PrtInputVariant = 'filled' | 'outline' | 'minimal'

export interface PrtInputProps {
  modelValue?: string | number
  type?: string
  id?: string
  placeholder?: string
  variant?: PrtInputVariant
  size?: PrtSize
  edges?: PrtEdges
  error?: boolean
  disabled?: boolean
  // layout-only additions (margin/width) — appearance goes through variants
  class?: string
}
