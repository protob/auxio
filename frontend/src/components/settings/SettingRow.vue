<script setup lang="ts">
import PrtChip from '@/components/ui/chip/PrtChip.vue'

defineProps<{
  label: string
  // the AUXIO_* variable this row documents
  env?: string
  description?: string
  // the row applies only when another row is on, so it dims rather than
  // disappears - the knob stays readable, just unreachable
  dimmed?: boolean
}>()
</script>

<template>
  <div class="flex items-start justify-between gap-8 py-5" :class="dimmed ? 'opacity-40 pointer-events-none' : ''">
    <div class="flex-1 min-w-0">
      <div class="flex flex-wrap items-center gap-2.25 mb-1.25">
        <span class="text-sm font-medium text-ink">{{ label }}</span>
        <PrtChip v-if="env" variant="ghost" seed="7" size="sm" class="font-mono">{{ env }}</PrtChip>
      </div>
      <div class="text-xs text-ink-muted leading-normal">
        <slot name="description">{{ description }}</slot>
      </div>
    </div>
    <!-- The rows that only document something (per-request transforms) have no
         control, and an empty flex child would still eat the gap. -->
    <div v-if="$slots.default" class="flex items-center gap-2 shrink-0">
      <slot />
    </div>
  </div>
</template>
