import { cva } from 'class-variance-authority'

// Three variants: spinner (the house icon mechanism), dots and squares
// (staggered pulse trios). Quiet ink-muted default; seed
// re-tints — and an ancestor's data-seed derives through for free, since
// the classes consume var(--seed).
export const loaderVariants = cva(
  'inline-block i-lucide-loader-circle animate-spin text-[var(--seed,var(--ink-muted))]',
  {
    variants: {
      size: {
        sm: 'w-4 h-4',
        base: 'w-5 h-5',
        lg: 'w-7 h-7',
      },
    },
    defaultVariants: {
      size: 'base',
    },
  },
)

// dots/squares container — marks inherit the color via bg-current
export const loaderTrioVariants = cva(
  'inline-flex items-center text-[var(--seed,var(--ink-muted))]',
  {
    variants: {
      size: {
        sm: 'gap-1',
        base: 'gap-1',
        lg: 'gap-1.5',
      },
    },
    defaultVariants: {
      size: 'base',
    },
  },
)

// one mark of the trio: dots are the legit circle (a dot IS a circle);
// squares are sharp — the technical option
export const loaderMarkVariants = cva('bg-current prt-loader-pulse', {
  variants: {
    variant: {
      dots: 'rounded-full',
      squares: 'rounded-none',
    },
    size: {
      sm: 'w-1 h-1',
      base: 'w-1.5 h-1.5',
      lg: 'w-2 h-2',
    },
  },
  defaultVariants: {
    variant: 'dots',
    size: 'base',
  },
})
