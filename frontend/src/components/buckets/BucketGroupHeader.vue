<script setup lang="ts">
import IconChevronRight from '~icons/lucide/chevron-right'


defineProps<{ label: string; count: number; open: boolean }>()
const emit = defineEmits<{ toggle: [] }>()
</script>

<template>
  <!-- Wrapped in row/cell so the group heading is a legal child of the table
       it divides; the button itself carries the expand state. -->
  <div role="row">
    <div role="cell">
      <!-- group headers separate by whitespace + type, not by a filled band -->
      <button
        type="button"
        class="group w-full flex items-center gap-2.25 pt-4.5 px-5 pb-2 border-b border-edge bg-transparent cursor-pointer select-none"
        :aria-expanded="open"
        @click="emit('toggle')"
      >
        <IconChevronRight class="w-3.5 h-3.5 text-ink-muted prt-motion" :class="open ? 'rotate-90' : ''" />
        <span class="text-[0.78125rem] font-semibold uppercase tracking-[0.03125rem] text-ink-muted group-hover:text-ink">{{ label }}</span>
        <span class="font-mono text-[0.6875rem] text-ink-faint">{{ count }}</span>
      </button>
    </div>
  </div>
</template>
