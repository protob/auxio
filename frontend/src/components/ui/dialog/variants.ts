import { cva } from 'class-variance-authority'

// The dialog floats over the scrim as a full-page overlay → --shadow-overlay
// (the heavier modal/sheet tier, 09-elevation), not the lighter --shadow-float
// used for menus/tooltips. The hairline border lives in the SFC's scoped
// block, not here: the UA sheet gives <dialog> `border: solid` (medium), so
// the reset and the 1px --edge replacement must sit at the same (scoped)
// specificity — a `border` utility would lose to the scoped reset.
export const dialogVariants = cva('w-full bg-surface-1 text-ink shadow-overlay', {
  variants: {
    // xl/2xl are for advanced multi-column forms, not confirms
    size: {
      sm: 'max-w-sm',
      base: 'max-w-lg',
      lg: 'max-w-2xl',
      xl: 'max-w-4xl',
      '2xl': 'max-w-6xl',
      full: 'max-w-[calc(100vw-2rem)]',
    },
    edges: {
      square: 'rounded-none',
      rounded: 'rounded-surface',
      pill: 'rounded-surface',
    },
  },
  defaultVariants: { size: 'base', edges: 'rounded' },
})

// header row: title left, X right; title speaks quietly (weight restraint)
export const dialogHeaderClass = 'flex items-start justify-between gap-4 px-5 pt-5'

export const dialogTitleClass = 'text-sm font-medium text-ink'

// icon-only button — bg-transparent is the tailwind-compat reset (trap registry)
export const dialogCloseClass =
  '-mt-1 -mr-1 inline-flex items-center justify-center h-7 w-7 rounded-control bg-transparent text-ink-faint hover:text-ink hover:bg-wash cursor-pointer outline-none prt-motion-colors focus-visible:ring-1 focus-visible:ring-accent'

export const dialogBodyClass = 'px-5 py-4 text-sm text-ink-muted'

export const dialogActionsClass = 'flex justify-end gap-2 px-5 pb-5'
