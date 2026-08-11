// no Vue imports in this file. Pure windowing math (MUI-style).

// the two gaps carry stable distinct keys — v-for churn re-mounts
// buttons mid-interaction otherwise (trap registry)
export type PrtPaginationCell = number | 'gap-start' | 'gap-end'

function range(start: number, end: number): number[] {
  if (end < start) return []
  return Array.from({ length: end - start + 1 }, (_, i) => start + i)
}

// Boundary pages + ellipsis + sibling window around the current page.
// Invariant: the output length is constant for a given
// (totalPages, siblingCount, boundaryCount) once totalPages is large
// enough — the layout never jitters while paging.
export function paginate(
  page: number,
  totalPages: number,
  siblingCount = 1,
  boundaryCount = 1,
): PrtPaginationCell[] {
  if (totalPages <= 0) return []

  const startPages = range(1, Math.min(boundaryCount, totalPages))
  const endPages = range(
    Math.max(totalPages - boundaryCount + 1, boundaryCount + 1),
    totalPages,
  )

  const siblingsStart = Math.max(
    Math.min(
      page - siblingCount,
      // keep the window size constant when page is near the end
      totalPages - boundaryCount - siblingCount * 2 - 1,
    ),
    boundaryCount + 2,
  )
  const firstEndPage = endPages[0]
  const siblingsEnd = Math.min(
    Math.max(
      page + siblingCount,
      // …and when page is near the start
      boundaryCount + siblingCount * 2 + 2,
    ),
    firstEndPage === undefined ? totalPages - 1 : firstEndPage - 2,
  )

  return [
    ...startPages,
    ...(siblingsStart > boundaryCount + 2
      ? (['gap-start'] as const)
      : boundaryCount + 1 < totalPages - boundaryCount
        ? [boundaryCount + 1]
        : []),
    ...range(siblingsStart, siblingsEnd),
    ...(siblingsEnd < totalPages - boundaryCount - 1
      ? (['gap-end'] as const)
      : totalPages - boundaryCount > boundaryCount
        ? [totalPages - boundaryCount]
        : []),
    ...endPages,
  ]
}
