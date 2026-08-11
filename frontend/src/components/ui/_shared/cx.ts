import { type ClassValue, clsx } from 'clsx'

// Deliberately NO tailwind-merge: the `class` prop is layout-only (margin/width/placement);
// appearance goes through variants. See CONVENTIONS.md.
export function cx(...inputs: ClassValue[]) {
  return clsx(inputs)
}
