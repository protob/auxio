import { cva } from 'class-variance-authority'

// The field is a real text input riding the input vocabulary
// (input/variants.ts) — sizes declare pl/pr explicitly because the right
// padding reserves room for the affordance overlay (chevron, and a clear
// button left of it when clearable). Don't fight px-* (02-styling-doctrine).
export const comboboxVariants = cva(
  'w-full appearance-none text-ink placeholder:text-ink-faint outline-none prt-motion-colors',
  {
    variants: {
      variant: {
        // hover gated on not-disabled — see input/variants.ts
        filled:
          'bg-surface-2 border border-transparent not-disabled:hover:bg-surface-3 focus-visible:bg-surface-2 focus-visible:border-accent',
        outline:
          'bg-surface-1 border border-edge-strong not-disabled:hover:border-ink-faint focus-visible:border-accent',
        minimal:
          'bg-transparent border border-transparent not-disabled:hover:bg-wash focus-visible:bg-surface-1 focus-visible:border-edge',
      },
      size: {
        sm: 'h-8 pl-3 text-sm',
        base: 'h-9 pl-3.5 text-sm',
        lg: 'h-11 pl-4 text-base',
      },
      edges: {
        square: 'rounded-none',
        rounded: 'rounded-control',
        pill: 'rounded-full',
      },
      error: {
        true: 'border-danger not-disabled:hover:border-danger focus-visible:border-danger',
      },
      disabled: {
        true: 'opacity-50 cursor-not-allowed',
      },
      // reserves overlay room: one icon (chevron) or two (clear + chevron)
      clearable: { true: '', false: '' },
    },
    compoundVariants: [
      { clearable: false, size: 'sm', class: 'pr-8' },
      { clearable: false, size: 'base', class: 'pr-9' },
      { clearable: false, size: 'lg', class: 'pr-10' },
      { clearable: true, size: 'sm', class: 'pr-13' },
      { clearable: true, size: 'base', class: 'pr-14' },
      { clearable: true, size: 'lg', class: 'pr-16' },
    ],
    defaultVariants: {
      variant: 'filled',
      size: 'base',
      edges: 'rounded',
      clearable: false,
    },
  },
)

export const comboboxChevronVariants = cva(
  'i-lucide-chevron-down pointer-events-none absolute top-1/2 -translate-y-1/2 text-ink-faint prt-motion',
  {
    variants: {
      size: {
        sm: 'right-2 text-xs',
        base: 'right-2.5 text-sm',
        lg: 'right-3 text-base',
      },
    },
    defaultVariants: { size: 'base' },
  },
)

// loading spinner swaps in for the chevron (same slot)
export const comboboxSpinnerVariants = cva(
  'i-lucide-loader-circle animate-spin pointer-events-none absolute top-1/2 -translate-y-1/2 text-ink-faint',
  {
    variants: {
      size: {
        sm: 'right-2 text-xs',
        base: 'right-2.5 text-sm',
        lg: 'right-3 text-base',
      },
    },
    defaultVariants: { size: 'base' },
  },
)

// clear affordance sits left of the chevron; bg-transparent is
// load-bearing (tailwind-compat keeps native button backgrounds)
export const comboboxClearVariants = cva(
  'i-lucide-x absolute top-1/2 -translate-y-1/2 bg-transparent border-0 p-0 cursor-pointer text-ink-faint hover:text-ink-muted prt-motion-colors',
  {
    variants: {
      size: {
        sm: 'right-7 text-xs',
        base: 'right-8 text-sm',
        lg: 'right-9 text-base',
      },
    },
    defaultVariants: { size: 'base' },
  },
)

// The open panel — select's substrate verbatim: FLAT + FULL-BLEED by
// decree (one bg-surface-2 surface, NO border/shadow/inner padding,
// rows edge to edge, rounded outer clipping). Positioning lives in the
// component's scoped block. The ONE deviation from select is behavioral
// (popover="manual"), not visual.
export const comboboxPanelVariants = cva('bg-surface-2 overflow-y-auto', {
  variants: {
    edges: {
      square: 'rounded-none',
      rounded: 'rounded-surface',
      pill: 'rounded-surface',
    },
  },
  defaultVariants: { edges: 'rounded' },
})

// Option rows — select's exact shape: full-bleed, never rounded, one
// `active` highlight source of truth (aria-activedescendant), px aligns
// with the input's pl so closed and open text line up.
export const comboboxOptionVariants = cva(
  'w-full appearance-none inline-flex items-center gap-2 text-left text-ink select-none prt-motion-colors',
  {
    variants: {
      size: {
        sm: 'px-3 py-1.5 text-sm',
        base: 'px-3.5 py-2 text-sm',
        lg: 'px-4 py-2.5 text-base',
      },
      active: {
        true: 'bg-surface-3',
        false: 'bg-transparent',
      },
      disabled: {
        true: 'opacity-50 cursor-not-allowed',
        false: 'cursor-pointer',
      },
    },
    defaultVariants: { size: 'base' },
  },
)

// Non-option rows (empty / loading / capped tail) — outside the
// activedescendant cycle, quiet mono metadata voice.
export const comboboxNoteVariants = cva(
  'flex items-center gap-2 text-xs font-mono text-ink-faint select-none',
  {
    variants: {
      size: {
        sm: 'px-3 py-2',
        base: 'px-3.5 py-2.5',
        lg: 'px-4 py-3',
      },
    },
    defaultVariants: { size: 'base' },
  },
)
