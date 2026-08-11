<!--
  PrtCombobox
  The autocomplete/combobox merge (one component, `freeText` is the axis):
  strict mode emits only on selection (typed values — numbers stay numbers,
  null = empty); freeText mode makes the TEXT the value, emitted on input,
  selection just fills. Search, clear and async loading land here — the
  "PrtCombobox territory" select pointed at.

  Substrate: select's anchored panel verbatim (CSS anchor positioning, flat +
  full-bleed doctrine, aria-activedescendant listbox, focus pinned) with ONE
  deliberate deviation: the panel is popover="manual", NOT auto —
  `popovertarget` is button-only, an <input> can't be a declarative invoker,
  so with auto every click INTO the input (cursor repositioning) would be an
  "outside" pointerdown and light-dismiss the panel. Manual means we own what
  auto gave select for free: Esc lives in the input's keydown, outside-close
  is one native document pointerdown listener while open. Do not "fix" this
  back to auto (08 trap registry).

  Wrapper-rooted: $attrs land on the INPUT (the focusable form element) —
  inputProps/blur reporting work unchanged. Inputs never inject() anything.
  Async recipe: listen to update:query, fetch, pass :options + :filter="false"
  + :loading. Home/End keep their text-editing meaning (deliberate deviation
  from select — the field is a text input).
  Slot contract: none — options render from `options`.
-->
<template>
  <div
    ref="root"
    :class="cx('relative w-full', props.class)"
    :style="{ 'anchor-name': anchorName }"
  >
    <input
      :id="id"
      ref="inputEl"
      type="text"
      autocomplete="off"
      role="combobox"
      aria-autocomplete="list"
      :aria-expanded="isOpen"
      :aria-controls="listboxId"
      :aria-activedescendant="
        isOpen && activeIndex >= 0 ? optionId(activeIndex) : undefined
      "
      :aria-invalid="error || undefined"
      :value="query"
      :placeholder="placeholder"
      :disabled="disabled"
      :class="
        comboboxVariants({
          variant,
          size,
          edges,
          error,
          disabled,
          clearable: showClear,
        })
      "
      v-bind="$attrs"
      @input="onInput"
      @click="onInputClick"
      @keydown="onKeydown"
    />
    <button
      v-if="showClear"
      type="button"
      tabindex="-1"
      :aria-label="clearLabel"
      :class="comboboxClearVariants({ size })"
      @mousedown.prevent
      @click="clear"
    />
    <span
      v-if="loading"
      :class="comboboxSpinnerVariants({ size })"
      aria-hidden="true"
    />
    <span
      v-else
      :class="[comboboxChevronVariants({ size }), isOpen ? 'rotate-180' : '']"
      aria-hidden="true"
    />
    <div
      :id="listboxId"
      ref="panel"
      popover="manual"
      role="listbox"
      class="prt-cbx-panel"
      :class="comboboxPanelVariants({ edges })"
      :style="{ 'position-anchor': anchorName }"
      @toggle="onToggle"
    >
      <button
        v-for="(opt, i) in visible"
        :id="optionId(i)"
        :key="String(opt.value)"
        type="button"
        role="option"
        :aria-selected="isSelected(opt)"
        :disabled="opt.disabled"
        :class="
          comboboxOptionVariants({
            size,
            active: i === activeIndex && !opt.disabled,
            disabled: !!opt.disabled,
          })
        "
        @mousedown.prevent
        @click="choose(opt)"
        @mousemove="activeIndex = i"
      >
        <span v-if="opt.icon" :class="[opt.icon, 'shrink-0']" aria-hidden="true" />
        <span class="truncate min-w-0">{{ opt.label }}</span>
        <span
          v-if="isSelected(opt)"
          class="i-lucide-check ml-auto shrink-0 text-accent"
          aria-hidden="true"
        />
      </button>
      <div v-if="loading" :class="comboboxNoteVariants({ size })" aria-hidden="true">
        <span class="i-lucide-loader-circle animate-spin shrink-0" />
        <span>{{ loadingText }}</span>
      </div>
      <div
        v-else-if="visible.length === 0"
        :class="comboboxNoteVariants({ size })"
        aria-hidden="true"
      >
        {{ emptyText }}
      </div>
      <div
        v-if="moreCount > 0"
        :class="comboboxNoteVariants({ size })"
        aria-hidden="true"
      >
        {{ moreText.replace('{n}', String(moreCount)) }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, shallowRef, useId, watch } from 'vue'
import { onClickOutside } from '@vueuse/core'
import { cx } from '../_shared/cx'
import type { PrtOption } from '../_shared/types'
import type { PrtComboboxProps } from './types'
import {
  comboboxChevronVariants,
  comboboxClearVariants,
  comboboxNoteVariants,
  comboboxOptionVariants,
  comboboxPanelVariants,
  comboboxSpinnerVariants,
  comboboxVariants,
} from './variants'

defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<PrtComboboxProps>(), {
  modelValue: null,
  freeText: false,
  clearable: false,
  loading: false,
  maxVisible: 100,
  variant: 'filled',
  size: 'base',
  edges: 'rounded',
  error: false,
  disabled: false,
  class: '',
  emptyText: 'No results',
  moreText: '+{n} more — keep typing',
  loadingText: 'Loading…',
  clearLabel: 'Clear',
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number | null]
  // the raw input text — the async hook (fetch on it, pass options + :filter="false" + :loading)
  'update:query': [query: string]
}>()

// per-instance anchor/ids — truly dynamic values (style binding / id
// context, not class construction)
const uid = useId()
const anchorName = `--prt-cbx-${uid}`
const listboxId = `prt-cbx-${uid}-listbox`
const optionId = (i: number) => `prt-cbx-${uid}-opt-${i}`

const root = shallowRef<HTMLElement | null>(null)
const inputEl = shallowRef<HTMLInputElement | null>(null)
const panel = shallowRef<HTMLElement | null>(null)
const isOpen = ref(false)
const activeIndex = ref(-1)

const selected = computed(() =>
  props.modelValue == null
    ? undefined
    : props.options.find((o) => o.value === props.modelValue),
)

// the input's text — strict mode displays the selected label, freeText
// mode IS the model
const query = ref(
  props.freeText ? String(props.modelValue ?? '') : (selected.value?.label ?? ''),
)
// filter only engages once the user typed since open — opening a filled
// strict field shows ALL options, not just the one matching its own label
const queryDirty = ref(false)

function isSelected(opt: PrtOption): boolean {
  return props.freeText
    ? opt.label === props.modelValue
    : opt.value === props.modelValue
}

const defaultFilter = (q: string, o: PrtOption) =>
  o.label.toLowerCase().includes(q.toLowerCase())

const filtered = computed(() => {
  if (props.filter === false) return props.options
  if (!queryDirty.value || query.value === '') return props.options
  const fn = props.filter ?? defaultFilter
  return props.options.filter((o) => fn(query.value, o))
})

const visible = computed(() => filtered.value.slice(0, props.maxVisible))
const moreCount = computed(() => filtered.value.length - visible.value.length)

const showClear = computed(
  () =>
    props.clearable &&
    !props.disabled &&
    (props.freeText ? query.value !== '' : props.modelValue != null),
)

// external model changes re-sync the text — never while the user is
// typing in an open strict field
watch([() => props.modelValue, selected], () => {
  if (props.freeText) {
    const text = String(props.modelValue ?? '')
    if (text !== query.value) query.value = text
  } else if (!isOpen.value) {
    query.value = selected.value?.label ?? ''
  }
})

// typing reshapes the list — keep the active row valid (first enabled)
watch(visible, () => {
  if (!isOpen.value) return
  const cur = visible.value[activeIndex.value]
  if (cur && !cur.disabled) return
  activeIndex.value = visible.value.findIndex((o) => !o.disabled)
})

function show() {
  if (props.disabled) return
  const el = panel.value
  if (el?.isConnected && !el.matches(':popover-open')) el.showPopover()
}

function hide() {
  const el = panel.value
  if (el?.isConnected && el.matches(':popover-open')) el.hidePopover()
}

// manual popovers fire `toggle` too — popover state stays the single
// isOpen source of truth, exactly like select
function onToggle(event: Event) {
  isOpen.value = (event as { newState?: string }).newState === 'open'
  if (isOpen.value) {
    queryDirty.value = false
    const sel = visible.value.findIndex((o) => isSelected(o) && !o.disabled)
    activeIndex.value = sel >= 0 ? sel : visible.value.findIndex((o) => !o.disabled)
    scrollActive()
  } else {
    activeIndex.value = -1
    queryDirty.value = false
    // close-without-selection must not leave half-typed garbage: strict
    // mode re-displays the committed label (toggle fires async — the
    // v-model echo of a selection has already landed by now)
    if (!props.freeText) query.value = selected.value?.label ?? ''
  }
}

// outside-close (the manual-popover deviation): VueUse owns the document
// listener + its teardown. The panel is a DOM child of `root` (top layer
// notwithstanding), so onClickOutside's composedPath check treats clicks on
// the input, affordances and panel as inside — same semantics as the old
// hand-rolled composedPath guard. hide() is a no-op when already closed.
onClickOutside(root, () => {
  if (isOpen.value) hide()
})

function onInput(event: Event) {
  query.value = (event.target as HTMLInputElement).value
  queryDirty.value = true
  emit('update:query', query.value)
  if (props.freeText) emit('update:modelValue', query.value)
  if (!isOpen.value) show()
}

function onInputClick() {
  if (!isOpen.value) show()
}

function move(delta: 1 | -1) {
  let i = activeIndex.value
  do i += delta
  while (i >= 0 && i < visible.value.length && visible.value[i]?.disabled)
  if (i >= 0 && i < visible.value.length) {
    activeIndex.value = i
    scrollActive()
  }
}

function scrollActive() {
  nextTick(() => {
    if (activeIndex.value < 0) return
    document
      .getElementById(optionId(activeIndex.value))
      ?.scrollIntoView({ block: 'nearest' })
  })
}

function choose(opt: PrtOption) {
  if (opt.disabled) return
  if (props.freeText) {
    // selection just fills — the text is the value
    query.value = opt.label
    emit('update:query', opt.label)
    emit('update:modelValue', opt.label)
  } else {
    query.value = opt.label
    emit('update:modelValue', opt.value)
  }
  hide()
}

function onKeydown(event: KeyboardEvent) {
  switch (event.key) {
    case 'ArrowDown':
      event.preventDefault()
      if (isOpen.value) move(1)
      else show()
      break
    case 'ArrowUp':
      event.preventDefault()
      if (isOpen.value) move(-1)
      else show()
      break
    case 'Enter': {
      if (!isOpen.value) return
      event.preventDefault()
      const opt = visible.value[activeIndex.value]
      if (opt) choose(opt)
      else hide()
      break
    }
    case 'Escape':
      if (!isOpen.value) return
      event.preventDefault()
      hide()
      break
    case 'Tab':
      if (isOpen.value) hide()
      break
  }
}

function clear() {
  query.value = ''
  queryDirty.value = false
  emit('update:query', '')
  emit('update:modelValue', props.freeText ? '' : null)
  inputEl.value?.focus()
}

onBeforeUnmount(() => {
  hide()
})
</script>

<style scoped>
/* popover UA resets + anchor placement + @starting-style — the select
   panel's exact scoped-block shape (the sanctioned <style> uses). LOOK
   classes stay in variants.ts. The panel tracks the input's width via
   anchor-size(). */
.prt-cbx-panel {
  position: fixed;
  inset: auto;
  /* UA [popover] defaults are padding:0.25em + border:solid — kill both
     or the full-bleed rows float in a phantom inset */
  padding: 0;
  border: 0;
  margin: 0;
  margin-top: 4px;
  width: anchor-size(width);
  max-height: 16rem;
  position-area: bottom;
  position-try-fallbacks: flip-block;
  opacity: 0;
  transition:
    opacity var(--motion-duration) var(--motion-ease),
    display var(--motion-duration) allow-discrete,
    overlay var(--motion-duration) allow-discrete;
}
.prt-cbx-panel:popover-open {
  opacity: 1;
  @starting-style {
    opacity: 0;
  }
}
</style>
