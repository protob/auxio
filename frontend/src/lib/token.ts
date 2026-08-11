import { TOKEN_KEY } from './constants'

// The only module that touches the session token. Both api/client.ts and
// stores/auth.ts read through here, which is what keeps the transport from
// importing the store.

export function readToken(): string {
  return sessionStorage.getItem(TOKEN_KEY) ?? ''
}

export function writeToken(token: string): void {
  sessionStorage.setItem(TOKEN_KEY, token)
}

export function clearToken(): void {
  sessionStorage.removeItem(TOKEN_KEY)
}
