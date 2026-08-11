import type { PrtSeed, PrtSize } from '../_shared/types'

// bar is the inline default; circle is the determinate ring. Determinate only — indeterminate/top-bar duty is PrtLoadingBar's
export type PrtProgressType = 'bar' | 'circle'

export interface PrtProgressProps {
  // 0–100, default 0
  value?: number
  type?: PrtProgressType
  seed?: PrtSeed
  size?: PrtSize
  // numeric label (tabular-nums). Default: bars hide it, circle shows it centered — pass false to hide
  showValue?: boolean
  // layout-only additions (width/margin)
  class?: string
}
