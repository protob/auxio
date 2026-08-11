<!--
  PrtProgress
  Determinate only: bar and circle. Indeterminate/top-bar
  duty is PrtLoadingBar's — a progress that can't state progress is a
  loading bar, not a progress.
  Slot contract: none.
-->
<template>
  <!-- circle: determinate ring, % centered by default -->
  <div
    v-if="type === 'circle'"
    :data-seed="seed ?? undefined"
    :class="cx('relative inline-flex', props.class)"
  >
    <svg :class="progressCircleVariants({ size })" viewBox="0 0 48 48" fill="none">
      <circle cx="24" cy="24" r="20" class="stroke-surface-2" stroke-width="4" />
      <circle
        cx="24"
        cy="24"
        r="20"
        class="stroke-[var(--seed,var(--accent))] prt-motion"
        stroke-width="4"
        :stroke-dasharray="CIRCUMFERENCE"
        :stroke-dashoffset="CIRCUMFERENCE * (1 - clamped / 100)"
        transform="rotate(-90 24 24)"
      />
    </svg>
    <span v-if="labelShown" :class="progressCircleLabelVariants({ size })"
      >{{ Math.round(clamped) }}%</span
    >
  </div>

  <!-- bar -->
  <div
    v-else
    :data-seed="seed ?? undefined"
    :class="cx('flex items-center gap-3', props.class)"
  >
    <div :class="progressTrackVariants({ size })">
      <div
        :class="[progressFillClass, 'prt-motion']"
        :style="{ width: clamped + '%' }"
      />
    </div>
    <span
      v-if="labelShown"
      class="shrink-0 text-xs tabular-nums text-ink-muted"
      >{{ Math.round(clamped) }}%</span
    >
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { cx } from '../_shared/cx'
import type { PrtProgressProps } from './types'
import {
  progressCircleLabelVariants,
  progressCircleVariants,
  progressFillClass,
  progressTrackVariants,
} from './variants'

const props = withDefaults(defineProps<PrtProgressProps>(), {
  value: 0,
  type: 'bar',
  size: 'base',
  showValue: undefined,
  class: '',
})

// r=20 in the 48-unit viewBox
const CIRCUMFERENCE = 2 * Math.PI * 20

const clamped = computed(() => Math.min(100, Math.max(0, props.value)))
// the centered % IS the circle's point — shown unless explicitly disabled;
// bars stay quiet unless asked
const labelShown = computed(() => props.showValue ?? props.type === 'circle')
</script>
