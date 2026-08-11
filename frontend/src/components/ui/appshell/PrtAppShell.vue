<!--
  PrtAppShell
  The deliberately THIN frame: a CSS Grid that positions header / sidebar / main /
  aside / footer and nothing else. It owns no state, imports nothing, and knows
  nothing about widths — the slotted PrtSidebar carries its own transitioning
  width and a grid `auto` track sizes to it, so collapse "just works" with zero
  shared JS state (the over-coupling the proposal rejected lived entirely in the
  state/breakpoint/width props this frame does NOT have).

  Optional regions cost nothing: an unfilled slot renders no element, its `auto`
  track collapses to 0, no layout jump, no :has() needed. The shell owns the
  viewport (height: 100svh) and makes <main> the only scroll container; drop a
  PrtSticky into #header for a pinning bar. Compose the frame from semantic HTML
  instead if you'd rather — this is a convenience, not a requirement.

  Slot contract:
    header  — top bar, spans full width (optional)
    sidebar — the nav column (PrtSidebar, or your own); sizes the left track
    default — main content (the scroll container)
    aside   — right rail / inspector, spans body height (optional)
    footer  — bottom bar, spans full width (optional)
-->
<template>
  <div :class="cx(appShellClass, props.class)">
    <header v-if="$slots.header" :class="appShellHeaderClass"><slot name="header" /></header>
    <div v-if="$slots.sidebar" :class="appShellSidebarClass"><slot name="sidebar" /></div>
    <main :class="appShellMainClass"><slot /></main>
    <aside v-if="$slots.aside" :class="appShellAsideClass"><slot name="aside" /></aside>
    <footer v-if="$slots.footer" :class="appShellFooterClass"><slot name="footer" /></footer>
  </div>
</template>

<script setup lang="ts">
import { cx } from '../_shared/cx'
import type { PrtAppShellProps } from './types'
import {
  appShellAsideClass,
  appShellClass,
  appShellFooterClass,
  appShellHeaderClass,
  appShellMainClass,
  appShellSidebarClass,
} from './variants'

const props = withDefaults(defineProps<PrtAppShellProps>(), { class: '' })
</script>

<style scoped>
/* grid-template-areas + auto-track offset math — the sanctioned <style> use
   (none of this is expressible as utility literals). The sidebar column is
   `auto`: it sizes to the slotted sidebar's own width, so the frame stays
   width-agnostic and the collapse animation rides the sidebar's `width`
   transition. Empty areas collapse their auto track to 0 with no jump. */
.prt-appshell {
  display: grid;
  grid-template-columns: auto 1fr auto;
  grid-template-rows: auto 1fr auto;
  grid-template-areas:
    'header  header header'
    'sidebar main   aside'
    'footer  footer footer';
  /* the shell owns the viewport by default; override the var to embed it in a
     sub-route or a bounded box (the demo sets it to a fixed height). */
  height: var(--prt-appshell-h, 100svh);
  background: var(--surface-0);
}
.prt-appshell-header {
  grid-area: header;
  min-width: 0;
}
.prt-appshell-sidebar {
  grid-area: sidebar;
  display: flex;
  min-height: 0;
}
/* <main> is the ONLY scroller (min-height/min-width:0 = the grid min-size fix,
   without it the track refuses to shrink and overflow never kicks in). */
.prt-appshell-main {
  grid-area: main;
  min-width: 0;
  min-height: 0;
  overflow: auto;
}
.prt-appshell-aside {
  grid-area: aside;
  min-width: 0;
  min-height: 0;
  overflow: auto;
}
.prt-appshell-footer {
  grid-area: footer;
  min-width: 0;
}
</style>
