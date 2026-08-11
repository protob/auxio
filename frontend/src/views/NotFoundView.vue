<script setup lang="ts">
import PrtBtn from '@/components/ui/btn/PrtBtn.vue'
import PrtEmptyState from '@/components/ui/emptystate/PrtEmptyState.vue'
</script>

<template>
  <!-- size, not a padding through `class`: the variant already sets px/py and a
       second one would race it on emission order. -->
  <PrtEmptyState icon="i-lucide-triangle-alert" title="That page does not exist." size="lg">
    <template #action>
      <!-- tag resolves through the globally registered router-link, and `to`
           falls through to it. -->
      <PrtBtn tag="router-link" to="/" seed="7" size="sm">Back to buckets</PrtBtn>
    </template>
  </PrtEmptyState>
</template>
