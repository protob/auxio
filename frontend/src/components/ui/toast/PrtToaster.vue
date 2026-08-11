<!--
  PrtToaster
  The queue host. Mount ONE at app root; fire from anywhere via useToast().
  popover=manual = top layer without z-index management.
  Shown while the queue has items, hidden when it empties.
  Slot contract: none.
-->
<template>
  <div ref="host" popover="manual" :class="toasterVariants({ position })">
    <PrtToast
      v-for="t in toasts"
      :key="t.id"
      :intent="t.intent"
      :variant="t.variant"
      :message="t.message"
      :description="t.description"
      @dismiss="dismiss(t.id)"
    />
  </div>
</template>

<script setup lang="ts">
import { shallowRef, watchEffect } from 'vue'
import type { PrtToasterProps } from './types'
import PrtToast from './PrtToast.vue'
import { useToast } from './useToast'
import { toasterVariants } from './variants'

withDefaults(defineProps<PrtToasterProps>(), {
  position: 'bottom-right',
})

const host = shallowRef<HTMLElement | null>(null)
const { toasts, dismiss } = useToast()

watchEffect(() => {
  const el = host.value
  if (!el?.isConnected) return
  const open = el.matches(':popover-open')
  if (toasts.length > 0 && !open) el.showPopover()
  else if (toasts.length === 0 && open) el.hidePopover()
})
</script>
