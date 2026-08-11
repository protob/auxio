<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import PrtInput from '@/components/ui/input/PrtInput.vue'
import IconSearch from '~icons/lucide/search'

const props = withDefaults(
  defineProps<{ label: string; placeholder?: string; size?: 'sm' | 'base' }>(),
  { placeholder: '', size: 'base' },
)

const query = defineModel<string>({ required: true })

// Complete literal strings per size: the icon offset and the padding that
// clears it have to move together, and neither may be assembled. The pad wins
// over the size variant's px- because UnoCSS emits pl- after px-, at equal
// specificity - no !important, and no tailwind-merge in the kit's cx().
const geometry = {
  sm: { icon: 'absolute left-2.5 text-ink-faint pointer-events-none', pad: 'pl-8' },
  base: { icon: 'absolute left-3 text-ink-faint pointer-events-none', pad: 'pl-9' },
} as const
const geo = computed(() => geometry[props.size])

// ObjectsView binds '/' to focus this, and the real input is two components
// down, so the wrapper has to hand it back.
const input = useTemplateRef<{ $el: HTMLInputElement }>('input')
defineExpose({ focus: () => input.value?.$el.focus() })
</script>

<template>
  <div class="relative flex items-center">
    <span :class="geo.icon" aria-hidden="true">
      <IconSearch class="w-3.5 h-3.5" />
    </span>
    <PrtInput
      ref="input"
      v-model="query"
      :size="size"
      :placeholder="placeholder"
      :aria-label="label"
      :class="geo.pad"
    />
  </div>
</template>
