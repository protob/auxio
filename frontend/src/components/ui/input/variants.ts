import { cva } from 'class-variance-authority'

// No focus ring: the border already changes colour on focus, and a ring sits
// directly outside it, so the two read as one thick edge rather than as an
// indicator. The border alone is the whole focus state.
export const inputVariants = cva(
  'w-full text-ink placeholder:text-ink-faint outline-none prt-motion-colors',
  {
    variants: {
      variant: {
        // hover gated on not-disabled: disabled inputs keep pointer events
        // (the not-allowed cursor must show on hover) but never light up
        filled:
          'bg-surface-2 border border-transparent not-disabled:hover:bg-surface-3 focus-visible:bg-surface-2 focus-visible:border-accent',
        outline:
          'bg-surface-1 border border-edge-strong not-disabled:hover:border-ink-faint focus-visible:border-accent',
        minimal:
          'bg-transparent border border-transparent not-disabled:hover:bg-wash focus-visible:bg-surface-1 focus-visible:border-edge',
      },
      size: {
        sm: 'h-8 px-3 text-sm',
        base: 'h-9 px-3.5 text-sm',
        lg: 'h-11 px-4 text-base',
      },
      edges: {
        square: 'rounded-none',
        rounded: 'rounded-control',
        pill: 'rounded-full',
      },
      error: {
        // hover gate matches the variant rows' (same specificity keeps the
        // danger pin winning by source order)
        true: 'border-danger not-disabled:hover:border-danger focus-visible:border-danger',
      },
      disabled: {
        true: 'opacity-50 cursor-not-allowed',
      },
    },
    defaultVariants: {
      variant: 'filled',
      size: 'base',
      edges: 'rounded',
    },
  },
)
