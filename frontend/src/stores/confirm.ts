import { ref, shallowRef } from 'vue'
import { defineStore } from 'pinia'

export interface ConfirmOptions {
  title: string
  message: string
  confirmLabel?: string
  danger?: boolean
}

export const useConfirmStore = defineStore('confirm', () => {
  const options = shallowRef<ConfirmOptions | null>(null)
  const open = ref(false)
  let settle: ((ok: boolean) => void) | null = null

  // Awaitable so call sites keep reading like `if (!await confirm(...)) return`,
  // which is the one thing window.confirm had going for it.
  function ask(opts: ConfirmOptions): Promise<boolean> {
    options.value = opts
    open.value = true
    return new Promise<boolean>((resolve) => { settle = resolve })
  }

  function answer(ok: boolean) {
    open.value = false
    settle?.(ok)
    settle = null
  }

  return { options, open, ask, answer }
})
