<script setup lang="ts">
import { shallowRef, watch } from 'vue'
import { useCopy } from '@/composables/useCopy'
import type { S3Object } from '@/types'
import { fileUrl } from '@/lib/objectPath'
import { formatBytes, formatDateTime } from '@/lib/format'
import PrtBtn from '@/components/ui/btn/PrtBtn.vue'
import PrtDrawer from '@/components/ui/drawer/PrtDrawer.vue'
import IconClipboard from '~icons/lucide/clipboard'
import IconDownload from '~icons/lucide/download'
import IconTrash2 from '~icons/lucide/trash-2'

const props = defineProps<{ object: S3Object | null; bucket: string }>()
const emit = defineEmits<{ close: []; delete: [key: string] }>()

const { copy } = useCopy()

// The sheet slides out over ~150ms after the prop goes null. Rendering from the
// prop would empty it mid-slide, so the body reads the last object it was given.
const shown = shallowRef<S3Object | null>(props.object)
watch(() => props.object, (o) => { if (o) shown.value = o })

function href(key: string): string {
  return fileUrl(props.bucket, key)
}

// Wider than the row thumbnail and unconstrained in height: the panel preview is
// the one place the image is meant to be looked at.
function previewUrl(key: string): string {
  return `${href(key)}?w=320&fmt=webp`
}

function onPreviewError(e: Event, key: string) {
  const img = e.target as HTMLImageElement
  img.onerror = null
  img.src = href(key)
}

function baseName(key: string): string {
  return key.split('/').pop() || key
}

function absoluteUrl(key: string): string {
  return window.location.origin + href(key)
}

function copyUrl() {
  if (!shown.value) return
  copy(absoluteUrl(shown.value.key), 'Link copied')
}
</script>

<template>
  <PrtDrawer
    :model-value="object !== null"
    placement="right"
    close-label="Close details"
    @update:model-value="emit('close')"
  >
    <template #title>Object details</template>

    <div v-if="shown" class="panel-body">
      <div v-if="shown.content_type?.startsWith('image/')" class="panel-preview">
        <img :src="previewUrl(shown.key)" :alt="baseName(shown.key)" class="preview-img" @error="onPreviewError($event, shown.key)" />
      </div>

      <div class="meta-list">
        <div class="meta-row">
          <span class="meta-label">Name</span>
          <span class="meta-value">{{ baseName(shown.key) }}</span>
        </div>
        <div class="meta-row">
          <span class="meta-label">Full path</span>
          <span class="meta-value meta-mono">{{ shown.key }}</span>
        </div>
        <div class="meta-row">
          <span class="meta-label">Size</span>
          <span class="meta-value">{{ formatBytes(shown.size) }} <span class="meta-dim">({{ shown.size.toLocaleString() }} bytes)</span></span>
        </div>
        <div class="meta-row">
          <span class="meta-label">Modified</span>
          <span class="meta-value">{{ formatDateTime(shown.last_modified) }}</span>
        </div>
        <div class="meta-row">
          <span class="meta-label">ETag</span>
          <span class="meta-value meta-mono">{{ shown.etag }}</span>
        </div>
        <div class="meta-row">
          <span class="meta-label">Type</span>
          <span class="meta-value">{{ shown.content_type || '—' }}</span>
        </div>
        <div class="meta-row url-row">
          <span class="meta-label">URL</span>
          <div class="url-field">
            <input :value="absoluteUrl(shown.key)" readonly class="url-input" />
            <button
              type="button"
              class="shrink-0 w-7 h-7 flex items-center justify-center rounded-control border border-edge bg-transparent cursor-pointer text-ink-muted prt-motion-colors hover:bg-surface-3 hover:text-ink"
              title="Copy URL"
              aria-label="Copy URL"
              @click="copyUrl"
            ><IconClipboard class="w-3.5 h-3.5" /></button>
          </div>
        </div>
      </div>

      <div class="panel-actions">
        <PrtBtn tag="a" variant="outline" size="sm" :href="href(shown.key)" :download="baseName(shown.key)">
          <IconDownload class="w-3.5 h-3.5" /> Download
        </PrtBtn>
        <PrtBtn variant="outline" seed="8" size="sm" @click="emit('delete', shown.key)">
          <IconTrash2 class="w-3.5 h-3.5" /> Delete
        </PrtBtn>
      </div>
    </div>
  </PrtDrawer>
</template>

<style scoped>
.panel-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.panel-preview {
  width: 100%;
  background: var(--surface-2);
  border-radius: var(--radius-control);
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-img {
  width: 100%;
  max-height: 200px;
  object-fit: contain;
}

.meta-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.meta-row {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.meta-label {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.034375rem;
  color: var(--ink-faint);
}

.meta-value {
  font-size: 13px;
  color: var(--ink);
  word-break: break-all;
}

.meta-mono {
  font-family: var(--font-mono);
  font-size: 12.5px;
}

.meta-dim {
  color: var(--ink-muted);
}

.url-row { gap: 6px; }

.url-field {
  display: flex;
  gap: 6px;
  align-items: center;
}

.url-input {
  flex: 1;
  font-size: 12px;
  font-family: var(--font-mono);
  padding: 5px 8px;
  background: var(--surface-2);
  border: 1px solid var(--edge);
  border-radius: var(--radius-control);
  color: var(--ink-muted);
  min-width: 0;
}

.panel-actions {
  display: flex;
  gap: 8px;
  padding-top: 4px;
  border-top: 1px solid var(--edge);
}
</style>
