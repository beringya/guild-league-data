export type ApiError = Error & { status?: number };

const API_BASE = "/api";

export function csrfToken(): string {
  const cookie = document.cookie.split("; ").find((item) => item.startsWith("nsh_csrf="));
  return cookie ? decodeURIComponent(cookie.split("=")[1]) : "";
}

export async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers);
  if (!(options.body instanceof FormData) && options.body !== undefined) {
    headers.set("Content-Type", "application/json");
  }
  const method = (options.method || "GET").toUpperCase();
  if (!["GET", "HEAD", "OPTIONS"].includes(method)) {
    headers.set("X-CSRF-Token", csrfToken());
  }
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
    credentials: "include"
  });
  if (!response.ok) {
    const payload = await response.json().catch(() => ({}));
    const error = new Error(payload.detail || "请求失败") as ApiError;
    error.status = response.status;
    throw error;
  }
  return response.json() as Promise<T>;
}

export function postJson<T>(path: string, body: unknown): Promise<T> {
  return api<T>(path, { method: "POST", body: JSON.stringify(body) });
}

export function putJson<T>(path: string, body: unknown): Promise<T> {
  return api<T>(path, { method: "PUT", body: JSON.stringify(body) });
}

export function upload<T>(path: string, data: FormData): Promise<T> {
  return api<T>(path, { method: "POST", body: data });
}
