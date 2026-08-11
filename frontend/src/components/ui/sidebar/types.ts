import type { PrtSeed, PrtSize } from '../_shared/types'

// One navigation row. A leaf links (href) or emits select; an item with
// `children` is an expandable group (grid 0fr→1fr in expanded mode;
// an anchor-positioned flyout when the rail is collapsed). One level of nesting.
export interface PrtSidebarItem {
  // stable key + what `active` matches against
  value: string | number
  label: string
  // 'i-lucide-*' class string — the rail shows the icon alone
  icon?: string
  // present → renders an <a href>; absent → a <button> that emits select
  href?: string
  disabled?: boolean
  // trailing count / status text (hidden in the rail)
  badge?: string
  // one level of sub-navigation
  children?: PrtSidebarItem[]
}

// A titled section. `items` is the rows; `label` is a faint heading (hidden in rail).
export interface PrtSidebarGroup {
  label?: string
  items: PrtSidebarItem[]
}

export interface PrtSidebarProps {
  // flat rows or titled sections (the PrtMenu idiom — detected by shape)
  items: PrtSidebarItem[] | PrtSidebarGroup[]
  // v-model (modelValue) — the DESKTOP icon-rail collapsed state (app persists it)
  modelValue?: boolean
  // which item.value is the current route — renders aria-current + data-active
  active?: string | number
  // what collapse means on desktop: shrink to an icon rail, or nothing
  collapsible?: 'icon' | 'none'
  // which edge it sits on — drives drawer placement + flyout direction + border
  side?: 'left' | 'right'
  // expanded width (any CSS length)
  width?: string
  // collapsed icon-rail width
  railWidth?: string
  // media query below which the sidebar becomes a PrtDrawer overlay
  mobileBreakpoint?: string
  // <nav aria-label> (a prop, never hardcoded)
  label?: string
  // rail-toggle aria-label when expanded
  collapseLabel?: string
  // rail-toggle aria-label when collapsed
  expandLabel?: string
  // decoration: tints the active row; cascades to descendants via data-seed
  seed?: PrtSeed
  size?: PrtSize
  // layout-only additions (margin/grid placement) — appearance goes through variants
  class?: string
}
