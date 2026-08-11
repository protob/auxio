// The menu-presentation trigger is icon-only, so it MUST carry bg-transparent
// (the tailwind-compat ButtonFace reset — the trap registry's button-bg trap),
// or the UA button face shows through.
export const modeToggleTriggerClass =
  'inline-flex items-center justify-center h-9 w-9 rounded-control bg-transparent text-ink-muted hover:bg-wash hover:text-ink prt-motion-colors prt-ring'
