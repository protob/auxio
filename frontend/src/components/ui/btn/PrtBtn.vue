<!--
  PrtBtn
  Slot contract:
    default — button content
    prepend — leading content (icon)
    append  — trailing content (icon)
-->
<template>
  <component
    :is="tag"
    :type="tag === 'button' ? 'button' : undefined"
    :disabled="tag === 'button' ? disabled || loading : undefined"
    :data-seed="seed ?? undefined"
    @click="guard"
    :class="
      cx(
        buttonVariants({ variant, size, edges, fullWidth, disabled, loading }),
        props.class,
      )
    "
  >
    <span v-if="loading" class="i-lucide-loader-circle animate-spin" aria-hidden="true" />
    <slot v-else name="prepend" />
    <slot />
    <slot name="append" />
  </component>
</template>

<script setup lang="ts">
import { cx } from '../_shared/cx'
import type { PrtBtnProps } from './types'
import { buttonVariants } from './variants'

const props = withDefaults(defineProps<PrtBtnProps>(), {
  variant: 'solid',
  size: 'base',
  edges: 'rounded',
  disabled: false,
  loading: false,
  fullWidth: false,
  tag: 'button',
  class: '',
})

// pointer events stay ON when disabled (the not-allowed cursor must show on
// hover). Native <button> blocks clicks via the disabled attr; this guard
// covers polymorphic roots (tag="a": stops the navigation).
function guard(e: MouseEvent) {
  if (props.disabled || props.loading) {
    e.preventDefault()
    e.stopPropagation()
  }
}
</script>
