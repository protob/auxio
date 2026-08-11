import type { PrtEdges, PrtOption, PrtSize } from '../_shared/types'

export type PrtComboboxVariant = 'filled' | 'outline' | 'minimal'

export interface PrtComboboxProps {
  // typed values like select: numbers stay numbers; null = empty.
  // freeText mode: the text IS the value (string).
  modelValue?: string | number | null
  options: PrtOption[]
  // false = consumer filters (the async case); fn = custom predicate;
  // default = case-insensitive contains on the label
  filter?: false | ((query: string, option: PrtOption) => boolean)
  // free-text mode: modelValue is the raw text, emitted on every input;
  // selecting an option just fills it. Default false = strict: modelValue
  // changes only on selection. This flag IS the autocomplete/combobox merge.
  freeText?: boolean
  clearable?: boolean
  // loading row in the panel + spinner replacing the chevron
  loading?: boolean
  // rendering cap for huge sets — a capped tail row invites more typing
  maxVisible?: number
  placeholder?: string
  id?: string
  variant?: PrtComboboxVariant
  size?: PrtSize
  edges?: PrtEdges
  error?: boolean
  disabled?: boolean
  // layout-only additions (margin/width) — appearance goes through variants
  class?: string
  // D13 — every user-facing string is a prop with an English default
  // shown when the filter leaves nothing
  emptyText?: string
  // capped tail row; `{n}` interpolates the hidden count
  moreText?: string
  // shown while `loading`
  loadingText?: string
  // aria-label of the clear affordance
  clearLabel?: string
}
