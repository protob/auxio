<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useConfirmStore } from '@/stores/confirm'
import PrtDialog from '@/components/ui/dialog/PrtDialog.vue'
import PrtBtn from '@/components/ui/btn/PrtBtn.vue'

const confirm = useConfirmStore()
const { open, options } = storeToRefs(confirm)

// Esc and the scrim are a decline, same as Cancel. The dialog element is what
// emits false, so this is the only close path that needs settling by hand.
function onOpenChange(next: boolean) {
  if (!next) confirm.answer(false)
}
</script>

<template>
  <!-- v-if on options, not on open: the store keeps the last options after it
       answers, so the dialog stays mounted and closes through v-model, which is
       what lets the native close run. -->
  <PrtDialog
    v-if="options"
    :model-value="open"
    width="25rem"
    @update:model-value="onOpenChange"
  >
    <template #title>{{ options.title }}</template>
    {{ options.message }}
    <template #actions>
      <PrtBtn variant="outline" size="sm" @click="confirm.answer(false)">Cancel</PrtBtn>
      <PrtBtn :seed="options.danger ? '8' : '7'" size="sm" autofocus @click="confirm.answer(true)">
        {{ options.confirmLabel ?? 'Confirm' }}
      </PrtBtn>
    </template>
  </PrtDialog>
</template>
