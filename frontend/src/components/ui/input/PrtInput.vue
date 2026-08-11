<!--
  PrtInput
  Slot contract: none — bare input element. Compose labels/help via PrtFormField.
-->
<template>
  <input
    :id="id"
    :type="type"
    :value="modelValue"
    :placeholder="placeholder"
    :disabled="disabled"
    :aria-invalid="error || undefined"
    :class="cx(inputVariants({ variant, size, edges, error, disabled }), props.class)"
    @input="onInput"
  />
</template>

<script setup lang="ts">
import { cx } from '../_shared/cx'
import type { PrtInputProps } from './types'
import { inputVariants } from './variants'

const props = withDefaults(defineProps<PrtInputProps>(), {
  modelValue: '',
  type: 'text',
  variant: 'filled',
  size: 'base',
  edges: 'rounded',
  error: false,
  disabled: false,
  class: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

function onInput(event: Event) {
  emit('update:modelValue', (event.target as HTMLInputElement).value)
}
</script>
