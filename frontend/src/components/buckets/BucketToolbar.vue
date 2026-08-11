<script setup lang="ts">
import SearchInput from '@/components/base/SearchInput.vue'
import PrtBtn from '@/components/ui/btn/PrtBtn.vue'

defineProps<{
  selectMode: boolean
  selectedCount: number
  hasGroups: boolean
  grouped: boolean
}>()

const query = defineModel<string>('query', { required: true })
const emit = defineEmits<{ enterSelect: []; exitSelect: []; toggleGrouped: [] }>()
</script>

<template>
  <div class="flex items-center justify-between gap-4 mb-3.5">
    <SearchInput
      v-model="query"
      label="Filter buckets"
      placeholder="Filter buckets…"
      class="flex-1 max-w-80"
    />
    <div class="flex items-center gap-4">
      <template v-if="!selectMode">
        <PrtBtn variant="outline" size="sm" @click="emit('enterSelect')">Select</PrtBtn>
        <!-- labeled by the view you get - "Ungroup" read like a destructive edit
             (Photoshop-style) when it only changes the view.
             The fixed width stops the label swap resizing a control under the
             cursor. -->
        <PrtBtn
          v-if="hasGroups"
          variant="outline"
          size="sm"
          class="min-w-28"
          title="Switch view only — bucket groups are not changed"
          @click="emit('toggleGrouped')"
        >
          {{ grouped ? 'Flat view' : 'Grouped view' }}
        </PrtBtn>
      </template>
      <template v-else>
        <span class="font-mono text-xs text-ink-muted whitespace-nowrap">{{ selectedCount }} selected</span>
        <PrtBtn variant="outline" size="sm" @click="emit('exitSelect')">Done</PrtBtn>
      </template>
    </div>
  </div>
</template>
