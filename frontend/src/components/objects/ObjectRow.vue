<script setup lang="ts">
import { computed } from 'vue'
import type { S3Object } from '@/types'
import { formatBytes, formatDate } from '@/lib/format'
import { fileUrl, objectName, thumbUrl } from '@/lib/objectPath'
import BaseCheckbox from '@/components/base/BaseCheckbox.vue'
import IconDownload from '~icons/lucide/download'
import IconEye from '~icons/lucide/eye'
import IconFile from '~icons/lucide/file'
import IconLink from '~icons/lucide/link'
import IconTrash2 from '~icons/lucide/trash-2'

const props = defineProps<{
  object: S3Object
  bucket: string
  prefix: string
  selected: boolean
}>()

const emit = defineEmits<{
  toggleSelect: []
  open: []
  copyUrl: []
  remove: []
}>()

const name = computed(() => objectName(props.object.key, props.prefix))
const href = computed(() => fileUrl(props.bucket, props.object.key))
const isImage = computed(() => props.object.content_type?.startsWith('image/') ?? false)

// Complete literal strings, one per action flavour. The two hover washes are
// both hover:bg-*, so they may never appear on the same element - which one
// won would be generated-CSS order, not class order.
const actionClass = 'inline-flex items-center justify-center w-6.5 h-6.5 rounded-control bg-transparent cursor-pointer text-ink-faint no-underline hover:no-underline prt-motion-colors hover:bg-surface-3 hover:text-accent'
const dangerActionClass = 'inline-flex items-center justify-center w-6.5 h-6.5 rounded-control bg-transparent cursor-pointer text-ink-faint no-underline hover:no-underline prt-motion-colors hover:bg-[var(--danger-wash)] hover:text-danger'

// One retry at full size, then stop: clearing onerror first is what keeps a
// 404 on the original from looping requests forever.
function onThumbError(e: Event) {
  const img = e.target as HTMLImageElement
  img.onerror = null
  img.src = href.value
}
</script>

<template>
  <!-- group, because the row actions rest at opacity 0 and come back on hover
       or focus-within: keyboard users cannot hover. -->
  <div
    class="group grid grid-cols-[var(--cols)] items-center w-full px-4.5 py-2.5 border-b border-edge last:border-b-0 prt-motion-colors hover:bg-surface-2"
    role="row"
  >
    <div class="flex items-center" role="cell">
      <BaseCheckbox :checked="selected" :label="`Select ${name}`" @change="emit('toggleSelect')" />
    </div>
    <div class="flex items-center min-w-0" role="cell">
      <!-- The ink lives on the button so the hover colour reaches the label by
           inheritance rather than through a second group. -->
      <button
        type="button"
        class="flex items-center w-full min-w-0 p-0 text-left bg-transparent cursor-pointer font-medium text-ink hover:text-accent"
        @click="emit('open')"
      >
        <span class="flex items-center gap-2.5 min-w-0 overflow-hidden text-ellipsis whitespace-nowrap">
          <img
            v-if="isImage"
            :src="thumbUrl(bucket, object.key)"
            class="rounded-control object-cover shrink-0"
            width="26"
            height="26"
            loading="lazy"
            :alt="name"
            @error="onThumbError"
          />
          <span
            v-else
            class="w-6.5 h-6.5 rounded-control flex items-center justify-center shrink-0 bg-surface-3 text-ink-muted"
          ><IconFile class="w-3.5 h-3.5" /></span>
          {{ name }}
        </span>
      </button>
    </div>
    <div class="font-mono text-[0.84375rem] text-ink-muted" role="cell">{{ formatBytes(object.size) }}</div>
    <div class="font-mono text-[0.84375rem] text-ink-muted" role="cell">{{ formatDate(object.last_modified) }}</div>
    <div class="font-mono text-[0.78125rem] text-ink-muted overflow-hidden text-ellipsis whitespace-nowrap" role="cell">
      {{ object.content_type || '—' }}
    </div>
    <div class="flex justify-end" role="cell">
      <div class="flex items-center gap-0.5 opacity-0 prt-motion group-hover:opacity-100 group-focus-within:opacity-100">
        <button type="button" :class="actionClass" title="Copy URL" :aria-label="`Copy URL for ${name}`" @click="emit('copyUrl')">
          <IconLink class="w-3.5 h-3.5" />
        </button>
        <a :href="href" :download="name" :class="actionClass" title="Download" :aria-label="`Download ${name}`">
          <IconDownload class="w-3.5 h-3.5" />
        </a>
        <a :href="href" target="_blank" rel="noopener" :class="actionClass" title="View" :aria-label="`View ${name}`">
          <IconEye class="w-3.5 h-3.5" />
        </a>
        <button type="button" :class="dangerActionClass" title="Delete" :aria-label="`Delete ${name}`" @click="emit('remove')">
          <IconTrash2 class="w-3.5 h-3.5" />
        </button>
      </div>
    </div>
  </div>
</template>
