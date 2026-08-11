<!--
  PrtChip
  Tag/label mark. Not a button: an interactive chip is a PrtBtn in chip
  clothing — this one only ever emits `dismiss` from its × button.
  Slot contract:
    default — chip content
-->
<template>
  <span
    :data-seed="seed ?? undefined"
    :class="cx(chipVariants({ variant, size, edges, disabled }), props.class)"
  >
    <span v-if="icon" :class="[icon, 'w-3.5 h-3.5 shrink-0']" aria-hidden="true" />
    <slot />
    <button
      v-if="dismissible"
      type="button"
      class="-mr-0.5 inline-flex items-center justify-center bg-transparent rounded-full not-disabled:hover:bg-[var(--seed-wash,var(--wash))] prt-motion-colors prt-ring"
      :disabled="disabled"
      @click="emit('dismiss')"
    >
      <span class="i-lucide-x w-3 h-3" aria-hidden="true" />
    </button>
  </span>
</template>

<script setup lang="ts">
import { cx } from '../_shared/cx'
import type { PrtChipProps } from './types'
import { chipVariants } from './variants'

const props = withDefaults(defineProps<PrtChipProps>(), {
  variant: 'solid',
  size: 'base',
  edges: 'rounded',
  dismissible: false,
  disabled: false,
  class: '',
})

const emit = defineEmits<{
  dismiss: []
}>()
</script>
