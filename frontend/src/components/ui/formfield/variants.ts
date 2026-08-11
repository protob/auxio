import { cva } from 'class-variance-authority'

export const fieldVariants = cva('flex flex-col gap-1.5', {
  variants: {
    disabled: { true: 'opacity-60' },
  },
})

export const labelVariants = cva('text-sm font-medium text-ink-muted', {
  variants: {
    error: { true: 'text-danger' },
  },
})

export const helpVariants = cva('text-xs', {
  variants: {
    tone: {
      help: 'text-ink-faint',
      error: 'text-danger flex items-center gap-1.5',
    },
  },
  defaultVariants: { tone: 'help' },
})
