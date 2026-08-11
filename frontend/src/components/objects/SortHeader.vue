<script setup lang="ts">
import { computed } from 'vue'
import type { SortKey } from '@/lib/sort'
import IconChevronDown from '~icons/lucide/chevron-down'
import IconChevronUp from '~icons/lucide/chevron-up'

const props = defineProps<{
  label: string
  column: SortKey
  active: SortKey
  dir: 'asc' | 'desc'
}>()

const emit = defineEmits<{ sort: [column: SortKey] }>()

const isActive = computed(() => props.active === props.column)
// aria-sort belongs on the cell, and only the sorted one may carry a direction.
const ariaSort = computed(() => (isActive.value ? (props.dir === 'asc' ? 'ascending' : 'descending') : 'none'))
</script>

<template>
  <div role="columnheader" :aria-sort="ariaSort">
    <!-- No type or size of its own: the header row's uppercase, tracking and
         faint ink all inherit, and the reset already hands a button its
         parent's font and colour. -->
    <button
      type="button"
      class="inline-flex items-center gap-1 p-0 bg-transparent cursor-pointer hover:text-ink"
      @click="emit('sort', column)"
    >
      {{ label }}
      <IconChevronUp v-if="isActive && dir === 'asc'" class="w-3 h-3" />
      <IconChevronDown v-else-if="isActive" class="w-3 h-3" />
    </button>
  </div>
</template>
