<!--
  PrtFormField
  Slot contract:
    default — replaces the built-in PrtInput; receives { inputId, error, disabled, inputProps }
              custom inputs in the slot integrate via v-bind="inputProps" —
              that object (id, name, data-field, aria-*, onBlur/onInput) is
              THE field→input contract; inputs never inject anything.
    label   — replaces the label text
    help    — replaces the helper text
  `name` connects the field to a surrounding PrtForm (error lookup by
  dot-path, blur/input reporting, deterministic `field-<name>` id).
  Explicit `error`/`errorMessage` props ALWAYS beat the injected error.
-->
<template>
  <div :class="cx(fieldVariants({ disabled }), props.class)">
    <label v-if="label || $slots.label" :for="inputId" :class="labelVariants({ error: resolvedError })">
      <slot name="label">
        {{ label }}<span v-if="required" class="text-danger"> *</span>
      </slot>
    </label>

    <p v-if="description" class="text-xs text-ink-faint">{{ description }}</p>

    <slot
      :input-id="inputId"
      :error="resolvedError"
      :disabled="disabled"
      :input-props="inputProps"
    >
      <PrtInput
        v-bind="inputProps"
        v-model="inputValue"
        :type="type"
        :placeholder="placeholder"
        :disabled="disabled"
        :error="resolvedError"
      />
    </slot>

    <p v-if="resolvedError && resolvedMessage" :id="`${inputId}-error`" :class="helpVariants({ tone: 'error' })">
      <span class="i-lucide-circle-alert shrink-0" aria-hidden="true" />
      {{ resolvedMessage }}
    </p>
    <p v-else-if="helperText || $slots.help" :class="helpVariants({ tone: 'help' })">
      <slot name="help">{{ helperText }}</slot>
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, useId } from 'vue'
import { cx } from '../_shared/cx'
import { prtFormKey } from '../_shared/form'
import { PrtInput } from '../input'
import type { PrtFormFieldProps } from './types'
import { fieldVariants, helpVariants, labelVariants } from './variants'

const props = withDefaults(defineProps<PrtFormFieldProps>(), {
  modelValue: '',
  type: 'text',
  required: false,
  error: false,
  disabled: false,
  class: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
}>()

const form = inject(prtFormKey, undefined) // undefined = Shape A, by design

const injectedError = computed(() =>
  props.name && form
    ? form.displayError(props.name, props.errorPattern)
    : undefined,
)
// Explicit props ALWAYS beat injection:
const resolvedMessage = computed(() => props.errorMessage ?? injectedError.value)
const resolvedError = computed(() => props.error || !!resolvedMessage.value)

const generatedId = useId()
// Deterministic when named (agents/E2E cache selectors; useId() churns):
const inputId = computed(
  () => props.id ?? (props.name ? `field-${props.name}` : generatedId),
)

const inputProps = computed(() => ({
  id: inputId.value,
  name: props.name,
  'data-field': props.name,
  'aria-invalid': resolvedError.value || undefined,
  'aria-describedby':
    resolvedError.value && resolvedMessage.value
      ? `${inputId.value}-error`
      : undefined,
  // Field → form protocol. On a Prt input these are undeclared attrs, so Vue
  // falls them through to the root <input> — inputs never inject form/field
  // context; this fallthrough is the whole channel.
  onBlur: props.name && form ? () => form.reportBlur(props.name!) : undefined,
  onInput: props.name && form ? () => form.reportInput(props.name!) : undefined,
}))

const inputValue = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})
</script>
