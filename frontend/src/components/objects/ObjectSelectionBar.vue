<script setup lang="ts">
import PrtBtn from '@/components/ui/btn/PrtBtn.vue'
import IconTrash2 from '~icons/lucide/trash-2'

defineProps<{ count: number; busy: boolean }>()
const emit = defineEmits<{ clear: []; remove: [] }>()
</script>

<template>
  <div class="flex items-center gap-2.5 mb-3 px-3.5 py-2.5 rounded-control bg-[var(--accent-wash)] border border-accent">
    <span class="mr-auto text-sm font-semibold text-accent">{{ count }} selected</span>
    <PrtBtn variant="outline" size="sm" @click="emit('clear')">Clear</PrtBtn>
    <PrtBtn variant="outline" seed="8" size="sm" :disabled="busy" @click="emit('remove')">
      <IconTrash2 class="w-3.5 h-3.5" />
      {{ busy ? 'Deleting…' : 'Delete selected' }}
    </PrtBtn>
  </div>
</template>
