<!--
  PrtLoader
  Slot contract: none — a loader is a glyph. Visibility is the consumer's
  v-if; text sits next to it in consumer markup.
  Variants: spinner (default) | dots | squares (staggered pulse trios).
-->
<template>
  <span
    v-if="variant === 'spinner'"
    :data-seed="seed ?? undefined"
    :class="cx(loaderVariants({ size }), props.class)"
    aria-hidden="true"
  />
  <span
    v-else
    :data-seed="seed ?? undefined"
    :class="cx(loaderTrioVariants({ size }), props.class)"
    aria-hidden="true"
  >
    <span :class="loaderMarkVariants({ variant, size })" />
    <span :class="loaderMarkVariants({ variant, size })" />
    <span :class="loaderMarkVariants({ variant, size })" />
  </span>
</template>

<script setup lang="ts">
import { cx } from '../_shared/cx'
import type { PrtLoaderProps } from './types'
import { loaderMarkVariants, loaderTrioVariants, loaderVariants } from './variants'

const props = withDefaults(defineProps<PrtLoaderProps>(), {
  variant: 'spinner',
  size: 'base',
  class: '',
})
</script>

<style>
/* keyframes are a sanctioned <style> use — utilities can't express them.
   stagger rides nth-child so snippets never carry delay classes. */
.prt-loader-pulse {
  animation: prt-loader-pulse 0.9s var(--motion-ease) infinite;
}
.prt-loader-pulse:nth-child(2) {
  animation-delay: 150ms;
}
.prt-loader-pulse:nth-child(3) {
  animation-delay: 300ms;
}
@keyframes prt-loader-pulse {
  0%,
  80%,
  100% {
    opacity: 0.25;
    transform: scale(0.85);
  }
  40% {
    opacity: 1;
    transform: scale(1);
  }
}
</style>
