import { cva } from 'class-variance-authority'

// The dropzone is a SURFACE (rounded-surface, not control radius).
// Drag-over speaks the kit's focus language — accent edge + 1px ring.
export const uploaderVariants = cva(
  'flex w-full flex-col items-center justify-center text-center prt-motion-colors',
  {
    variants: {
      variant: {
        dashed:
          'bg-surface-1 border border-dashed border-edge-strong',
        outline:
          'bg-surface-1 border border-edge-strong',
      },
      size: {
        sm: 'px-4 py-5 gap-1',
        base: 'px-6 py-7 gap-1.5',
        lg: 'px-8 py-10 gap-2',
      },
      edges: {
        square: 'rounded-none',
        rounded: 'rounded-surface',
        pill: 'rounded-surface',
      },
      dragging: {
        true: 'border-accent ring-1 ring-accent bg-wash',
      },
      error: {
        true: 'border-danger',
      },
      disabled: {
        true: 'opacity-50 cursor-not-allowed',
      },
    },
    defaultVariants: {
      variant: 'dashed',
      size: 'base',
      edges: 'rounded',
    },
  },
)

// the one interactive element in the dropzone — keyboard path for free
export const uploaderButtonVariants = cva(
  'bg-transparent font-medium text-ink underline underline-offset-2 decoration-edge-strong not-disabled:hover:text-accent not-disabled:hover:decoration-accent prt-motion-colors prt-ring rounded-mark',
  {
    variants: {
      size: {
        sm: 'text-xs',
        base: 'text-sm',
        lg: 'text-base',
      },
    },
    defaultVariants: { size: 'base' },
  },
)

// quiet rows, hairline separation
export const uploaderRowVariants = cva(
  'flex items-center min-w-0 border-t border-edge first:border-t-0',
  {
    variants: {
      size: {
        sm: 'gap-2.5 py-1.5',
        base: 'gap-3 py-2',
        lg: 'gap-3 py-2.5',
      },
    },
    defaultVariants: { size: 'base' },
  },
)
