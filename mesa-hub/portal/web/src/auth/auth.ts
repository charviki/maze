/**
 * Authentication service — OIDC-ready interface.
 *
 * Current: hardcoded users + localStorage.
 * Future: swap internals for OIDC token + redirect flow
 * (Auth0, Keycloak, etc.) without touching AuthGate or LoginPage.
 */

const STORAGE_KEY = 'maze:auth';

interface AuthSession {
  user: string;
  loginAt: number;
}

const HARDCODED_USERS = [{ username: 'admin', password: 'admin' }];

export function login(username: string, password: string): boolean {
  const match = HARDCODED_USERS.find((u) => u.username === username && u.password === password);
  if (!match) return false;

  const session: AuthSession = { user: username, loginAt: Date.now() };
  localStorage.setItem(STORAGE_KEY, JSON.stringify(session));
  return true;
}

export function logout(): void {
  localStorage.removeItem(STORAGE_KEY);
}

export function isAuthenticated(): boolean {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return false;
  try {
    // JSON.parse returns unknown; validate shape before use
    const parsed: unknown = JSON.parse(raw);
    if (typeof parsed !== 'object' || parsed === null) return false;
    const session = parsed as Record<string, unknown>;
    return typeof session.user === 'string' && session.user.length > 0;
  } catch {
    return false;
  }
}

export function getCurrentUser(): string | null {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return null;
  try {
    const parsed: unknown = JSON.parse(raw);
    if (typeof parsed !== 'object' || parsed === null) return null;
    const session = parsed as Record<string, unknown>;
    return typeof session.user === 'string' ? session.user : null;
  } catch {
    return null;
  }
}
