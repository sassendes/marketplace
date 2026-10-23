import type { AuthResponse } from "./types";

const REFRESH_KEY = "mp.refresh";

let accessToken: string | null = null;
let refreshing: Promise<boolean> | null = null;
let onSession: (auth: AuthResponse | null) => void = () => {};

export class ApiError extends Error {
  status: number;
  body: unknown;
  constructor(status: number, message: string, body: unknown) {
    super(message);
    this.status = status;
    this.body = body;
  }
}

function readRefresh(): string | null {
  try {
    return localStorage.getItem(REFRESH_KEY);
  } catch {
    return null;
  }
}

function writeRefresh(token: string | null) {
  try {
    if (token) localStorage.setItem(REFRESH_KEY, token);
    else localStorage.removeItem(REFRESH_KEY);
  } catch {
    return;
  }
}

export function setSessionListener(fn: (auth: AuthResponse | null) => void) {
  onSession = fn;
}

export function applySession(auth: AuthResponse | null) {
  accessToken = auth?.access_token ?? null;
  writeRefresh(auth?.refresh_token ?? null);
  onSession(auth);
}

export function hasStoredSession(): boolean {
  return readRefresh() !== null;
}

export async function refreshSession(): Promise<boolean> {
  const token = readRefresh();
  if (!token) return false;
  if (!refreshing) {
    refreshing = (async () => {
      try {
        const res = await fetch("/api/auth/refresh", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ refresh_token: token }),
        });
        if (!res.ok) {
          if (res.status === 401) applySession(null);
          return false;
        }
        applySession((await res.json()) as AuthResponse);
        return true;
      } catch {
        return false;
      } finally {
        refreshing = null;
      }
    })();
  }
  return refreshing;
}

export async function api<T>(path: string, init: RequestInit & { json?: unknown } = {}, retry = true): Promise<T> {
  const headers = new Headers(init.headers);
  let body = init.body;
  if (init.json !== undefined) {
    headers.set("Content-Type", "application/json");
    body = JSON.stringify(init.json);
  }
  if (accessToken) headers.set("Authorization", `Bearer ${accessToken}`);
  const res = await fetch(`/api${path}`, { ...init, headers, body });
  if (res.status === 401 && retry && readRefresh() && !path.startsWith("/auth/")) {
    if (await refreshSession()) return api<T>(path, init, false);
  }
  const text = await res.text();
  let data: unknown = null;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = text;
    }
  }
  if (!res.ok) {
    const msg =
      data && typeof data === "object" && "error" in data ? String((data as { error: string }).error) : `Request failed (${res.status})`;
    throw new ApiError(res.status, msg, data);
  }
  return data as T;
}

export function errorMessage(err: unknown): string {
  if (err instanceof ApiError) {
    if (err.status === 429) return "Too many attempts. Wait a moment and try again.";
    if (err.status >= 500) return "Something went wrong on our side. Your data is safe, try again in a moment.";
    return err.message.charAt(0).toUpperCase() + err.message.slice(1);
  }
  return "We could not reach Mercatus. Check your internet connection and try again.";
}
