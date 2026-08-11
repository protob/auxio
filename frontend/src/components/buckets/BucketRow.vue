<script setup lang="ts">
import { computed } from 'vue'
import type { Bucket } from '@/types'
import { formatBytes } from '@/lib/format'
import BaseCheckbox from '@/components/base/BaseCheckbox.vue'
import PrtBtn from '@/components/ui/btn/PrtBtn.vue'
import IconDatabase from '~icons/lucide/database'
import IconStar from '~icons/lucide/star'
import IconStarSolid from '~icons/prt/star-solid'
import IconTrash2 from '~icons/lucide/trash-2'

const props = defineProps<{
  bucket: Bucket
  selectMode: boolean
  selected: boolean
  showGroupChip: boolean
  deleting: boolean
  toggling: boolean
}>()

const emit = defineEmits<{
  open: []
  toggleSelect: []
  togglePin: []
  togglePublic: []
  remove: []
}>()

// In select mode the name picks the bucket instead of opening it: the row is
// then a selection target, and opening one would lose the selection.
function onNameClick() {
  props.selectMode ? emit('toggleSelect') : emit('open')
}

const blocked = computed(() => props.deleting || props.bucket.object_count > 0)

// aria-disabled doesn't block clicks the way disabled would, so this guards itself.
function onDelete() {
  if (blocked.value) return
  emit('remove')
}
</script>

<template>
  <!-- group, because the pin star and the action buttons rest at opacity 0 and
       come back on hover or focus-within of the whole row. -->
  <div
    class="group grid grid-cols-[var(--cols)] items-center w-full gap-3 px-5 py-3 border-b border-edge prt-motion-colors"
    :class="selectMode && selected ? 'bg-[var(--accent-wash)]' : 'hover:bg-surface-2'"
    role="row"
  >
    <div v-if="selectMode" class="flex items-center justify-center" role="cell">
      <BaseCheckbox :checked="selected" :label="`Select ${bucket.name}`" @change="emit('toggleSelect')" />
    </div>

    <div class="flex items-center gap-2.5" role="cell">
      <!-- Pinned and unpinned are one binding: a pinned star is always visible
           and stays amber under the pointer, so it must not carry the reveal or
           the neutral hover at all. -->
      <button
        type="button"
        class="flex w-5.5 h-5.5 items-center justify-center shrink-0 rounded-control bg-transparent cursor-pointer prt-motion hover:bg-wash"
        :class="bucket.pinned
          ? 'opacity-100 text-warning'
          : 'opacity-0 text-ink-faint hover:text-ink-muted group-hover:opacity-100 group-focus-within:opacity-100'"
        :title="bucket.pinned ? 'Unpin' : 'Pin'"
        :aria-label="`${bucket.pinned ? 'Unpin' : 'Pin'} ${bucket.name}`"
        :aria-pressed="bucket.pinned"
        @click="emit('togglePin')"
      >
        <IconStarSolid v-if="bucket.pinned" class="w-3.5 h-3.5" />
        <IconStar v-else class="w-3.5 h-3.5" />
      </button>
      <!-- bare icon, no tile - a colored square on every row would be the
           heaviest thing on the page. Accent only when the bucket holds data;
           empties stay faint so content is visible at a glance. -->
      <span
        class="flex items-center justify-center shrink-0"
        :class="bucket.object_count === 0 ? 'text-ink-faint' : 'text-accent'"
      >
        <IconDatabase class="w-4 h-4" />
      </span>
      <!-- Only the name is clickable, not the whole row: a click handler on the
           row div is unreachable without a pointer. -->
      <button
        type="button"
        class="p-0 text-left bg-transparent cursor-pointer text-[0.90625rem] font-medium text-ink hover:underline"
        @click="onNameClick"
      >{{ bucket.name }}</button>
      <!-- chip is redundant under a group header; only flat view passes this -->
      <span
        v-if="showGroupChip && bucket.group"
        class="px-2 py-px rounded-[0.3125rem] bg-surface-3 whitespace-nowrap text-[0.65625rem] font-semibold uppercase tracking-[0.02625rem] text-ink-muted"
      >{{ bucket.group }}</span>
    </div>

    <div
      class="font-mono text-right"
      role="cell"
      :class="bucket.object_count === 0 ? 'text-ink-faint' : 'text-ink-muted'"
    >
      {{ bucket.object_count === 0 ? '—' : bucket.object_count.toLocaleString() }}
    </div>
    <div
      class="font-mono text-right"
      role="cell"
      :class="bucket.total_size === 0 ? 'text-ink-faint' : 'text-ink-muted'"
    >
      {{ bucket.total_size === 0 ? '—' : formatBytes(bucket.total_size) }}
    </div>

    <div role="cell">
      <!-- status = colored dot + quiet text; a filled pill per row is pure weight -->
      <span class="inline-flex items-center gap-1.75 text-[0.78125rem] font-medium text-ink-muted">
        <span class="w-1.5 h-1.5 rounded-full" :class="bucket.public ? 'bg-accent' : 'bg-ink-faint'"></span>
        {{ bucket.public ? 'Public' : 'Private' }}
      </span>
    </div>

    <!-- rest state is calm: actions surface on the row you are on (or tabbing
         through), instead of twelve delete buttons competing for attention -->
    <div
      class="flex items-center justify-end gap-1.5 opacity-0 prt-motion group-hover:opacity-100 group-focus-within:opacity-100"
      role="cell"
    >
      <template v-if="!selectMode">
        <PrtBtn
          variant="outline"
          size="sm"
          :disabled="toggling"
          :aria-label="`${bucket.public ? 'Make private' : 'Make public'}: ${bucket.name}`"
          @click="emit('togglePublic')"
        >
          {{ toggling ? '…' : bucket.public ? 'Make private' : 'Make public' }}
        </PrtBtn>
        <PrtBtn
          variant="outline"
          seed="8"
          size="sm"
          :aria-disabled="blocked"
          :aria-label="bucket.object_count > 0
            ? `Delete ${bucket.name} - unavailable, bucket must be empty`
            : `Delete ${bucket.name}`"
          :title="bucket.object_count > 0 ? 'Bucket must be empty' : undefined"
          @click="onDelete"
        >
          <IconTrash2 class="w-3.5 h-3.5" />
          {{ deleting ? 'Deleting…' : 'Delete' }}
        </PrtBtn>
      </template>
    </div>
  </div>
</template>
