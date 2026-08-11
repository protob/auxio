<!--
  PrtPagination
  Windowed page navigation (boundary pages + ellipsis + sibling window —
  paginate.ts, pure TS). Never renders every page; the window size stays
  constant while paging, so the row doesn't jitter. Cells are their own
  quiet vocabulary (not PrtBtn); current page = aria-current + accent.
  Arrows are icon-only buttons (bg-transparent — the tailwind-compat
  reset rides the variant). All aria strings are props with English
  defaults.
  Slot contract: none.
-->
<template>
  <nav :aria-label="label" :class="cx('inline-flex items-center gap-1', props.class)">
    <button
      v-if="showEdges"
      type="button"
      :aria-label="firstLabel"
      :disabled="disabled || page <= 1"
      :class="paginationCellVariants({ size, edges, active: false, disabled: disabled || page <= 1 })"
      @click="go(1)"
    >
      <span class="i-lucide-chevrons-left" aria-hidden="true" />
    </button>
    <button
      v-if="showArrows"
      type="button"
      :aria-label="prevLabel"
      :disabled="disabled || page <= 1"
      :class="paginationCellVariants({ size, edges, active: false, disabled: disabled || page <= 1 })"
      @click="go(page - 1)"
    >
      <span class="i-lucide-chevron-left" aria-hidden="true" />
    </button>

    <template v-for="cell in cells" :key="cell">
      <span
        v-if="typeof cell === 'string'"
        :class="paginationGapVariants({ size })"
        aria-hidden="true"
      >…</span>
      <button
        v-else
        type="button"
        :aria-label="`${pageLabel} ${cell}`"
        :aria-current="cell === page ? 'page' : undefined"
        :disabled="disabled"
        :class="paginationCellVariants({ size, edges, active: cell === page, disabled })"
        @click="go(cell)"
      >{{ cell }}</button>
    </template>

    <button
      v-if="showArrows"
      type="button"
      :aria-label="nextLabel"
      :disabled="disabled || page >= totalPages"
      :class="paginationCellVariants({ size, edges, active: false, disabled: disabled || page >= totalPages })"
      @click="go(page + 1)"
    >
      <span class="i-lucide-chevron-right" aria-hidden="true" />
    </button>
    <button
      v-if="showEdges"
      type="button"
      :aria-label="lastLabel"
      :disabled="disabled || page >= totalPages"
      :class="paginationCellVariants({ size, edges, active: false, disabled: disabled || page >= totalPages })"
      @click="go(totalPages)"
    >
      <span class="i-lucide-chevrons-right" aria-hidden="true" />
    </button>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { cx } from '../_shared/cx'
import { paginate } from './paginate'
import type { PrtPaginationProps } from './types'
import { paginationCellVariants, paginationGapVariants } from './variants'

const props = withDefaults(defineProps<PrtPaginationProps>(), {
  siblingCount: 1,
  boundaryCount: 1,
  showArrows: true,
  showEdges: false,
  size: 'base',
  edges: 'rounded',
  disabled: false,
  class: '',
  label: 'Pagination',
  prevLabel: 'Previous page',
  nextLabel: 'Next page',
  firstLabel: 'First page',
  lastLabel: 'Last page',
  pageLabel: 'Page',
})

const emit = defineEmits<{
  'update:page': [page: number]
}>()

const cells = computed(() =>
  paginate(props.page, props.totalPages, props.siblingCount, props.boundaryCount),
)

function go(target: number) {
  if (props.disabled) return
  const next = Math.min(Math.max(target, 1), props.totalPages)
  if (next !== props.page) emit('update:page', next)
}
</script>
