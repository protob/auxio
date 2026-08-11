import { cva } from 'class-variance-authority'

// The active fill rides the seed system (seed → accent fallback); the resting
// segments are quiet ink — color is a deliberate pluck, set per group via
// `data-seed`, never a built class.

// the joined bar. overflow-hidden clips the active fill to the rounded corners.
export const segmentedRootVariants = cva(
  'inline-flex items-stretch overflow-hidden rounded-control border border-edge bg-surface-1',
  {
    variants: {
      size: {
        sm: 'h-8',
        base: 'h-9',
        lg: 'h-11',
      },
      disabled: {
        true: 'opacity-50 cursor-not-allowed',
        false: '',
      },
    },
    defaultVariants: { size: 'base', disabled: false },
  },
)

// one segment. border-l + first:border-l-0 draws the dividers between cells.
// INSET focus ring (not prt-ring) — the root's overflow-hidden would clip an
// offset ring. active fill = seed/accent; resting = quiet ink with a hover wash.
export const segmentVariants = cva(
  'relative inline-flex items-center justify-center gap-2 whitespace-nowrap font-medium select-none border-l border-edge first:border-l-0 prt-motion-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-accent disabled:cursor-not-allowed disabled:opacity-40',
  {
    variants: {
      size: {
        sm: 'px-3 text-sm',
        base: 'px-4 text-sm',
        lg: 'px-5 text-base',
      },
      active: {
        true: 'bg-[var(--seed,var(--accent))] text-[var(--seed-ink,var(--accent-ink))]',
        false: 'text-ink-muted not-disabled:hover:bg-[var(--wash)] not-disabled:hover:text-ink',
      },
    },
    defaultVariants: { size: 'base', active: false },
  },
)
