import { cva } from 'class-variance-authority'

// Underline tabs, Linear school: quiet transparent row over a full-width
// hairline; triggers are ghost (text color is the only state voice), the
// 2px accent indicator is a separate anchor-positioned element (its
// geometry lives in the SFC's scoped block — the LOOK class is below).
export const tabsListClass = 'relative flex items-center gap-1 border-b border-edge'

export const tabsTriggerVariants = cva(
  'inline-flex items-center gap-1.5 bg-transparent select-none outline-none rounded-control prt-motion-colors focus-visible:ring-1 focus-visible:ring-accent',
  {
    variants: {
      size: {
        sm: 'h-8 px-2.5 text-xs',
        base: 'h-9 px-3 text-sm',
        lg: 'h-10 px-3.5 text-base',
      },
      active: {
        true: 'text-ink',
        false: 'text-ink-muted not-disabled:hover:text-ink',
      },
      disabled: {
        true: 'opacity-50 cursor-not-allowed',
        false: 'cursor-pointer',
      },
    },
    defaultVariants: { size: 'base', active: false, disabled: false },
  },
)

// the sliding underline — geometry (anchor()) is scoped; this is color only
export const tabsIndicatorClass = 'bg-accent'

export const tabsPanelClass = 'pt-4 text-sm text-ink-muted'
