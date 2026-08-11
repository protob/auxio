import { cva } from 'class-variance-authority'

// A drawer is an edge sheet: square outer corners (it has no free corners
// to round — three edges touch the viewport), hairline on the floating
// edge only. The hairline lives in the SFC's scoped block (same UA
// `border: solid` replacement story as PrtDialog); placement pinning,
// slide transitions and UA resets live there too. LOOK = surface, shadow
// and the size steps, all literal per placement × size (no string math).
// Full-page edge sheet → --shadow-overlay (the heavier modal/sheet tier,
// 09-elevation), same as PrtDialog.
export const drawerVariants = cva('bg-surface-1 text-ink shadow-overlay', {
  variants: {
    placement: {
      left: 'max-w-full',
      right: 'max-w-full',
      bottom: 'max-h-[100dvh]',
    },
    // xl/2xl are the working-panel widths — real side panels are
    // inspectors/edit forms, not thin option strips; lg→full was the gap
    size: {
      sm: '',
      base: '',
      lg: '',
      xl: '',
      '2xl': '',
      full: '',
    },
  },
  compoundVariants: [
    { placement: 'left', size: 'sm', class: 'w-72' },
    { placement: 'left', size: 'base', class: 'w-96' },
    { placement: 'left', size: 'lg', class: 'w-[28rem]' },
    { placement: 'left', size: 'xl', class: 'w-[36rem]' },
    { placement: 'left', size: '2xl', class: 'w-[48rem]' },
    { placement: 'left', size: 'full', class: 'w-screen' },
    { placement: 'right', size: 'sm', class: 'w-72' },
    { placement: 'right', size: 'base', class: 'w-96' },
    { placement: 'right', size: 'lg', class: 'w-[28rem]' },
    { placement: 'right', size: 'xl', class: 'w-[36rem]' },
    { placement: 'right', size: '2xl', class: 'w-[48rem]' },
    { placement: 'right', size: 'full', class: 'w-screen' },
    { placement: 'bottom', size: 'sm', class: 'h-64' },
    { placement: 'bottom', size: 'base', class: 'h-96' },
    { placement: 'bottom', size: 'lg', class: 'h-[32rem]' },
    { placement: 'bottom', size: 'xl', class: 'h-[40rem]' },
    { placement: 'bottom', size: '2xl', class: 'h-[48rem]' },
    { placement: 'bottom', size: 'full', class: 'h-dvh' },
  ],
  defaultVariants: { placement: 'right', size: 'base' },
})

export const drawerHeaderClass = 'flex items-start justify-between gap-4 px-5 pt-5'

export const drawerTitleClass = 'text-sm font-medium text-ink'

// icon-only button — bg-transparent is the tailwind-compat reset (trap registry)
export const drawerCloseClass =
  '-mt-1 -mr-1 inline-flex items-center justify-center h-7 w-7 rounded-control bg-transparent text-ink-faint hover:text-ink hover:bg-wash cursor-pointer outline-none prt-motion-colors focus-visible:ring-1 focus-visible:ring-accent'

export const drawerBodyClass = 'px-5 py-4 text-sm text-ink-muted overflow-y-auto'
