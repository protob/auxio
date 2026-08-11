<!--
  PrtTabs
  Underline tabs (Linear school): transparent row, full-width hairline,
  ghost triggers, 2px accent indicator. Controlled only — v-model is the
  single source of truth, no internal fallback state.
  The indicator is ONE anchor-positioned element: only the ACTIVE trigger
  carries `anchor-name` (bound undefined on the rest — two elements with
  one name = last-in-DOM wins and the indicator sticks; trap registry),
  and the indicator tracks it with `left/width: anchor()/anchor-size()`.
  Where the browser interpolates anchor moves the underline SLIDES; where
  it doesn't it jumps — both correct, zero JS measurement either way.
  Keyboard: Tab + Enter/Space (arrow roving deferred — accessibility deprioritized).
  Slot contract:
    panel-{value} — optional panel per item, rendered under the row;
                    omit all panels to use tabs as pure nav and switch
                    content yourself. Slot props: { item }.
-->
<template>
  <div :class="cx('w-full', props.class)">
    <div :class="tabsListClass" role="tablist">
      <button
        v-for="item in items"
        :id="tabId(item.value)"
        :key="item.value"
        type="button"
        role="tab"
        :aria-selected="item.value === modelValue"
        :aria-controls="hasPanel(item) ? panelDomId(item.value) : undefined"
        :disabled="item.disabled"
        :class="
          tabsTriggerVariants({
            size,
            active: item.value === modelValue,
            disabled: !!item.disabled,
          })
        "
        :style="
          item.value === modelValue ? { 'anchor-name': anchorName } : undefined
        "
        @click="select(item)"
      >
        <span v-if="item.icon" :class="[item.icon, 'shrink-0']" aria-hidden="true" />
        <span class="truncate">{{ item.label }}</span>
      </button>
      <span
        v-if="activeItem"
        class="prt-tabs-indicator"
        :class="tabsIndicatorClass"
        :style="{ 'position-anchor': anchorName }"
        aria-hidden="true"
      />
    </div>
    <template v-for="item in items" :key="item.value">
      <div
        v-if="hasPanel(item) && item.value === modelValue"
        :id="panelDomId(item.value)"
        role="tabpanel"
        :aria-labelledby="tabId(item.value)"
        :class="tabsPanelClass"
      >
        <slot :name="`panel-${item.value}`" :item="item" />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, useId, useSlots } from 'vue'
import { cx } from '../_shared/cx'
import type { PrtTabItem, PrtTabsProps } from './types'
import { tabsIndicatorClass, tabsListClass, tabsPanelClass, tabsTriggerVariants } from './variants'

const props = withDefaults(defineProps<PrtTabsProps>(), {
  size: 'base',
  class: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
}>()

// per-instance anchor/ids — truly dynamic values (style binding / id
// context, not class construction)
const uid = useId()
const anchorName = `--prt-tabs-${uid}`
const tabId = (value: string | number) => `prt-tabs-${uid}-tab-${value}`
const panelDomId = (value: string | number) => `prt-tabs-${uid}-panel-${value}`

const slots = useSlots()
const activeItem = computed(() =>
  props.items.find((i) => i.value === props.modelValue),
)

function hasPanel(item: PrtTabItem) {
  return !!slots[`panel-${item.value}`]
}

function select(item: PrtTabItem) {
  if (item.disabled) return
  emit('update:modelValue', item.value)
}
</script>

<style scoped>
/* anchor geometry — the sanctioned <style> use (anchor()/anchor-size()
   are not expressible as utility literals). The indicator straddles the
   row hairline: 2px tall, its last pixel covering the border line. */
.prt-tabs-indicator {
  position: absolute;
  left: anchor(left);
  width: anchor-size(width);
  top: calc(anchor(bottom) - 1px);
  height: 2px;
  transition:
    left var(--motion-duration) var(--motion-ease),
    width var(--motion-duration) var(--motion-ease);
}
</style>
