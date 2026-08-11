<!--
  PrtCheckbox
  Slot contract:
    default — inline label content (overrides the `label` prop)
  Wrapper-rooted control: $attrs land on the NATIVE input (id, name,
  data-field, aria-*, onBlur/onInput from PrtFormField's inputProps).
  Inputs never inject() anything.
-->
<template>
  <label
    :data-seed="seed ?? undefined"
    :class="cx(checkboxRootVariants({ disabled }), props.class)"
  >
    <input
      :id="id"
      type="checkbox"
      class="sr-only peer"
      :checked="modelValue"
      :disabled="disabled"
      v-bind="$attrs"
      @change="onChange"
    />
    <span
      :class="checkboxBoxVariants({ edges, size, checked: modelValue, error })"
      aria-hidden="true"
    >
      <span v-if="modelValue" :class="checkboxGlyphVariants({ size })" />
    </span>
    <span v-if="label || $slots.default" :class="checkboxLabelVariants({ size })">
      <slot>{{ label }}</slot>
    </span>
  </label>
</template>

<script setup lang="ts">
import { cx } from '../_shared/cx'
import type { PrtCheckboxProps } from './types'
import {
  checkboxBoxVariants,
  checkboxGlyphVariants,
  checkboxLabelVariants,
  checkboxRootVariants,
} from './variants'

defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<PrtCheckboxProps>(), {
  modelValue: false,
  size: 'base',
  edges: 'rounded',
  error: false,
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
