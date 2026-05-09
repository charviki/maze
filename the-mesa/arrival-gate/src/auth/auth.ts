import {
  clearTokens,
  fetchWithAuthSession,
  getAccessToken as getStoredAccessToken,
  getRefreshToken as getStoredRefreshToken,
  isTokenAuthenticated,
  refreshAccessToken as refreshSharedAccessToken,
  storeTokens,
} from '@maze/fabrication';

const USER_KEY = 'maze:auth_user';

interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number | string;
}

export async function login(username: string, password: string): Promise<boolean> {
  try {
    const response = await fetch('/api/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    });
    if (!response.ok) return false;
    const data = (await response.json()) as LoginResponse;
    if (!data.accessToken || !data.refreshToken) {
      return false;
    }
    // 登录成功后立即落绝对过期时间，后续请求才能稳定执行预刷新。
    storeTokens(data);
    localStorage.setItem(USER_KEY, username);
    return true;
  } catch {
    return false;
  }
}

export async function logout(): Promise<void> {
  const refreshToken = getStoredRefreshToken();
  let accessToken = getStoredAccessToken();

  // accessToken 已失效时，先用 refreshToken 尝试获取新的，确保 logout API 能鉴权通过
  if (refreshToken && !accessToken) {
    try {
      await refreshSharedAccessToken();
      accessToken = getStoredAccessToken();
    } catch {
      // refresh 失败仍继续调 logout，尽最大努力通知后端
    }
  }

  if (refreshToken) {
    const headers = new Headers({ 'Content-Type': 'application/json' });
    if (accessToken) {
      headers.set('Authorization', `Bearer ${accessToken}`);
    }
    await fetch('/api/v1/auth/logout', {
      method: 'POST',
      headers,
      body: JSON.stringify({ refreshToken }),
    }).catch(() => {});
  }
  clearTokens();
  localStorage.removeItem(USER_KEY);
}

export function isAuthenticated(): boolean {
  return isTokenAuthenticated();
}

export function getAccessToken(): string | null {
  return getStoredAccessToken();
}

export function getRefreshToken(): string | null {
  return getStoredRefreshToken();
}

export function getCurrentUser(): string | null {
  return localStorage.getItem(USER_KEY);
}

export async function refreshAccessToken(): Promise<boolean> {
  const refreshed = await refreshSharedAccessToken();
  if (!refreshed) {
    localStorage.removeItem(USER_KEY);
  }
  return refreshed;
}

export async function fetchWithAuth(
  input: RequestInfo | URL,
  init?: RequestInit,
): Promise<Response> {
  const response = await fetchWithAuthSession(input, init);
  if (response.status === 401 && !isTokenAuthenticated()) {
    localStorage.removeItem(USER_KEY);
  }
  return response;
}
