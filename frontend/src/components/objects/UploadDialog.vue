<script setup lang="ts">
import { nextTick, ref, onMounted } from 'vue'
import { uploadFile as uploadObject } from '@/api/upload'
import { UPLOAD_CONCURRENCY } from '@/lib/constants'
import type { UploadProgress } from '@/types'
import { errorMessage } from '@/lib/errors'
import PrtDialog from '@/components/ui/dialog/PrtDialog.vue'
import PrtBtn from '@/components/ui/btn/PrtBtn.vue'
import PrtProgress from '@/components/ui/progress/PrtProgress.vue'
import PrtUploader from '@/components/ui/uploader/PrtUploader.vue'
import type { PrtUploaderFile } from '@/components/ui/uploader/types'
import IconFolder from '~icons/lucide/folder'

const props = defineProps<{ bucket: string; prefix: string; initialFiles?: File[] }>()
const emit = defineEmits<{ close: []; complete: [] }>()

const uploads = ref<UploadProgress[]>([])

// PrtUploader is a picker here, not a list: the queue below is the upload
// queue, so nothing is bound to its v-model and every batch arrives via @add.
const directory = ref(false)

onMounted(() => {
  if (props.initialFiles && props.initialFiles.length > 0) {
    addFilesArray(props.initialFiles)
  }
})

const queue: number[] = []
let active = 0

function addFilesArray(files: File[]) {
  for (const file of files) {
    const key = props.prefix + (file.webkitRelativePath || file.name)
    queue.push(uploads.value.push({ file, key, progress: 0, status: 'pending' }) - 1)
  }
  pump()
}

function onAdd(entries: PrtUploaderFile[]) {
  addFilesArray(entries.map(e => e.file))
}

function pump() {
  while (active < UPLOAD_CONCURRENCY && queue.length > 0) {
    const index = queue.shift()
    if (index === undefined) return
    active++
    void uploadOne(index).finally(() => {
      active--
      pump()
    })
  }
}

async function uploadOne(index: number) {
  const upload = uploads.value[index]
  if (!upload) return
  upload.status = 'uploading'
  try {
    await uploadObject(props.bucket, upload.key, upload.file, (p) => {
      upload.progress = p
    })
    upload.status = 'complete'
    upload.progress = 100
  } catch (e) {
    upload.status = 'error'
    upload.error = errorMessage(e, 'Upload failed')
  }
}

// nothing to clean up before a retry - the PUT is idempotent
function retry(u: UploadProgress) {
  const idx = uploads.value.indexOf(u)
  if (idx === -1) return
  u.error = undefined
  u.progress = 0
  queue.push(idx)
  pump()
}

// webkitdirectory is a prop, so the folder button flips it and waits for the
// attribute to land before opening the picker it shares with the dropzone.
async function browseFolder(open: () => void) {
  directory.value = true
  await nextTick()
  open()
}

function browseFiles(open: () => void) {
  directory.value = false
  open()
}

const allDone = () => uploads.value.length > 0 && uploads.value.every(u => u.status === 'complete' || u.status === 'error')

// Closing after everything landed still has to reload the listing behind the
// dialog, which is what 'complete' means to the parent.
function dismiss() {
  if (allDone()) emit('complete')
  else emit('close')
}

// exhaustive by construction: a new status is a compile error here
const badgeLabel = (u: UploadProgress) => ({
  pending: 'Pending', uploading: `${u.progress}%`, complete: 'Done', error: 'Error',
}[u.status])
</script>

<template>
  <!-- model-value is a constant true and the parent mounts this with v-if: one
       instance per open is what clears the queue, and useModalDialog opens on
       mount when the model is already true. Every close path emits false. -->
  <PrtDialog :model-value="true" width="30rem" @update:model-value="dismiss">
    <template #title>Upload Files</template>
    <div class="dialog-body">
      <!-- No accept/maxSize/maxFiles: auxio puts no client-side limit on an
           upload, so the kit has nothing to reject and @reject cannot fire. -->
      <PrtUploader :directory="directory" multiple paste :show-list="false" @add="onAdd">
        <template #default="{ open }">
          <span class="i-lucide-upload w-7 h-7 text-ink-faint" aria-hidden="true" />
          <p class="text-sm text-ink-muted">
            Drop files here or
            <button type="button" class="drop-link" @click="browseFiles(open)">browse</button>
          </p>
          <PrtBtn variant="outline" size="sm" class="mt-0.5" @click="browseFolder(open)">
            <IconFolder class="w-3.5 h-3.5" />
            Upload folder
          </PrtBtn>
        </template>
      </PrtUploader>

      <div v-if="uploads.length > 0" class="upload-list">
        <!-- Keyed by destination key, not index: retrying reuses the entry, and
             an index key would move the progress bar onto its neighbour. -->
        <div v-for="u in uploads" :key="u.key" class="upload-item">
          <div class="upload-info">
            <span class="upload-name">{{ u.file.webkitRelativePath || u.file.name }}</span>
            <span :class="['upload-badge', `badge-${u.status}`]">{{ badgeLabel(u) }}</span>
          </div>
          <PrtProgress v-if="u.status === 'uploading'" :value="u.progress" size="sm" />
          <div v-if="u.status === 'error'" class="upload-error" role="alert">
            {{ u.error }} <button class="retry-link" @click="retry(u)">Retry</button>
          </div>
        </div>
      </div>
    </div>

    <template v-if="allDone()" #actions>
      <PrtBtn seed="7" @click="emit('complete')">Done</PrtBtn>
    </template>
  </PrtDialog>
</template>

<style scoped>
.dialog-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* The dropzone, its icon and its hover wash are PrtUploader's; only the two
   openers inside the slot are auxio's. */
.drop-link {
  color: var(--accent);
  font-weight: 500;
  background: none;
  border: none;
  padding: 0;
  font-size: inherit;
  cursor: pointer;
}
.drop-link:hover { text-decoration: underline; }

.upload-list { display: flex; flex-direction: column; gap: 8px; }

.upload-item {
  padding: 10px 12px;
  background: var(--surface-2);
  border: 1px solid var(--edge);
  border-radius: var(--radius-control);
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.upload-info { display: flex; align-items: center; justify-content: space-between; gap: 8px; }

.upload-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.upload-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 7px;
  border-radius: 99px;
  flex-shrink: 0;
}
.badge-pending { background: var(--surface-3); color: var(--ink-muted); }
.badge-uploading { background: var(--accent-wash); color: var(--accent); }
.badge-complete { background: var(--accent-wash); color: var(--accent); }
.badge-error { background: var(--danger-wash); color: var(--danger); }

.upload-error { font-size: 12px; color: var(--danger); }
.retry-link {
  font-size: 12px;
  font-weight: 600;
  color: var(--accent);
  background: none;
  border: none;
  padding: 0;
  margin-left: 4px;
  cursor: pointer;
}
.retry-link:hover { text-decoration: underline; }
</style>
