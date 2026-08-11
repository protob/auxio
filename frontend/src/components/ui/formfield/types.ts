export interface PrtFormFieldProps {
  modelValue?: string | number
  label?: string
  description?: string
  helperText?: string
  errorMessage?: string
  type?: string
  id?: string
  // Connects the field to a surrounding PrtForm: error lookup by dot-path,
  // name/data-field on the input, deterministic id. Without it (or without
  // a PrtForm ancestor) the field renders unwired: no error lookup, no id from name.
  name?: string
  // Match indexed/array error paths, e.g. /^tags\.\d+$/
  errorPattern?: RegExp
  placeholder?: string
  required?: boolean
  error?: boolean
  disabled?: boolean
  // layout-only additions (margin/width) — appearance goes through variants
  class?: string
}
