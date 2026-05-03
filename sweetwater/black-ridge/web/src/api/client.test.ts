import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createAgentApi } from './client';

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

function mockError(status: number, data?: unknown) {
  return {
    ok: false,
    status,
    text: () => Promise.resolve(data ? JSON.stringify(data) : ''),
    json: () => Promise.resolve(data),
    clone() {
      return this;
    },
  };
}

describe('api.listSessions', () => {
  it('应正确解析响应', async () => {
    const mockData = [
      {
        id: 'session-1',
        name: 'session-1',
        status: 'running',
        createdAt: '2024-01-01',
        windowCount: 1,
      },
    ];
    mockFetch.mockResolvedValueOnce(mockOk({ sessions: mockData }));

    const api = createAgentApi();
    const result = await api.listSessions();

    expect(result.status).toBe('ok');
    expect(result.data).toHaveLength(1);
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/sessions'),
      expect.anything(),
    );
  });
});

describe('api.createSession', () => {
  it('应发送 POST 并返回会话', async () => {
    const newSession = {
      id: 'new-1',
      name: 'new-1',
      status: 'running',
      createdAt: '2024-01-01',
      windowCount: 1,
    };
    mockFetch.mockResolvedValueOnce(mockOk(newSession));

    const api = createAgentApi();
    const result = await api.createSession({ name: 'new-1', command: 'bash' });

    expect(result.status).toBe('ok');
    expect(mockFetch).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({ method: 'POST' }),
    );
  });
});

describe('api.deleteSession', () => {
  it('应发送 DELETE', async () => {
    mockFetch.mockResolvedValueOnce(mockOk({}));

    const api = createAgentApi();
    const result = await api.deleteSession('session-1');

    expect(result.status).toBe('ok');
  });
});

describe('配置接口', () => {
  it('getSessionConfig 应返回配置', async () => {
    mockFetch.mockResolvedValueOnce(
      mockOk({
        sessionId: 'session-1',
        templateId: 'claude',
        workingDir: '/home',
        scope: 'project',
        files: [{ path: '.claude/settings.json', content: '{}', exists: true, hash: 'md5:abc' }],
      }),
    );

    const api = createAgentApi();
    const result = await api.getSessionConfig('session-1');

    expect(result.status).toBe('ok');
  });

  it('updateTemplateConfig 应发送 PUT', async () => {
    mockFetch.mockResolvedValueOnce(
      mockOk({
        templateId: 'claude',
        scope: 'global',
        files: [{ path: '~/.claude/settings.json', content: '{}', exists: true, hash: 'md5:abc' }],
      }),
    );

    const api = createAgentApi();
    const result = await api.updateTemplateConfig('claude', {
      files: [{ path: '~/.claude/settings.json', content: '{}', baseHash: 'md5:empty' }],
    });

    expect(result.status).toBe('ok');
  });

  it('409 冲突应保留错误信息', async () => {
    mockFetch.mockResolvedValueOnce(
      mockError(409, {
        code: 'config_conflict',
        message: '配置已变更',
        conflicts: [{ path: '~/.claude/settings.json', currentHash: 'md5:def' }],
      }),
    );

    const api = createAgentApi();
    const result = await api.updateTemplateConfig('claude', {
      files: [{ path: '~/.claude/settings.json', content: '{}', baseHash: 'md5:abc' }],
    });

    expect(result.status).toBe('error');
    expect(result.code).toBe('config_conflict');
    expect(result.conflicts?.[0].path).toBe('~/.claude/settings.json');
  });
});

describe('网络错误', () => {
  it('fetch 异常应返回 error', async () => {
    mockFetch.mockRejectedValueOnce(new TypeError('Failed to fetch'));

    const api = createAgentApi();
    const result = await api.listSessions();

    expect(result.status).toBe('error');
    expect(result.message).toContain('Failed to fetch');
  });
});
