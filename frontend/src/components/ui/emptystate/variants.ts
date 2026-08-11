import { cva } from 'class-variance-authority'

// NO frame by default — an empty state sits inside cards/listings that
// already have frames (a bordered variant waits until a real page wants
// one). Quiet centered column, generous padding.
export const emptyStateVariants = cva(
  'flex w-full flex-col items-center justify-center text-center',
  {
    variants: {
      size: {
        sm: 'px-4 py-8 gap-1',
        base: 'px-6 py-12 gap-1.5',
        lg: 'px-8 py-16 gap-2',
      },
    },
    defaultVariants: { size: 'base' },
  },
)
