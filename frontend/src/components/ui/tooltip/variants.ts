// A tooltip is a label, not a state carrier: no intents — but the surface is
// seedable (decoration): quiet surface-3 by default, the whole derived set
// (bg + ink) re-tints from one seed.
// It floats → shadow-float (borders-before-shadows applies to static things).
// Placement/flip and the popover UA resets live in PrtTooltip.vue's <style>
// block (anchor positioning + @starting-style — not reliably expressible as
// utility literals); this string is the LOOK only.
export const tooltipClass =
  'pointer-events-none w-max max-w-60 px-2.5 py-1.5 text-xs font-medium bg-[var(--seed,var(--surface-3))] text-[var(--seed-ink,var(--ink))] border border-edge rounded-control shadow-float'
