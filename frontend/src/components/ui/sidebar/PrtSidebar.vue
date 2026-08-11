<!--
  PrtSidebar
  Desktop-first application nav: a full panel that collapses to an icon RAIL
  (the headline interaction) and, below `mobileBreakpoint`, becomes a PrtDrawer
  overlay (the de-emphasized fallback — RWD supported, never the architecture).
  Standalone-first: drop it next to your own <main>, or into PrtAppShell's
  #sidebar slot — same component, no shell-aware variant.

  Coupling is pure CSS: the panel owns its width as a transitioning custom
  property (two fixed lengths — `interpolate-size` is not in Firefox, same call
  as PrtCollapsible), so a grid `auto` track sizes to it with zero JS shared
  state. State rides `v-model` (modelValue = the collapsed boolean, per the kit's
  modelValue convention) + a `data-state` attribute; nothing is
  injected (the kit's no-inject spirit — only PrtForm provides()).

  Reuse over rebuild: mobile overlay = PrtDrawer (focus-trap, scrim, scroll-lock,
  animated close — never re-solved); rail leaf labels = PrtTooltip; rail group
  flyouts = the PrtMenu popover+anchor substrate (flat full-bleed panel doctrine).
  Expandable groups REPLICATE the grid 0fr→1fr region (PrtFacet precedent) rather
  than nest PrtCollapsible, whose hairline-bar trigger would clash with a nav row.

  Slot contract:
    header  — pinned top (logo/brand). Slot props: { collapsed, mobile, toggle }
    footer  — pinned bottom (profile). Slot props: { collapsed, mobile }
    default — (escape) replace the whole nav body
  emits: update:modelValue (collapsed) · select (value) · navigate (item)
  expose: { collapsed, mobile, toggle, openMobile, closeMobile, toggleCollapsed }
-->
<template>
  <component
    :is="mobile ? PrtDrawer : 'aside'"
    v-bind="wrapperProps"
  >
    <div class="flex h-full min-h-0 flex-col">
      <!-- header: brand slot + the rail toggle -->
      <div v-if="(!railed && $slots.header) || showToggle" :class="headerClass">
        <slot
          v-if="!railed"
          name="header"
          :collapsed="railed"
          :mobile="mobile"
          :toggle="toggle"
        />
        <button
          v-if="showToggle"
          type="button"
          :class="sidebarToggleClass"
          :aria-expanded="!railed"
          :aria-label="railed ? expandLabel : collapseLabel"
          @click="toggleCollapsed"
        >
          <span
            :class="railed ? 'i-lucide-panel-left-open' : 'i-lucide-panel-left-close'"
            aria-hidden="true"
          />
        </button>
      </div>

      <!-- the scrolling nav -->
      <nav :class="sidebarNavClass" :aria-label="label">
        <slot>
          <template v-for="(group, gi) in groups" :key="gi">
            <div
              v-if="group.label && !railed"
              :class="sidebarGroupLabelVariants({ size })"
            >
              {{ group.label }}
            </div>
            <ul role="list" class="flex flex-col gap-0.5" :class="gi > 0 ? 'mt-1' : ''">
              <li v-for="(item, ii) in group.items" :key="item.value">
                <!-- ── leaf row ─────────────────────────────────────────── -->
                <template v-if="!item.children?.length">
                  <PrtTooltip
                    v-if="railed"
                    :content="item.label"
                    :placement="flyoutPlacement"
                    class="block"
                  >
                    <component
                      :is="item.href ? 'a' : 'button'"
                      v-bind="rowAttrs(item)"
                      :class="sidebarItemVariants({ size, active: isActive(item), disabled: item.disabled, rail: true })"
                      @click="onItemClick(item)"
                    >
                      <span
                        v-if="isActive(item)"
                        :class="[sidebarItemMarkClass, markEdge]"
                        aria-hidden="true"
                      />
                      <span
                        v-if="item.icon"
                        :class="[item.icon, 'shrink-0 h-4 w-4']"
                        aria-hidden="true"
                      />
                    </component>
                  </PrtTooltip>
                  <component
                    :is="item.href ? 'a' : 'button'"
                    v-else
                    v-bind="rowAttrs(item)"
                    :class="sidebarItemVariants({ size, active: isActive(item), disabled: item.disabled, rail: false })"
                    @click="onItemClick(item)"
                  >
                    <span
                      v-if="isActive(item)"
                      :class="[sidebarItemMarkClass, markEdge]"
                      aria-hidden="true"
                    />
                    <span
                      v-if="item.icon"
                      :class="[item.icon, 'shrink-0 h-4 w-4']"
                      aria-hidden="true"
                    />
                    <span class="min-w-0 truncate">{{ item.label }}</span>
                    <span v-if="item.badge" :class="sidebarBadgeClass">{{ item.badge }}</span>
                  </component>
                </template>

                <!-- ── group, expanded: own trigger + grid 0fr→1fr region ── -->
                <template v-else-if="!railed">
                  <button
                    type="button"
                    :class="sidebarItemVariants({ size, active: groupHasActive(item), disabled: item.disabled, rail: false })"
                    :aria-expanded="groupOpen(item)"
                    :aria-controls="regionId(gi, ii)"
                    @click="toggleGroup(item)"
                  >
                    <span
                      v-if="item.icon"
                      :class="[item.icon, 'shrink-0 h-4 w-4']"
                      aria-hidden="true"
                    />
                    <span class="min-w-0 truncate">{{ item.label }}</span>
                    <span
                      :class="[sidebarChevronClass, groupOpen(item) ? 'rotate-180' : '']"
                      aria-hidden="true"
                    />
                  </button>
                  <div
                    :id="regionId(gi, ii)"
                    class="prt-sidebar-group-region"
                    :data-open="groupOpen(item) || undefined"
                  >
                    <div class="prt-sidebar-group-inner">
                      <ul role="list" class="flex flex-col gap-0.5 pt-0.5">
                        <li v-for="child in item.children" :key="child.value">
                          <component
                            :is="child.href ? 'a' : 'button'"
                            v-bind="rowAttrs(child)"
                            :class="sidebarSubItemVariants({ size, active: isActive(child), disabled: child.disabled })"
                            @click="onItemClick(child)"
                          >
                            <span class="min-w-0 truncate">{{ child.label }}</span>
                            <span v-if="child.badge" :class="sidebarBadgeClass">{{ child.badge }}</span>
                          </component>
                        </li>
                      </ul>
                    </div>
                  </div>
                </template>

                <!-- ── group, rail: icon button → anchored flyout panel ──── -->
                <template v-else>
                  <button
                    type="button"
                    :popovertarget="flyoutId(gi, ii)"
                    :class="sidebarItemVariants({ size, active: groupHasActive(item), disabled: item.disabled, rail: true })"
                    :style="{ 'anchor-name': flyoutAnchor(gi, ii) }"
                    :aria-label="item.label"
                  >
                    <span
                      v-if="groupHasActive(item)"
                      :class="[sidebarItemMarkClass, markEdge]"
                      aria-hidden="true"
                    />
                    <span
                      v-if="item.icon"
                      :class="[item.icon, 'shrink-0 h-4 w-4']"
                      aria-hidden="true"
                    />
                  </button>
                  <div
                    :id="flyoutId(gi, ii)"
                    popover="auto"
                    class="prt-sidebar-flyout"
                    :class="sidebarFlyoutPanelClass"
                    :data-side="side"
                    :style="{ 'position-anchor': flyoutAnchor(gi, ii) }"
                  >
                    <div :class="sidebarFlyoutLabelClass">{{ item.label }}</div>
                    <component
                      :is="child.href ? 'a' : 'button'"
                      v-for="child in item.children"
                      :key="child.value"
                      v-bind="rowAttrs(child)"
                      :class="sidebarFlyoutItemVariants({ active: isActive(child), disabled: child.disabled })"
                      @click="onFlyoutClick(child, gi, ii)"
                    >
                      <span class="min-w-0 truncate">{{ child.label }}</span>
                      <span v-if="child.badge" :class="sidebarBadgeClass">{{ child.badge }}</span>
                    </component>
                  </div>
                </template>
              </li>
            </ul>
          </template>
        </slot>
      </nav>

      <!-- footer -->
      <div v-if="!railed && $slots.footer" :class="sidebarFooterClass">
        <slot name="footer" :collapsed="railed" :mobile="mobile" />
      </div>
    </div>
  </component>
</template>

<script setup lang="ts">
import { useMediaQuery } from '@vueuse/core'
import { computed, ref, useId } from 'vue'
import { cx } from '../_shared/cx'
import PrtDrawer from '../drawer/PrtDrawer.vue'
import PrtTooltip from '../tooltip/PrtTooltip.vue'
import type { PrtSidebarGroup, PrtSidebarItem, PrtSidebarProps } from './types'
import {
  sidebarBadgeClass,
  sidebarChevronClass,
  sidebarFlyoutItemVariants,
  sidebarFlyoutLabelClass,
  sidebarFlyoutPanelClass,
  sidebarFooterClass,
  sidebarGroupLabelVariants,
  sidebarHeaderClass,
  sidebarItemMarkClass,
  sidebarItemVariants,
  sidebarNavClass,
  sidebarRootVariants,
  sidebarSubItemVariants,
  sidebarToggleClass,
} from './variants'

const props = withDefaults(defineProps<PrtSidebarProps>(), {
  modelValue: false,
  collapsible: 'icon',
  side: 'left',
  width: '16rem',
  railWidth: '3.5rem',
  mobileBreakpoint: '(width < 48rem)',
  label: 'Main',
  collapseLabel: 'Collapse sidebar',
  expandLabel: 'Expand sidebar',
  size: 'base',
  class: '',
})

const emit = defineEmits<{
  'update:modelValue': [collapsed: boolean]
  select: [value: string | number]
  navigate: [item: PrtSidebarItem]
}>()

// per-instance ids/anchors for the rail flyouts (id + style context, never a
// class — same sanctioned dynamic channel as PrtMenu/PrtTooltip)
const uid = useId()
const regionId = (g: number, i: number) => `prt-sb-${uid}-rgn-${g}-${i}`
const flyoutId = (g: number, i: number) => `prt-sb-${uid}-fly-${g}-${i}`
const flyoutAnchor = (g: number, i: number) => `--prt-sb-${uid}-a-${g}-${i}`

// below the breakpoint the panel hands off to PrtDrawer. Desktop-first: a
// VIEWPORT query, not a container query — the sidebar is a viewport-edge citizen.
const mobile = useMediaQuery(() => props.mobileBreakpoint)
const mobileOpen = ref(false)

// rail only on desktop, only when allowed, only when collapsed
const railed = computed(
  () => props.collapsible === 'icon' && props.modelValue && !mobile.value,
)
const showToggle = computed(() => props.collapsible === 'icon' && !mobile.value)

const flyoutPlacement = computed<'left' | 'right'>(() =>
  props.side === 'left' ? 'right' : 'left',
)
const markEdge = computed(() => (props.side === 'left' ? 'left-0' : 'right-0'))

// flat or grouped input → always sections internally (the PrtMenu idiom)
const groups = computed<PrtSidebarGroup[]>(() => {
  const raw = props.items
  const first = raw[0]
  if (!first) return []
  return 'items' in first
    ? (raw as PrtSidebarGroup[])
    : [{ items: raw as PrtSidebarItem[] }]
})

function isActive(item: PrtSidebarItem) {
  return props.active !== undefined && item.value === props.active
}
function groupHasActive(item: PrtSidebarItem) {
  return isActive(item) || !!item.children?.some((c) => isActive(c))
}

// group disclosure: explicit override, else open when it holds the active route
const openState = ref<Record<string, boolean>>({})
function groupOpen(item: PrtSidebarItem) {
  return openState.value[item.value] ?? groupHasActive(item)
}
function toggleGroup(item: PrtSidebarItem) {
  if (item.disabled) return
  openState.value = { ...openState.value, [item.value]: !groupOpen(item) }
}

function rowAttrs(item: PrtSidebarItem) {
  return item.href
    ? {
        href: item.disabled ? undefined : item.href,
        'aria-current': isActive(item) ? ('page' as const) : undefined,
        'aria-disabled': item.disabled || undefined,
        'data-active': isActive(item) || undefined,
      }
    : {
        type: 'button' as const,
        disabled: item.disabled,
        'aria-current': isActive(item) ? ('page' as const) : undefined,
        'data-active': isActive(item) || undefined,
      }
}

function onItemClick(item: PrtSidebarItem) {
  if (item.disabled) return
  emit('navigate', item)
  if (!item.href) emit('select', item.value)
  if (mobile.value) mobileOpen.value = false
}
function onFlyoutClick(item: PrtSidebarItem, g: number, i: number) {
  if (item.disabled) return
  onItemClick(item)
  const el = document.getElementById(flyoutId(g, i))
  if (el?.matches(':popover-open')) el.hidePopover()
}

// ── the two roots ────────────────────────────────────────────────────────
const headerClass = computed(() =>
  cx(sidebarHeaderClass, railed.value ? 'justify-center px-2' : 'justify-between'),
)

const wrapperProps = computed(() =>
  mobile.value
    ? {
        modelValue: mobileOpen.value,
        'onUpdate:modelValue': (v: boolean) => (mobileOpen.value = v),
        placement: props.side,
        width: props.width,
        showClose: false,
        'data-seed': props.seed ?? undefined,
        class: cx('prt-sidebar', props.class),
      }
    : {
        class: cx(sidebarRootVariants({ side: props.side }), props.class),
        'data-state': railed.value ? 'collapsed' : 'expanded',
        'data-seed': props.seed ?? undefined,
        style: {
          '--prt-sb-w-expanded': props.width,
          '--prt-sb-w-rail': props.railWidth,
        },
      },
)

// ── desktop rail vs mobile open: two orthogonal states (see header comment) ─
function toggleCollapsed() {
  emit('update:modelValue', !props.modelValue)
}
function openMobile() {
  mobileOpen.value = true
}
function closeMobile() {
  mobileOpen.value = false
}
function toggle() {
  if (mobile.value) mobileOpen.value = !mobileOpen.value
  else toggleCollapsed()
}

defineExpose({ collapsed: railed, mobile, toggle, openMobile, closeMobile, toggleCollapsed })
</script>

<style scoped>
/* width is the source of truth — a grid `auto` track (PrtAppShell) sizes to it,
   so the frame needs ZERO knowledge of widths. Two fixed lengths, transitioned;
   never animated to `auto` (interpolate-size is Chrome-only — PrtCollapsible
   call). The vars come from props via the style binding above. */
.prt-sidebar[data-state] {
  width: var(--prt-sidebar-w);
  transition: width var(--motion-duration) var(--motion-ease);
}
.prt-sidebar[data-state='expanded'] {
  --prt-sidebar-w: var(--prt-sb-w-expanded, 16rem);
}
.prt-sidebar[data-state='collapsed'] {
  --prt-sidebar-w: var(--prt-sb-w-rail, 3.5rem);
}

/* group disclosure — the grid 0fr→1fr trick REPLICATED (PrtFacet precedent),
   not PrtCollapsible (its hairline-bar trigger clashes with a nav row). The
   overflow:hidden + min-height:0 MUST sit on the CHILD (trap registry). */
.prt-sidebar-group-region {
  display: grid;
  grid-template-rows: 0fr;
  transition: grid-template-rows var(--motion-duration) var(--motion-ease);
}
.prt-sidebar-group-region[data-open] {
  grid-template-rows: 1fr;
}
.prt-sidebar-group-inner {
  overflow: hidden;
  min-height: 0;
}

/* rail flyout — popover UA resets + anchor placement + @starting-style (the
   sanctioned <style> uses; same block shape as PrtMenu). LOOK stays in variants.
   UA [popover] ships padding:0.25em + border:solid — zero both or the full-bleed
   rows float in a phantom inset (trap registry). */
.prt-sidebar-flyout {
  position: fixed;
  inset: auto;
  padding: 0;
  border: 0;
  margin: 0;
  opacity: 0;
  transition:
    opacity var(--motion-duration) var(--motion-ease),
    display var(--motion-duration) allow-discrete,
    overlay var(--motion-duration) allow-discrete;
}
.prt-sidebar-flyout:popover-open {
  opacity: 1;
  @starting-style {
    opacity: 0;
  }
}
.prt-sidebar-flyout[data-side='left'] {
  position-area: right span-bottom;
  margin-left: 6px;
  position-try-fallbacks: flip-inline;
}
.prt-sidebar-flyout[data-side='right'] {
  position-area: left span-bottom;
  margin-right: 6px;
  position-try-fallbacks: flip-inline;
}
</style>
