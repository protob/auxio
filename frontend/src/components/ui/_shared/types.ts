// No Vue imports — portability rule.

export type PrtSize = 'sm' | 'base' | 'lg'

export type PrtEdges = 'square' | 'rounded' | 'pill'

export type PrtSeedValue = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10

// A seed is ONE value a whole coordinated style set derives from
// (background, ink, hover, wash — computed in tokens.css, never in components).
// The union is closed on purpose: a wrong seed is a TypeScript error.
// Accepts number and string forms so both `seed="7"` (template-clean) and `:seed="item.seed"` type-check.
export type PrtSeed = PrtSeedValue | `${PrtSeedValue}`

// True-state vocabulary for notices/toasts/badges. NOT decoration —
// decoration is seeds (CONVENTIONS "Seeds vs states"). Closed on purpose.
// `success` rides the accent tokens; the others ride --info/--warning/--danger with their --*-ink pairs.
export type PrtIntent = 'info' | 'success' | 'warning' | 'danger'

export interface PrtOption {
  value: string | number
  label: string
  disabled?: boolean
  // icon class string, e.g. 'i-lucide-home'
  icon?: string
}

// A choice in a switcher/selector family member (mode toggle, language switcher, future theme switcher).
// A typed superset of PrtOption: the value is generic (e.g. a ModePreference union), plus presentation hints.
// Shipping the TYPE (not a `useSelection` core) is deliberate — existing primitives already take options + v-model.
export interface PrtChoiceOption<T extends string = string> {
  value: T
  label: string
  // terse label for compact (segmented/buttons) presentations
  shortLabel?: string
  // icon class string, e.g. 'i-lucide-sun' (custom 'i-prt-*' too)
  icon?: string
  // future theme swatch colour
  swatch?: string
  // BCP-47 code — language options set this for :lang()/SR pronunciation
  lang?: string
  disabled?: boolean
}
