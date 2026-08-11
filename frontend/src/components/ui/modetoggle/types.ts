import type { PrtSize } from '../_shared/types'
import type { PrtMode } from '../_shared/useMode'

export interface PrtModeToggleProps {
  // 'icon' (default — one button, click toggles) or 'segmented' (both visible)
  presentation?: 'icon' | 'segmented'
  // label overrides, e.g. { dark: 'Night' } — defaults from useMode().modeOptions
  labels?: Partial<Record<PrtMode, string>>
  // group/trigger aria-label — default 'Color mode'
  ariaLabel?: string
  size?: PrtSize
}
