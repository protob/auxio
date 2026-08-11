<script setup lang="ts">
import PrtBtn from '@/components/ui/btn/PrtBtn.vue'
import PrtPagination from '@/components/ui/pagination/PrtPagination.vue'

defineProps<{
  grouped: boolean
  total: number
  filteredTotal: number
  rangeStart: number
  rangeEnd: number
  totalPages: number
  groupCount: number
  anyGroupOpen: boolean
}>()

const page = defineModel<number>('page', { required: true })
const emit = defineEmits<{ collapseAll: []; expandAll: [] }>()
</script>

<template>
  <!-- The card always ends in this band, in both views - only its content
       adapts, so toggling views never shifts the layout. min-h keeps the band
       the same height whether the right slot holds the pager, a button, or
       nothing; no border-top because the last row already draws the hairline. -->
  <div class="flex items-center justify-between gap-3 min-h-12.5 px-5 py-2">
    <template v-if="!grouped">
      <span class="font-mono text-xs text-ink-muted">
        {{ rangeStart }}–{{ rangeEnd }} of {{ filteredTotal }}<template v-if="filteredTotal !== total"> (filtered from {{ total }})</template>
      </span>
      <PrtPagination v-if="totalPages > 1" v-model:page="page" :total-pages="totalPages" size="sm" />
    </template>
    <template v-else>
      <span class="font-mono text-xs text-ink-muted">
        {{ groupCount }} {{ groupCount === 1 ? 'group' : 'groups' }} · {{ filteredTotal }}<template v-if="filteredTotal !== total"> of {{ total }}</template> buckets
      </span>
      <!-- min-w so swapping "Collapse all"/"Expand all" does not resize a
           control under the cursor. -->
      <PrtBtn variant="outline" size="sm" class="min-w-26" @click="anyGroupOpen ? emit('collapseAll') : emit('expandAll')">
        {{ anyGroupOpen ? 'Collapse all' : 'Expand all' }}
      </PrtBtn>
    </template>
  </div>
</template>
