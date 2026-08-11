import type { PrtIntent } from '../_shared/types'

// House look: quiet surface + hairline edge + a 2px tone edge on the left
// (the technical callout mark — what keeps the neutral ones from reading
// generic). wash = tinted background; ghost = borderless wash block.
// Color as a deliberate pluck — never solid bg-el-* slabs.

// 'edge' on purpose, NOT 'outline': in this kit outline means a full tone
// border (btn/chip) — and notice's 'outline' honors exactly that meaning:
// full tone frame, transparent fill. Calmest seeded (decorative frame);
// for alerts prefer 'edge' — a full intent frame shouts.
export type PrtNoticeVariant = 'edge' | 'outline' | 'wash' | 'ghost'

// tone = intent (true state) or 'seed' (decoration; intent beats seed)
export type PrtNoticeTone = PrtIntent | 'seed'

export const noticeBase = 'flex items-start gap-3 p-4 rounded-surface'

// literal lookup tables — extraction manifests, same discipline as CVA
export const noticeVariantBase: Record<PrtNoticeVariant, string> = {
  edge: 'bg-surface-1 border border-edge border-l-2',
  outline: 'bg-transparent border',
  wash: 'border border-edge',
  ghost: '',
}

export const noticeToneClasses: Record<
  PrtNoticeVariant,
  Record<PrtNoticeTone, string>
> = {
  edge: {
    info: 'border-l-info',
    success: 'border-l-accent',
    warning: 'border-l-warning',
    danger: 'border-l-danger',
    seed: 'border-l-[var(--seed,var(--edge))]',
  },
  outline: {
    info: 'border-info',
    success: 'border-accent',
    warning: 'border-warning',
    danger: 'border-danger',
    seed: 'border-[var(--seed,var(--edge))]',
  },
  wash: {
    info: 'bg-[var(--info-wash)]',
    success: 'bg-[var(--accent-wash)]',
    warning: 'bg-[var(--warning-wash)]',
    danger: 'bg-[var(--danger-wash)]',
    seed: 'bg-[var(--seed-wash,var(--wash))]',
  },
  ghost: {
    info: 'bg-[var(--info-wash)]',
    success: 'bg-[var(--accent-wash)]',
    warning: 'bg-[var(--warning-wash)]',
    danger: 'bg-[var(--danger-wash)]',
    seed: 'bg-[var(--seed-wash,var(--wash))]',
  },
}

// titleColor/bodyColor serve the `mono` treatment (the typographic notice):
// title speaks the tone at full strength; body rides the derived dimmed
// tokens (--seed-body / --*-body, ~75% toward the surface) so the title
// still leads. lint:rack measures the body case — bright pads (9 in light)
// sit below the floor; `mono="title"` is the safer ship there.
export const noticeTone: Record<
  PrtNoticeTone,
  { icon: string; iconColor: string; titleColor: string; bodyColor: string }
> = {
  info: {
    icon: 'i-lucide-info',
    iconColor: 'text-info',
    titleColor: 'text-info',
    bodyColor: 'text-[var(--info-body)]',
  },
  success: {
    icon: 'i-lucide-circle-check',
    iconColor: 'text-accent',
    titleColor: 'text-accent',
    bodyColor: 'text-[var(--accent-body)]',
  },
  warning: {
    icon: 'i-lucide-triangle-alert',
    iconColor: 'text-warning',
    titleColor: 'text-warning',
    bodyColor: 'text-[var(--warning-body)]',
  },
  danger: {
    icon: 'i-lucide-circle-x',
    iconColor: 'text-danger',
    titleColor: 'text-danger',
    bodyColor: 'text-[var(--danger-body)]',
  },
  // decoration carries no semantics — no default icon; `icon` prop opts in
  seed: {
    icon: '',
    iconColor: 'text-[var(--seed,var(--ink-muted))]',
    titleColor: 'text-[var(--seed,var(--ink))]',
    bodyColor: 'text-[var(--seed-body,var(--ink-muted))]',
  },
}
