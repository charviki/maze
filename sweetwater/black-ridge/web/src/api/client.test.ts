import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createAgentApi } from './client';

const mockFetch = vi.fn();
vi.stubGlobal('fetch', mockFetch);

beforeEach(() => {
  mockFetch.mockReset();
});

describe('api.listSessions - 成功请求', () => {
  it('应正确解析成功的 API 响应', async () => {
    const mockData = [
      {
        id: 'session-1',
        name: 'session-1',
        status: 'running',
        created_at: '2024-01-01',
        window_count: 1,
      },
    ];
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve(JSON.stringify({ status: 'ok', data: mockData })),
    });

    const api = createAgentApi();
    const result = await api.listSessions();

    expect(result.status).toBe('ok');
    expect(result.data).toHaveLength(1);
    expect(result.data![0].id).toBe('session-1');
    expect(mockFetch).toHaveBeenCalledWith(
      '/api/v1/sessions',
      expect.objectContaining({
        headers: expect.objectContaining({ 'Content-Type': 'application/json' }),
      }),
    );
  });
});

describe('api.createSession - 成功创建', () => {
  it('应发送 POST 请求并返回创建的会话', async () => {
    const newSession = {
      id: 'new-1',
      name: 'new-1',
      status: 'running',
      created_at: '2024-01-01',
      window_count: 1,
    };
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve(JSON.stringify({ status: 'ok', data: newSession })),
    });

    const api = createAgentApi();
    const result = await api.createSession({ name: 'new-1', command: 'bash' });

    expect(result.status).toBe('ok');
    expect(result.data!.name).toBe('new-1');
    expect(mockFetch).toHaveBeenCalledWith(
      '/api/v1/sessions',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ name: 'new-1', command: 'bash' }),
      }),
    );
  });
});

describe('api.deleteSession - 成功删除', () => {
  it('应发送 DELETE 请求', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve(JSON.stringify({ status: 'ok' })),
    });

    const api = createAgentApi();
    const result = await api.deleteSession('session-1');

    expect(result.status).toBe('ok');
    expect(mockFetch).toHaveBeenCalledWith(
      '/api/v1/sessions/session-1',
      expect.objectContaining({ method: 'DELETE' }),
    );
  });
});

describe('配置接口', () => {
  it('getSessionConfig 应请求 /api/v1/sessions/:id/config', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () =>
        Promise.resolve(
          JSON.stringify({
            status: 'ok',
            data: {
              session_id: 'session-1',
              template_id: 'claude',
              working_dir: '/home/agent/session-1',
              scope: 'project',
              files: [
                { path: '.claude/settings.json', content: '{}', exists: true, hash: 'md5:abc' },
              ],
            },
          }),
        ),
    });

    const api = createAgentApi();
    const result = await api.getSessionConfig('session-1');

    expect(result.status).toBe('ok');
    expect(result.data?.files[0].path).toBe('.claude/settings.json');
    expect(mockFetch).toHaveBeenCalledWith('/api/v1/sessions/session-1/config', expect.anything());
  });

  it('updateTemplateConfig 应发送 PUT 到 /api/v1/templates/:id/config', async () => {
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

    const api = createAgentApi();
    const result = await api.updateTemplateConfig('claude', {
      files: [{ path: '~/.claude/settings.json', content: '{}', base_hash: 'md5:empty' }],
    });

    expect(result.status).toBe('ok');
    expect(mockFetch).toHaveBeenCalledWith(
      '/api/v1/templates/claude/config',
      expect.objectContaining({ method: 'PUT' }),
    );
  });

  it('409 冲突应保留错误 code 与 conflicts', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 409,
      text: () =>
        Promise.resolve(
          JSON.stringify({
            status: 'error',
            code: 'config_conflict',
            message: '配置已变更，请重新加载后再修改',
            conflicts: [{ path: '~/.claude/settings.json', current_hash: 'md5:def' }],
          }),
        ),
    });

    const api = createAgentApi();
    const result = await api.updateTemplateConfig('claude', {
      files: [{ path: '~/.claude/settings.json', content: '{}', base_hash: 'md5:abc' }],
    });

    expect(result.status).toBe('error');
    expect(result.code).toBe('config_conflict');
    expect(result.conflicts?.[0].path).toBe('~/.claude/settings.json');
  });
});

describe('网络错误', () => {
  it('fetch 抛出异常时 request 函数应捕获并返回 error 状态', async () => {
    mockFetch.mockRejectedValueOnce(new TypeError('Failed to fetch'));

    const api = createAgentApi();
    const result = await api.listSessions();

    expect(result.status).toBe('error');
    expect(result.message).toContain('Failed to fetch');
  });
});
