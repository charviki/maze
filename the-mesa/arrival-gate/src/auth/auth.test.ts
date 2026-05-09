import { beforeEach, describe, expect, it } from 'vitest';
import { http, HttpResponse } from 'msw';
import { getAccessTokenExpiresAt } from '@maze/fabrication';
import {
  fetchWithAuth,
  getAccessToken,
  getCurrentUser,
  isAuthenticated,
  login,
  logout,
  refreshAccessToken,
} from './auth';
import { server } from '../test/mocks/server';

describe('auth', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('login 成功后会保存用户与 token 绝对过期时间', async () => {
    server.use(
      http.post('*/api/v1/auth/login', async ({ request }) => {
        const body = (await request.json()) as { username: string; password: string };
        expect(body).toEqual({ username: 'admin', password: 'admin' });
        return HttpResponse.json({
          accessToken: 'access-1',
          refreshToken: 'refresh-1',
          expiresIn: 3600,
        });
      }),
    );

    await expect(login('admin', 'admin')).resolves.toBe(true);
    expect(getAccessToken()).toBe('access-1');
    expect(getCurrentUser()).toBe('admin');
    expect(getAccessTokenExpiresAt()).toBeGreaterThan(Date.now());
  });

  it('login 失败时不保存本地状态', async () => {
    server.use(
      http.post('*/api/v1/auth/login', () =>
        HttpResponse.json({ message: 'invalid credentials' }, { status: 401 }),
      ),
    );

    await expect(login('admin', 'wrong')).resolves.toBe(false);
    expect(isAuthenticated()).toBe(false);
    expect(getCurrentUser()).toBeNull();
  });

  it('logout 有 accessToken 时直接调 logout API 并清理本地状态', async () => {
    localStorage.setItem('maze:access_token', 'access-1');
    localStorage.setItem('maze:refresh_token', 'refresh-1');
    localStorage.setItem('maze:auth_user', 'admin');

    server.use(
      http.post('*/api/v1/auth/logout', async ({ request }) => {
        const body = (await request.json()) as { refreshToken: string };
        expect(body.refreshToken).toBe('refresh-1');
        return HttpResponse.json({});
      }),
    );

    await logout();

    expect(isAuthenticated()).toBe(false);
    expect(getCurrentUser()).toBeNull();
  });

  it('logout 无 accessToken 时先 refresh 再调 logout API', async () => {
    localStorage.setItem('maze:refresh_token', 'refresh-1');
    localStorage.setItem('maze:auth_user', 'admin');

    const refreshCalled = { value: false };
    server.use(
      http.post('*/api/v1/auth/refresh', async ({ request }) => {
        const body = (await request.json()) as { refreshToken: string };
        expect(body.refreshToken).toBe('refresh-1');
        refreshCalled.value = true;
        return HttpResponse.json({
          accessToken: 'access-new',
          refreshToken: 'refresh-new',
          expiresIn: 3600,
        });
      }),
      http.post('*/api/v1/auth/logout', async ({ request }) => {
        const authHeader = request.headers.get('Authorization');
        expect(authHeader).toBe('Bearer access-new');
        const body = (await request.json()) as { refreshToken: string };
        expect(body.refreshToken).toBe('refresh-new');
        return HttpResponse.json({});
      }),
    );

    await logout();

    expect(refreshCalled.value).toBe(true);
    expect(isAuthenticated()).toBe(false);
    expect(getCurrentUser()).toBeNull();
  });

  it('logout 在 refresh 也失败时仍清理本地状态', async () => {
    localStorage.setItem('maze:refresh_token', 'refresh-1');
    localStorage.setItem('maze:auth_user', 'admin');

    server.use(
      http.post('*/api/v1/auth/refresh', () =>
        HttpResponse.json({ reason: 'TOKEN_EXPIRED' }, { status: 401 }),
      ),
      http.post('*/api/v1/auth/logout', () => HttpResponse.json({})),
    );

    await logout();

    expect(isAuthenticated()).toBe(false);
    expect(getCurrentUser()).toBeNull();
  });

  it('refreshAccessToken 成功时会轮换本地 token', async () => {
    localStorage.setItem('maze:refresh_token', 'refresh-1');
    localStorage.setItem('maze:auth_user', 'admin');

    server.use(
      http.post('*/api/v1/auth/refresh', async ({ request }) => {
        const body = (await request.json()) as { refreshToken: string };
        expect(body.refreshToken).toBe('refresh-1');
        return HttpResponse.json({
          accessToken: 'access-2',
          refreshToken: 'refresh-2',
          expiresIn: 1200,
        });
      }),
    );

    await expect(refreshAccessToken()).resolves.toBe(true);
    expect(getAccessToken()).toBe('access-2');
    expect(localStorage.getItem('maze:refresh_token')).toBe('refresh-2');
    expect(getCurrentUser()).toBe('admin');
  });

  it('refreshAccessToken 失败时会清理入口页用户信息', async () => {
    localStorage.setItem('maze:refresh_token', 'refresh-1');
    localStorage.setItem('maze:auth_user', 'admin');

    server.use(
      http.post('*/api/v1/auth/refresh', () =>
        HttpResponse.json({ reason: 'TOKEN_EXPIRED' }, { status: 401 }),
      ),
    );

    await expect(refreshAccessToken()).resolves.toBe(false);
    expect(isAuthenticated()).toBe(false);
    expect(getCurrentUser()).toBeNull();
  });

  it('fetchWithAuth 在 401 TOKEN_EXPIRED 后会刷新并重试一次', async () => {
    localStorage.setItem('maze:access_token', 'access-1');
    localStorage.setItem('maze:refresh_token', 'refresh-1');
    localStorage.setItem('maze:auth_user', 'admin');

    const capturedAuth: string[] = [];
    let attempt = 0;
    server.use(
      http.post('*/api/v1/auth/refresh', () =>
        HttpResponse.json({
          accessToken: 'access-2',
          refreshToken: 'refresh-2',
          expiresIn: 3600,
        }),
      ),
      http.get('*/api/v1/protected', ({ request }) => {
        capturedAuth.push(request.headers.get('Authorization') ?? '');
        attempt += 1;
        if (attempt === 1) {
          return HttpResponse.json(
            { reason: 'TOKEN_EXPIRED', message: 'token expired' },
            { status: 401 },
          );
        }
        return HttpResponse.json({ ok: true });
      }),
    );

    const response = await fetchWithAuth('/api/v1/protected');

    expect(response.ok).toBe(true);
    expect(capturedAuth).toEqual(['Bearer access-1', 'Bearer access-2']);
  });
});
