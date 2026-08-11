import { cva } from 'class-variance-authority'

export const checkboxRootVariants = cva(
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

// Selection-control house rule: the CHECKED mark is identity-colored —
// var(--seed, var(--accent)) fill, var(--seed-ink, var(--accent-ink)) glyph.
// Unchecked is quiet surface + edge, like an unseeded button.
// The box rides --radius-mark (3px), NOT --radius-control: 6px on a ~16px
// box reads near-circular (blobby). Square is the technical option.
export const checkboxBoxVariants = cva(
  'inline-flex items-center justify-center shrink-0 border prt-motion-colors peer-focus-visible:ring-2 peer-focus-visible:ring-accent peer-focus-visible:ring-offset-2 peer-focus-visible:ring-offset-surface-0',
  {
    variants: {
      edges: {
        square: 'rounded-none',
        rounded: 'rounded-mark',
      },
      size: {
        sm: 'h-3.5 w-3.5',
        base: 'h-4 w-4',
        lg: 'h-5 w-5',
      },
      checked: {
        true: 'bg-[var(--seed,var(--accent))] border-transparent',
        false: 'bg-surface-2 border-edge hover:border-ink-faint',
      },
      error: {
        true: '',
      },
    },
    compoundVariants: [
      { checked: false, error: true, class: 'border-danger hover:border-danger' },
      { checked: true, error: true, class: 'ring-1 ring-danger' },
    ],
    defaultVariants: { edges: 'rounded', size: 'base', checked: false },
  },
)

export const checkboxGlyphVariants = cva(
  'i-lucide-check text-[var(--seed-ink,var(--accent-ink))]',
  {
    variants: {
      size: {
        sm: 'text-[0.625rem]',
        base: 'text-xs',
        lg: 'text-sm',
      },
    },
    defaultVariants: { size: 'base' },
  },
)

export const checkboxLabelVariants = cva('text-ink', {
  variants: {
    size: {
      sm: 'text-xs',
      base: 'text-sm',
      lg: 'text-base',
    },
  },
  defaultVariants: { size: 'base' },
})
