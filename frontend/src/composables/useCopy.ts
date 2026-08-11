import { useClipboard } from '@vueuse/core'
import { useToast } from '@/components/ui/toast/useToast'

// navigator.clipboard is undefined on plain HTTP, which is how this dashboard is
// reached on a LAN. legacy: true falls back to execCommand there, and the toast
// covers the case where even that is refused - a silent no-op leaves the
// operator pasting whatever was in the clipboard before.
export function useCopy() {
  const { copy: write, isSupported } = useClipboard({ legacy: true })
  const { push } = useToast()

  // duration: 0 - a failure waits to be dismissed rather than timing out.
  const fail = (message: string) => push({ intent: 'danger', message, duration: 0 })

  async function copy(text: string, message: string) {
    if (!isSupported.value) {
      fail('Clipboard is not available in this browser')
      return
    }
    try {
      await write(text)
      push({ intent: 'success', message })
    } catch {
      fail('Could not copy to clipboard')
    }
  }

  return { copy }
}
