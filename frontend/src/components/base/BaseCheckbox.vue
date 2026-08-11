<script setup lang="ts">
import { computed } from 'vue'
import IconCheck from '~icons/lucide/check'
import IconMinus from '~icons/lucide/minus'

const props = withDefaults(
  defineProps<{ checked: boolean; indeterminate?: boolean; label?: string }>(),
  { indeterminate: false },
)
const emit = defineEmits<{ change: [] }>()

// 'mixed' is the tri-state value the checkbox role defines; a plain false would
// tell a screen reader "nothing selected" when some rows are.
const ariaChecked = computed(() => (props.indeterminate ? 'mixed' : props.checked))
</script>

<template>
  <!-- Not PrtCheckbox: the kit's is a two-state native input, and select-all
       needs the mixed state. A button, not a span: the rows are div grids, so
       without this the whole selection UI is unreachable from the keyboard.
       Both fill states sit in one binding - two literals that can never be on
       the element together, rather than two background classes racing. -->
  <button
    type="button"
    role="checkbox"
    class="w-4.25 h-4.25 flex items-center justify-center shrink-0 p-0 cursor-pointer border-1.5 rounded-[0.3125rem] prt-motion-colors"
    :class="checked || indeterminate ? 'bg-accent border-accent' : 'bg-surface-1 border-edge-strong'"
    :aria-checked="ariaChecked"
    :aria-label="label"
    @click.stop="emit('change')"
    @keydown.space.prevent="emit('change')"
  >
    <IconMinus v-if="indeterminate" class="w-3 h-3 text-accent-ink" />
    <IconCheck v-else-if="checked" class="w-3 h-3 text-accent-ink" />
  </button>
</template>
