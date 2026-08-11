<script setup lang="ts">
import { computed, ref, useTemplateRef } from 'vue'
import { useIntersectionObserver } from '@vueuse/core'
import { storeToRefs } from 'pinia'
import { useBucketsStore } from '@/stores/buckets'
import PrtBtn from '@/components/ui/btn/PrtBtn.vue'
import PrtCombobox from '@/components/ui/combobox/PrtCombobox.vue'
import IconX from '~icons/lucide/x'

defineProps<{ selectedCount: number; busy: boolean }>()
const group = defineModel<string>('group', { required: true })
const emit = defineEmits<{
  assignGroup: []
  ungroup: []
  setPublic: [value: boolean]
  remove: []
  done: []
}>()

const { groupNames } = storeToRefs(useBucketsStore())
const groupOptions = computed(() => groupNames.value.map(g => ({ value: g, label: g })))

// The combobox owns Enter while its panel is open - it picks the active option -
// so only a closed panel means "assign what is typed".
function onGroupEnter(event: KeyboardEvent) {
  if ((event.target as HTMLElement).getAttribute('aria-expanded') === 'true') return
  emit('assignGroup')
}

// Firefox has no `scroll-state` container query, so "is the bar pinned?" comes
// from an IntersectionObserver on a zero-height sentinel at the bar's resting
// top edge - out of .prt-appshell-main means pinned.
const stuck = ref(false)
const sentinel = useTemplateRef<HTMLElement>('sentinel')
const scrollRoot = computed(() => sentinel.value?.closest<HTMLElement>('.prt-appshell-main') ?? undefined)

// grow the root's top edge by 30px (= the page gutter's padding-top) so the flag
// flips exactly when the bar pins flush to the top, matching its `top: -30px`.
useIntersectionObserver(
  sentinel,
  ([entry]) => { if (entry) stuck.value = !entry.isIntersecting },
  { root: scrollRoot, rootMargin: '30px 0px 0px 0px', threshold: 0 },
)
</script>

<template>
  <!-- probe: when this scrolls out the top of the shell's main, the bar is stuck -->
  <div ref="sentinel" class="bulk-sentinel" aria-hidden="true"></div>
  <div class="bulk-bar-sticky">
    <div class="bulk-bar flex items-center gap-2.5" :class="{ 'is-stuck': stuck }">
      <span class="text-xs font-semibold text-ink-muted whitespace-nowrap">With selected:</span>
      <!-- The width is on the wrapper: the combobox root is w-full, and a w-55
           passed through `class` would race it rather than replace it. -->
      <div class="w-55">
        <PrtCombobox
          :model-value="group"
          :options="groupOptions"
          free-text
          clearable
          size="sm"
          placeholder="type a new or existing group…"
          aria-label="Group name"
          @update:model-value="group = String($event ?? '')"
          @keydown.enter="onGroupEnter"
        />
      </div>
      <PrtBtn seed="7" size="sm" :disabled="busy || !selectedCount || !group.trim()" @click="emit('assignGroup')">
        Assign group
      </PrtBtn>
      <PrtBtn variant="outline" size="sm" :disabled="busy || !selectedCount" @click="emit('ungroup')">Ungroup</PrtBtn>
      <span class="w-px h-5 mx-0.5 bg-edge"></span>
      <PrtBtn variant="outline" size="sm" :disabled="busy || !selectedCount" @click="emit('setPublic', true)">Make public</PrtBtn>
      <PrtBtn variant="outline" size="sm" :disabled="busy || !selectedCount" @click="emit('setPublic', false)">Make private</PrtBtn>
      <span class="w-px h-5 mx-0.5 bg-edge"></span>
      <PrtBtn variant="outline" seed="8" size="sm" :disabled="busy || !selectedCount" @click="emit('remove')">Delete</PrtBtn>
      <!-- exit affordance for the pinned bar (the toolbar's Done has scrolled away) -->
      <PrtBtn v-if="stuck" variant="outline" size="sm" class="ml-auto" aria-label="Done" @click="emit('done')">
        <IconX class="w-4 h-4" />
      </PrtBtn>
    </div>
  </div>
</template>

<style scoped>
/* Two states, one mechanism: at rest an inline rounded panel matching the card;
   pinned, the sticky wrapper's -40px/-30px cancel the shell's page gutter
   (AppLayout's px-10 py-7.5) and .is-stuck restyles it as a full-bleed docked
   bar. */
.bulk-sentinel { height: 1px; margin-bottom: -1px; }

.bulk-bar-sticky {
  position: sticky;
  top: -30px;
  z-index: 20;
  margin: 0 -40px 14px;
}
.bulk-bar {
  margin: 0 40px;
  padding: 11px 20px;
  background: var(--surface-2);
  border: 1px solid var(--edge);
  border-radius: var(--radius-surface);
  transition: box-shadow 0.15s ease;
}

.bulk-bar.is-stuck {
  margin: 0;
  padding-left: 40px;
  padding-right: 40px;
  border-color: transparent;
  border-bottom-color: var(--edge-strong);
  border-radius: 0;
  box-shadow: 0 6px 16px oklch(0% 0 0 / 0.22);
}
</style>
