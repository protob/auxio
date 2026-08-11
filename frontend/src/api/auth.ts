import { ofetch } from 'ofetch'
import { LoginResponse } from './schemas'

// Login is the one call that runs without a token, so it does not go through
// api/client.ts - it has nothing to attach and no 401 handler to trigger.
export async function login(username: string, password: string): Promise<string> {
  // restarts drop every session, so server-down is common - it gets its own
  // message rather than a blanket "invalid credentials"
  const data = await ofetch<unknown>('/api/login', {
    method: 'POST',
    retry: 0,
    body: { username, password },
  }).catch((err: unknown) => {
    const status = (err as { status?: number }).status
    throw new Error(status ? 'Invalid credentials' : 'Cannot reach the server')
  })
  return LoginResponse.parse(data).token
}
