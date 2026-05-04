import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createAgentApiClient } from './api';

const mockFetch = vi.fn();
vi.stubGlobal('fetch', mockFetch);

beforeEach(() => {
  mockFetch.mockReset();
});

function mockOk(data: unknown) {
  return {
    ok: true,
    status: 200,
    text: () => Promise.resolve(JSON.stringify(data)),
    json: () => Promise.resolve(data),
    clone() {
      return this;
    },
  };
}

describe('createAgentApiClient', () => {
  it('本地配置接口应请求 local-config 路径', async () => {
    const api = createAgentApiClient('/api/v1/nodes/test/');
    mockFetch.mockResolvedValueOnce(mockOk({ workingDir: '/home/agent', env: {} }));

    const result = await api.getLocalConfig();

    expect(result.status).toBe('ok');
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/nodes/test/local-config'),
      expect.anything(),
    );
  });

  it('更新本地配置应请求 local-config 路径', async () => {
    const api = createAgentApiClient('/api/v1/nodes/test/');
    mockFetch.mockResolvedValueOnce(mockOk({ workingDir: '/home/agent', env: { FOO: 'bar' } }));

    const result = await api.updateLocalConfig({ env: { FOO: 'bar' } });

    expect(result.status).toBe('ok');
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/nodes/test/local-config'),
      expect.objectContaining({ method: 'PUT' }),
    );
  });
});
