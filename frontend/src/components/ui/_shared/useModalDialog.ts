import { onBeforeUnmount, onMounted, watch, type Ref } from 'vue'

// Shared modal machinery for PrtDialog / PrtDrawer — both root a native <dialog> opened with showModal().
// The platform provides top layer, ::backdrop, focus containment, Esc + light dismiss (`closedby`); this
// composable only keeps v-model and element state in sync, plus the JS-assisted animated close the drawer
// needs (CSS `overlay` is not in Firefox, so a pure-CSS exit from the top layer is impossible there —
// this path behaves identically in both browsers).
//
// Animated close: a close request (Esc / light dismiss) fires `cancel`; we preventDefault once, set
// `data-closing` so the exit transition runs, await `transitionend` (duration-based timeout fallback),
// then close for real. Programmatic closes (v-model → false, the X button) take the same path.
// Re-opening mid-animation aborts it cleanly.
//
// Known v1 limit (deliberate): a dialog opened DECLARATIVELY (command/commandfor) does not flip the
// consumer's v-model to true — declarative usage simply doesn't need a v-model; Esc/scrim/X still close correctly.
export interface UseModalDialogOptions {
  modelValue: () => boolean
  emit: (open: boolean) => void
  // drawer: true — intercept close requests and play the exit transition
  animatedClose?: boolean
}

export function useModalDialog(
  el: Ref<HTMLDialogElement | null>,
  opts: UseModalDialogOptions,
) {
  let closing = false
  let stopAnimation: (() => void) | undefined

  function motionDuration(dialog: HTMLDialogElement): number {
    const raw = getComputedStyle(dialog).getPropertyValue('--motion-duration').trim()
    const value = Number.parseFloat(raw)
    if (!Number.isFinite(value)) return 150
    return raw.endsWith('ms') ? value : value * 1000
  }

  function abortClose() {
    closing = false
    stopAnimation?.()
    stopAnimation = undefined
    const dialog = el.value
    if (dialog) delete dialog.dataset.closing
  }

  function animateClose(dialog: HTMLDialogElement) {
    if (closing) return
    closing = true
    dialog.dataset.closing = ''
    const finish = () => {
      if (!closing) return
      abortClose()
      dialog.close()
    }
    const onEnd = (event: TransitionEvent) => {
      // children transition too — only the dialog's own end counts
      if (event.target === dialog) finish()
    }
    const timer = setTimeout(finish, motionDuration(dialog) + 50)
    dialog.addEventListener('transitionend', onEnd)
    stopAnimation = () => {
      clearTimeout(timer)
      dialog.removeEventListener('transitionend', onEnd)
    }
  }

  function syncOpen(open: boolean) {
    const dialog = el.value
    if (!dialog?.isConnected) return
    if (open) {
      abortClose() // re-open mid-exit-animation: keep it open, drop the close
      if (!dialog.open) dialog.showModal()
    } else if (dialog.open && !closing) {
      if (opts.animatedClose) animateClose(dialog)
      else dialog.close()
    }
  }

  function onCancel(event: Event) {
    if (!opts.animatedClose) return
    // Esc and closedby="any" light dismiss route through the close-request
    // path; intercept and play the exit transition instead (a second
    // request while it runs is also swallowed — the animation will close)
    event.preventDefault()
    const dialog = el.value
    if (dialog) animateClose(dialog)
  }

  function onClose() {
    // covers every close path: Esc, light dismiss, requestClose(),
    // form method="dialog", and our own animated finish
    abortClose()
    if (opts.modelValue()) opts.emit(false)
  }

  onMounted(() => {
    const dialog = el.value
    if (!dialog) return
    dialog.addEventListener('cancel', onCancel)
    dialog.addEventListener('close', onClose)
    if (opts.modelValue()) syncOpen(true)
  })

  watch(opts.modelValue, syncOpen)

  onBeforeUnmount(() => {
    abortClose()
    const dialog = el.value
    if (!dialog) return
    dialog.removeEventListener('cancel', onCancel)
    dialog.removeEventListener('close', onClose)
    if (dialog.open) dialog.close()
  })

  // programmatic close — takes the animated path when configured
  return { close: () => syncOpen(false) }
}
