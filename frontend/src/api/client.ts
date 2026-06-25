import type { ImportPreview } from '../types'

const csrfHeader = 'X-CSRF-Token'

export class ApiError extends Error {
  status: number
  payload: unknown

  constructor(status: number, payload: unknown) {
    super(`API request failed: ${status}`)
    this.status = status
    this.payload = payload
  }
}

export function getCSRFToken() {
  const match = document.cookie.match(/(?:^|;\s*)nsh_csrf=([^;]+)/)
  return match ? decodeURIComponent(match[1]) : ''
}

async function request<T>(url: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  if (!(init.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  const method = (init.method || 'GET').toUpperCase()
  if (!['GET', 'HEAD', 'OPTIONS'].includes(method)) {
    headers.set(csrfHeader, getCSRFToken())
  }
  const response = await fetch(url, { ...init, headers, credentials: 'include' })
  const text = await response.text()
  const payload = text ? JSON.parse(text) : null
  if (!response.ok) {
    throw new ApiError(response.status, payload)
  }
  return payload as T
}

export const api = {
  get: <T>(url: string) => request<T>(url),
  post: <T>(url: string, body?: unknown) => request<T>(url, { method: 'POST', body: body === undefined ? undefined : JSON.stringify(body) }),
  put: <T>(url: string, body?: unknown) => request<T>(url, { method: 'PUT', body: body === undefined ? undefined : JSON.stringify(body) }),
  del: <T>(url: string) => request<T>(url, { method: 'DELETE' }),
  uploadPreview: (file: File) => {
    const form = new FormData()
    form.append('file', file)
    return request<ImportPreview>('/api/battles/import/preview', { method: 'POST', body: form, headers: { [csrfHeader]: getCSRFToken() } })
  },
  uploadAvatar: (url: string, file: File) => {
    const form = new FormData()
    form.append('file', file)
    return request<{ asset_path: string }>(url, { method: 'PUT', body: form, headers: { [csrfHeader]: getCSRFToken() } })
  }
}
