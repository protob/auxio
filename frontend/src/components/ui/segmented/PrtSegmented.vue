<!--
  PrtSegmented
  A segmented control: a group of joined cells where you pick ONE (`type=single`)
  or SEVERAL (`type=multiple`). NOT a radio (single-only), NOT a toggle (binary),
  NOT tabs (navigation) — the single/multiple duality is its identity.

  modelValue IS the contract: a value for `single`, an array for `multiple`
  (always emitted as a fresh array). The active fill rides the seed cascade.

  Slot contract: none — segments render from `options`.
  emits: update:modelValue, change
-->
<template>
  <div
    :data-seed="seed ?? undefined"
    :role="type === 'single' ? 'radiogroup' : 'group'"
    :aria-label="label"
    :class="cx(segmentedRootVariants({ size, disabled }), props.class)"
  >
    <button
      v-for="(opt, i) in options"
      :key="String(opt.value)"
      :ref="(el) => setItem(el, i)"
      type="button"
      :class="segmentVariants({ size, active: isActive(opt.value) })"
      :role="type === 'single' ? 'radio' : undefined"
      :aria-checked="type === 'single' ? isActive(opt.value) : undefined"
      :aria-pressed="type === 'multiple' ? isActive(opt.value) : undefined"
      :tabindex="i === focusedIndex ? 0 : -1"
      :disabled="disabled || opt.disabled"
      @click="select(opt)"
      @focus="focusedIndex = i"
      @keydown="onKeydown"
    >
      <span v-if="opt.icon" :class="[opt.icon, 'h-4 w-4']" aria-hidden="true" />
      <span v-if="opt.label" :class="iconOnly && opt.icon ? 'sr-only' : undefined">{{ opt.label }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { cx } from '../_shared/cx'
import {
  firstFocusable as initialFocus,
  isActive as isOptionActive,
  moveFocus as nextFocusable,
  toggle,
} from './navigation'
import type { PrtOption, PrtSegmentedProps } from './types'
import { segmentedRootVariants, segmentVariants } from './variants'

const props = withDefaults(defineProps<PrtSegmentedProps>(), {
  type: 'single',
  size: 'base',
  disabled: false,
  class: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number | (string | number)[]]
  change: [value: string | number | (string | number)[]]
}>()

function isActive(v: string | number) {
  return isOptionActive(props.modelValue, props.type, v)
}

function select(opt: PrtOption) {
  if (props.disabled || opt.disabled) return
  if (props.type === 'multiple') {
    const next = toggle(props.modelValue, opt.value)
    emit('update:modelValue', next)
    emit('change', next)
  } else {
    emit('update:modelValue', opt.value)
    emit('change', opt.value)
  }
}

// ── roving focus (arrow keys move between segments, skipping disabled) ────────
const items: HTMLElement[] = []
function setItem(el: Element | { $el?: Element } | null, i: number) {
  if (el instanceof HTMLElement) items[i] = el
}

const focusedIndex = ref(initialFocus(props.options, props.modelValue, props.type))

function moveFocus(dir: 1 | -1) {
  const i = nextFocusable(props.options, focusedIndex.value, dir, props.disabled)
  if (i >= 0) {
    focusedIndex.value = i
    items[i]?.focus()
  }
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'ArrowRight' || e.key === 'ArrowDown') {
    e.preventDefault()
    moveFocus(1)
  } else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
    e.preventDefault()
    moveFocus(-1)
  }
  // Space/Enter activate natively (these are real <button>s)
}
</script>
