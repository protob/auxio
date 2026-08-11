<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import LoginForm from '@/components/auth/LoginForm.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import ConfirmHost from '@/components/base/ConfirmHost.vue'
import PrtToaster from '@/components/ui/toast/PrtToaster.vue'

const { isAuthenticated } = storeToRefs(useAuthStore())
</script>

<template>
  <LoginForm v-if="!isAuthenticated" />
  <AppLayout v-else />
  <ConfirmHost />
  <!-- The kit's host declares no live region and $attrs land on it, so
       aria-live and role are set here: a background failure is read out
       without stealing focus. -->
  <PrtToaster position="bottom-right" role="status" aria-live="polite" />
</template>
