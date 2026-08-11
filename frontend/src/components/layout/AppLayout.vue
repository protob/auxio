<script setup lang="ts">
import { onMounted } from 'vue'
import PrtAppShell from '@/components/ui/appshell/PrtAppShell.vue'
import { useBucketsStore } from '@/stores/buckets'
import { useStatsStore } from '@/stores/stats'
import AppSidebar from './AppSidebar.vue'

const bucketsStore = useBucketsStore()
const statsStore = useStatsStore()

// The shell loads both because the sidebar shows them on every route; the views
// re-request their own data and the store dedupes nothing, which is fine at
// this size.
onMounted(() => {
  bucketsStore.load()
  statsStore.load()
})
</script>

<template>
  <PrtAppShell>
    <template #sidebar>
      <AppSidebar />
    </template>

    <!-- The page gutter is auxio's, not the shell's: PrtAppShell's <main> owns
         scrolling, and it is the scroll container the sticky bulk bar pins
         against, so the padding has to sit on a child of it. -->
    <div class="px-10 py-7.5">
      <router-view />
    </div>
  </PrtAppShell>
</template>

<style>
/* Unscoped, and the same selector the kit's dialogs already ship: their lock
   targets <html>, but PrtAppShell makes .prt-appshell-main the only scroller,
   so on this layout the kit's rule has nothing to hold still. Without this the
   listing scrolls under an open dialog. */
html:has(dialog:modal) .prt-appshell-main {
  overflow: hidden;
  scrollbar-gutter: stable;
}
</style>
