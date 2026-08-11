// Every whitespace-separated term in `query` must appear in `text` as a
// case-insensitive subsequence: subsequence absorbs dropped characters
// ("ecm" → "ecom"), multi-term absorbs partial words ("saas av").
export function fuzzyMatch(text: string, query: string): boolean {
  const haystack = text.toLowerCase()
  const terms = query.toLowerCase().split(/\s+/).filter(Boolean)
  if (terms.length === 0) return true
  return terms.every(term => isSubsequence(term, haystack))
}

function isSubsequence(needle: string, hay: string): boolean {
  let i = 0
  for (let j = 0; j < hay.length && i < needle.length; j++) {
    if (hay[j] === needle[i]) i++
  }
  return i === needle.length
}
