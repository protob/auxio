import { reactive } from 'vue'
import type { PrtToastInput, PrtToastItem } from './types'

// Toast machinery — rides the component dir (house pattern: a contract moves
// to _shared only when other components must see it; none do). Module-scope
// state: copy-source means this is a per-app singleton by construction.
// The consumer mounts ONE <PrtToaster /> at app root; useToast() anywhere.
// One method on purpose: push({ intent, message }) — no helper matrix.
const toasts = reactive<PrtToastItem[]>([])
const timers = new Map<number, ReturnType<typeof setTimeout>>()
let nextId = 1

function push(input: PrtToastInput): number {
  const item: PrtToastItem = {
    id: nextId++,
    intent: input.intent ?? 'info',
    variant: input.variant ?? 'quiet',
    message: input.message,
    description: input.description,
    duration: input.duration ?? 5000,
  }
  toasts.push(item)
  if (item.duration > 0) {
    timers.set(
      item.id,
      setTimeout(() => dismiss(item.id), item.duration),
    )
  }
  return item.id
}

function dismiss(id: number) {
  const timer = timers.get(id)
  if (timer) {
    clearTimeout(timer)
    timers.delete(id)
  }
  const i = toasts.findIndex((t) => t.id === id)
  if (i !== -1) toasts.splice(i, 1)
}

export function useToast() {
  return { toasts, push, dismiss }
}
