import type { PrtFormContext } from '../_shared/form'

export interface PrtFormProps {
  // The context returned by useForm() — composable-first.
  form: PrtFormContext
  // layout-only additions (margin/width) — appearance goes through variants
  class?: string
}
