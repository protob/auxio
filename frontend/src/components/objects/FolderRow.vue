<script setup lang="ts">
import { computed } from 'vue'
import { folderName } from '@/lib/objectPath'
import IconFolder from '~icons/lucide/folder'

const props = defineProps<{ folder: string; prefix: string }>()
const emit = defineEmits<{ open: [] }>()

const name = computed(() => folderName(props.folder, props.prefix))
</script>

<template>
  <!-- The checkbox cell stays empty: a folder is a virtual prefix, not an
       object, so there is nothing to select or delete. -->
  <div
    class="grid grid-cols-[var(--cols)] items-center w-full px-4.5 py-2.5 border-b border-edge prt-motion-colors hover:bg-surface-2"
    role="row"
  >
    <div role="cell"></div>
    <div class="flex items-center min-w-0" role="cell">
      <!-- Only the name opens the folder: a click handler on the row div is
           unreachable without a pointer. The ink lives on the button so the
           hover colour reaches the label by inheritance. -->
      <button
        type="button"
        class="flex items-center gap-2.5 min-w-0 w-full p-0 bg-transparent cursor-pointer overflow-hidden text-ellipsis whitespace-nowrap font-medium text-ink hover:text-accent"
        @click="emit('open')"
      >
        <!-- warning, not a literal colour: a hard-coded hex cannot follow the mode,
             and amber reads unreadable on the light surface. -->
        <span class="w-6.5 h-6.5 rounded-control flex items-center justify-center shrink-0 bg-[var(--warning-wash)] text-warning">
          <IconFolder class="w-3.5 h-3.5" />
        </span>
        {{ name }}/
      </button>
    </div>
    <div class="font-mono text-[0.84375rem] text-ink-muted" role="cell">—</div>
    <div class="font-mono text-[0.84375rem] text-ink-muted" role="cell">—</div>
    <div class="font-mono text-[0.84375rem] text-ink-muted" role="cell">Folder</div>
    <div role="cell"></div>
  </div>
</template>
