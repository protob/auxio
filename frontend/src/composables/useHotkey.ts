import { onKeyStroke } from '@vueuse/core'

export interface HotkeyOptions {
  // Escape needs to reach a dialog even while the cursor is in its text field.
  whileTyping?: boolean
}

function isTyping(target: EventTarget | null): boolean {
  const el = target as HTMLElement | null
  if (!el) return false
  return el.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(el.tagName)
}

// Bare-key shortcuts only. A chord belongs to the browser: intercepting Cmd-N
// costs the operator a new window and buys nothing.
export function useHotkey(key: string, handler: () => void, options: HotkeyOptions = {}) {
  onKeyStroke(key, (e) => {
    if (e.ctrlKey || e.metaKey || e.altKey) return
    if (!options.whileTyping && isTyping(e.target)) return
    e.preventDefault()
    handler()
  })
}
