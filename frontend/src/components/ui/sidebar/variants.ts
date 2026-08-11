import { cva } from 'class-variance-authority'

// The panel column. Width + the rail transition live in the SFC scoped block
// (truly-dynamic widths ride style-bound CSS vars). Here: the chrome — surface,
// the edge hairline on the side it sits, the flex column that pins header/footer
// and scrolls the nav body.
export const sidebarRootVariants = cva(
  'prt-sidebar flex flex-col min-h-0 bg-surface-1 text-ink overflow-hidden',
  {
    variants: {
      side: {
        left: 'border-r border-edge',
        right: 'border-l border-edge',
      },
    },
    defaultVariants: { side: 'left' },
  },
)

// section heading — faint mono uppercase metadata type (house voice); the SFC
// hides it (and collapses its row) in the rail.
export const sidebarGroupLabelVariants = cva(
  'px-3 pt-4 pb-1.5 text-[0.6875rem] font-medium tracking-[0.103125rem] uppercase text-ink-faint select-none',
  {
    variants: {
      size: {
        sm: 'pt-3 text-[0.625rem]',
        base: '',
        lg: 'pt-5 text-xs',
      },
    },
    defaultVariants: { size: 'base' },
  },
)

// a nav row (leaf link/button OR a group trigger face). The active fill rides
// the seed cascade — default (no seed) lands on wash + accent ink, a seed tints
// the lot. rail centers the icon and hides the label (handled in the SFC).
export const sidebarItemVariants = cva(
  'group/item relative w-full inline-flex items-center gap-2.5 rounded-control text-left outline-none prt-motion-colors focus-visible:ring-1 focus-visible:ring-accent',
  {
    variants: {
      size: {
        sm: 'h-8 px-2.5 text-xs',
        base: 'h-9 px-3 text-sm',
        lg: 'h-10 px-3 text-base',
      },
      // neutral by default — a quiet elevated surface, NOT a colored fill
      // (color is a deliberate pluck). A seed tints it; the lone accent is the
      // small "you are here" tick (sidebarItemMarkClass).
      // inactive rows are bg-transparent — i.e. the sidebar background itself.
      // (bg-transparent is load-bearing: the tailwind-compat reset leaves a UA
      // ButtonFace bg on <button> otherwise — the grey "2010" boxes.) Only the
      // current item carries a fill; the lone accent is the leading tick.
      active: {
        true: 'bg-[var(--seed-wash,var(--surface-3))] text-[var(--seed,var(--ink))] font-medium',
        false: 'bg-transparent text-ink-muted hover:bg-wash hover:text-ink',
      },
      disabled: {
        true: 'opacity-50 cursor-not-allowed pointer-events-none',
        false: 'cursor-pointer',
      },
      rail: {
        true: 'justify-center px-0',
        false: '',
      },
    },
    defaultVariants: { size: 'base', active: false, disabled: false, rail: false },
  },
)

// the active rail mark — a short accent bar pinned to the active row's leading
// edge (the Linear/Vercel "you are here" tick). Side flips it; rail keeps it.
export const sidebarItemMarkClass =
  'absolute inset-y-1.5 w-0.5 rounded-mark bg-[var(--seed,var(--accent))]'

// indent for nested children under an expanded group
export const sidebarSubItemVariants = cva(
  'group/item relative w-full inline-flex items-center gap-2.5 rounded-control text-left outline-none prt-motion-colors focus-visible:ring-1 focus-visible:ring-accent',
  {
    variants: {
      size: {
        sm: 'h-7 pl-9 pr-2.5 text-xs',
        base: 'h-8 pl-10 pr-3 text-sm',
        lg: 'h-9 pl-11 pr-3 text-base',
      },
      active: {
        true: 'bg-transparent text-[var(--seed,var(--ink))] font-medium',
        false: 'bg-transparent text-ink-faint hover:bg-wash hover:text-ink-muted',
      },
      disabled: {
        true: 'opacity-50 cursor-not-allowed pointer-events-none',
        false: 'cursor-pointer',
      },
    },
    defaultVariants: { size: 'base', active: false, disabled: false },
  },
)

// pinned header / footer regions inside the column (logo, profile). A hairline
// separates them from the scrolling nav.
export const sidebarHeaderClass =
  'shrink-0 flex items-center gap-2.5 min-h-[3.25rem] px-3 border-b border-edge'
export const sidebarFooterClass =
  'shrink-0 flex items-center gap-2.5 min-h-[3.25rem] px-3 border-t border-edge'

// the scroll region holding the nav lists
export const sidebarNavClass = 'flex-1 min-h-0 overflow-y-auto overscroll-contain px-2 py-2'

// trailing badge text
export const sidebarBadgeClass =
  'ml-auto shrink-0 text-[0.6875rem] font-mono tabular-nums text-ink-faint'

// chevron on an expandable group trigger
export const sidebarChevronClass =
  'i-lucide-chevron-down ml-auto shrink-0 text-ink-faint prt-motion'

// ── the collapsed-rail flyout (a group's children float beside the icon) ──
// Same flat + full-bleed panel doctrine + popover/anchor substrate as PrtMenu.
export const sidebarFlyoutPanelClass =
  'prt-sidebar-flyout bg-surface-2 rounded-surface shadow-float ring-1 ring-edge min-w-[12rem] overflow-hidden'
export const sidebarFlyoutLabelClass =
  'px-3 pt-2.5 pb-1.5 text-[0.6875rem] font-medium tracking-[0.103125rem] uppercase text-ink-faint'
// The only outline-none in the kit that ringed nothing back. Inset, unlike the
// rail items': these rows are full-bleed in a panel that clips, so an outset
// ring loses its outer edge.
export const sidebarFlyoutItemVariants = cva(
  'w-full inline-flex items-center gap-2.5 px-3 h-9 text-sm text-left outline-none prt-motion-colors focus-visible:ring-1 focus-visible:ring-inset focus-visible:ring-accent',
  {
    variants: {
      active: {
        true: 'bg-[var(--seed-wash,var(--surface-3))] text-[var(--seed,var(--ink))] font-medium',
        false: 'bg-transparent text-ink-muted hover:bg-surface-3 hover:text-ink',
      },
      disabled: {
        true: 'opacity-50 cursor-not-allowed pointer-events-none',
        false: 'cursor-pointer',
      },
    },
    defaultVariants: { active: false, disabled: false },
  },
)

// the rail collapse/expand toggle button (lives in the header region)
export const sidebarToggleClass =
  'shrink-0 inline-flex items-center justify-center h-8 w-8 rounded-control bg-transparent text-ink-faint hover:bg-wash hover:text-ink outline-none prt-motion-colors focus-visible:ring-1 focus-visible:ring-accent cursor-pointer'
