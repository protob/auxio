<script setup lang="ts">
import { useLocalStorage } from '@vueuse/core'
import { storeToRefs } from 'pinia'
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import SearchInput from '@/components/base/SearchInput.vue'
import PrtModeToggle from '@/components/ui/modetoggle/PrtModeToggle.vue'
import PrtSidebar from '@/components/ui/sidebar/PrtSidebar.vue'
import type { PrtSidebarGroup, PrtSidebarItem } from '@/components/ui/sidebar/types'
import { sidebarItemVariants } from '@/components/ui/sidebar/variants'
import { SIDEBAR_RAILED_KEY } from '@/lib/constants'
import { formatBytes } from '@/lib/format'
import { fuzzyMatch } from '@/lib/fuzzy'
import { groupBuckets, groupKeyOf } from '@/lib/groupRows'
import { byPinnedThenName } from '@/lib/sort'
import { useAuthStore } from '@/stores/auth'
import { useBucketsStore } from '@/stores/buckets'
import { useStatsStore } from '@/stores/stats'
import type { Bucket } from '@/types'
import IconLogOut from '~icons/lucide/log-out'
import IconSettings from '~icons/lucide/settings'

const router = useRouter()
const route = useRoute()
const { logout } = useAuthStore()
const { buckets } = storeToRefs(useBucketsStore())
const { stats } = storeToRefs(useStatsStore())

const railed = useLocalStorage(SIDEBAR_RAILED_KEY, false)

const filter = ref('')

const filtered = computed(() => {
  const q = filter.value.trim()
  return q ? buckets.value.filter(b => fuzzyMatch(b.name, q)) : buckets.value
})

const leaf = (b: Bucket): PrtSidebarItem => ({
  value: `/${b.name}`,
  label: b.name,
  icon: 'i-lucide-database',
  badge: b.object_count > 0 ? String(b.object_count) : undefined,
})

const storageItems = computed<PrtSidebarItem[]>(() => {
  // "no matches" as a disabled row rather than a slot: the item shape is the
  // only way into the nav body short of replacing it wholesale.
  if (buckets.value.length > 0 && filtered.value.length === 0) {
    return [{ value: 'no-match', label: 'No buckets match.', disabled: true }]
  }
  if (!buckets.value.some(b => groupKeyOf(b) !== '')) {
    return [...filtered.value].sort(byPinnedThenName).map(leaf)
  }
  return groupBuckets(filtered.value, byPinnedThenName).map(g => ({
    value: `group:${g.key}`,
    label: g.label,
    icon: 'i-lucide-folder',
    children: g.buckets.map(leaf),
  }))
})

const nav = computed<PrtSidebarGroup[]>(() => [
  {
    items: [
      {
        value: '/',
        label: 'All Buckets',
        icon: 'i-lucide-database',
        badge: buckets.value.length ? String(buckets.value.length) : undefined,
      },
    ],
  },
  { label: 'Storage', items: storageItems.value },
])

// Item values are routes, so `active` is the route and @select is a push.
const active = computed(() => {
  const bucket = route.params.bucket
  if (typeof bucket === 'string') return `/${bucket}`
  return route.name === 'buckets' ? '/' : ''
})

function onSelect(value: string | number) {
  if (typeof value === 'string' && value.startsWith('/')) router.push(value)
}
</script>

<template>
  <PrtSidebar
    v-model="railed"
    :items="nav"
    :active="active"
    size="sm"
    label="Buckets"
    width="var(--sidebar-width)"
    @select="onSelect"
  >
    <template #header>
      <div class="flex-1 min-w-0 flex flex-col gap-2.5 py-3">
        <button
          type="button"
          class="flex items-center gap-2.5 bg-transparent cursor-pointer"
          aria-label="Go to buckets"
          @click="router.push('/')"
        >
          <span class="shrink-0 w-7 h-7 inline-flex items-center justify-center rounded-control bg-accent text-accent-ink text-sm font-bold">A</span>
          <span class="text-base font-semibold text-ink">Auxio</span>
        </button>

        <SearchInput v-model="filter" size="sm" label="Filter buckets" placeholder="Filter buckets…" />
      </div>
    </template>

    <template #footer>
      <div class="flex-1 min-w-0 flex flex-col gap-0.5 py-3">
        <div v-if="stats" class="mb-2 px-3 py-2.5 rounded-control bg-surface-1 border border-edge">
          <div class="flex items-baseline justify-between gap-2">
            <span class="text-xs font-medium text-ink-muted">Storage used</span>
            <span class="font-mono text-xs text-ink">{{ formatBytes(stats.total_size) }}</span>
          </div>
          <div class="mt-1 font-mono text-[0.6875rem] text-ink-faint">
            {{ stats.objects.toLocaleString() }} {{ stats.objects === 1 ? 'object' : 'objects' }}<template v-if="stats.cache_size > 0"> · {{ formatBytes(stats.cache_size) }} cache</template>
          </div>
        </div>

        <router-link
          to="/settings"
          :class="sidebarItemVariants({ size: 'sm', active: route.name === 'settings' })"
        >
          <IconSettings class="w-4 h-4 shrink-0" />
          Settings
        </router-link>

        <!-- Not sidebarItemVariants: its hover wash and the danger wash are both
             `hover:bg-*`, and which one wins is generated-CSS order, not class
             order. -->
        <button
          type="button"
          class="group/item w-full inline-flex items-center gap-2.5 h-8 px-2.5 rounded-control text-left text-xs text-ink-muted bg-transparent cursor-pointer outline-none prt-motion-colors hover:bg-[var(--danger-wash)] hover:text-danger focus-visible:ring-1 focus-visible:ring-accent"
          @click="logout"
        >
          <IconLogOut class="w-4 h-4 shrink-0" />
          Sign out
        </button>

        <div class="mt-1 flex items-center justify-between h-8 px-2.5">
          <span class="text-xs text-ink-muted">Theme</span>
          <PrtModeToggle />
        </div>
      </div>
    </template>
  </PrtSidebar>
</template>
