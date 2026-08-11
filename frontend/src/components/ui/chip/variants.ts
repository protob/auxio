import { cva } from 'class-variance-authority'

// Variant rows share btn/variants.ts' vocabulary (solid/outline/ghost,
// same seed-consuming literals) MINUS the hover rows: a chip is a static
// label, not a button — only its dismiss × reacts.
export const chipVariants = cva(
  'inline-flex items-center justify-center font-medium select-none',
  {
    variants: {
      variant: {
        solid:
          'bg-[var(--seed,var(--surface-2))] text-[var(--seed-ink,var(--ink))] border border-edge',
        outline:
          'bg-transparent border border-[var(--seed,var(--edge-strong))] text-[var(--seed,var(--ink))]',
        ghost:
          'bg-transparent text-[var(--seed,var(--ink))]',
      },
      size: {
        sm: 'h-5 px-2 text-xs gap-1',
        base: 'h-6 px-2.5 text-xs gap-1.5',
        lg: 'h-7 px-3 text-sm gap-1.5',
      },
      edges: {
        square: 'rounded-none',
        rounded: 'rounded-control',
        pill: 'rounded-full',
      },
      disabled: {
        true: 'opacity-50 cursor-not-allowed',
      },
    },
    defaultVariants: {
      variant: 'solid',
      size: 'base',
      edges: 'rounded',
    },
  },
)
