import type { PrtIntent, PrtSeed } from '../_shared/types'
import type { PrtNoticeVariant } from './variants'

export interface PrtNoticeProps {
  // true state (semantic tokens); beats seed when both are set. Default tone is info when neither is set
  intent?: PrtIntent
  // decoration — pad 1–10 from the rack (tip boxes, callouts without semantics)
  seed?: PrtSeed
  // edge (default): left tone edge on quiet surface · outline: full tone frame,
  // transparent fill (btn/chip meaning; calmest seeded) · wash: tinted bg + hairline ·
  // ghost: borderless wash block
  variant?: PrtNoticeVariant
  // heading line; #title slot overrides
  title?: string
  // body; default slot overrides
  description?: string
  // override the tone icon; '' hides it (seed tone has no icon by default)
  icon?: string
  // monochrome ("typographic") treatment — orthogonal to variant: 'title' puts the tone
  // color on the title, body stays muted ink (the safer ship); true also dims the body
  // into the tone (derived --seed-body / --*-body)
  mono?: boolean | 'title'
  // renders the × button; clicking emits 'dismiss' (visibility is the parent's v-if)
  dismissible?: boolean
  // layout-only additions (margin, width, grid placement)
  class?: string
}
