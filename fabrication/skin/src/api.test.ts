import { describe, expect, it } from 'vitest';
import { http, HttpResponse } from 'msw';
import { createAgentApiClient } from './api';
import { server } from './test/mocks/server';

describe('createAgentApiClient', () => {
  it('本地配置接口应请求 local-config 路径', async () => {
    const api = createAgentApiClient('/api/v1/nodes/test/');
    server.use(
      http.get('*/api/v1/nodes/test/local-config', () =>
        HttpResponse.json({ workingDir: '/home/agent', env: {} }),
      ),
    );

    const result = await api.getLocalConfig();

    expect(result.status).toBe('ok');
  });

  it('更新本地配置应请求 local-config 路径', async () => {
    const api = createAgentApiClient('/api/v1/nodes/test/');
    server.use(
      http.put('*/api/v1/nodes/test/local-config', () =>
        HttpResponse.json({ workingDir: '/home/agent', env: { FOO: 'bar' } }),
      ),
    );

    const result = await api.updateLocalConfig({ env: { FOO: 'bar' } });

    expect(result.status).toBe('ok');
  });
});
