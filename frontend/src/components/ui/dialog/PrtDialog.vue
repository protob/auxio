<!--
  PrtDialog
  Modal dialog on the native <dialog> substrate (browser floor 2026-06:
  closedby Cr 134+/Fx 141+, requestClose 134/139, command/commandfor
  135/144). The platform provides top layer, ::backdrop scrim, focus
  containment, Esc + outside click — there is NO PrtOverlay component;
  the shared machinery is _shared/useModalDialog.ts (v-model ↔ element
  sync). Close is instant by policy (native OS dialogs close instantly);
  the drawer is the one with the animated exit.
  Scroll lock: the UNscoped style block targets html:has(dialog:modal)
  with scrollbar-gutter reserved — no layout shift either way, no JS,
  no body-padding hacks (scoped styles can't reach <html>; the rule is
  duplicated in PrtDrawer on purpose — copied components stay
  self-contained, and the duplicate is idempotent).
  Confirm is a demo preset, not a component.
  Wrapper-rooted: $attrs (id for commandfor, aria-*) land on the
  <dialog> element.
  Slot contract:
    title   — header text (header renders if title or showClose)
    default — body content
    actions — footer row, right-aligned
-->
<template>
  <dialog
    ref="el"
    class="prt-dialog"
    :class="cx(dialogVariants({ size: width ? null : size, edges }), props.class)"
    :style="width ? { maxWidth: width } : undefined"
    :closedby="dismissible ? 'any' : 'closerequest'"
    v-bind="$attrs"
  >
    <header v-if="$slots.title || showClose" :class="dialogHeaderClass">
      <h2 v-if="$slots.title" :class="dialogTitleClass"><slot name="title" /></h2>
      <span v-else aria-hidden="true" />
      <button
        v-if="showClose"
        type="button"
        :class="dialogCloseClass"
        @click="close()"
      >
        <span class="i-lucide-x" aria-hidden="true" />
        <span class="sr-only">{{ closeLabel }}</span>
      </button>
    </header>
    <div :class="dialogBodyClass"><slot /></div>
    <footer v-if="$slots.actions" :class="dialogActionsClass">
      <slot name="actions" />
    </footer>
  </dialog>
</template>

<script setup lang="ts">
import { shallowRef } from 'vue'
import { cx } from '../_shared/cx'
import { useModalDialog } from '../_shared/useModalDialog'
import type { PrtDialogProps } from './types'
import {
  dialogActionsClass,
  dialogBodyClass,
  dialogCloseClass,
  dialogHeaderClass,
  dialogTitleClass,
  dialogVariants,
} from './variants'

defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<PrtDialogProps>(), {
  modelValue: false,
  size: 'base',
  width: '',
  edges: 'rounded',
  dismissible: true,
  showClose: true,
  closeLabel: 'Close',
  class: '',
})

const emit = defineEmits<{
  'update:modelValue': [open: boolean]
}>()

const el = shallowRef<HTMLDialogElement | null>(null)

const { close } = useModalDialog(el, {
  modelValue: () => props.modelValue,
  emit: (open) => emit('update:modelValue', open),
  animatedClose: false,
})
</script>

<style>
/* UNscoped on purpose — scoped styles can't target html. Locks page
   scroll while ANY modal dialog is open; the reserved gutter means no
   layout shift on open or close. Idempotent across PrtDialog/PrtDrawer. */
html:has(dialog:modal) {
  overflow: hidden;
  scrollbar-gutter: stable;
}
</style>

<style scoped>
/* <dialog> UA resets + entry animation — the sanctioned <style> uses.
   LOOK classes stay in variants.ts. UA gives dialog `padding: 1em` and
   `border: solid` (medium): padding is zeroed (sections carry spacing);
   the border is REPLACED here, same specificity, with the house
   hairline. UA centering (inset 0 + margin auto) is kept. */
.prt-dialog {
  padding: 0;
  border: 1px solid var(--edge);
  max-height: calc(100dvh - 4rem);
  overflow-y: auto;
  opacity: 0;
  translate: 0 8px;
  transition:
    opacity var(--motion-duration) var(--motion-ease),
    translate var(--motion-duration) var(--motion-ease),
    display var(--motion-duration) allow-discrete,
    overlay var(--motion-duration) allow-discrete;
}
.prt-dialog[open] {
  opacity: 1;
  translate: 0 0;
  @starting-style {
    opacity: 0;
    translate: 0 8px;
  }
}
.prt-dialog::backdrop {
  background: var(--scrim);
  opacity: 0;
  transition:
    opacity var(--motion-duration) var(--motion-ease),
    display var(--motion-duration) allow-discrete,
    overlay var(--motion-duration) allow-discrete;
}
.prt-dialog[open]::backdrop {
  opacity: 1;
  @starting-style {
    opacity: 0;
  }
}
</style>
