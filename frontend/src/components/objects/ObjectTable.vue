<script setup lang="ts">
import type { S3Object } from '@/types'
import type { SortKey } from '@/lib/sort'
import { OBJECT_MAX_KEYS } from '@/lib/constants'
import BaseCheckbox from '@/components/base/BaseCheckbox.vue'
import PrtBtn from '@/components/ui/btn/PrtBtn.vue'
import PrtEmptyState from '@/components/ui/emptystate/PrtEmptyState.vue'
import SortHeader from './SortHeader.vue'
import FolderRow from './FolderRow.vue'
import ObjectRow from './ObjectRow.vue'

defineProps<{
  objects: S3Object[]
  folders: string[]
  bucket: string
  prefix: string
  isSelected: (key: string) => boolean
  allSelected: boolean
  someSelected: boolean
  sortKey: SortKey
  sortDir: 'asc' | 'desc'
  searchQuery: string
  truncated: boolean
}>()

const emit = defineEmits<{
  sort: [column: SortKey]
  toggleSelect: [key: string]
  toggleSelectAll: []
  openFolder: [folder: string]
  openObject: [object: S3Object]
  copyUrl: [key: string]
  remove: [key: string]
  upload: []
}>()
</script>

<template>
  <!-- The column track lives here; FolderRow and ObjectRow read it back via
       grid-cols-[var(--cols)]. No column gap: the checkbox and name columns
       are already padded. -->
  <div class="bg-surface-1 border border-edge rounded-surface overflow-hidden [--cols:44px_minmax(160px,1fr)_90px_140px_150px_112px]">
    <!-- The truncated notice and empty state sit outside role="table" - neither
         is a row. -->
    <div role="table" aria-label="Objects">
      <div
        class="grid grid-cols-[var(--cols)] items-center w-full px-4.5 py-2.5 border-b border-edge text-[0.6875rem] font-semibold uppercase tracking-[0.04125rem] text-ink-faint"
        role="row"
      >
        <div class="flex items-center" role="columnheader">
          <BaseCheckbox
            :checked="allSelected"
            :indeterminate="someSelected && !allSelected"
            label="Select all files"
            @change="emit('toggleSelectAll')"
          />
        </div>
        <SortHeader label="Name" column="name" :active="sortKey" :dir="sortDir" @sort="emit('sort', $event)" />
        <SortHeader label="Size" column="size" :active="sortKey" :dir="sortDir" @sort="emit('sort', $event)" />
        <SortHeader label="Modified" column="modified" :active="sortKey" :dir="sortDir" @sort="emit('sort', $event)" />
        <div role="columnheader">Type</div>
        <div role="columnheader"></div>
      </div>

      <!-- Folders render above the files and in the order S3 returned them: they
           are prefixes, so neither the object sort nor the selection applies. -->
      <FolderRow
        v-for="folder in folders"
        :key="folder"
        :folder="folder"
        :prefix="prefix"
        @open="emit('openFolder', folder)"
      />

      <ObjectRow
        v-for="obj in objects"
        :key="obj.key"
        :object="obj"
        :bucket="bucket"
        :prefix="prefix"
        :selected="isSelected(obj.key)"
        @toggle-select="emit('toggleSelect', obj.key)"
        @open="emit('openObject', obj)"
        @copy-url="emit('copyUrl', obj.key)"
        @remove="emit('remove', obj.key)"
      />
    </div>

    <!-- The listing is one page deep and there is no "next" control, so without
         this line the missing objects look like they do not exist. -->
    <div v-if="truncated" class="px-4.5 py-2.5 border-t border-edge text-[0.78125rem] text-ink-muted">
      Showing the first {{ OBJECT_MAX_KEYS.toLocaleString() }} objects in this folder.
    </div>

    <!-- Same component as the view-level states, so the card keeps one empty
         vocabulary whether the bucket is empty or the filter matched nothing. -->
    <PrtEmptyState
      v-if="objects.length === 0 && folders.length === 0"
      icon="i-lucide-file"
      :title='searchQuery ? `No files match "${searchQuery}"` : "This bucket is empty"'
      size="lg"
    >
      <template v-if="!searchQuery" #action>
        <PrtBtn seed="7" size="sm" @click="emit('upload')">Upload files</PrtBtn>
      </template>
    </PrtEmptyState>
  </div>
</template>
