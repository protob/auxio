<!--
  PrtToast
  Presentational toast — the floating sibling of PrtNotice. Usable standalone
  inline; queueing/timers live in useToast + PrtToaster.
  Slot contract:
    default — the toast line; overrides the `message` prop
-->
<template>
  <div
    :class="
      cx('prt-toast', toastBase, toastVariantClasses[variant][intent], props.class)
    "
  >
    <span
      :class="[
        toastIntent[intent].icon,
        variant === 'solid' ? 'text-current' : toastIntent[intent].iconColor,
        'w-4 h-4 mt-0.5 shrink-0',
      ]"
      aria-hidden="true"
    />
    <div class="flex-1 min-w-0">
      <div :class="['text-sm font-medium', toastText[variant].message]">
        <slot>{{ message }}</slot>
      </div>
      <div v-if="description" :class="['mt-1 text-sm', toastText[variant].description]">
        {{ description }}
      </div>
    </div>
    <button
      v-if="dismissible"
      type="button"
      :class="[
        'shrink-0 -m-1 inline-flex items-center justify-center p-1 bg-transparent rounded-control prt-motion-colors prt-ring',
        toastText[variant].dismiss,
      ]"
      @click="emit('dismiss')"
    >
      <span class="i-lucide-x w-4 h-4" aria-hidden="true" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { cx } from '../_shared/cx'
import type { PrtToastProps } from './types'
import { toastBase, toastIntent, toastText, toastVariantClasses } from './variants'

const props = withDefaults(defineProps<PrtToastProps>(), {
  intent: 'info',
  variant: 'quiet',
  message: '',
  dismissible: true,
  class: '',
})

const emit = defineEmits<{
  dismiss: []
}>()
</script>

<style scoped>
/* entry transition via @starting-style — a sanctioned <style> use */
.prt-toast {
  transition:
    opacity var(--motion-duration) var(--motion-ease),
    translate var(--motion-duration) var(--motion-ease);
  @starting-style {
    opacity: 0;
    translate: 0 8px;
  }
}
</style>
