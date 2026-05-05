import { describe, it, expect } from 'vitest';
import { http, HttpResponse } from 'msw';
import { createAgentApi } from './agent';
import { server } from '../test/mocks/server';

const agentApi = createAgentApi('', 'agent-1');

describe('agent API - 通过 Manager 代理成功请求', () => {
  it('listSessions 应请求代理路径', async () => {
    const sessions = [
      { id: 's1', name: 's1', status: 'running', createdAt: '2024-01-01', windowCount: 1 },
    ];
    server.use(http.get('*/api/v1/nodes/agent-1/sessions', () => HttpResponse.json({ sessions })));

    const result = await agentApi.listSessions();

    expect(result.status).toBe('ok');
    expect(result.data).toHaveLength(1);
  });

  it('createSession 应发送 POST', async () => {
    const newSession = {
      id: 'new',
      name: 'new',
      status: 'running',
      createdAt: '2024-01-01',
      windowCount: 1,
    };
    server.use(http.post('*/api/v1/nodes/agent-1/sessions', () => HttpResponse.json(newSession)));

    const result = await agentApi.createSession({ name: 'new', restoreStrategy: 'auto' });

    expect(result.status).toBe('ok');
  });

  it('getSession 应请求指定 session', async () => {
    const session = {
      id: 'abc',
      name: 'abc',
      status: 'running',
      createdAt: '2024-01-01',
      windowCount: 1,
    };
    server.use(http.get('*/api/v1/nodes/agent-1/sessions/abc', () => HttpResponse.json(session)));

    const result = await agentApi.getSession('abc');

    expect(result.status).toBe('ok');
  });

  it('saveSessions 应发送 POST', async () => {
    server.use(
      http.post('*/api/v1/nodes/agent-1/sessions/save', () =>
        HttpResponse.json({ savedAt: '2025-04-21T12:00:00Z' }),
      ),
    );

    const result = await agentApi.saveSessions();

    expect(result.status).toBe('ok');
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
    server.use(
      http.get('*/api/v1/nodes/agent-1/sessions/saved', () =>
        HttpResponse.json({ sessions: savedSessions }),
      ),
    );

    const result = await agentApi.getSavedSessions();

    expect(result.status).toBe('ok');
    expect(result.data).toHaveLength(1);
  });

  it('restoreSession 应发送 POST', async () => {
    server.use(
      http.post('*/api/v1/nodes/agent-1/sessions/saved-1/restore', () => HttpResponse.json({})),
    );

    const result = await agentApi.restoreSession('saved-1');

    expect(result.status).toBe('ok');
  });

  it('deleteSession 应发送 DELETE', async () => {
    server.use(
      http.delete('*/api/v1/nodes/agent-1/sessions/old-session', () => HttpResponse.json({})),
    );

    const result = await agentApi.deleteSession('old-session');

    expect(result.status).toBe('ok');
  });

  it('getSessionConfig 应返回配置', async () => {
    server.use(
      http.get('*/api/v1/nodes/agent-1/sessions/sess-1/config', () =>
        HttpResponse.json({
          sessionId: 'sess-1',
          templateId: 'claude',
          workingDir: '/home',
          scope: 'project',
          files: [{ path: '.claude/settings.json', content: '{}', exists: true, hash: 'md5:abc' }],
        }),
      ),
    );

    const result = await agentApi.getSessionConfig('sess-1');

    expect(result.status).toBe('ok');
  });

  it('updateSessionConfig 应发送 PUT', async () => {
    server.use(
      http.put('*/api/v1/nodes/agent-1/sessions/sess-1/config', () =>
        HttpResponse.json({
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
      ),
    );

    const result = await agentApi.updateSessionConfig('sess-1', {
      files: [{ path: '.claude/settings.json', content: '{"theme":"dark"}', baseHash: 'md5:abc' }],
    });

    expect(result.status).toBe('ok');
  });

  it('getTemplateConfig 应返回模板配置', async () => {
    server.use(
      http.get('*/api/v1/nodes/agent-1/templates/claude/config', () =>
        HttpResponse.json({
          templateId: 'claude',
          scope: 'global',
          files: [
            { path: '~/.claude/settings.json', content: '{}', exists: true, hash: 'md5:abc' },
          ],
        }),
      ),
    );

    const result = await agentApi.getTemplateConfig('claude');

    expect(result.status).toBe('ok');
  });
});

describe('agent API - HTTP 错误', () => {
  it('500 应返回 error', async () => {
    server.use(
      http.get('*/api/v1/nodes/agent-1/sessions', () => new HttpResponse(null, { status: 500 })),
    );

    const result = await agentApi.listSessions();

    expect(result.status).toBe('error');
    expect(result.message).toBe('HTTP 500');
  });

  it('404 应返回 error', async () => {
    server.use(
      http.get(
        '*/api/v1/nodes/agent-1/sessions/nonexistent',
        () => new HttpResponse(null, { status: 404 }),
      ),
    );

    const result = await agentApi.getSession('nonexistent');

    expect(result.status).toBe('error');
    expect(result.message).toBe('HTTP 404');
  });

  it('409 应保留 code 和 conflicts', async () => {
    server.use(
      http.put('*/api/v1/nodes/agent-1/sessions/sess-1/config', () =>
        HttpResponse.json(
          {
            code: 'config_conflict',
            message: '配置已变更',
            conflicts: [{ path: '.claude/settings.json', currentHash: 'md5:def' }],
          },
          { status: 409 },
        ),
      ),
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
    server.use(http.get('*/api/v1/nodes/agent-1/sessions', () => HttpResponse.error()));

    const result = await agentApi.listSessions();

    expect(result.status).toBe('error');
  });
});

describe('agent API - 代理路径验证', () => {
  // 验证不同 nodeName 生成的 API 路径确实不同，
  // 防止实现错误地忽略 nodeName 导致所有请求打到同一个 handler
  it('不同 nodeName 应生成不同路径', async () => {
    let alphaHit = false;
    let betaHit = false;

    server.use(
      http.get('*/api/v1/nodes/agent-alpha/sessions', () => {
        alphaHit = true;
        return HttpResponse.json({ sessions: [] });
      }),
      http.get('*/api/v1/nodes/agent-beta/sessions', () => {
        betaHit = true;
        return HttpResponse.json({ sessions: [] });
      }),
    );

    const api1 = createAgentApi('', 'agent-alpha');
    const api2 = createAgentApi('', 'agent-beta');

    await api1.listSessions();
    await api2.listSessions();

    // 两个 handler 必须分别被命中，证明路径确实不同
    expect(alphaHit).toBe(true);
    expect(betaHit).toBe(true);
  });

  // 验证 nodeName 中的空格被 encodeURIComponent 编码为 %20，
  // 防止实现遗漏编码步骤（直接拼接会导致请求路径包含裸空格）
  it('nodeName 含特殊字符应被编码', async () => {
    let hitUrl = '';

    server.use(
      http.get('*/api/v1/nodes/:nodeName/sessions', ({ request }) => {
        hitUrl = new URL(request.url).pathname;
        return HttpResponse.json({ sessions: [] });
      }),
    );

    const api = createAgentApi('', 'agent name');
    const result = await api.listSessions();

    expect(result.status).toBe('ok');
    // 编码后的路径应包含 %20 而非裸空格
    expect(hitUrl).toContain('agent%20name');
    expect(hitUrl).not.toContain('agent name');
  });
});
