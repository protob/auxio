<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useBucketsStore } from '@/stores/buckets'
import { errorMessage } from '@/lib/errors'
import { BucketCreate } from '@/api/schemas'
import PrtBtn from '@/components/ui/btn/PrtBtn.vue'
import PrtCheckbox from '@/components/ui/checkbox/PrtCheckbox.vue'
import PrtCombobox from '@/components/ui/combobox/PrtCombobox.vue'
import PrtDialog from '@/components/ui/dialog/PrtDialog.vue'
import PrtForm from '@/components/ui/form/PrtForm.vue'
import PrtFormField from '@/components/ui/formfield/PrtFormField.vue'
import PrtInput from '@/components/ui/input/PrtInput.vue'
import PrtNotice from '@/components/ui/notice/PrtNotice.vue'
import { useForm } from '@/components/ui/form/useForm'

const emit = defineEmits<{ close: []; created: [] }>()

const buckets = useBucketsStore()
const { groupNames } = storeToRefs(buckets)
const groupOptions = computed(() => groupNames.value.map(g => ({ value: g, label: g })))

const state = reactive({ name: '', group: '', public: true })
const submitError = ref('')

const form = useForm({
  state,
  schema: BucketCreate,
  onSubmit: async (input) => {
    submitError.value = ''
    try {
      await buckets.create(input.name, input.public, input.group)
      emit('created')
    } catch (e) {
      submitError.value = errorMessage(e, 'Failed to create bucket')
    }
  },
})

// auto-suggest the group from the bucket-name prefix until the user types a group
let groupTouched = false
function onGroupInput(value: string | number | null) {
  groupTouched = true
  state.group = String(value ?? '')
}
watch(() => state.name, (n) => {
  if (groupTouched) return
  const m = n.match(/^([a-z0-9]+)[-_]/)
  state.group = m?.[1] ?? ''
})
</script>

<template>
  <!-- model-value is a constant true and the parent mounts this with v-if: one
       instance per open is what resets the form, and useModalDialog opens on
       mount when the model is already true. Every close path emits false. -->
  <PrtDialog :model-value="true" @update:model-value="emit('close')">
    <template #title>New Bucket</template>
    <!-- PrtDialog's body slot brings its own px-5 py-4; the form owns the rest. -->
    <PrtForm :form="form">
      <!-- name has to match a BucketCreate field - that's what the dot-path
           error lookup resolves against. -->
      <PrtFormField
        name="name"
        label="Bucket name"
        helper-text="Lowercase letters, numbers, and hyphens. 3–63 characters."
      >
        <template #default="{ inputId, error, inputProps }">
          <PrtInput
            v-bind="inputProps"
            :id="inputId"
            v-model="state.name"
            :error="error"
            placeholder="my-bucket"
            autofocus
          />
        </template>
      </PrtFormField>

      <PrtFormField name="group">
        <template #label>Group <span class="font-normal text-ink-faint">(optional)</span></template>
        <!-- freeText: the typed text is the value, so an existing group can be
             picked from the list and a new one just typed. -->
        <template #default="{ inputId, inputProps }">
          <PrtCombobox
            v-bind="inputProps"
            :id="inputId"
            :model-value="state.group"
            :options="groupOptions"
            free-text
            clearable
            placeholder="e.g. ecom"
            @update:model-value="onGroupInput"
          />
        </template>
      </PrtFormField>

      <PrtCheckbox
        v-model="state.public"
        :class="state.public
          ? 'w-full p-3.5 rounded-control border border-accent bg-[var(--accent-wash)] prt-motion-colors'
          : 'w-full p-3.5 rounded-control border border-edge bg-surface-2 hover:border-edge-strong prt-motion-colors'"
      >
        <span class="block text-sm font-semibold text-ink">Public access</span>
        <span class="block mt-0.5 text-xs text-ink-muted">Uncheck to make this bucket private</span>
      </PrtCheckbox>

      <PrtNotice v-if="submitError" intent="danger" variant="wash" role="alert" :description="submitError" />

      <div class="flex justify-end gap-2 pt-1">
        <PrtBtn variant="outline" type="button" @click="emit('close')">Cancel</PrtBtn>
        <!-- Not disabled on invalid: useForm guards double-submit itself and
             re-runs the schema on submit, so a click on a bad name reports the
             reason instead of doing nothing. -->
        <PrtBtn seed="7" type="submit" :loading="form.submitting">
          {{ form.submitting ? 'Creating…' : 'Create bucket' }}
        </PrtBtn>
      </div>
    </PrtForm>
  </PrtDialog>
</template>
