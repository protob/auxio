import { cva } from 'class-variance-authority'

// Color rides the seed system: classes reference var(--seed*) with
// surface-token fallbacks, so ONE row serves both the seeded and the
// unseeded (quiet surface) state — and an unseeded button inside a
// `data-seed` ancestor derives its tint from it through the cascade for free.
// This table IS the extraction manifest UnoCSS reads.
export const buttonVariants = cva(
  'inline-flex items-center justify-center gap-2 font-medium select-none prt-motion prt-ring aria-disabled:opacity-50 aria-disabled:cursor-not-allowed',
  {
    variants: {
      variant: {
        // hover gated on not-disabled: disabled buttons keep pointer events
        // (the not-allowed cursor must show on hover) but never light up
        solid:
          'bg-[var(--seed,var(--surface-2))] text-[var(--seed-ink,var(--ink))] border border-edge not-disabled:not-[[aria-disabled=true]]:hover:bg-[var(--seed-hover,var(--surface-3))]',
        outline:
          'bg-transparent border border-[var(--seed,var(--edge-strong))] text-[var(--seed,var(--ink))] not-disabled:not-[[aria-disabled=true]]:hover:bg-[var(--seed-wash,var(--wash))]',
        ghost:
          'bg-transparent text-[var(--seed,var(--ink))] not-disabled:not-[[aria-disabled=true]]:hover:bg-[var(--seed-wash,var(--wash))]',
      },
      size: {
        sm: 'h-8 px-3 text-sm',
        base: 'h-9 px-4 text-sm',
        lg: 'h-11 px-6 text-base',
      },
      edges: {
        square: 'rounded-none',
        rounded: 'rounded-control',
        pill: 'rounded-full',
      },
      fullWidth: { true: 'w-full' },
      disabled: { true: 'opacity-50 cursor-not-allowed' },
      loading: { true: 'cursor-wait' },
    },
    defaultVariants: {
      variant: 'solid',
      size: 'base',
      edges: 'rounded',
    },
  },
)
