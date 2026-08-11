<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useLocalStorage } from '@vueuse/core'
import { useBucketsStore } from '@/stores/buckets'
import { useConfirmStore } from '@/stores/confirm'
import { useToast } from '@/components/ui/toast/useToast'
import { useHotkey } from '@/composables/useHotkey'
import { useCollapsedGroups } from '@/composables/useCollapsedGroups'
import { usePagination } from '@/composables/usePagination'
import { useSelection } from '@/composables/useSelection'
import BucketCreateDialog from '@/components/buckets/BucketCreateDialog.vue'
import BucketToolbar from '@/components/buckets/BucketToolbar.vue'
import BucketBulkBar from '@/components/buckets/BucketBulkBar.vue'
import BucketTable from '@/components/buckets/BucketTable.vue'
import BucketTableFooter from '@/components/buckets/BucketTableFooter.vue'
import PrtBtn from '@/components/ui/btn/PrtBtn.vue'
import PrtEmptyState from '@/components/ui/emptystate/PrtEmptyState.vue'
import PrtLoader from '@/components/ui/loader/PrtLoader.vue'
import PrtNotice from '@/components/ui/notice/PrtNotice.vue'
import type { Bucket } from '@/types'
import { byPinnedThenName } from '@/lib/sort'
import { fuzzyMatch } from '@/lib/fuzzy'
import { buildGroupRows, groupKeyOf, UNGROUPED, type GroupedRow } from '@/lib/groupRows'
import { BUCKET_PAGE_SIZE, COLLAPSED_KEY, GROUP_BY_KEY } from '@/lib/constants'
import { errorMessage } from '@/lib/errors'
import IconPlus from '~icons/lucide/plus'

const router = useRouter()
const bucketsStore = useBucketsStore()
const { buckets, loading, error } = storeToRefs(bucketsStore)
const confirm = useConfirmStore()
const { push } = useToast()
// duration: 0 - a failure usually names a bucket the operator has to go act on,
// so it waits to be dismissed rather than timing out.
const fail = (message: string) => push({ intent: 'danger', message, duration: 0 })

const showCreate = ref(false)
const deleting = ref<string | null>(null)
const toggling = ref<string | null>(null)
const query = ref('')

const filteredBuckets = computed(() => {
  const q = query.value.trim()
  if (!q) return buckets.value
  return buckets.value.filter(b => fuzzyMatch(b.name, q))
})

const sortedFlat = computed(() => [...filteredBuckets.value].sort(byPinnedThenName))
const { page, totalPages, pageItems, rangeStart, rangeEnd } = usePagination(sortedFlat, BUCKET_PAGE_SIZE)

watch(query, () => { page.value = 1 })

const groupBy = useLocalStorage(GROUP_BY_KEY, true)
const { isCollapsed, toggle: toggleGroup, setAll: setCollapsed } = useCollapsedGroups(COLLAPSED_KEY)

const hasGroups = computed(() => buckets.value.some(b => groupKeyOf(b) !== ''))
const grouped = computed(() => groupBy.value && hasGroups.value)

// group keys visible under the current filter - drives the footer summary and
// the collapse/expand-all control
const visibleGroupKeys = computed(() => {
  const keys = new Set<string>()
  for (const b of filteredBuckets.value) keys.add(groupKeyOf(b) || UNGROUPED)
  return [...keys]
})
const groupCount = computed(() => visibleGroupKeys.value.filter(k => k !== UNGROUPED).length)
const anyGroupOpen = computed(() => visibleGroupKeys.value.some(k => !isCollapsed(k)))
function collapseAll() { setCollapsed(visibleGroupKeys.value) }
function expandAll() { setCollapsed([]) }

const displayRows = computed<GroupedRow[]>(() => {
  if (!grouped.value) {
    return pageItems.value.map(b => ({ type: 'bucket', bucket: b } as GroupedRow))
  }
  // grouped view ignores pagination - every filtered bucket, bucketed by group
  return buildGroupRows(filteredBuckets.value, k => !isCollapsed(k), byPinnedThenName)
})

onMounted(() => bucketsStore.load())

useHotkey('n', () => { showCreate.value = true })
useHotkey('Escape', () => { if (selectMode.value) exitSelect() })

function openBucket(name: string) { router.push(`/${name}`) }

async function handleDelete(bucket: Bucket) {
  const ok = await confirm.ask({
    title: 'Delete bucket',
    message: `Delete bucket "${bucket.name}" and everything in it?`,
    confirmLabel: 'Delete',
    danger: true,
  })
  if (!ok) return
  deleting.value = bucket.name
  try {
    await bucketsStore.remove(bucket.name)
  } catch (e) {
    fail(errorMessage(e, 'Failed to delete bucket'))
  } finally {
    deleting.value = null
  }
}

async function handleTogglePublic(bucket: Bucket) {
  toggling.value = bucket.name
  try {
    await bucketsStore.update(bucket.name, { public: !bucket.public })
  } catch (e) {
    fail(errorMessage(e, 'Failed to update bucket'))
  } finally {
    toggling.value = null
  }
}

async function togglePin(b: Bucket) {
  try { await bucketsStore.update(b.name, { pinned: !b.pinned }) }
  catch (e) { fail(errorMessage(e, 'Failed to pin')) }
}

const selectMode = ref(false)
const selection = useSelection()
const bulkGroup = ref('')
const bulkBusy = ref(false)

function enterSelect() { selectMode.value = true }
function exitSelect() {
  selectMode.value = false
  selection.clear()
  bulkGroup.value = ''
}

// the bucket rows currently rendered (respects filter, grouping, pagination)
const visibleNames = computed(() =>
  displayRows.value.flatMap(r => (r.type === 'bucket' ? [r.bucket.name] : [])))
const allVisibleSelected = computed(() => selection.allSelected(visibleNames.value))

async function runBulk(fn: () => Promise<void>) {
  if (selection.count.value === 0) return
  bulkBusy.value = true
  try { await fn() }
  catch (e) { fail(errorMessage(e, 'Bulk action failed')) }
  finally { bulkBusy.value = false }
}

// A partial failure leaves exactly the failed names selected, so the operator
// can see what is left and retry without reselecting.
async function reportFailures(work: Promise<string[]>) {
  const failed = await work
  if (failed.length === 0) return
  selection.set(failed)
  fail(`${failed.length} of the selected buckets could not be updated: ${failed.join(', ')}`)
}

function assignGroup() {
  const g = bulkGroup.value.trim().toLowerCase()
  runBulk(() => reportFailures(bucketsStore.updateMany([...selection.ids.value], { group: g })))
}
const ungroupSelected = () => runBulk(() => reportFailures(bucketsStore.updateMany([...selection.ids.value], { group: '' })))
const bulkSetPublic = (pub: boolean) => runBulk(() => reportFailures(bucketsStore.updateMany([...selection.ids.value], { public: pub })))

async function bulkDelete() {
  const names = [...selection.ids.value]
  if (names.length === 0) return
  const ok = await confirm.ask({
    title: 'Delete buckets',
    message: `Delete ${names.length} bucket${names.length > 1 ? 's' : ''} and all their objects?`,
    confirmLabel: 'Delete',
    danger: true,
  })
  if (!ok) return
  bulkBusy.value = true
  try {
    const failed = await bucketsStore.removeMany(names)
    selection.set(failed)
    if (failed.length) {
      fail(`Could not delete ${failed.length} bucket${failed.length > 1 ? 's' : ''} (must be empty): ${failed.join(', ')}`)
    }
  } finally {
    bulkBusy.value = false
  }
}
</script>

<template>
  <div>
    <div class="flex items-start justify-between gap-4 mb-6">
      <h1 class="text-[1.5625rem] font-semibold text-ink tracking-[-0.05rem]">Buckets</h1>
      <PrtBtn seed="7" @click="showCreate = true">
        <IconPlus class="w-3.5 h-3.5" /> New Bucket
      </PrtBtn>
    </div>

    <PrtNotice v-if="error" intent="danger" variant="wash" role="alert" class="mb-4" :description="error" />

    <!-- Loading borrows the empty state's chrome through its icon slot, which is
         what keeps the two the same height: the list must not jump when one
         replaces the other. -->
    <PrtEmptyState v-if="loading" title="Loading buckets…">
      <template #icon>
        <PrtLoader />
      </template>
    </PrtEmptyState>

    <PrtEmptyState v-else-if="buckets.length === 0" icon="i-lucide-database" title="No buckets yet">
      <template #action>
        <PrtBtn seed="7" size="sm" @click="showCreate = true">Create your first bucket</PrtBtn>
      </template>
    </PrtEmptyState>

    <template v-else>
      <BucketToolbar
        v-model:query="query"
        :select-mode="selectMode"
        :selected-count="selection.count.value"
        :has-groups="hasGroups"
        :grouped="grouped"
        @enter-select="enterSelect"
        @exit-select="exitSelect"
        @toggle-grouped="groupBy = !groupBy"
      />

      <BucketBulkBar
        v-if="selectMode"
        v-model:group="bulkGroup"
        :selected-count="selection.count.value"
        :busy="bulkBusy"
        @assign-group="assignGroup"
        @ungroup="ungroupSelected"
        @set-public="bulkSetPublic"
        @remove="bulkDelete"
        @done="exitSelect"
      />

      <PrtEmptyState
        v-if="filteredBuckets.length === 0"
        icon="i-lucide-search"
        :title='`No buckets match "${query}"`'
      />

      <BucketTable
        v-else
        :rows="displayRows"
        :select-mode="selectMode"
        :grouped="grouped"
        :all-visible-selected="allVisibleSelected"
        :is-selected="selection.has"
        :deleting="deleting"
        :toggling="toggling"
        @open="openBucket"
        @toggle-select="selection.toggle"
        @toggle-select-all="selection.toggleAll(visibleNames)"
        @toggle-group="toggleGroup"
        @toggle-pin="togglePin"
        @toggle-public="handleTogglePublic"
        @remove="handleDelete"
      >
        <template #footer>
          <BucketTableFooter
            v-model:page="page"
            :grouped="grouped"
            :total="buckets.length"
            :filtered-total="filteredBuckets.length"
            :range-start="rangeStart"
            :range-end="rangeEnd"
            :total-pages="totalPages"
            :group-count="groupCount"
            :any-group-open="anyGroupOpen"
            @collapse-all="collapseAll"
            @expand-all="expandAll"
          />
        </template>
      </BucketTable>
    </template>

    <BucketCreateDialog v-if="showCreate" @close="showCreate = false" @created="showCreate = false; bucketsStore.load()" />
  </div>
</template>
