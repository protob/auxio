// Anything is throwable in JS; collapse unknown -> human string in one place.
// ApiError carries the server's detail as its message, so it needs no branch.
export function errorMessage(e: unknown, fallback = 'Something went wrong'): string {
  if (e instanceof Error && e.message) return e.message
  return fallback
}
