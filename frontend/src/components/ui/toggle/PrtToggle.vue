<!--
  PrtToggle
  Slot contract:
    default — inline label content (overrides the `label` prop)
  Wrapper-rooted control: $attrs land on the NATIVE input (id, name,
  data-field, aria-*, onBlur/onInput from PrtFormField's inputProps).
  Inputs never inject() anything.
-->
<template>
  <label
    :data-seed="seed ?? undefined"
    :class="cx(toggleRootVariants({ disabled }), props.class)"
  >
    <input
      :id="id"
      type="checkbox"
      role="switch"
      class="sr-only peer"
      :checked="modelValue"
      :disabled="disabled"
      v-bind="$attrs"
      @change="onChange"
    />
    <span :class="toggleTrackVariants({ size, checked: modelValue })" aria-hidden="true">
      <span :class="toggleKnobVariants({ size, checked: modelValue })" />
    </span>
    <span v-if="label || $slots.default" :class="toggleLabelVariants({ size })">
      <slot>{{ label }}</slot>
    </span>
  </label>
</template>

<script setup lang="ts">
import { cx } from '../_shared/cx'
import type { PrtToggleProps } from './types'
import {
  toggleKnobVariants,
  toggleLabelVariants,
  toggleRootVariants,
  toggleTrackVariants,
} from './variants'

defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<PrtToggleProps>(), {
  modelValue: false,
  size: 'base',
  disabled: false,
  class: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

function onChange(event: Event) {
  emit('update:modelValue', (event.target as HTMLInputElement).checked)
}
</script>
