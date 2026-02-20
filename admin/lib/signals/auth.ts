import { effect, signal } from "@preact/signals";
import { api } from "../api/client.ts";

const STORAGE_KEY = "admin_token";

// Auth state
export const authToken = signal<string | null>(null);
export const isAuthenticated = signal(false);

// Initialize from localStorage on load
if (typeof globalThis.localStorage !== "undefined") {
  const stored = globalThis.localStorage.getItem(STORAGE_KEY);
  if (stored) {
    authToken.value = stored;
    api.setAuthToken(stored);
    isAuthenticated.value = true;
  }
}

// Sync token to API client and localStorage
effect(() => {
  const token = authToken.value;
  api.setAuthToken(token);

  if (typeof globalThis.localStorage !== "undefined") {
    if (token) {
      globalThis.localStorage.setItem(STORAGE_KEY, token);
      isAuthenticated.value = true;
    } else {
      globalThis.localStorage.removeItem(STORAGE_KEY);
      isAuthenticated.value = false;
    }
  }
});

// Actions
export function setAuthToken(token: string | null) {
  authToken.value = token;
}

export function clearAuthToken() {
  authToken.value = null;
}

// Parse JWT to get basic info (without verification)
export function parseToken(
  token: string,
): { exp?: number; email?: string; role?: string; user_id?: string } | null {
  try {
    const parts = token.split(".");
    if (parts.length !== 3) return null;
    const payload = JSON.parse(atob(parts[1]));
    return payload;
  } catch {
    return null;
  }
}

// Check if token is expired
export function isTokenExpired(token: string): boolean {
  const payload = parseToken(token);
  if (!payload?.exp) return false;
  return Date.now() >= payload.exp * 1000;
}
