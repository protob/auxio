<!--
  PrtDrawer
  Edge sheet on the native <dialog> substrate (same machinery as
  PrtDialog: _shared/useModalDialog.ts, no overlay component, scroll
  lock via html:has(dialog:modal)). The dialog element IS the panel:
  UA centering is overridden and the sheet pins to an edge per
  `placement`. Unlike the dialog, the close is ANIMATED — the slide is
  part of the gesture: useModalDialog intercepts close requests, sets
  data-closing so the translate transition returns the sheet off-screen,
  then closes for real (identical in Chrome and Firefox — CSS `overlay`
  is not in Firefox, so this is the one sanctioned JS-assisted exit).
  Wrapper-rooted: $attrs (id for commandfor, aria-*) land on <dialog>.
  Slot contract:
    title   — header text (header renders if title or showClose)
    default — body content (scrolls; header stays)
-->
<template>
  <dialog
    ref="el"
    class="prt-drawer"
    :class="cx(drawerVariants({ placement, size: customWidth ? null : size }), props.class)"
    :style="customWidth ? { width: customWidth } : undefined"
    :data-placement="placement"
    :closedby="dismissible ? 'any' : 'closerequest'"
    v-bind="$attrs"
  >
    <div class="flex h-full min-h-0 flex-col">
      <header v-if="$slots.title || showClose" :class="drawerHeaderClass">
        <h2 v-if="$slots.title" :class="drawerTitleClass"><slot name="title" /></h2>
        <span v-else aria-hidden="true" />
        <button
          v-if="showClose"
          type="button"
          :class="drawerCloseClass"
          @click="close()"
        >
          <span class="i-lucide-x" aria-hidden="true" />
          <span class="sr-only">{{ closeLabel }}</span>
        </button>
      </header>
      <div class="flex-1 min-h-0" :class="drawerBodyClass"><slot /></div>
    </div>
  </dialog>
</template>

<script setup lang="ts">
import { computed, shallowRef } from 'vue'
import { cx } from '../_shared/cx'
import { useModalDialog } from '../_shared/useModalDialog'
import type { PrtDrawerProps } from './types'
import {
  drawerBodyClass,
  drawerCloseClass,
  drawerHeaderClass,
  drawerTitleClass,
  drawerVariants,
} from './variants'

defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<PrtDrawerProps>(), {
  modelValue: false,
  placement: 'right',
  size: 'base',
  width: '',
  dismissible: true,
  showClose: true,
  closeLabel: 'Close',
  class: '',
})

// exact width is a side-sheet concept — a bottom sheet is always
// viewport-wide, its main axis is height (size steps)
const customWidth = computed(() =>
  props.placement !== 'bottom' && props.width ? props.width : '',
)

const emit = defineEmits<{
  'update:modelValue': [open: boolean]
}>()

const el = shallowRef<HTMLDialogElement | null>(null)

const { close } = useModalDialog(el, {
  modelValue: () => props.modelValue,
  emit: (open) => emit('update:modelValue', open),
  animatedClose: true,
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
/* <dialog> UA resets + edge pinning + the slide — the sanctioned <style>
   uses. LOOK classes stay in variants.ts. UA centering (inset 0 +
   margin auto + max sizes) is fully overridden: the dialog IS the sheet.
   Each placement declares its off-screen vector ONCE (--prt-drawer-out);
   entry (@starting-style), rest state and the data-closing exit all
   reuse it. */
.prt-drawer {
  padding: 0;
  margin: 0;
  border: 0;
  max-width: none;
  max-height: none;
  translate: var(--prt-drawer-out);
  transition:
    translate var(--motion-duration) var(--motion-ease),
    display var(--motion-duration) allow-discrete,
    overlay var(--motion-duration) allow-discrete;
}
/* the cross axis is EXPLICIT per placement: the UA sheet gives <dialog>
   width/height fit-content, so two-sided insets alone never stretch the
   sheet — without these the drawer renders content-sized (trap registry).
   The size variant owns the main axis (width on side sheets, height on
   the bottom sheet). */
.prt-drawer[data-placement='left'] {
  inset: 0 auto 0 0;
  height: 100dvh;
  border-right: 1px solid var(--edge);
  --prt-drawer-out: -100% 0;
}
.prt-drawer[data-placement='right'] {
  inset: 0 0 0 auto;
  height: 100dvh;
  border-left: 1px solid var(--edge);
  --prt-drawer-out: 100% 0;
}
.prt-drawer[data-placement='bottom'] {
  inset: auto 0 0 0;
  width: 100vw;
  border-top: 1px solid var(--edge);
  --prt-drawer-out: 0 100%;
}
.prt-drawer[open] {
  translate: 0 0;
  @starting-style {
    translate: var(--prt-drawer-out);
  }
}
/* the animated close: data-closing returns the sheet off-screen while
   the dialog is still open; useModalDialog closes on transitionend */
.prt-drawer[open][data-closing] {
  translate: var(--prt-drawer-out);
}
.prt-drawer::backdrop {
  background: var(--scrim);
  opacity: 0;
  transition:
    opacity var(--motion-duration) var(--motion-ease),
    display var(--motion-duration) allow-discrete,
    overlay var(--motion-duration) allow-discrete;
}
.prt-drawer[open]::backdrop {
  opacity: 1;
  @starting-style {
    opacity: 0;
  }
}
.prt-drawer[open][data-closing]::backdrop {
  opacity: 0;
}
</style>
