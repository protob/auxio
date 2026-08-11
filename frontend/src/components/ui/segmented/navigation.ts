// no Vue imports in this file. Pure selection + roving-focus math; the
// component stays the imperative shell (it owns emit, the focusedIndex ref, and
// the items[i].focus() call). This nav WRAPS — distinct from the listbox-nav
// triad (select/menu/combobox), whose move does NOT wrap. Never share a fixture.

import type { PrtOption } from '../_shared/types'

type Value = string | number
type Model = Value | Value[] | undefined

// is `v` selected? `multiple` → membership; `single` → identity. Pure.
export function isActive(modelValue: Model, type: 'single' | 'multiple', v: Value): boolean {
  return type === 'multiple'
    ? Array.isArray(modelValue) && modelValue.includes(v)
    : modelValue === v
}

// `multiple`-mode toggle of `value` against the model. Returns a FRESH array —
// never mutates `modelValue` (a non-array model is treated as empty).
export function toggle(modelValue: Model, value: Value): Value[] {
  const next = Array.isArray(modelValue) ? [...modelValue] : []
  const i = next.indexOf(value)
  if (i >= 0) next.splice(i, 1)
  else next.push(value)
  return next
}

// initial roving index: active-and-enabled, else first-enabled, else 0.
export function firstFocusable(
  options: PrtOption[],
  modelValue: Model,
  type: 'single' | 'multiple',
): number {
  const active = options.findIndex((o) => isActive(modelValue, type, o.value) && !o.disabled)
  if (active >= 0) return active
  const enabled = options.findIndex((o) => !o.disabled)
  return enabled >= 0 ? enabled : 0
}

// Next focusable index from `from` in `dir`, skipping disabled, WITH wrap
// (`(i+dir+n)%n`, bounded to one lap). Returns -1 when nothing is focusable
// (the group is disabled, or every option is) — the caller then does nothing.
export function moveFocus(
  options: PrtOption[],
  from: number,
  dir: 1 | -1,
  groupDisabled: boolean,
): number {
  if (groupDisabled) return -1
  const n = options.length
  let i = from
  for (let step = 0; step < n; step++) {
    i = (i + dir + n) % n
    const opt = options[i]
    if (opt && !opt.disabled) return i
  }
  return -1
}
