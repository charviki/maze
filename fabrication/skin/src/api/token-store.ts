const ACCESS_TOKEN_KEY = 'maze:access_token';
const REFRESH_TOKEN_KEY = 'maze:refresh_token';
const ACCESS_TOKEN_EXPIRES_AT_KEY = 'maze:access_token_expires_at';

export const DEFAULT_TOKEN_REFRESH_WINDOW_MS = 60_000;

interface StoredTokens {
  accessToken: string;
  refreshToken: string;
  expiresIn?: number | string | null;
}

function readStorage(key: string): string | null {
  if (typeof window === 'undefined') {
    return null;
  }
  return window.localStorage.getItem(key);
}

function writeStorage(key: string, value: string): void {
  if (typeof window === 'undefined') {
    return;
  }
  window.localStorage.setItem(key, value);
}

function removeStorage(key: string): void {
  if (typeof window === 'undefined') {
    return;
  }
  window.localStorage.removeItem(key);
}

function normalizeExpiresInSeconds(expiresIn?: number | string | null): number | null {
  if (expiresIn == null) {
    return null;
  }
  const seconds = typeof expiresIn === 'string' ? Number(expiresIn) : expiresIn;
  if (!Number.isFinite(seconds)) {
    return null;
  }
  return Math.max(0, seconds);
}

export function getAccessToken(): string | null {
  return readStorage(ACCESS_TOKEN_KEY);
}

export function setAccessToken(token: string): void {
  writeStorage(ACCESS_TOKEN_KEY, token);
}

export function getRefreshToken(): string | null {
  return readStorage(REFRESH_TOKEN_KEY);
}

export function setRefreshToken(token: string): void {
  writeStorage(REFRESH_TOKEN_KEY, token);
}

export function getAccessTokenExpiresAt(): number | null {
  const raw = readStorage(ACCESS_TOKEN_EXPIRES_AT_KEY);
  if (!raw) {
    return null;
  }
  const expiresAt = Number(raw);
  return Number.isFinite(expiresAt) ? expiresAt : null;
}

export function setAccessTokenExpiresAt(expiresAt: number): void {
  writeStorage(ACCESS_TOKEN_EXPIRES_AT_KEY, String(expiresAt));
}

export function storeTokens(tokens: StoredTokens): void {
  const expiresInSeconds = normalizeExpiresInSeconds(tokens.expiresIn);

  setAccessToken(tokens.accessToken);
  setRefreshToken(tokens.refreshToken);
  if (expiresInSeconds == null) {
    // 旧会话没有 expiresIn 时不强行猜测，交给 401 兜底刷新处理。
    removeStorage(ACCESS_TOKEN_EXPIRES_AT_KEY);
    return;
  }
  setAccessTokenExpiresAt(Date.now() + expiresInSeconds * 1000);
}

export function clearTokens(): void {
  removeStorage(ACCESS_TOKEN_KEY);
  removeStorage(REFRESH_TOKEN_KEY);
  removeStorage(ACCESS_TOKEN_EXPIRES_AT_KEY);
}

export function isAccessTokenExpired(now = Date.now()): boolean {
  const expiresAt = getAccessTokenExpiresAt();
  if (expiresAt == null) {
    return false;
  }
  return expiresAt <= now;
}

export function willAccessTokenExpireSoon(
  bufferMs = DEFAULT_TOKEN_REFRESH_WINDOW_MS,
  now = Date.now(),
): boolean {
  const expiresAt = getAccessTokenExpiresAt();
  if (expiresAt == null) {
    return false;
  }
  return expiresAt - now <= bufferMs;
}

export function isAuthenticated(): boolean {
  // 刷新令牌仍在时，页面可通过预刷新恢复 access token，不应过早把会话判定为未登录。
  return !!getAccessToken() || !!getRefreshToken();
}
