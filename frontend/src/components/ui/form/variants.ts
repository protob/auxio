import { cva } from 'class-variance-authority'

// a form is the one place a layout default earns its keep.
export const formVariants = cva('flex flex-col gap-4')
