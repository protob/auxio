import { cva } from 'class-variance-authority'

// Hairline-thin tracks (restraint over thick slabs).
export const progressTrackVariants = cva(
  'relative w-full overflow-hidden rounded-full bg-surface-2',
  {
    variants: {
      size: {
        sm: 'h-1',
        base: 'h-1.5',
        lg: 'h-2.5',
      },
    },
    defaultVariants: {
      size: 'base',
    },
  },
)

// fill width is a truly dynamic value — it rides a style binding in the
// component, never a generated class
export const progressFillClass =
  'h-full rounded-full bg-[var(--seed,var(--accent))]'

// circle: svg size per scale — dashoffset is the truly
// dynamic value and rides an attribute binding in the component
export const progressCircleVariants = cva('block', {
  variants: {
    size: {
      sm: 'w-10 h-10',
      base: 'w-14 h-14',
      lg: 'w-20 h-20',
    },
  },
  defaultVariants: {
    size: 'base',
  },
})

export const progressCircleLabelVariants = cva(
  'absolute inset-0 flex items-center justify-center tabular-nums text-ink-muted',
  {
    variants: {
      size: {
        sm: 'text-[10px]',
        base: 'text-xs',
        lg: 'text-sm',
      },
    },
    defaultVariants: {
      size: 'base',
    },
  },
)
