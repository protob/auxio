<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { errorMessage } from '@/lib/errors'
import PrtBtn from '@/components/ui/btn/PrtBtn.vue'
import PrtFormField from '@/components/ui/formfield/PrtFormField.vue'
import PrtInput from '@/components/ui/input/PrtInput.vue'
import PrtNotice from '@/components/ui/notice/PrtNotice.vue'

const { login } = useAuthStore()

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function handleSubmit() {
  if (!username.value || !password.value) {
    error.value = 'Please enter your username and password'
    return
  }
  loading.value = true
  error.value = ''
  try {
    await login(username.value, password.value)
  } catch (e) {
    error.value = errorMessage(e, 'Sign-in failed')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-10 bg-surface-0">
    <div class="w-90 max-w-full">
      <div class="flex items-center gap-2.75 mb-10">
        <span class="w-8 h-8 inline-flex items-center justify-center rounded-lg bg-accent text-accent-ink text-base font-bold">A</span>
        <span class="text-lg font-semibold text-ink tracking-[-0.0225rem]">Auxio</span>
      </div>
      <h1 class="mb-8 text-[1.625rem] font-semibold text-ink tracking-[-0.05rem]">Sign in</h1>

      <!-- No PrtForm: the only rule is "both fields present" and the message
           that matters comes back from the server, so there is no schema for a
           form context to validate against. The fields take the slot because
           PrtFormField's own $attrs land on its wrapper div, not on the input,
           and autocomplete/autofocus have to reach the input. -->
      <form class="flex flex-col gap-4" @submit.prevent="handleSubmit">
        <PrtFormField label="Username">
          <template #default="{ inputId }">
            <PrtInput
              :id="inputId"
              v-model="username"
              placeholder="admin"
              autocomplete="username"
              autofocus
            />
          </template>
        </PrtFormField>

        <PrtFormField label="Password">
          <template #default="{ inputId }">
            <PrtInput
              :id="inputId"
              v-model="password"
              type="password"
              placeholder="••••••••"
              autocomplete="current-password"
            />
          </template>
        </PrtFormField>

        <PrtNotice v-if="error" intent="danger" variant="wash" role="alert" :description="error" />

        <!-- type wins over the kit's own `type="button"` on fallthrough, which
             is what keeps Enter in either field submitting the form. -->
        <PrtBtn seed="7" size="lg" full-width type="submit" class="mt-2" :loading="loading">
          {{ loading ? 'Signing in…' : 'Sign in' }}
        </PrtBtn>
      </form>
    </div>
  </div>
</template>
