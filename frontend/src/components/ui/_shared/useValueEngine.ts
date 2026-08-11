import { onBeforeUnmount, ref, watchEffect, type Ref } from 'vue'

/*
Headless value engine — slider now, rotary knob later. Geometry-blind:
ALL behavior lives here (pointer capture, absolute/relative drag,
tapers, Shift-fine, detents, keyboard, wheel, reset, ARIA); the
renderer owns only geometry and supplies it through the adapter
(`phaseAt` / `pixelDelta`). A future PrtKnob is this engine + a rotary
adapter (vertical-drag relative mode, `trackClick: 'none'`) — see
0_notes/research/tier-D/slider-and-knobs/a/00-consensus.md and
0_docs/10-value-controls.md before building it.

Core discipline — the phase/value split: phase is 0..1 physical
position, value is domain units (Hz, dB, €). Pointer math and
rendering touch ONLY phase; the taper is a pure function pair between
the two domains.

Perf fast path (sanctioned dynamic channel, 08 known-benign list):
during drag the engine writes `--prt-slider-phase` (and `-b` for a
second thumb) straight onto the host element via `style.setProperty`,
rAF-throttled — NO Vue reactivity in the pointermove path. v-model
commits ride the same rAF.
*/

export interface ValueTaperPair {
  // value (domain units) → phase 0..1
  toPhase(value: number): number
  // phase 0..1 → value (domain units)
  toValue(phase: number): number
}

export type ValueTaper = 'linear' | 'log' | 'db' | ValueTaperPair

export interface ValueEngineGeometry {
  // absolute phase (clamped 0..1) at the pointer position — renderer maps from its track rect
  phaseAt(event: PointerEvent): number
  // signed pixels along the INCREASING axis from a to b (vertical: up is positive)
  pixelDelta(a: { x: number; y: number }, b: { x: number; y: number }): number
}

export interface ValueEngineOptions {
  // committed values, domain units — 1 entry per thumb (dual = [lo, hi])
  values: () => number[]
  min: () => number
  max: () => number
  // value-domain quantization; also flips keyboard into value-domain steps
  step: () => number | undefined
  // reset target for dbl-click / Ctrl-click — reset disabled when undefined
  defaultValue: () => number | undefined
  taper: () => ValueTaper
  interactionMode: () => 'absolute' | 'relative'
  trackClick: () => 'jump' | 'pickup' | 'none'
  // px of drag for full range in relative mode
  sensitivity: () => number
  // Shift multiplier for drag/wheel/phase-keyboard
  fineScale: () => number
  // magnetic value points — epsilon clamp with a hidden raw accumulator (drag only)
  detents: () => number[] | undefined
  wheel: () => boolean
  // dual-thumb crossing policy
  collision: () => 'clamp' | 'push' | 'swap'
  vertical: () => boolean
  disabled: () => boolean
  // feeds the readout and aria-valuetext
  format: () => ((value: number) => string) | undefined
  // rAF-cadence commit during drag (v-model)
  onInput: (values: number[]) => void
  // commit on release / keyboard / wheel
  onChange: (values: number[]) => void
  onStart?: () => void
  onEnd?: () => void
  // move focus to a thumb after a track press (keyboard continuation)
  focusThumb?: (index: number) => void
}

// phase radius a detent holds the rendered value inside (raw keeps accumulating)
const DETENT_EPSILON = 0.025
// phase step for arrow keys when no explicit `step` exists (page = 10×)
const PHASE_ARROW = 0.01

const clamp01 = (p: number) => Math.min(1, Math.max(0, p))
// kill float noise (0.30000000000000004 → 0.3) without visible rounding
const tidy = (v: number) => Math.round(v * 1e9) / 1e9
const clampVal = (v: number, min: number, max: number) => Math.min(max, Math.max(min, v))

// ── Pure math (geometry-blind, opts-blind) ───────────────────────────
// Extracted to module scope so the taper/quantize/collide contract is
// testable directly (no engine instance, no DOM, no lifecycle hooks) —
// the drag plumbing below binds these to the reactive `opts` getters.
// Behaviour-preserving extraction; see 0_notes/plans/testing/03-value-engine.md.

// Build the phase⇄value taper pair for a given range.
export function makeTaper(taper: ValueTaper, min: number, max: number): ValueTaperPair {
  const span = max - min || 1
  if (typeof taper === 'object') return taper
  if (taper === 'log' && min > 0) {
    const ratio = max / min
    return {
      toPhase: (v) => clamp01(Math.log(Math.max(v, min) / min) / Math.log(ratio)),
      toValue: (p) => min * ratio ** clamp01(p),
    }
  }
  if (taper === 'db') {
    // audio-fader taper: linear-in-amplitude^(1/4) — unity gain lands
    // around phase 0.7–0.85 like a console fader; min is the floor the
    // consumer's format may render as −∞
    const ampMin = 10 ** (min / 20)
    const ampMax = 10 ** (max / 20)
    return {
      toPhase: (v) => clamp01(((10 ** (v / 20) - ampMin) / (ampMax - ampMin)) ** 0.25),
      toValue: (p) => 20 * Math.log10(ampMin + (ampMax - ampMin) * clamp01(p) ** 4),
    }
  }
  // 'linear' — and the guarded fallback for 'log' with min <= 0
  return {
    toPhase: (v) => clamp01((v - min) / span),
    toValue: (p) => min + clamp01(p) * span,
  }
}

// Snap to `step` (when set) and clamp to range; `tidy` kills float noise.
export function quantize(v: number, min: number, max: number, step?: number): number {
  if (!step) return tidy(clampVal(v, min, max))
  const snapped = min + Math.round((v - min) / step) * step
  return tidy(clampVal(snapped, min, max))
}

// Drag pipeline: raw phase → (detent ε-clamp) → value (quantized).
export function valueFromRaw(
  raw: number,
  taper: ValueTaper,
  min: number,
  max: number,
  step?: number,
  detents?: number[],
): number {
  const pair = makeTaper(taper, min, max)
  if (detents) {
    for (const dv of detents) {
      if (Math.abs(raw - pair.toPhase(dv)) <= DETENT_EPSILON) return tidy(clampVal(dv, min, max))
    }
  }
  return quantize(pair.toValue(raw), min, max, step)
}

// Dual-thumb crossing policy — returns the full next array (never mutates input).
export function collide(
  values: number[],
  index: number,
  v: number,
  mode: 'clamp' | 'push' | 'swap',
): { values: number[]; index: number } {
  if (values.length < 2) return { values: [v], index }
  const next = [...values]
  const sibling = index === 0 ? next[1] : next[0]
  if (mode === 'clamp') {
    next[index] = index === 0 ? Math.min(v, sibling) : Math.max(v, sibling)
    return { values: next, index }
  }
  if (mode === 'push') {
    next[index] = v
    if (index === 0 && v > next[1]) next[1] = v
    if (index === 1 && v < next[0]) next[0] = v
    return { values: next, index }
  }
  // swap: values may cross — reorder and hand the drag to the other index
  next[index] = v
  if (next[0] > next[1]) {
    const [lo, hi] = [next[1], next[0]]
    next[0] = lo
    next[1] = hi
    return { values: next, index: index === 0 ? 1 : 0 }
  }
  return { values: next, index }
}

// Keyboard/wheel increment — pure (smaller surface than a harness, mirrors
// `collide`). With an explicit `step` the move is value-domain: exactly `step`
// per arrow, ×10 for page. Without one it is phase-domain (re-quantized through
// the taper) so a log taper gives uniform phase deltas, non-uniform value
// deltas. Always clamped to [min,max]. Home/End live at the DOM seam, not here.
export function keyboardStep(
  values: number[],
  index: number,
  dir: 1 | -1,
  page: boolean,
  fine: boolean,
  opts: { taper: ValueTaper; min: number; max: number; step?: number; fineScale: number },
): number {
  const { taper, min, max, step, fineScale } = opts
  if (step) {
    // explicit step is a contract — value-domain, exactly step (×10 page)
    return quantize(values[index] + dir * step * (page ? 10 : 1), min, max, step)
  }
  // taper without step: phase-domain so increments feel uniform on a log scale
  const pair = makeTaper(taper, min, max)
  const delta = dir * (page ? PHASE_ARROW * 10 : PHASE_ARROW) * (fine ? fineScale : 1)
  return quantize(pair.toValue(clamp01(pair.toPhase(values[index]) + delta)), min, max, step)
}

// Dual-thumb aria-valuemin/valuemax bounds. clamp/push pin each thumb to its
// sibling (lower thumb's max = upper's value, and vice-versa); swap and the
// single-thumb case keep the full [min,max].
export function thumbBounds(
  values: number[],
  index: number,
  collision: 'clamp' | 'push' | 'swap',
  min: number,
  max: number,
): { valuemin: number; valuemax: number } {
  const clampish = values.length > 1 && collision !== 'swap'
  return {
    valuemin: clampish && index === 1 ? values[0] : min,
    valuemax: clampish && index === 0 ? values[1] : max,
  }
}

export function useValueEngine(
  host: Ref<HTMLElement | null>,
  geometry: ValueEngineGeometry,
  opts: ValueEngineOptions,
) {
  // ── taper + value pipeline (bind the pure module math to opts) ────────
  const taperPair = (): ValueTaperPair => makeTaper(opts.taper(), opts.min(), opts.max())
  const toPhase = (v: number) => taperPair().toPhase(v)
  const quantizeValue = (v: number) => quantize(v, opts.min(), opts.max(), opts.step())
  const rawToValue = (raw: number) =>
    valueFromRaw(raw, opts.taper(), opts.min(), opts.max(), opts.step(), opts.detents())
  const collideAt = (values: number[], index: number, v: number) =>
    collide(values, index, v, opts.collision())

  // ── fast path: phase CSS vars on the host element ────────────────────
  function writeVars(values: number[]) {
    const el = host.value
    if (!el) return
    el.style.setProperty('--prt-slider-phase', String(toPhase(values[0])))
    if (values.length > 1)
      el.style.setProperty('--prt-slider-phase-b', String(toPhase(values[1])))
  }

  // ── drag session (plain locals — NO reactivity in the move path) ─────
  // reactive only at session boundaries: readout + z-order + start/end
  const dragging = ref<number | null>(null)
  // last-interacted thumb — z-order (overlapping thumbs dead-click otherwise)
  const top = ref(0)

  let session: {
    index: number
    raw: number
    last: { x: number; y: number }
    lastAbs: number
    mode: 'absolute' | 'relative'
    el: HTMLElement
    pointerId: number
    values: number[] // working copy — the source of truth while dragging
  } | null = null

  let raf = 0
  let pending: number[] | null = null

  function scheduleCommit(values: number[]) {
    pending = values
    if (raf) return
    raf = requestAnimationFrame(() => {
      raf = 0
      if (!pending) return
      writeVars(pending)
      opts.onInput(pending)
      pending = null
    })
  }

  function beginDrag(
    el: HTMLElement,
    event: PointerEvent,
    index: number,
    raw: number,
    mode: 'absolute' | 'relative',
    values: number[],
  ) {
    session = {
      index,
      raw,
      last: { x: event.clientX, y: event.clientY },
      lastAbs: geometry.phaseAt(event),
      mode,
      el,
      pointerId: event.pointerId,
      values,
    }
    el.setPointerCapture(event.pointerId)
    el.addEventListener('pointermove', onMove)
    el.addEventListener('pointerup', endDrag)
    el.addEventListener('pointercancel', endDrag)
    // Firefox release-path quirks (Radix #2777/#1760): lostpointercapture
    // is the teardown of last resort — Esc-during-drag and window blur
    // leak a stuck "dragging" state otherwise
    el.addEventListener('lostpointercapture', endDrag)
    dragging.value = index
    top.value = index
    opts.onStart?.()
  }

  function onMove(event: PointerEvent) {
    if (!session) return
    const fine = event.shiftKey ? opts.fineScale() : 1
    const cur = { x: event.clientX, y: event.clientY }
    if (session.mode === 'absolute') {
      // telescoping absolute deltas: tracks the pointer exactly while
      // unmodified (grab offset preserved — no jump on thumb grab), and
      // degrades into relative fine motion the moment Shift is held
      const abs = geometry.phaseAt(event)
      session.raw = clamp01(session.raw + (abs - session.lastAbs) * fine)
      session.lastAbs = abs
    } else {
      session.raw = clamp01(
        session.raw + (geometry.pixelDelta(session.last, cur) / opts.sensitivity()) * fine,
      )
    }
    session.last = cur
    const hit = collideAt(session.values, session.index, rawToValue(session.raw))
    session.values = hit.values
    session.index = hit.index
    scheduleCommit(hit.values)
  }

  function endDrag() {
    if (!session) return
    const { el, pointerId, values } = session
    session = null
    el.removeEventListener('pointermove', onMove)
    el.removeEventListener('pointerup', endDrag)
    el.removeEventListener('pointercancel', endDrag)
    el.removeEventListener('lostpointercapture', endDrag)
    if (el.hasPointerCapture(pointerId)) el.releasePointerCapture(pointerId)
    if (raf) {
      cancelAnimationFrame(raf)
      raf = 0
      pending = null
    }
    writeVars(values)
    opts.onInput(values)
    opts.onChange(values)
    dragging.value = null
    opts.onEnd?.()
    // re-sync from the model after the echo settles (covers parents that
    // reject or transform the committed value)
    requestAnimationFrame(() => {
      if (!session) writeVars(opts.values())
    })
  }

  // ── pointer entries ──────────────────────────────────────────────────
  function onThumbPointerdown(event: PointerEvent, index: number) {
    if (opts.disabled() || event.button !== 0) return
    event.preventDefault()
    event.stopPropagation()
    opts.focusThumb?.(index)
    if ((event.ctrlKey || event.metaKey) && opts.defaultValue() !== undefined) {
      reset(index)
      return
    }
    const values = [...opts.values()]
    beginDrag(
      event.currentTarget as HTMLElement,
      event,
      index,
      toPhase(values[index]),
      opts.interactionMode(),
      values,
    )
  }

  function onTrackPointerdown(event: PointerEvent) {
    if (opts.disabled() || event.button !== 0) return
    if (opts.trackClick() === 'none') return
    event.preventDefault()
    const values = [...opts.values()]
    const at = geometry.phaseAt(event)
    // dual: nearest thumb picks up the press (ties go to the upper thumb
    // so a collapsed pair at min stays grabbable)
    let index = 0
    if (values.length > 1) {
      const dLo = Math.abs(at - toPhase(values[0]))
      const dHi = Math.abs(at - toPhase(values[1]))
      index = dHi <= dLo ? 1 : 0
    }
    opts.focusThumb?.(index)
    let raw = toPhase(values[index])
    let working = values
    if (opts.trackClick() === 'jump') {
      // hybrid (consensus §3.3): jump to the pressed position, then the
      // SAME drag continues without further jumps
      raw = at
      const hit = collideAt(values, index, rawToValue(raw))
      working = hit.values
      index = hit.index
      writeVars(working)
      opts.onInput(working)
    }
    // pickup never jumps — and the continuation is always relative
    // (an absolute continuation would teleport on the first move)
    const mode = opts.trackClick() === 'jump' ? opts.interactionMode() : 'relative'
    beginDrag(event.currentTarget as HTMLElement, event, index, raw, mode, working)
  }

  // ── keyboard ─────────────────────────────────────────────────────────
  // bind the pure module step to the reactive opts (same idiom as collideAt)
  const stepFromKey = (values: number[], index: number, dir: 1 | -1, page: boolean, fine: boolean) =>
    keyboardStep(values, index, dir, page, fine, {
      taper: opts.taper(),
      min: opts.min(),
      max: opts.max(),
      step: opts.step(),
      fineScale: opts.fineScale(),
    })

  function commitDiscrete(values: number[], index: number, v: number) {
    const hit = collideAt(values, index, v)
    writeVars(hit.values)
    opts.onInput(hit.values)
    opts.onChange(hit.values)
  }

  function onThumbKeydown(event: KeyboardEvent, index: number) {
    if (opts.disabled()) return
    const values = [...opts.values()]
    let next: number
    switch (event.key) {
      case 'ArrowRight':
      case 'ArrowUp':
        next = stepFromKey(values, index, 1, false, event.shiftKey)
        break
      case 'ArrowLeft':
      case 'ArrowDown':
        next = stepFromKey(values, index, -1, false, event.shiftKey)
        break
      case 'PageUp':
        next = stepFromKey(values, index, 1, true, event.shiftKey)
        break
      case 'PageDown':
        next = stepFromKey(values, index, -1, true, event.shiftKey)
        break
      case 'Home':
        next = quantizeValue(opts.min())
        break
      case 'End':
        next = quantizeValue(opts.max())
        break
      default:
        return
    }
    event.preventDefault()
    top.value = index
    commitDiscrete(values, index, next)
  }

  // ── wheel (off by default; bound { passive: false } only when on) ────
  function onWheel(event: WheelEvent) {
    if (opts.disabled()) return
    event.preventDefault()
    // Shift turns deltaY into deltaX on most mice — read whichever moved
    const dy = event.deltaY !== 0 ? event.deltaY : event.deltaX
    if (dy === 0) return
    const values = [...opts.values()]
    const index = values.length > 1 ? top.value : 0
    const next = stepFromKey(values, index, dy < 0 ? 1 : -1, false, event.shiftKey)
    commitDiscrete(values, index, next)
  }

  let wheelEl: HTMLElement | null = null
  const stopWheel = watchEffect(() => {
    const el = host.value
    const on = opts.wheel()
    if (wheelEl) {
      wheelEl.removeEventListener('wheel', onWheel)
      wheelEl = null
    }
    if (el && on) {
      // element-level binding = adjusts only while hovered; explicit
      // passive: false so preventDefault keeps the page from scrolling
      el.addEventListener('wheel', onWheel, { passive: false })
      wheelEl = el
    }
  })

  // ── reset (dbl-click / Ctrl-click → defaultValue) ────────────────────
  function reset(index: number) {
    const dv = opts.defaultValue()
    if (dv === undefined || opts.disabled()) return
    top.value = index
    commitDiscrete([...opts.values()], index, quantizeValue(dv))
  }

  function onThumbDblclick(_event: MouseEvent, index: number) {
    reset(index)
  }

  // ── ARIA (spread onto each thumb) ────────────────────────────────────
  function thumbAria(index: number): Record<string, string | number | undefined> {
    const values = opts.values()
    const v = values[index]
    const format = opts.format()
    const nonLinear = opts.taper() !== 'linear'
    const bounds = thumbBounds(values, index, opts.collision(), opts.min(), opts.max())
    return {
      role: 'slider',
      tabindex: opts.disabled() ? -1 : 0,
      'aria-valuemin': bounds.valuemin,
      'aria-valuemax': bounds.valuemax,
      'aria-valuenow': v,
      'aria-valuetext': format ? format(v) : nonLinear ? String(v) : undefined,
      'aria-orientation': opts.vertical() ? 'vertical' : undefined,
      'aria-disabled': opts.disabled() ? 'true' : undefined,
    }
  }

  // model/taper/range changes outside a drag re-sync the phase vars
  // (also paints the initial position on mount)
  const stopSync = watchEffect(() => {
    const values = opts.values()
    // track the inputs the phase depends on
    opts.min()
    opts.max()
    opts.taper()
    if (!session) writeVars(values)
  })

  onBeforeUnmount(() => {
    endDrag()
    stopWheel()
    stopSync()
    if (wheelEl) wheelEl.removeEventListener('wheel', onWheel)
  })

  return {
    // index of the thumb being dragged (readout), null when idle
    dragging,
    // last-interacted thumb — bind z-order from it
    top,
    // value → 0..1 through the active taper (tick positioning)
    phaseOf: toPhase,
    thumbAria,
    onThumbPointerdown,
    onThumbKeydown,
    onThumbDblclick,
    onTrackPointerdown,
    reset,
  }
}
