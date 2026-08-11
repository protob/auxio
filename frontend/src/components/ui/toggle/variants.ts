import { cva } from 'class-variance-authority'

export const toggleRootVariants = cva(
  'inline-flex items-center gap-2 select-none',
  {
    variants: {
      disabled: {
        true: 'opacity-60 cursor-not-allowed',
        false: 'cursor-pointer',
      },
    },
    defaultVariants: { disabled: false },
  },
)

// Selection-control house rule: ON is identity-colored
// (var(--seed, var(--accent)) track); OFF is quiet surface.
export const toggleTrackVariants = cva(
  'relative inline-flex shrink-0 items-center rounded-full p-0.5 prt-motion-colors peer-focus-visible:ring-2 peer-focus-visible:ring-accent peer-focus-visible:ring-offset-2 peer-focus-visible:ring-offset-surface-0',
  {
    variants: {
      size: {
        sm: 'h-5 w-9',
        base: 'h-6 w-11',
        lg: 'h-8 w-14',
      },
      checked: {
        true: 'bg-[var(--seed,var(--accent))]',
        false: 'bg-surface-3',
      },
    },
    defaultVariants: { size: 'base', checked: false },
  },
)

// Knob translate per size: complete literals in compound variants — NEVER computed.
export const toggleKnobVariants = cva(
  'pointer-events-none inline-block rounded-full prt-motion',
  {
    variants: {
      size: {
        sm: 'h-4 w-4',
        base: 'h-5 w-5',
        lg: 'h-7 w-7',
      },
      checked: {
        true: 'bg-[var(--seed-ink,var(--accent-ink))]',
        false: 'translate-x-0 bg-ink-muted',
      },
    },
    compoundVariants: [
      { size: 'sm', checked: true, class: 'translate-x-4' },
      { size: 'base', checked: true, class: 'translate-x-5' },
      { size: 'lg', checked: true, class: 'translate-x-6' },
    ],
    defaultVariants: { size: 'base', checked: false },
  },
)

export const toggleLabelVariants = cva('text-ink', {
  variants: {
    size: {
      sm: 'text-xs',
      base: 'text-sm',
      lg: 'text-base',
    },
  },
  defaultVariants: { size: 'base' },
})
