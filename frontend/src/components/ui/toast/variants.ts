import { cva } from 'class-variance-authority'
import type { PrtIntent } from '../_shared/types'
import type { PrtToastVariant } from './types'

// The host is a popover=manual: top layer without z-index wars.
// Rows reset the popover UA centering with LONGHANDS on
// all four insets — never shorthand+longhand mixes (output-order trap).
export const toasterVariants = cva(
  'fixed m-0 p-0 border-0 bg-transparent overflow-visible flex flex-col gap-2',
  {
    variants: {
      position: {
        'top-left': 'top-4 left-4 right-auto bottom-auto',
        'top-center': 'top-4 left-1/2 right-auto bottom-auto -translate-x-1/2',
        'top-right': 'top-4 right-4 bottom-auto left-auto',
        'bottom-left': 'bottom-4 left-4 right-auto top-auto',
        'bottom-center': 'bottom-4 left-1/2 right-auto top-auto -translate-x-1/2',
        'bottom-right': 'bottom-4 right-4 top-auto left-auto',
      },
    },
    defaultVariants: {
      position: 'bottom-right',
    },
  },
)

// the floating sibling of PrtNotice: same anatomy, shadow because it floats
export const toastBase =
  'pointer-events-auto flex items-start gap-3 w-80 p-3.5 rounded-surface border'

// variant × intent — literal lookup table, same discipline as notice's
// noticeToneClasses. quiet: one unobtrusive surface, intent speaks through
// the icon. solid: the attention-grabber — full intent background with its
// hand-designed *-ink pair (the tokens exist exactly for text ON intent
// surfaces; warning takes dark ink in dark mode — the amber rule).
export const toastVariantClasses: Record<
  PrtToastVariant,
  Record<PrtIntent, string>
> = {
  quiet: {
    info: 'bg-surface-1 border-edge shadow-float',
    success: 'bg-surface-1 border-edge shadow-float',
    warning: 'bg-surface-1 border-edge shadow-float',
    danger: 'bg-surface-1 border-edge shadow-float',
  },
  solid: {
    info: 'bg-info border-transparent text-info-ink shadow-float',
    success: 'bg-accent border-transparent text-accent-ink shadow-float',
    warning: 'bg-warning border-transparent text-warning-ink shadow-float',
    danger: 'bg-danger border-transparent text-danger-ink shadow-float',
  },
}

// inner anatomy per variant — solid inherits the root's *-ink via
// text-current (opacity steps the hierarchy; the ink is already designed)
export const toastText: Record<
  PrtToastVariant,
  { message: string; description: string; dismiss: string }
> = {
  quiet: {
    message: 'text-ink',
    description: 'text-ink-muted',
    dismiss: 'text-ink-faint hover:text-ink hover:bg-wash',
  },
  solid: {
    message: 'text-current',
    description: 'text-current opacity-80',
    dismiss: 'text-current opacity-70 hover:opacity-100 hover:bg-wash',
  },
}

// literal lookup table — an extraction manifest, same discipline as CVA.
// iconColor serves the quiet variant; solid icons ride text-current.
export const toastIntent: Record<PrtIntent, { icon: string; iconColor: string }> = {
  info: { icon: 'i-lucide-info', iconColor: 'text-info' },
  success: { icon: 'i-lucide-circle-check', iconColor: 'text-accent' },
  warning: { icon: 'i-lucide-triangle-alert', iconColor: 'text-warning' },
  danger: { icon: 'i-lucide-circle-x', iconColor: 'text-danger' },
}
