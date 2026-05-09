import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { http, HttpResponse, delay } from 'msw';
import { createSdkConfiguration } from './sdk-config';
import { Configuration } from './gen/runtime';
import { server } from '../test/mocks/server';
import { storeTokens } from './token-store';

describe('createSdkConfiguration', () => {
  it('should return a Configuration instance', () => {
    const config = createSdkConfiguration('/api/v1');
    expect(config).toBeInstanceOf(Configuration);
  });

  it('should set basePath', () => {
    const config = createSdkConfiguration('/api/v1');
    expect(config.basePath).toBe('/api/v1');
  });

  it('should inject custom fetchApi', async () => {
    server.use(http.get('*/test', () => HttpResponse.json({ test: true })));

    const config = createSdkConfiguration('');
    expect(config.fetchApi).toBeDefined();

    const result = await config.fetchApi!('/test');
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

    const config = createSdkConfiguration('');
    const promise = config.fetchApi!('/test');
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

      const config = createSdkConfiguration('');
      await config.fetchApi!('/test');

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

      const config = createSdkConfiguration('');
      await config.fetchApi!('/test');

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

      const config = createSdkConfiguration('');
      const response = await config.fetchApi!('/test');

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

      const config = createSdkConfiguration('');
      const response = await config.fetchApi!('/test');

      expect(response.ok).toBe(true);
      expect(capturedAuth).toEqual(['Bearer stale-token', 'Bearer retried-token']);
    });
  });
});
