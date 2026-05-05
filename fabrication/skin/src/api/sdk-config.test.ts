import { describe, it, expect, vi } from 'vitest';
import { http, HttpResponse, delay } from 'msw';
import { createSdkConfiguration } from './sdk-config';
import { Configuration } from './gen/runtime';
import { server } from '../test/mocks/server';

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
});
