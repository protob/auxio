<!--
  PrtEmptyState
  Neutral chrome for "nothing here yet / no results / first run":
  icon + one line + one action. No intent, no dismiss (PrtNotice's job), no seed.
  Slot contract:
    icon    — custom visual (replaces the icon class span)
    default — description override
    action  — the ONE action (a PrtBtn, typically); no action props,
              composition wins
-->
<template>
  <div :class="cx(emptyStateVariants({ size }), props.class)">
    <slot name="icon">
      <span
        v-if="icon"
        :class="[
          icon,
          'text-ink-faint',
          size === 'lg' ? 'w-10 h-10' : size === 'sm' ? 'w-6 h-6' : 'w-8 h-8',
        ]"
        aria-hidden="true"
      />
    </slot>
    <p :class="['font-medium text-ink', size === 'lg' ? 'text-base' : 'text-sm']">
      {{ title }}
    </p>
    <slot>
      <p
        v-if="description"
        :class="[
          'text-ink-muted max-w-sm',
          size === 'sm' ? 'text-xs' : 'text-sm',
        ]"
      >
        {{ description }}
      </p>
    </slot>
    <div v-if="$slots.action" class="mt-3">
      <slot name="action" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { cx } from '../_shared/cx'
import type { PrtEmptyStateProps } from './types'
import { emptyStateVariants } from './variants'

const props = withDefaults(defineProps<PrtEmptyStateProps>(), {
  size: 'base',
  class: '',
})
</script>
