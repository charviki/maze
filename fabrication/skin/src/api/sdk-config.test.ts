import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { http, HttpResponse, delay } from 'msw';
import { createSdkConfiguration, createAuthFetch } from './sdk-config';
import { client } from './gen/client.gen';
import { server } from '../test/mocks/server';
import { storeTokens } from './token-store';

describe('createSdkConfiguration', () => {
  it('should configure client baseUrl', () => {
    createSdkConfiguration('/api/v1');
    // hey-api setConfig 会裁剪 baseUrl 尾斜杠
    expect(client.getConfig().baseUrl).toBe('/api/v1');
  });

  it('should inject custom fetch via createAuthFetch', async () => {
    server.use(http.get('*/test', () => HttpResponse.json({ test: true })));

    const fetchFn = createAuthFetch();
    const result = await fetchFn('/test');
    expect(result.ok).toBe(true);
  });

  it('should abort request after timeout', async () => {
    vi.useFakeTimers();

    server.use(
      http.get('*/test', async () => {
        await delay('infinite');
        return HttpResponse.json({});
      }),
    );

    const fetchFn = createAuthFetch();
    const promise = fetchFn('/test');
    vi.advanceTimersByTime(30000);
    await expect(promise).rejects.toThrow();

    vi.useRealTimers();
  });

  describe('Authorization header', () => {
    beforeEach(() => {
      localStorage.clear();
    });

    afterEach(() => {
      localStorage.clear();
    });

    it('should inject Authorization header when token exists', async () => {
      localStorage.setItem('maze:access_token', 'test-jwt-token');

      let capturedAuth: string | null = null;
      server.use(
        http.get('*/test', ({ request }) => {
          capturedAuth = request.headers.get('Authorization');
          return HttpResponse.json({ ok: true });
        }),
      );

      const fetchFn = createAuthFetch();
      await fetchFn('/test');

      expect(capturedAuth).toBe('Bearer test-jwt-token');
    });

    it('should not inject Authorization header when no token', async () => {
      let capturedAuth: string | null = undefined as unknown as string | null;
      server.use(
        http.get('*/test', ({ request }) => {
          capturedAuth = request.headers.get('Authorization');
          return HttpResponse.json({ ok: true });
        }),
      );

      const fetchFn = createAuthFetch();
      await fetchFn('/test');

      expect(capturedAuth).toBeNull();
    });

    it('should refresh before sdk request when token is expiring', async () => {
      storeTokens({
        accessToken: 'expiring-token',
        refreshToken: 'refresh-token',
        expiresIn: 0,
      });

      let capturedAuth: string | null = null;
      server.use(
        http.post('*/api/v1/auth/refresh', () =>
          HttpResponse.json({
            accessToken: 'fresh-token',
            refreshToken: 'refresh-token-2',
            expiresIn: 3600,
          }),
        ),
        http.get('*/test', ({ request }) => {
          capturedAuth = request.headers.get('Authorization');
          return HttpResponse.json({ ok: true });
        }),
      );

      const fetchFn = createAuthFetch();
      const response = await fetchFn('/test');

      expect(response.ok).toBe(true);
      expect(capturedAuth).toBe('Bearer fresh-token');
    });

    it('should retry sdk request once after 401 TOKEN_EXPIRED', async () => {
      storeTokens({
        accessToken: 'stale-token',
        refreshToken: 'refresh-token',
        expiresIn: 3600,
      });

      const capturedAuth: string[] = [];
      let attempt = 0;
      server.use(
        http.post('*/api/v1/auth/refresh', () =>
          HttpResponse.json({
            accessToken: 'retried-token',
            refreshToken: 'refresh-token-2',
            expiresIn: 3600,
          }),
        ),
        http.get('*/test', ({ request }) => {
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

      const fetchFn = createAuthFetch();
      const response = await fetchFn('/test');

      expect(response.ok).toBe(true);
      expect(capturedAuth).toEqual(['Bearer stale-token', 'Bearer retried-token']);
    });
  });
});
