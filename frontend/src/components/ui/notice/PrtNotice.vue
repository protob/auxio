<!--
  PrtNotice
  Inline alert / callout. Intents are TRUE states (semantic tokens); seed is
  decoration for non-semantic callouts (tip boxes) — intent beats seed.
  Inline-only by design — notice is notice; a page hero/banner block is a
  different thing entirely (protosymbiont territory, never this component).
  `mono` is a tone TREATMENT, orthogonal to variant: the title (and the
  body when `true`) speaks the tone color — the programming-book callout.
  Slot contract:
    title   — heading line; overrides the `title` prop
    default — body; overrides the `description` prop
-->
<template>
  <div
    :data-seed="seed ?? undefined"
    :class="
      cx(
        noticeBase,
        noticeVariantBase[variant],
        noticeToneClasses[variant][tone],
        props.class,
      )
    "
  >
    <span
      v-if="resolvedIcon"
      :class="[resolvedIcon, noticeTone[tone].iconColor, 'w-4 h-4 mt-0.5 shrink-0']"
      aria-hidden="true"
    />
    <div class="flex-1 min-w-0">
      <div v-if="$slots.title || title" :class="['text-sm font-medium', titleColor]">
        <slot name="title">{{ title }}</slot>
      </div>
      <div
        v-if="$slots.default || description"
        :class="['text-sm', bodyColor, $slots.title || title ? 'mt-1' : '']"
      >
        <slot>{{ description }}</slot>
      </div>
    </div>
    <button
      v-if="dismissible"
      type="button"
      class="shrink-0 -m-1 inline-flex items-center justify-center p-1 bg-transparent rounded-control text-ink-faint hover:text-ink hover:bg-wash prt-motion-colors prt-ring"
      @click="emit('dismiss')"
    >
      <span class="i-lucide-x w-4 h-4" aria-hidden="true" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { cx } from '../_shared/cx'
import type { PrtNoticeProps } from './types'
import {
  noticeBase,
  noticeTone,
  noticeToneClasses,
  noticeVariantBase,
} from './variants'

const props = withDefaults(defineProps<PrtNoticeProps>(), {
  variant: 'edge',
  dismissible: false,
  mono: false,
  class: '',
})

const emit = defineEmits<{
  dismiss: []
}>()

// intent (true state) beats seed (decoration); neither = info
const tone = computed(() => props.intent ?? (props.seed ? 'seed' : 'info'))

// explicit icon (incl. '' to hide) beats the tone icon
const resolvedIcon = computed(() =>
  props.icon !== undefined ? props.icon : noticeTone[tone.value].icon,
)

// mono treatment: any mono colors the title; only `true` dims the body
const titleColor = computed(() =>
  props.mono ? noticeTone[tone.value].titleColor : 'text-ink',
)
const bodyColor = computed(() =>
  props.mono === true ? noticeTone[tone.value].bodyColor : 'text-ink-muted',
)
</script>
