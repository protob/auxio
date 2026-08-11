import { cva } from 'class-variance-authority'

// Pagination cells are their OWN vocabulary, not PrtBtn (restyling PrtBtn
// through `class` would be the layout-only-class violation): quiet ghost
// cells, tabular numerals (house type rule), accent + wash where
// "current" lives. bg/cursor ride the variants, not the base — one class
// per CSS property per render (CVA concatenates, cx doesn't merge).
export const paginationCellVariants = cva(
  'inline-flex items-center justify-center tabular-nums select-none outline-none prt-motion-colors focus-visible:ring-1 focus-visible:ring-accent',
  {
    variants: {
      size: {
        sm: 'h-7 min-w-7 px-1 text-xs',
        base: 'h-8 min-w-8 px-1.5 text-sm',
        lg: 'h-9 min-w-9 px-2 text-base',
      },
      edges: {
        square: 'rounded-none',
        rounded: 'rounded-control',
        pill: 'rounded-full',
      },
      active: {
        true: 'bg-wash text-accent',
        false: 'bg-transparent text-ink-muted',
      },
      disabled: {
        true: 'opacity-50 cursor-not-allowed',
        false: 'cursor-pointer',
      },
    },
    compoundVariants: [
      { active: false, disabled: false, class: 'hover:bg-wash hover:text-ink' },
    ],
    defaultVariants: { size: 'base', edges: 'rounded', active: false, disabled: false },
  },
)

// the ellipsis is a SPAN, never a button
export const paginationGapVariants = cva(
  'inline-flex items-center justify-center font-mono text-ink-faint select-none',
  {
    variants: {
      size: {
        sm: 'h-7 min-w-7 text-xs',
        base: 'h-8 min-w-8 text-sm',
        lg: 'h-9 min-w-9 text-base',
      },
    },
    defaultVariants: { size: 'base' },
  },
)
