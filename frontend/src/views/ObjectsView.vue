<script setup lang="ts">
import { ref, computed, watch, onMounted, useTemplateRef } from 'vue'
import { storeToRefs } from 'pinia'
import { useObjectsStore } from '@/stores/objects'
import { useStatsStore } from '@/stores/stats'
import { useConfirmStore } from '@/stores/confirm'
import { useToast } from '@/components/ui/toast/useToast'
import { useHotkey } from '@/composables/useHotkey'
import { useCopy } from '@/composables/useCopy'
import { useSelection } from '@/composables/useSelection'
import { useTableSort } from '@/composables/useTableSort'
import SearchInput from '@/components/base/SearchInput.vue'
import ObjectBreadcrumbs from '@/components/objects/ObjectBreadcrumbs.vue'
import ObjectSelectionBar from '@/components/objects/ObjectSelectionBar.vue'
import NewFolderBar from '@/components/objects/NewFolderBar.vue'
import ObjectTable from '@/components/objects/ObjectTable.vue'
import UploadDialog from '@/components/objects/UploadDialog.vue'
import ObjectDetailsPanel from '@/components/objects/ObjectDetailsPanel.vue'
import PrtBtn from '@/components/ui/btn/PrtBtn.vue'
import PrtEmptyState from '@/components/ui/emptystate/PrtEmptyState.vue'
import PrtLoader from '@/components/ui/loader/PrtLoader.vue'
import PrtNotice from '@/components/ui/notice/PrtNotice.vue'
import type { S3Object } from '@/types'
import { objectComparators } from '@/lib/sort'
import { fileUrl, folderName, objectName } from '@/lib/objectPath'
import { errorMessage } from '@/lib/errors'
import IconFolderPlus from '~icons/lucide/folder-plus'
import IconUpload from '~icons/lucide/upload'

const props = defineProps<{ bucket: string }>()

const objectsStore = useObjectsStore()
const { objects, prefixes, loading, error, truncated } = storeToRefs(objectsStore)
const stats = useStatsStore()
const confirm = useConfirmStore()
const { push } = useToast()
// duration: 0 - a failure usually names a key the operator has to go act on, so
// it waits to be dismissed rather than timing out.
const fail = (message: string) => push({ intent: 'danger', message, duration: 0 })
const { copy } = useCopy()

const prefix = ref('')
const searchQuery = ref('')
const searchInput = useTemplateRef<{ focus: () => void }>('searchInput')
const deleting = ref(false)
const showUpload = ref(false)
const showNewFolder = ref(false)
const droppedFiles = ref<File[]>([])
const selectedObject = ref<S3Object | null>(null)

const selection = useSelection()

const filteredObjects = computed(() => {
  const q = searchQuery.value.toLowerCase()
  if (!q) return objects.value
  return objects.value.filter(o => objectName(o.key, prefix.value).toLowerCase().includes(q))
})

const filteredPrefixes = computed(() => {
  const q = searchQuery.value.toLowerCase()
  if (!q) return prefixes.value
  return prefixes.value.filter(p => folderName(p, prefix.value).toLowerCase().includes(q))
})

// The name column sorts on the displayed name, which is key minus prefix, so
// the comparator table has to be rebuilt whenever the prefix moves.
const comparators = computed(() => objectComparators(prefix.value))
const { key: sortKey, dir: sortDir, sorted: sortedObjects, toggle: toggleSort } = useTableSort(filteredObjects, comparators)

const visibleKeys = computed(() => sortedObjects.value.map(o => o.key))
const allSelected = computed(() => selection.allSelected(visibleKeys.value))

onMounted(load)

// Switching buckets has to drop the prefix too - otherwise the new bucket opens
// at a path that only existed in the old one.
watch(() => props.bucket, () => { prefix.value = ''; load() })

useHotkey('u', () => { showUpload.value = true })
useHotkey('/', () => searchInput.value?.focus())
useHotkey('Escape', onEscape, { whileTyping: true })

function onEscape() {
  if (selectedObject.value) { selectedObject.value = null; return }
  if (showUpload.value) { showUpload.value = false; return }
  if (showNewFolder.value) showNewFolder.value = false
}

function load() {
  selection.clear()
  objectsStore.load(props.bucket, prefix.value)
}

function navigateTo(path: string) {
  prefix.value = path
  load()
}

async function handleDeleteSelected() {
  const keys = [...selection.ids.value]
  if (keys.length === 0) return
  const ok = await confirm.ask({
    title: 'Delete objects',
    message: `Delete ${keys.length} object${keys.length > 1 ? 's' : ''}?`,
    confirmLabel: 'Delete',
    danger: true,
  })
  if (!ok) return
  deleting.value = true
  try {
    const result = await objectsStore.remove(props.bucket, keys)
    load()
    // after load(), which clears it: the failed keys stay selected so a retry
    // needs no reselection
    selection.set(result.errors.map(e => e.key))
    if (result.errors.length) {
      fail(`Could not delete ${result.errors.length} object(s): ${result.errors.map(e => e.key).join(', ')}`)
    }
  } catch (e) {
    fail(errorMessage(e, 'Failed to delete'))
  } finally {
    deleting.value = false
  }
}

async function handleDeleteSingle(key: string) {
  try {
    const result = await objectsStore.remove(props.bucket, [key])
    load()
    if (result.errors.length) {
      fail(`Could not delete ${key}: ${result.errors.map(e => e.error).join(', ')}`)
    }
  } catch (e) {
    fail(errorMessage(e, 'Failed to delete'))
  }
}

async function handlePanelDelete(key: string) {
  await handleDeleteSingle(key)
  selectedObject.value = null
}

function copyUrl(key: string) {
  copy(window.location.origin + fileUrl(props.bucket, key), 'Link copied')
}

function copyCurrentPath() {
  copy(`/${props.bucket}/${prefix.value}`, 'Prefix copied')
}

// Folders are virtual (S3 semantics): navigating into the prefix is the whole
// operation; it materializes when the first object is uploaded there.
function createFolder(name: string) {
  prefix.value = prefix.value + name + '/'
  showNewFolder.value = false
  load()
}

// Uploads bypass the objects store (they stream through api/upload), so unlike
// a normal store mutation they don't trigger a stats refresh - call it here.
function onUploadComplete() {
  showUpload.value = false
  droppedFiles.value = []
  load()
  stats.load()
}

function handleDrop(e: DragEvent) {
  const files = e.dataTransfer?.files
  if (files && files.length > 0) {
    droppedFiles.value = Array.from(files)
    showUpload.value = true
  }
}
</script>

<template>
  <div @dragover.prevent @drop.prevent="handleDrop">
    <div class="flex items-center justify-between gap-4 mb-4">
      <ObjectBreadcrumbs :bucket="bucket" :prefix="prefix" @navigate="navigateTo" @copy-path="copyCurrentPath" />
      <div class="flex items-center gap-2">
        <PrtBtn variant="outline" size="sm" @click="showNewFolder = true">
          <IconFolderPlus class="w-3.5 h-3.5" /> New folder
        </PrtBtn>
        <PrtBtn seed="7" @click="showUpload = true">
          <IconUpload class="w-3.5 h-3.5" /> Upload
        </PrtBtn>
      </div>
    </div>

    <ObjectSelectionBar
      v-if="selection.count.value > 0"
      :count="selection.count.value"
      :busy="deleting"
      @clear="selection.clear()"
      @remove="handleDeleteSelected"
    />

    <NewFolderBar v-if="showNewFolder" @create="createFolder" @cancel="showNewFolder = false" />

    <!-- The `/` hotkey focuses this, which is why SearchInput exposes focus(). -->
    <SearchInput
      ref="searchInput"
      v-model="searchQuery"
      label="Filter files"
      placeholder="Filter files… (/)"
      class="mb-3 max-w-90"
    />

    <PrtNotice v-if="error" intent="danger" variant="wash" role="alert" class="mb-3" :description="error" />

    <!-- Same chrome as the table's own empty state, so the listing does not jump
         when one replaces the other. -->
    <PrtEmptyState v-if="loading" title="Loading…">
      <template #icon>
        <PrtLoader />
      </template>
    </PrtEmptyState>

    <ObjectTable
      v-else
      :objects="sortedObjects"
      :folders="filteredPrefixes"
      :bucket="bucket"
      :prefix="prefix"
      :is-selected="selection.has"
      :all-selected="allSelected"
      :some-selected="selection.count.value > 0"
      :sort-key="sortKey"
      :sort-dir="sortDir"
      :search-query="searchQuery"
      :truncated="truncated"
      @sort="toggleSort"
      @toggle-select="selection.toggle"
      @toggle-select-all="selection.toggleAll(visibleKeys)"
      @open-folder="navigateTo"
      @open-object="selectedObject = $event"
      @copy-url="copyUrl"
      @remove="handleDeleteSingle"
      @upload="showUpload = true"
    />

    <!-- Always mounted: the panel owns its own open state so it can animate out
         as well as in. -->
    <ObjectDetailsPanel
      :object="selectedObject"
      :bucket="bucket"
      @close="selectedObject = null"
      @delete="handlePanelDelete"
    />

    <UploadDialog
      v-if="showUpload"
      :bucket="bucket"
      :prefix="prefix"
      :initial-files="droppedFiles"
      @close="showUpload = false; droppedFiles = []"
      @complete="onUploadComplete"
    />
  </div>
</template>
