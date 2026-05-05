import { describe, it, expect } from 'vitest';
import { http, HttpResponse } from 'msw';
import { createAgentApi } from './client';
import { server } from '../test/mocks/server';

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
    server.use(http.get('*/api/v1/sessions', () => HttpResponse.json({ sessions: mockData })));

    const api = createAgentApi();
    const result = await api.listSessions();

    expect(result.status).toBe('ok');
    expect(result.data).toHaveLength(1);
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
    server.use(http.post('*/api/v1/sessions', () => HttpResponse.json(newSession)));

    const api = createAgentApi();
    const result = await api.createSession({ name: 'new-1', command: 'bash' });

    expect(result.status).toBe('ok');
  });
});

describe('api.deleteSession', () => {
  it('应发送 DELETE', async () => {
    server.use(http.delete('*/api/v1/sessions/session-1', () => HttpResponse.json({})));

    const api = createAgentApi();
    const result = await api.deleteSession('session-1');

    expect(result.status).toBe('ok');
  });
});

describe('配置接口', () => {
  it('getSessionConfig 应返回配置', async () => {
    server.use(
      http.get('*/api/v1/sessions/session-1/config', () =>
        HttpResponse.json({
          sessionId: 'session-1',
          templateId: 'claude',
          workingDir: '/home',
          scope: 'project',
          files: [{ path: '.claude/settings.json', content: '{}', exists: true, hash: 'md5:abc' }],
        }),
      ),
    );

    const api = createAgentApi();
    const result = await api.getSessionConfig('session-1');

    expect(result.status).toBe('ok');
  });

  it('updateTemplateConfig 应发送 PUT', async () => {
    server.use(
      http.put('*/api/v1/templates/claude/config', () =>
        HttpResponse.json({
          templateId: 'claude',
          scope: 'global',
          files: [
            { path: '~/.claude/settings.json', content: '{}', exists: true, hash: 'md5:abc' },
          ],
        }),
      ),
    );

    const api = createAgentApi();
    const result = await api.updateTemplateConfig('claude', {
      files: [{ path: '~/.claude/settings.json', content: '{}', baseHash: 'md5:empty' }],
    });

    expect(result.status).toBe('ok');
  });

  it('409 冲突应保留错误信息', async () => {
    server.use(
      http.put('*/api/v1/templates/claude/config', () =>
        HttpResponse.json(
          {
            code: 'config_conflict',
            message: '配置已变更',
            conflicts: [{ path: '~/.claude/settings.json', currentHash: 'md5:def' }],
          },
          { status: 409 },
        ),
      ),
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
    server.use(http.get('*/api/v1/sessions', () => HttpResponse.error()));

    const api = createAgentApi();
    const result = await api.listSessions();

    expect(result.status).toBe('error');
  });
});
