import { ref, watch } from 'vue'
import type { PrtChoiceOption } from './types'

// The only two modes — dark-first. No 'system': the owner pins one of two, never defers to the OS.
export type PrtMode = 'light' | 'dark'

const STORAGE_KEY = 'prt-mode'

function readInitial(): PrtMode {
  if (typeof localStorage === 'undefined') return 'dark'
  // dark is canonical — only an explicit 'light' pin opts out (matches boot.html)
  return localStorage.getItem(STORAGE_KEY) === 'light' ? 'light' : 'dark'
}

// module-level singleton — every consumer shares one mode
const mode = ref<PrtMode>(readInitial())

let started = false

// Dark/light mode — two states, dark-first. `system` was removed (2026-06-16, owner: never used).
// The first-paint <head> snippet (starters/_boot) sets data-mode BEFORE paint; this composable only
// RECONCILES after mount — never the source of first paint (the FOUC trap).
// Components never know modes exist (light is a token-value override).
export function useMode() {
  if (!started && typeof document !== 'undefined') {
    started = true
    watch(
      mode,
      (m) => {
        document.documentElement.dataset.mode = m
        // themes native form controls / scrollbars to match
        document.documentElement.style.colorScheme = m
      },
      { immediate: true },
    )
  }

  function setMode(m: PrtMode) {
    mode.value = m
    if (typeof localStorage !== 'undefined') localStorage.setItem(STORAGE_KEY, m)
  }

  function toggle() {
    setMode(mode.value === 'dark' ? 'light' : 'dark')
  }

  // labels are D13 English defaults; a preset can localize via a prop
  const modeOptions: PrtChoiceOption<PrtMode>[] = [
    { value: 'light', label: 'Light', icon: 'i-lucide-sun' },
    { value: 'dark', label: 'Dark', icon: 'i-lucide-moon' },
  ]

  return { mode, setMode, toggle, modeOptions }
}
