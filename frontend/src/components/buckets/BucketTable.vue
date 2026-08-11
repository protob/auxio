<script setup lang="ts">
import type { Bucket } from '@/types'
import type { GroupedRow } from '@/lib/groupRows'
import BaseCheckbox from '@/components/base/BaseCheckbox.vue'
import BucketRow from './BucketRow.vue'
import BucketGroupHeader from './BucketGroupHeader.vue'

defineProps<{
  rows: GroupedRow[]
  selectMode: boolean
  grouped: boolean
  allVisibleSelected: boolean
  isSelected: (name: string) => boolean
  deleting: string | null
  toggling: string | null
}>()

const emit = defineEmits<{
  open: [name: string]
  toggleSelect: [name: string]
  toggleSelectAll: []
  toggleGroup: [key: string]
  togglePin: [bucket: Bucket]
  togglePublic: [bucket: Bucket]
  remove: [bucket: Bucket]
}>()
</script>

<template>
  <!-- The column track lives here; BucketRow reads it back via
       grid-cols-[var(--cols)]. -->
  <div
    class="bg-surface-1 border border-edge rounded-surface overflow-hidden"
    :class="selectMode
      ? '[--cols:34px_minmax(180px,1fr)_100px_120px_120px_240px]'
      : '[--cols:minmax(180px,1fr)_100px_120px_120px_240px]'"
  >
    <!-- The footer sits outside role="table" - it is not a row. -->
    <div role="table" aria-label="Buckets">
      <div
        class="grid grid-cols-[var(--cols)] items-center w-full gap-3 px-5 py-2.5 border-b border-edge text-[0.6875rem] font-semibold uppercase tracking-[0.04125rem] text-ink-faint"
        role="row"
      >
        <div v-if="selectMode" class="flex items-center justify-center" role="columnheader">
          <BaseCheckbox :checked="allVisibleSelected" label="Select all" @change="emit('toggleSelectAll')" />
        </div>
        <div role="columnheader">Name</div>
        <div class="text-right" role="columnheader">Objects</div>
        <div class="text-right" role="columnheader">Size</div>
        <div role="columnheader">Access</div>
        <div role="columnheader"></div>
      </div>

      <template
        v-for="row in rows"
        :key="row.type === 'header' ? `h:${row.key}` : `b:${row.bucket.name}`"
      >
        <BucketGroupHeader
          v-if="row.type === 'header'"
          :label="row.label"
          :count="row.count"
          :open="row.open"
          @toggle="emit('toggleGroup', row.key)"
        />
        <BucketRow
          v-else
          :bucket="row.bucket"
          :select-mode="selectMode"
          :selected="isSelected(row.bucket.name)"
          :show-group-chip="!grouped"
          :deleting="deleting === row.bucket.name"
          :toggling="toggling === row.bucket.name"
          @open="emit('open', row.bucket.name)"
          @toggle-select="emit('toggleSelect', row.bucket.name)"
          @toggle-pin="emit('togglePin', row.bucket)"
          @toggle-public="emit('togglePublic', row.bucket)"
          @remove="emit('remove', row.bucket)"
        />
      </template>
    </div>

    <slot name="footer" />
  </div>
</template>
