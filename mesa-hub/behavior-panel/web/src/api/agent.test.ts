import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createAgentApi } from './agent';

const mockFetch = vi.fn();
vi.stubGlobal('fetch', mockFetch);

beforeEach(() => {
  mockFetch.mockReset();
});

// createAgentApi 现在通过 Manager 代理路径调用，不再直连 Agent
const agentApi = createAgentApi('', 'agent-1');

// 期望的代理基础路径：Manager 地址 + /api/v1/nodes/{nodeName}/sessions
const proxyBase = '/api/v1/nodes/agent-1/sessions';

describe('agent API - 通过 Manager 代理成功请求', () => {
  it('listSessions 应请求代理路径 /api/v1/nodes/:name/sessions', async () => {
    const sessions = [
      { id: 's1', name: 's1', status: 'running', created_at: '2024-01-01', window_count: 1 },
    ];
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve(JSON.stringify({ status: 'ok', data: sessions })),
    });

    const result = await agentApi.listSessions();

    expect(result.status).toBe('ok');
    expect(result.data).toHaveLength(1);
    expect(result.data![0].id).toBe('s1');
    expect(mockFetch).toHaveBeenCalledWith(
      proxyBase,
      expect.objectContaining({
        headers: expect.objectContaining({ 'Content-Type': 'application/json' }),
      }),
    );
  });

  it('createSession 应发送 POST 到代理路径', async () => {
    const newSession = {
      id: 'new',
      name: 'new',
      status: 'running',
      created_at: '2024-01-01',
      window_count: 1,
    };
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve(JSON.stringify({ status: 'ok', data: newSession })),
    });

    const result = await agentApi.createSession({ name: 'new', restore_strategy: 'auto' });

    expect(result.status).toBe('ok');
    expect(result.data!.name).toBe('new');
    expect(mockFetch).toHaveBeenCalledWith(
      proxyBase,
      expect.objectContaining({
        method: 'POST',
      }),
    );
  });

  it('getSession 应请求代理路径 /api/v1/nodes/:name/sessions/:id', async () => {
    const session = {
      id: 'abc',
      name: 'abc',
      status: 'running',
      created_at: '2024-01-01',
      window_count: 1,
    };
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve(JSON.stringify({ status: 'ok', data: session })),
    });

    const result = await agentApi.getSession('abc');

    expect(result.status).toBe('ok');
    expect(result.data!.id).toBe('abc');
    expect(mockFetch).toHaveBeenCalledWith(
      `${proxyBase}/abc`,
      expect.objectContaining({
        headers: expect.objectContaining({ 'Content-Type': 'application/json' }),
      }),
    );
  });

  it('saveSessions 应发送 POST 到代理路径', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () =>
        Promise.resolve(
          JSON.stringify({ status: 'ok', data: { saved_at: '2025-04-21T12:00:00Z' } }),
        ),
    });

    const result = await agentApi.saveSessions();

    expect(result.status).toBe('ok');
    expect(result.data!.saved_at).toBe('2025-04-21T12:00:00Z');
    expect(mockFetch).toHaveBeenCalledWith(
      `${proxyBase}/save`,
      expect.objectContaining({ method: 'POST' }),
    );
  });

  it('getSavedSessions 应请求代理路径 /api/v1/nodes/:name/sessions/saved', async () => {
    const savedSessions = [
      {
        session_name: 'saved-1',
        pipeline: [
          { id: 'sys-cd', type: 'cd', phase: 'system', order: 0, key: '/home/agent', value: '' },
        ],
        restore_strategy: 'manual',
        working_dir: '/home/agent',
        terminal_snapshot: '$ claude\n> Hello',
        saved_at: '2025-01-01T00:00:00Z',
      },
    ];
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve(JSON.stringify({ status: 'ok', data: savedSessions })),
    });

    const result = await agentApi.getSavedSessions();

    expect(result.status).toBe('ok');
    expect(result.data).toHaveLength(1);
    expect(mockFetch).toHaveBeenCalledWith(`${proxyBase}/saved`, expect.anything());
  });

  it('restoreSession 应发送 POST 到代理路径', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve(JSON.stringify({ status: 'ok', data: null })),
    });

    const result = await agentApi.restoreSession('saved-1');

    expect(result.status).toBe('ok');
    expect(mockFetch).toHaveBeenCalledWith(
      `${proxyBase}/saved-1/restore`,
      expect.objectContaining({ method: 'POST' }),
    );
  });

  it('deleteSession 应发送 DELETE 到代理路径', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve(JSON.stringify({ status: 'ok', data: null })),
    });

    const result = await agentApi.deleteSession('old-session');

    expect(result.status).toBe('ok');
    expect(mockFetch).toHaveBeenCalledWith(
      `${proxyBase}/old-session`,
      expect.objectContaining({ method: 'DELETE' }),
    );
  });

  it('getSessionConfig 应请求 session 配置代理路径', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () =>
        Promise.resolve(
          JSON.stringify({
            status: 'ok',
            data: {
              session_id: 'sess-1',
              template_id: 'claude',
              working_dir: '/home/agent/sess-1',
              scope: 'project',
              files: [
                { path: '.claude/settings.json', content: '{}', exists: true, hash: 'md5:abc' },
              ],
            },
          }),
        ),
    });

    const result = await agentApi.getSessionConfig('sess-1');

    expect(result.status).toBe('ok');
    expect(result.data?.files[0].path).toBe('.claude/settings.json');
    expect(mockFetch).toHaveBeenCalledWith(`${proxyBase}/sess-1/config`, expect.anything());
  });

  it('updateSessionConfig 应发送 PUT 到 session 配置代理路径', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () =>
        Promise.resolve(
          JSON.stringify({
            status: 'ok',
            data: {
              session_id: 'sess-1',
              template_id: 'claude',
              working_dir: '/home/agent/sess-1',
              scope: 'project',
              files: [
                {
                  path: '.claude/settings.json',
                  content: '{"theme":"dark"}',
                  exists: true,
                  hash: 'md5:def',
                },
              ],
            },
          }),
        ),
    });

    const result = await agentApi.updateSessionConfig('sess-1', {
      files: [{ path: '.claude/settings.json', content: '{"theme":"dark"}', base_hash: 'md5:abc' }],
    });

    expect(result.status).toBe('ok');
    expect(mockFetch).toHaveBeenCalledWith(
      `${proxyBase}/sess-1/config`,
      expect.objectContaining({ method: 'PUT' }),
    );
  });

  it('getTemplateConfig 应请求模板配置代理路径', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () =>
        Promise.resolve(
          JSON.stringify({
            status: 'ok',
            data: {
              template_id: 'claude',
              scope: 'global',
              files: [
                { path: '~/.claude/settings.json', content: '{}', exists: true, hash: 'md5:abc' },
              ],
            },
          }),
        ),
    });

    const result = await agentApi.getTemplateConfig('claude');

    expect(result.status).toBe('ok');
    expect(result.data?.files[0].path).toBe('~/.claude/settings.json');
    expect(mockFetch).toHaveBeenCalledWith(
      '/api/v1/nodes/agent-1/templates/claude/config',
      expect.anything(),
    );
  });
});

describe('agent API - HTTP 错误', () => {
  it('非 2xx 响应应返回 error 状态和 HTTP 状态码', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      text: () => Promise.resolve(''),
    });

    const result = await agentApi.listSessions();

    expect(result.status).toBe('error');
    expect(result.message).toBe('HTTP 500');
  });

  it('404 响应应返回 error 状态', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 404,
      text: () => Promise.resolve(''),
    });

    const result = await agentApi.getSession('nonexistent');

    expect(result.status).toBe('error');
    expect(result.message).toBe('HTTP 404');
  });

  it('409 配置冲突应保留服务端错误体', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 409,
      text: () =>
        Promise.resolve(
          JSON.stringify({
            status: 'error',
            code: 'config_conflict',
            message: '配置已变更，请重新加载后再修改',
            conflicts: [{ path: '.claude/settings.json', current_hash: 'md5:def' }],
          }),
        ),
    });

    const result = await agentApi.updateSessionConfig('sess-1', {
      files: [{ path: '.claude/settings.json', content: '{}', base_hash: 'md5:abc' }],
    });

    expect(result.status).toBe('error');
    expect(result.code).toBe('config_conflict');
    expect(result.conflicts?.[0].path).toBe('.claude/settings.json');
  });
});

describe('agent API - 网络错误', () => {
  it('fetch 抛出异常时应返回 error 状态和错误信息', async () => {
    mockFetch.mockRejectedValueOnce(new TypeError('Failed to fetch'));

    const result = await agentApi.listSessions();

    expect(result.status).toBe('error');
    expect(result.message).toContain('Failed to fetch');
  });
});

describe('agent API - 代理路径验证', () => {
  it('不同 nodeName 应生成不同的代理路径', () => {
    const api1 = createAgentApi('', 'agent-alpha');
    const api2 = createAgentApi('', 'agent-beta');

    // 验证内部 base 路径正确（通过 fetch 调用验证）
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve(JSON.stringify({ status: 'ok', data: [] })),
    });
    void api1.listSessions();
    expect(mockFetch).toHaveBeenCalledWith('/api/v1/nodes/agent-alpha/sessions', expect.anything());

    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve(JSON.stringify({ status: 'ok', data: [] })),
    });
    void api2.listSessions();
    expect(mockFetch).toHaveBeenCalledWith('/api/v1/nodes/agent-beta/sessions', expect.anything());
  });

  it('nodeName 含特殊字符应被编码', () => {
    const api = createAgentApi('', 'agent name');
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve(JSON.stringify({ status: 'ok', data: [] })),
    });
    void api.listSessions();
    expect(mockFetch).toHaveBeenCalledWith(
      '/api/v1/nodes/agent%20name/sessions',
      expect.anything(),
    );
  });
});
