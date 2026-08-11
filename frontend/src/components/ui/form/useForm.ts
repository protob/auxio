import { computed, reactive, readonly, toRaw } from 'vue'
import type {
  PrtFormContext,
  PrtFormError,
  StandardSchemaV1,
} from '../_shared/form'
import { toDotPath } from '../_shared/form'

export interface UseFormOptions<S extends StandardSchemaV1> {
  // The caller's reactive object. Input-shaped. Never wrapped or mutated by the kit.
  state: object
  schema?: S
  // Typed submit handler — receives the PARSED OUTPUT, not the state.
  onSubmit?: (output: StandardSchemaV1.InferOutput<S>) => void | Promise<void>
  onError?: (errors: Readonly<Record<string, string>>) => void
  // Input revalidation debounce.
  debounceMs?: number
}

export function useForm<S extends StandardSchemaV1>(
  options: UseFormOptions<S>,
): PrtFormContext {
  const debounceMs = options.debounceMs ?? 300

  const errors = reactive<Record<string, string>>({})
  const blurred = reactive(new Set<string>())
  const meta = reactive({ submitCount: 0, validating: false, submitting: false })

  // Epoch guard: clear() and newer runs invalidate in-flight async validation
  // (the Nuxt UI clear/debounce race, #5700).
  let epoch = 0
  let debounceTimer: ReturnType<typeof setTimeout> | undefined

  async function run(): Promise<StandardSchemaV1.Result<unknown> | null> {
    if (!options.schema) return null
    const e = ++epoch
    meta.validating = true
    try {
      const result = await Promise.resolve(
        options.schema['~standard'].validate(toRaw(options.state)),
      )
      if (e !== epoch) return null // superseded — a newer run or clear() owns the truth
      // A run re-derives the whole map (server errors share it and clear here too).
      for (const key of Object.keys(errors)) delete errors[key]
      if (result.issues) {
        for (const issue of result.issues) {
          const name = issue.path ? toDotPath(issue.path) : ''
          if (!(name in errors)) errors[name] = issue.message // first message wins
        }
      }
      return result
    } finally {
      if (e === epoch) meta.validating = false
    }
  }

  function displayError(name: string, pattern?: RegExp): string | undefined {
    if (meta.submitCount === 0 && !blurred.has(name)) return undefined
    if (errors[name]) return errors[name]
    if (pattern) {
      for (const key of Object.keys(errors)) {
        if (pattern.test(key)) return errors[key]
      }
    }
    return undefined
  }

  function reportBlur(name: string) {
    blurred.add(name)
    void run()
  }

  function reportInput(_name: string) {
    if (!options.schema) return
    clearTimeout(debounceTimer)
    debounceTimer = setTimeout(() => void run(), debounceMs)
  }

  async function validate(): Promise<boolean> {
    if (!options.schema) return true
    const result = await run()
    return result !== null && !result.issues
  }

  async function submit() {
    if (meta.submitting) return // double-submit guard — never a disabled button
    meta.submitCount++
    let output: unknown
    if (options.schema) {
      const result = await run()
      if (result === null) return // superseded mid-flight
      if (result.issues) {
        options.onError?.(readonly(errors))
        return
      }
      output = result.value // parsed/transformed; state stays what was typed
    } else {
      output = toRaw(options.state) // schema-less form: pass-through
    }
    meta.submitting = true
    try {
      await options.onSubmit?.(output as StandardSchemaV1.InferOutput<S>)
    } finally {
      meta.submitting = false
    }
  }

  function setErrors(
    errs: PrtFormError[] | Record<string, string>,
    mode: 'merge' | 'replace' = 'merge',
  ) {
    if (mode === 'replace') for (const key of Object.keys(errors)) delete errors[key]
    if (Array.isArray(errs)) {
      for (const { name, message } of errs) errors[name] = message
    } else {
      Object.assign(errors, errs)
    }
  }

  function clearErrors(name?: string) {
    epoch++ // invalidate in-flight runs
    clearTimeout(debounceTimer)
    if (name) delete errors[name]
    else for (const key of Object.keys(errors)) delete errors[key]
  }

  function reset() {
    clearErrors()
    blurred.clear()
    meta.submitCount = 0
  }

  // the context is reactive() — computeds unwrap, so consumers read
  // form.errors / form.submitting with no .value anywhere.
  return reactive({
    errors: readonly(errors),
    submitCount: computed(() => meta.submitCount),
    validating: computed(() => meta.validating),
    submitting: computed(() => meta.submitting),
    displayError,
    reportBlur,
    reportInput,
    validate,
    submit,
    setErrors,
    clearErrors,
    reset,
  }) as PrtFormContext
}
