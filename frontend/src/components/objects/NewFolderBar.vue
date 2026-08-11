<script setup lang="ts">
import { ref } from 'vue'
import PrtBtn from '@/components/ui/btn/PrtBtn.vue'
import PrtInput from '@/components/ui/input/PrtInput.vue'

const emit = defineEmits<{ create: [name: string]; cancel: [] }>()

const name = ref('')

function submit() {
  const clean = name.value.trim().replace(/\/+$/, '')
  if (!clean) return
  name.value = ''
  emit('create', clean)
}
</script>

<template>
  <div class="flex items-center gap-2 mb-3 px-3.5 py-2.5 rounded-control bg-surface-1 border border-edge">
    <label class="text-sm text-ink-muted whitespace-nowrap" for="new-folder-name">New folder name:</label>
    <!-- autofocus rather than a mounted focus() call: the bar is v-if'd into
         existence by the toolbar, so the element is new on every open. -->
    <PrtInput
      id="new-folder-name"
      v-model="name"
      size="sm"
      placeholder="folder-name"
      class="flex-1 max-w-50"
      autofocus
      @keydown.enter="submit"
      @keydown.esc="emit('cancel')"
    />
    <PrtBtn seed="7" size="sm" :disabled="!name.trim()" @click="submit">Create</PrtBtn>
    <PrtBtn variant="outline" size="sm" @click="emit('cancel')">Cancel</PrtBtn>
  </div>
</template>
