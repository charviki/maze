import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { http, HttpResponse } from 'msw';
import { createRequest } from './request';
import { server } from '../test/mocks/server';
import { storeTokens } from '../api/token-store';

describe('createRequest', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  afterEach(() => {
    localStorage.clear();
  });

  it('should return ok response for successful request', async () => {
    server.use(http.get('*/api/test', () => HttpResponse.json({ data: 'hello' })));

    const request = createRequest();
    const result = await request<{ data: string }>('/api/test');

    expect(result.status).toBe('ok');
    expect(result.data).toEqual({ data: 'hello' });
  });

  it('should return error response for failed request', async () => {
    server.use(
      http.get('*/api/test', () => HttpResponse.json({ message: 'not found' }, { status: 404 })),
    );

    const request = createRequest();
    const result = await request('/api/test');

    expect(result.status).toBe('error');
    expect(result.message).toBe('not found');
  });

  describe('Authorization header', () => {
    it('should inject Authorization header when token exists', async () => {
      localStorage.setItem('maze:access_token', 'test-jwt-token');

      let capturedAuth: string | null = null;
      server.use(
        http.get('*/api/test', ({ request }) => {
          capturedAuth = request.headers.get('Authorization');
          return HttpResponse.json({ ok: true });
        }),
      );

      const request = createRequest();
      await request('/api/test');

      expect(capturedAuth).toBe('Bearer test-jwt-token');
    });

    it('should not inject Authorization header when no token', async () => {
      let capturedAuth: string | null = undefined as unknown as string | null;
      server.use(
        http.get('*/api/test', ({ request }) => {
          capturedAuth = request.headers.get('Authorization');
          return HttpResponse.json({ ok: true });
        }),
      );

      const request = createRequest();
      await request('/api/test');

      expect(capturedAuth).toBeNull();
    });

    it('should refresh before request when access token is about to expire', async () => {
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
        http.get('*/api/test', ({ request }) => {
          capturedAuth = request.headers.get('Authorization');
          return HttpResponse.json({ ok: true });
        }),
      );

      const request = createRequest();
      const result = await request('/api/test');

      expect(result.status).toBe('ok');
      expect(capturedAuth).toBe('Bearer fresh-token');
    });

    it('should retry once after 401 TOKEN_EXPIRED', async () => {
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
        http.get('*/api/test', ({ request }) => {
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

      const request = createRequest();
      const result = await request('/api/test');

      expect(result.status).toBe('ok');
      expect(capturedAuth).toEqual(['Bearer stale-token', 'Bearer retried-token']);
    });
  });
});
