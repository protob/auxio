<!--
  PrtTooltip
  Positioning substrate (Tier B menus/dialogs inherit this pattern):
  pure CSS Anchor Positioning (browser floor 2026-06: Cr 125+ / Fx 147+)
  on a popover=hint in the top layer — zero JS layout, no positioning
  library, ever. Show/hide is explicit JS hover/focus (interestfor is
  below the floor); placement/flip is position-area + position-try-fallbacks
  in the scoped style block. The per-instance anchor name is a truly
  dynamic value, so it rides style bindings.
  Slot contract:
    default — the trigger (wrapped in an inline-flex span, the anchor)
    content — tooltip body; overrides the `content` prop
-->
<template>
  <span
    class="inline-flex"
    :class="props.class"
    :style="{ 'anchor-name': anchorName }"
    @mouseenter="open"
    @mouseleave="close"
    @focusin="open"
    @focusout="close"
  >
    <slot />
    <div
      ref="tip"
      popover="hint"
      role="tooltip"
      class="prt-tip"
      :class="tooltipClass"
      :data-seed="seed ?? undefined"
      :data-placement="placement"
      :style="{ 'position-anchor': anchorName }"
    >
      <slot name="content">{{ content }}</slot>
    </div>
  </span>
</template>

<script setup lang="ts">
import { onBeforeUnmount, shallowRef, useId } from 'vue'
import { useTimeoutFn } from '@vueuse/core'
import type { PrtTooltipProps } from './types'
import { tooltipClass } from './variants'

const props = withDefaults(defineProps<PrtTooltipProps>(), {
  content: '',
  placement: 'top',
  // instant by design — a delayed tooltip is terrible UX (house rule);
  // opt INTO a delay per instance, never out of one
  delay: 0,
  disabled: false,
  class: '',
})

const tip = shallowRef<HTMLElement | null>(null)
const anchorName = `--prt-tt-${useId()}`

// instant by default (delay 0 → setTimeout(0)); the delay is
// read reactively at start, start() resets a pending show, VueUse stops it on
// unmount.
const showTimer = useTimeoutFn(
  () => {
    const el = tip.value
    if (el?.isConnected && !el.matches(':popover-open')) el.showPopover()
  },
  () => props.delay,
  { immediate: false },
)

function open() {
  if (props.disabled) return
  showTimer.start()
}

function close() {
  showTimer.stop()
  const el = tip.value
  if (el?.isConnected && el.matches(':popover-open')) el.hidePopover()
}

onBeforeUnmount(close)
</script>

<style scoped>
/* popover UA resets + anchor placement + @starting-style — the sanctioned
   <style> uses: utilities can't express these reliably (shorthand/longhand
   margin-inset ordering, at-rules). LOOK classes stay in variants.ts. */
.prt-tip {
  position: fixed;
  inset: auto;
  margin: 0;
  opacity: 0;
  transition:
    opacity var(--motion-duration) var(--motion-ease),
    display var(--motion-duration) allow-discrete,
    overlay var(--motion-duration) allow-discrete;
}
.prt-tip:popover-open {
  opacity: 1;
  @starting-style {
    opacity: 0;
  }
}
.prt-tip[data-placement='top'] {
  position-area: top;
  margin-bottom: 6px;
  position-try-fallbacks: flip-block;
}
.prt-tip[data-placement='bottom'] {
  position-area: bottom;
  margin-top: 6px;
  position-try-fallbacks: flip-block;
}
.prt-tip[data-placement='left'] {
  position-area: left;
  margin-right: 6px;
  position-try-fallbacks: flip-inline;
}
.prt-tip[data-placement='right'] {
  position-area: right;
  margin-left: 6px;
  position-try-fallbacks: flip-inline;
}
</style>
