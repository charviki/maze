import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createAgentApi } from './agent';

const mockFetch = vi.fn();
vi.stubGlobal('fetch', mockFetch);

beforeEach(() => {
  mockFetch.mockReset();
});

const agentApi = createAgentApi('', 'agent-1');

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

describe('agent API - 通过 Manager 代理成功请求', () => {
  it('listSessions 应请求代理路径', async () => {
    const sessions = [
      { id: 's1', name: 's1', status: 'running', createdAt: '2024-01-01', windowCount: 1 },
    ];
    mockFetch.mockResolvedValueOnce(mockOk({ sessions }));

    const result = await agentApi.listSessions();

    expect(result.status).toBe('ok');
    expect(result.data).toHaveLength(1);
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/nodes/agent-1/sessions'),
      expect.anything(),
    );
  });

  it('createSession 应发送 POST', async () => {
    const newSession = {
      id: 'new',
      name: 'new',
      status: 'running',
      createdAt: '2024-01-01',
      windowCount: 1,
    };
    mockFetch.mockResolvedValueOnce(mockOk(newSession));

    const result = await agentApi.createSession({ name: 'new', restoreStrategy: 'auto' });

    expect(result.status).toBe('ok');
    expect(mockFetch).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({ method: 'POST' }),
    );
  });

  it('getSession 应请求指定 session', async () => {
    const session = {
      id: 'abc',
      name: 'abc',
      status: 'running',
      createdAt: '2024-01-01',
      windowCount: 1,
    };
    mockFetch.mockResolvedValueOnce(mockOk(session));

    const result = await agentApi.getSession('abc');

    expect(result.status).toBe('ok');
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/nodes/agent-1/sessions/abc'),
      expect.anything(),
    );
  });

  it('saveSessions 应发送 POST', async () => {
    mockFetch.mockResolvedValueOnce(mockOk({ savedAt: '2025-04-21T12:00:00Z' }));

    const result = await agentApi.saveSessions();

    expect(result.status).toBe('ok');
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/save'),
      expect.objectContaining({ method: 'POST' }),
    );
  });

  it('getSavedSessions 应返回保存的会话', async () => {
    const savedSessions = [
      {
        sessionName: 'saved-1',
        pipeline: '',
        restoreStrategy: 'manual',
        workingDir: '/home',
        terminalSnapshot: '',
        savedAt: '2025-01-01',
      },
    ];
    mockFetch.mockResolvedValueOnce(mockOk({ sessions: savedSessions }));

    const result = await agentApi.getSavedSessions();

    expect(result.status).toBe('ok');
    expect(result.data).toHaveLength(1);
  });

  it('restoreSession 应发送 POST', async () => {
    mockFetch.mockResolvedValueOnce(mockOk({}));

    const result = await agentApi.restoreSession('saved-1');

    expect(result.status).toBe('ok');
  });

  it('deleteSession 应发送 DELETE', async () => {
    mockFetch.mockResolvedValueOnce(mockOk({}));

    const result = await agentApi.deleteSession('old-session');

    expect(result.status).toBe('ok');
  });

  it('getSessionConfig 应返回配置', async () => {
    mockFetch.mockResolvedValueOnce(
      mockOk({
        sessionId: 'sess-1',
        templateId: 'claude',
        workingDir: '/home',
        scope: 'project',
        files: [{ path: '.claude/settings.json', content: '{}', exists: true, hash: 'md5:abc' }],
      }),
    );

    const result = await agentApi.getSessionConfig('sess-1');

    expect(result.status).toBe('ok');
  });

  it('updateSessionConfig 应发送 PUT', async () => {
    mockFetch.mockResolvedValueOnce(
      mockOk({
        sessionId: 'sess-1',
        templateId: 'claude',
        workingDir: '/home',
        scope: 'project',
        files: [
          {
            path: '.claude/settings.json',
            content: '{"theme":"dark"}',
            exists: true,
            hash: 'md5:def',
          },
        ],
      }),
    );

    const result = await agentApi.updateSessionConfig('sess-1', {
      files: [{ path: '.claude/settings.json', content: '{"theme":"dark"}', baseHash: 'md5:abc' }],
    });

    expect(result.status).toBe('ok');
    expect(mockFetch).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({ method: 'PUT' }),
    );
  });

  it('getTemplateConfig 应返回模板配置', async () => {
    mockFetch.mockResolvedValueOnce(
      mockOk({
        templateId: 'claude',
        scope: 'global',
        files: [{ path: '~/.claude/settings.json', content: '{}', exists: true, hash: 'md5:abc' }],
      }),
    );

    const result = await agentApi.getTemplateConfig('claude');

    expect(result.status).toBe('ok');
  });
});

describe('agent API - HTTP 错误', () => {
  it('500 应返回 error', async () => {
    mockFetch.mockResolvedValueOnce(mockError(500));

    const result = await agentApi.listSessions();

    expect(result.status).toBe('error');
    expect(result.message).toBe('HTTP 500');
  });

  it('404 应返回 error', async () => {
    mockFetch.mockResolvedValueOnce(mockError(404));

    const result = await agentApi.getSession('nonexistent');

    expect(result.status).toBe('error');
    expect(result.message).toBe('HTTP 404');
  });

  it('409 应保留 code 和 conflicts', async () => {
    mockFetch.mockResolvedValueOnce(
      mockError(409, {
        code: 'config_conflict',
        message: '配置已变更',
        conflicts: [{ path: '.claude/settings.json', currentHash: 'md5:def' }],
      }),
    );

    const result = await agentApi.updateSessionConfig('sess-1', {
      files: [{ path: '.claude/settings.json', content: '{}', baseHash: 'md5:abc' }],
    });

    expect(result.status).toBe('error');
    expect(result.code).toBe('config_conflict');
    expect(result.conflicts?.[0].path).toBe('.claude/settings.json');
  });
});

describe('agent API - 网络错误', () => {
  it('fetch 异常应返回 error', async () => {
    mockFetch.mockRejectedValueOnce(new TypeError('Failed to fetch'));

    const result = await agentApi.listSessions();

    expect(result.status).toBe('error');
    expect(result.message).toContain('Failed to fetch');
  });
});

describe('agent API - 代理路径验证', () => {
  it('不同 nodeName 应生成不同路径', async () => {
    const api1 = createAgentApi('', 'agent-alpha');
    const api2 = createAgentApi('', 'agent-beta');

    mockFetch.mockResolvedValue(mockOk({ sessions: [] }));

    await api1.listSessions();
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/nodes/agent-alpha/sessions'),
      expect.anything(),
    );

    await api2.listSessions();
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/nodes/agent-beta/sessions'),
      expect.anything(),
    );
  });

  it('nodeName 含特殊字符应被编码', async () => {
    const api = createAgentApi('', 'agent name');
    mockFetch.mockResolvedValueOnce(mockOk({ sessions: [] }));

    await api.listSessions();
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/nodes/agent%20name/sessions'),
      expect.anything(),
    );
  });
});
