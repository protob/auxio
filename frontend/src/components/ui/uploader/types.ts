import type { PrtEdges, PrtSize } from '../_shared/types'

// dashed is the classic dropzone affordance (reads blueprint-technical
// in a 1px hairline); outline is the quiet option
export type PrtUploaderVariant = 'dashed' | 'outline'

// reason CODES — the consumer renders any message
export type PrtUploaderRejectReason = 'type' | 'size' | 'count'

export interface PrtUploaderRejection {
  file: File
  reason: PrtUploaderRejectReason
}

// The kit's wrapper around a picked File. The raw `File` rides inside
// because that is the upload payload (fetch/FormData) — the wrapper
// exists only for stable keys, display fields and the preview URL.
// The component does NO transport; its contract ends at "here are the
// Files the user chose".
export interface PrtUploaderFile {
  // internal — stable key for list rendering/removal
  id: string
  // the payload — the consumer uploads this
  file: File
  // webkitRelativePath when present (directory mode), else file.name
  name: string
  size: number
  type: string
  // object URL for image/* (revoked automatically when the entry leaves)
  previewUrl?: string
}

export interface PrtUploaderProps {
  // controlled — new arrays out, never mutation
  modelValue?: PrtUploaderFile[]
  // passes through to the input AND drives per-file validation
  accept?: string
  multiple?: boolean
  // in multiple mode; commits past it reject with 'count'
  maxFiles?: number
  // bytes; larger files reject with 'size'
  maxSize?: number
  // webkitdirectory mode — implies multiple; the picker may ignore
  // `accept` on some platforms, validation still runs per file
  directory?: boolean
  // paste-to-upload while the dropzone has focus
  paste?: boolean
  // image thumbnails (object URLs) in the list
  preview?: boolean
  // false = consumer renders its own list from modelValue
  showList?: boolean
  id?: string
  variant?: PrtUploaderVariant
  size?: PrtSize
  // rounded/pill both map to the surface radius — a pill dropzone is
  // nonsense (the textarea precedent)
  edges?: PrtEdges
  error?: boolean
  disabled?: boolean
  // layout-only additions (margin/width) — appearance goes through variants
  class?: string
  // String props with English defaults. The headline composes as
  // "<buttonLabel> <titleText>" ("Click to upload" + "or drag and drop");
  // only the button part is interactive.
  buttonLabel?: string
  titleText?: string
  // accept/size summary; default derives from accept + maxSize
  hintText?: string
  // aria-label of each row's × button
  removeLabel?: string
}
