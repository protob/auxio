import type { PrtIntent } from '../_shared/types'

// quiet (default): surface + intent icon, unobtrusive · solid: full intent
// background riding the *-ink pairs — the attention-grabber
export type PrtToastVariant = 'quiet' | 'solid'

export interface PrtToastItem {
  id: number
  intent: PrtIntent
  variant: PrtToastVariant
  // the toast line (caller-provided, no defaults)
  message: string
  // optional second line
  description?: string
  // ms; 0 = sticky until dismissed
  duration: number
}

export type PrtToastInput = Partial<Omit<PrtToastItem, 'id' | 'message'>> & {
  message: string
}

export type PrtToasterPosition =
  | 'top-left'
  | 'top-center'
  | 'top-right'
  | 'bottom-left'
  | 'bottom-center'
  | 'bottom-right'

export interface PrtToasterProps {
  position?: PrtToasterPosition
}

export interface PrtToastProps {
  intent?: PrtIntent
  // quiet (default) is the unobtrusive surface look; solid grabs attention
  variant?: PrtToastVariant
  // the toast line; default slot overrides
  message?: string
  // optional second line
  description?: string
  dismissible?: boolean
  // layout-only additions
  class?: string
}
