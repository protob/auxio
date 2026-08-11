export { default as PrtForm } from './PrtForm.vue'
export { useForm } from './useForm'
export type { UseFormOptions } from './useForm'
export type { PrtFormProps } from './types'
export { formVariants } from './variants'
// re-export the contract for consumer convenience (single import site)
export { prtFormKey, toDotPath } from '../_shared/form'
export type {
  PrtFormContext,
  PrtFormError,
  StandardSchemaV1,
} from '../_shared/form'
