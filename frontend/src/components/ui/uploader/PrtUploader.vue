<!--
  PrtUploader
  Selection/validation/preview ONLY — no transport. Upload is the app's
  job (fetch/FormData); this contract ends at "here are the Files the
  user chose". Controlled by modelValue (PrtUploaderFile[]) — new arrays
  out, never mutation.

  Click path: a REAL <button> inside the dropzone → input.showPicker()
  (stable in both browsers at our floor). The button is the only opener
  and the only tab stop (the sr-only input is tabindex="-1").
  Drop reads dataTransfer.files SYNCHRONOUSLY in the drop handler
  (the DataTransfer is neutered after — no awaits first).
  Drag-over state rides a depth counter: dragleave fires on every child
  crossing, a boolean flickers (classic trap).

  Validation per file: type (accept match) → size (maxSize bytes) →
  count (against modelValue.length + accepted so far; single mode
  replaces instead). One update:modelValue + one add + one reject per
  gesture; rejections are reason CODES ('type' | 'size' | 'count') —
  the consumer renders any message. Image previews are object URLs (never
  FileReader data-URLs), revoked automatically when an entry leaves the
  model — by × here, by consumer array surgery, or on unmount.

  webkitdirectory: the picker may ignore `accept` on some platforms —
  validation still runs per file; don't trust the picker. File System
  Access picker APIs are NOT in Firefox — never used.

  Wrapper-rooted: $attrs land on the hidden file input (the form
  element: name/data-field/id for label-for). It never takes focus, so
  form-layer blur reporting won't fire for it — uploader errors surface
  on submit (acceptable).

  Slot contract:
    default — replaces the icon/title/hint block; slot props:
      { open: () => void, dragging: boolean }
    file — custom list row when showList; slot props:
      { file: PrtUploaderFile, index: number, remove: () => void }
-->
<template>
  <div :class="cx('w-full', props.class)">
    <div
      :class="
        uploaderVariants({
          variant,
          size,
          edges,
          dragging: isDragging,
          error,
          disabled,
        })
      "
      @dragenter.prevent="onDragEnter"
      @dragover.prevent
      @dragleave="onDragLeave"
      @drop.prevent="onDrop"
      @paste="onPaste"
    >
      <input
        :id="id"
        ref="inputEl"
        type="file"
        class="sr-only"
        tabindex="-1"
        :accept="accept"
        :multiple="multiple || directory"
        :webkitdirectory="directory || undefined"
        :disabled="disabled"
        :aria-invalid="error || undefined"
        v-bind="$attrs"
        @change="onChange"
      />
      <slot :open="openPicker" :dragging="isDragging">
        <span
          :class="[
            'i-lucide-upload text-ink-faint',
            size === 'lg' ? 'w-8 h-8' : size === 'sm' ? 'w-5 h-5' : 'w-6 h-6',
          ]"
          aria-hidden="true"
        />
        <!-- real gap, not a whitespace text node — the compiler's
             whitespace condensing eats leading spaces ("uploador") -->
        <p
          :class="[
            'flex flex-wrap items-baseline justify-center gap-x-1.5',
            size === 'lg' ? 'text-base' : 'text-sm',
          ]"
        >
          <button
            type="button"
            :disabled="disabled"
            :class="uploaderButtonVariants({ size })"
            @click="openPicker"
          >
            {{ buttonLabel }}
          </button>
          <span v-if="titleText" class="text-ink-muted">{{ titleText }}</span>
        </p>
        <p v-if="hint" class="text-xs font-mono text-ink-faint">{{ hint }}</p>
      </slot>
    </div>
    <ul
      v-if="showList && modelValue.length > 0"
      class="mt-3 flex flex-col list-none p-0 m-0"
    >
      <li
        v-for="(entry, i) in modelValue"
        :key="entry.id"
        :class="uploaderRowVariants({ size })"
      >
        <slot name="file" :file="entry" :index="i" :remove="() => removeAt(i)">
          <img
            v-if="entry.previewUrl"
            :src="entry.previewUrl"
            :alt="entry.name"
            class="w-10 h-10 shrink-0 rounded-control object-cover bg-surface-2"
          />
          <span
            v-else
            class="w-10 h-10 shrink-0 rounded-control bg-surface-2 inline-flex items-center justify-center"
          >
            <span class="i-lucide-file w-4 h-4 text-ink-faint" aria-hidden="true" />
          </span>
          <span class="flex-1 min-w-0 flex flex-col">
            <span class="text-sm text-ink truncate">{{ entry.name }}</span>
            <span class="text-xs font-mono text-ink-faint tabular-nums">
              {{ formatFileSize(entry.size) }}
            </span>
          </span>
          <button
            type="button"
            :aria-label="removeLabel"
            :disabled="disabled"
            class="shrink-0 inline-flex items-center justify-center w-7 h-7 rounded-full bg-transparent text-ink-muted not-disabled:hover:text-ink not-disabled:hover:bg-wash prt-motion-colors prt-ring"
            @click="removeAt(i)"
          >
            <span class="i-lucide-x w-4 h-4" aria-hidden="true" />
          </button>
        </slot>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, shallowRef, watch } from 'vue'
import { cx } from '../_shared/cx'
import type {
  PrtUploaderFile,
  PrtUploaderProps,
  PrtUploaderRejection,
} from './types'
import {
  uploaderButtonVariants,
  uploaderRowVariants,
  uploaderVariants,
} from './variants'

defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<PrtUploaderProps>(), {
  modelValue: () => [],
  accept: '',
  multiple: false,
  directory: false,
  paste: false,
  preview: true,
  showList: true,
  variant: 'dashed',
  size: 'base',
  edges: 'rounded',
  error: false,
  disabled: false,
  class: '',
  buttonLabel: 'Click to upload',
  titleText: 'or drag and drop',
  removeLabel: 'Remove',
})

const emit = defineEmits<{
  'update:modelValue': [value: PrtUploaderFile[]]
  add: [files: PrtUploaderFile[]]
  remove: [file: PrtUploaderFile, index: number]
  reject: [rejections: PrtUploaderRejection[]]
}>()

const inputEl = shallowRef<HTMLInputElement | null>(null)
const isDragging = ref(false)
let dragDepth = 0

// hintText === undefined derives a summary; '' hides the line entirely
const hint = computed(() => {
  if (props.hintText !== undefined) return props.hintText
  const parts: string[] = []
  if (props.accept)
    parts.push(
      props.accept
        .split(',')
        .map((t) => t.trim())
        .join(', '),
    )
  if (props.maxSize !== undefined) parts.push('up to ' + formatFileSize(props.maxSize))
  return parts.join(' · ')
})

// Object-URL lifecycle: every URL this instance created is revoked the
// moment its entry leaves the model — whichever path removed it.
const createdUrls = new Set<string>()
watch(
  () => props.modelValue,
  (entries) => {
    const live = new Set(entries.map((entry) => entry.previewUrl))
    for (const url of createdUrls) {
      if (!live.has(url)) {
        URL.revokeObjectURL(url)
        createdUrls.delete(url)
      }
    }
  },
)
onBeforeUnmount(() => {
  for (const url of createdUrls) URL.revokeObjectURL(url)
})

let seq = 0
function uid(): string {
  seq += 1
  return Date.now().toString(36) + '-' + seq.toString(36)
}

function wrap(file: File): PrtUploaderFile {
  const rel = (file as File & { webkitRelativePath?: string }).webkitRelativePath
  const entry: PrtUploaderFile = {
    id: uid(),
    file,
    name: rel || file.name,
    size: file.size,
    type: file.type,
  }
  if (props.preview && file.type.startsWith('image/')) {
    entry.previewUrl = URL.createObjectURL(file)
    createdUrls.add(entry.previewUrl)
  }
  return entry
}

function isValidFileType(file: File, accept: string): boolean {
  const fileType = file.type.toLowerCase()
  const fileName = file.name.toLowerCase()
  return accept
    .split(',')
    .map((t) => t.trim().toLowerCase())
    .filter(Boolean)
    .some((t) => {
      if (t.endsWith('/*')) return fileType.startsWith(t.slice(0, -1))
      if (t.startsWith('.')) return fileName.endsWith(t)
      return t === fileType
    })
}

function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(k)), units.length - 1)
  return parseFloat((bytes / k ** i).toFixed(1)) + ' ' + units[i]
}

// type → size → count, per file; ONE model emit + ONE add + ONE reject
// per gesture. Single mode replaces (base = []) instead of counting.
function process(list: FileList | File[]) {
  if (props.disabled) return
  const candidates = Array.from(list)
  if (candidates.length === 0) return
  const isMulti = props.multiple || props.directory
  const base = isMulti ? [...props.modelValue] : []
  const cap = isMulti ? (props.maxFiles ?? Number.POSITIVE_INFINITY) : 1
  const accepted: PrtUploaderFile[] = []
  const rejections: PrtUploaderRejection[] = []
  for (const file of candidates) {
    if (props.accept && !isValidFileType(file, props.accept)) {
      rejections.push({ file, reason: 'type' })
    } else if (props.maxSize !== undefined && file.size > props.maxSize) {
      rejections.push({ file, reason: 'size' })
    } else if (base.length + accepted.length >= cap) {
      rejections.push({ file, reason: 'count' })
    } else {
      accepted.push(wrap(file))
    }
  }
  if (accepted.length > 0) {
    emit('update:modelValue', [...base, ...accepted])
    emit('add', accepted)
  }
  if (rejections.length > 0) emit('reject', rejections)
}

function openPicker() {
  if (props.disabled) return
  inputEl.value?.showPicker()
}

function onChange(event: Event) {
  const input = event.target as HTMLInputElement
  if (input.files) process(input.files)
  // re-picking the same file must fire change again
  input.value = ''
}

function onDragEnter() {
  if (props.disabled) return
  dragDepth += 1
  isDragging.value = true
}

function onDragLeave() {
  if (dragDepth > 0) dragDepth -= 1
  if (dragDepth === 0) isDragging.value = false
}

function onDrop(event: DragEvent) {
  dragDepth = 0
  isDragging.value = false
  if (props.disabled) return
  // synchronous read — the DataTransfer is neutered after the handler
  const files = event.dataTransfer?.files
  if (files) process(files)
}

function onPaste(event: ClipboardEvent) {
  if (!props.paste || props.disabled) return
  const files = event.clipboardData?.files
  if (files && files.length > 0) {
    event.preventDefault()
    process(files)
  }
}

function removeAt(index: number) {
  const entry = props.modelValue[index]
  if (!entry) return
  emit(
    'update:modelValue',
    props.modelValue.filter((_, i) => i !== index),
  )
  emit('remove', entry, index)
}
</script>
