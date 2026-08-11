<!--
  PrtForm
  Slot contract: default — plain content; no slot props (the context lives
  in script via useForm()). Requires :form (D4 — composable-first).
  The context is destructured once: a form context is a stable reactive
  object for the component's lifetime — swapping a whole context mid-flight
  is not a supported pattern.
-->
<template>
  <form :class="cx(formVariants(), props.class)" @submit.prevent="form.submit()">
    <slot />
  </form>
</template>

<script setup lang="ts">
import { provide } from 'vue'
import { cx } from '../_shared/cx'
import { prtFormKey } from '../_shared/form'
import type { PrtFormProps } from './types'
import { formVariants } from './variants'

const props = defineProps<PrtFormProps>()
const { form } = props

provide(prtFormKey, form)
</script>
