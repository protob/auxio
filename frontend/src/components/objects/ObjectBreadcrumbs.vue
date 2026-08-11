<script setup lang="ts">
import { computed } from 'vue'
import { breadcrumbs } from '@/lib/objectPath'
import IconClipboard from '~icons/lucide/clipboard'

const props = defineProps<{ bucket: string; prefix: string }>()
const emit = defineEmits<{ navigate: [path: string]; copyPath: [] }>()

const crumbs = computed(() => breadcrumbs(props.bucket, props.prefix))
</script>

<template>
  <!-- Not PrtBreadcrumb: it links via real hrefs and emits nothing, but the
       prefix is view state, not a route. Every crumb keeps a trailing slash,
       even the last, because the trail spells out an S3 prefix - generated so
       the slash isn't selectable or read out as a crumb. -->
  <nav class="flex items-center gap-0.5" aria-label="Breadcrumb">
    <button
      v-for="(crumb, i) in crumbs"
      :key="crumb.path"
      type="button"
      class="px-1.5 py-1 rounded-control text-[0.9375rem] font-semibold bg-transparent after:content-['/'] after:ml-1.5 after:font-normal after:text-ink-faint"
      :class="i === crumbs.length - 1
        ? 'text-ink cursor-default'
        : 'text-ink-muted cursor-pointer hover:text-ink hover:bg-surface-2'"
      :aria-current="i === crumbs.length - 1 ? 'page' : undefined"
      @click="emit('navigate', crumb.path)"
    >{{ crumb.label }}</button>
    <!-- "Prefix", not "path": this copies /bucket/prefix/, while the row action
         copies an absolute URL. -->
    <button
      type="button"
      class="ml-1 w-6 h-6 inline-flex items-center justify-center rounded-control bg-transparent cursor-pointer text-ink-faint hover:bg-surface-2 hover:text-ink-muted"
      title="Copy prefix"
      aria-label="Copy prefix"
      @click="emit('copyPath')"
    >
      <IconClipboard class="w-3.5 h-3.5" />
    </button>
  </nav>
</template>
