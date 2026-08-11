import { ofetch } from 'ofetch'
import type { FetchOptions } from 'ofetch'
import type { z } from 'zod'
import { readToken } from '@/lib/token'

export class ApiError extends Error {
  constructor(public status: number, detail: string) {
    super(detail)
    this.name = 'ApiError'
  }
}

// stores/auth.ts registers itself here rather than being imported: the store
// imports the transport, so the transport cannot import the store.
let unauthorizedHandler: (() => void) | null = null

export function onUnauthorized(handler: () => void): void {
  unauthorizedHandler = handler
}

// Uploads run on XHR for progress events, so they cannot reach the interceptor.
export function notifyUnauthorized(): void {
  unauthorizedHandler?.()
}

// Backend speaks two error dialects: huma {detail,title} and chi {error}. XHR
// hands this a raw string where ofetch has already parsed the JSON, hence the
// string branch.
export function errorDetail(body: unknown, fallback: string): string {
  if (typeof body === 'string') {
    try {
      return errorDetail(JSON.parse(body), body || fallback)
    } catch {
      return body || fallback
    }
  }
  if (body && typeof body === 'object') {
    const b = body as Record<string, unknown>
    const found = b.detail ?? b.error ?? b.title
    if (typeof found === 'string' && found !== '') return found
  }
  return fallback
}

const client = ofetch.create({
  baseURL: '/api',
  // A 500 on a delete should surface, not be tried three more times.
  retry: 0,
  onRequest({ options }) {
    options.headers.set('Authorization', `Bearer ${readToken()}`)
  },
  onResponseError({ response }) {
    if (response.status === 401) unauthorizedHandler?.()
    throw new ApiError(
      response.status,
      errorDetail(response._data, response.statusText || `HTTP ${response.status}`),
    )
  },
})

// `client` is not exported: routing every call through a schema is what stops a
// new endpoint handing back an `any` that spreads through three files.
export async function request<T>(
  path: string,
  schema: z.ZodType<T>,
  options?: FetchOptions<'json'>,
): Promise<T> {
  const data = await client<unknown>(path, options)
  return schema.parse(data)
}
