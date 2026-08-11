import type { PrtSize } from '../_shared/types'

export interface PrtTabItem {
  value: string | number
  label: string
  // icon class string, e.g. 'i-lucide-settings-2'
  icon?: string
  disabled?: boolean
}

export interface PrtTabsProps {
  // v-model — controlled only, no internal fallback state
  modelValue: string | number
  items: PrtTabItem[]
  size?: PrtSize
  // layout-only additions (margin/width) — appearance goes through variants
  class?: string
}
