import {
  clearTokens,
  getAccessToken,
  getRefreshToken,
  isAccessTokenExpired,
  storeTokens,
  willAccessTokenExpireSoon,
} from './token-store';

const AUTH_REFRESH_ENDPOINT = '/api/v1/auth/refresh';
const LOGIN_REDIRECT_PATH = '/arrival-gate/';

interface AuthTokenResponse {
  accessToken?: string;
  refreshToken?: string;
  expiresIn?: number | string | null;
}

interface AuthFailureResponse {
  reason?: string;
}

let inFlightRefresh: Promise<boolean> | null = null;

function resolveRequestInput(input: RequestInfo | URL): RequestInfo | URL {
  if (input instanceof Request) {
    return input;
  }
  if (input instanceof URL) {
    return input;
  }
  if (typeof input !== 'string' || /^[a-zA-Z][a-zA-Z\d+\-.]*:/.test(input)) {
    return input;
  }
  if (typeof window === 'undefined') {
    return new URL(input, 'http://localhost');
  }
  return new URL(input, window.location.origin);
}

function buildAuthorizedRequest(input: RequestInfo | URL, init?: RequestInit): Request {
  const headers = new Headers(init?.headers);
  const accessToken = getAccessToken();

  if (accessToken) {
    headers.set('Authorization', `Bearer ${accessToken}`);
  }

  return new Request(resolveRequestInput(input), { ...init, headers });
}

function createUnauthorizedResponse(reason = 'TOKEN_EXPIRED'): Response {
  return new Response(JSON.stringify({ reason, message: '认证已失效，请重新登录' }), {
    status: 401,
    headers: { 'Content-Type': 'application/json' },
  });
}

async function readAuthFailureReason(response: Response): Promise<string | null> {
  if (response.status !== 401) {
    return null;
  }

  try {
    const payload = (await response.clone().json()) as AuthFailureResponse;
    return typeof payload.reason === 'string' ? payload.reason : null;
  } catch {
    return null;
  }
}

function clearSessionAndRedirect(): void {
  clearTokens();
  if (typeof window === 'undefined') {
    return;
  }
  if (window.location.pathname !== LOGIN_REDIRECT_PATH) {
    window.location.href = LOGIN_REDIRECT_PATH;
  }
}

async function performRefreshRequest(): Promise<boolean> {
  const refreshToken = getRefreshToken();
  if (!refreshToken) {
    clearTokens();
    return false;
  }

  try {
    const response = await fetch(AUTH_REFRESH_ENDPOINT, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refreshToken }),
    });
    if (!response.ok) {
      clearTokens();
      return false;
    }

    const data = (await response.json()) as AuthTokenResponse;
    if (!data.accessToken || !data.refreshToken) {
      clearTokens();
      return false;
    }

    // 服务端返回的 expiresIn 才是当前 access token 的真实 TTL，本地必须同步绝对过期时间。
    storeTokens({
      accessToken: data.accessToken,
      refreshToken: data.refreshToken,
      expiresIn: data.expiresIn,
    });
    return true;
  } catch {
    clearTokens();
    return false;
  }
}

export async function refreshAccessToken(): Promise<boolean> {
  if (inFlightRefresh) {
    return inFlightRefresh;
  }

  inFlightRefresh = performRefreshRequest().finally(() => {
    inFlightRefresh = null;
  });
  return inFlightRefresh;
}

export async function ensureFreshAccessToken(): Promise<boolean> {
  if (!willAccessTokenExpireSoon()) {
    return true;
  }
  if (!getRefreshToken()) {
    return !isAccessTokenExpired();
  }
  return refreshAccessToken();
}

export async function fetchWithAuthSession(
  input: RequestInfo | URL,
  init?: RequestInit,
): Promise<Response> {
  if (!(await ensureFreshAccessToken())) {
    clearSessionAndRedirect();
    return createUnauthorizedResponse();
  }

  let response = await fetch(buildAuthorizedRequest(input, init));
  if ((await readAuthFailureReason(response)) !== 'TOKEN_EXPIRED') {
    return response;
  }

  // 只有明确的 TOKEN_EXPIRED 才触发一次兜底刷新，避免把权限错误或无效令牌误重试。
  if (await refreshAccessToken()) {
    response = await fetch(buildAuthorizedRequest(input, init));
    return response;
  }

  clearSessionAndRedirect();
  return createUnauthorizedResponse();
}
