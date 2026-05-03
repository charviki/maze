import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { TemplateManager, ToastProvider } from '@maze/fabrication';
import type { IAgentApiClient, SessionTemplate } from '@maze/fabrication';

const builtinClaudeTemplate: SessionTemplate = {
  id: 'claude',
  name: 'Claude Code',
  command: 'IS_SANDBOX=1 claude --dangerously-skip-permissions',
  description: 'Anthropic Claude CLI Agent',
  icon: '🤖',
  builtin: true,
  defaults: {
    env: {},
    files: [
      { path: '~/.claude.json', content: '{}' },
      { path: '~/.claude/settings.json', content: '{}' },
    ],
  },
  sessionSchema: {
    envDefs: [],
    fileDefs: [
      { path: 'CLAUDE.md', label: 'Project Memory', required: false, defaultContent: '' },
      {
        path: '.claude/settings.json',
        label: 'Project Settings',
        required: false,
        defaultContent: '',
      },
    ],
  },
};

function renderTemplateManager(apiClient: IAgentApiClient) {
  render(
    <ToastProvider>
      <TemplateManager open onClose={() => {}} apiClient={apiClient} />
    </ToastProvider>,
  );
}

function createMockApi(overrides: Partial<IAgentApiClient> = {}): IAgentApiClient {
  const ok = <T,>(data: T) => Promise.resolve({ status: 'ok' as const, data });
  const noop = () => Promise.resolve({ status: 'ok' as const, data: undefined });

  return {
    listSessions: vi.fn(() => ok([])),
    createSession: vi.fn(noop),
    getSession: vi.fn(noop),
    deleteSession: vi.fn(noop),
    getOutput: vi.fn(noop),
    sendInput: vi.fn(noop),
    sendSignal: vi.fn(noop),
    getSavedSessions: vi.fn(() => ok([])),
    restoreSession: vi.fn(noop),
    saveSessions: vi.fn(() => ok({ savedAt: new Date().toISOString() })),
    buildWsUrl: vi.fn(() => 'ws://localhost/test'),
    listTemplates: vi.fn(() => ok([builtinClaudeTemplate])),
    createTemplate: vi.fn((tpl: SessionTemplate) => ok(tpl)),
    getTemplate: vi.fn((id: string) => ok({ ...builtinClaudeTemplate, id })),
    getTemplateConfig: vi.fn(() =>
      ok({
        templateId: 'claude',
        scope: 'global' as const,
        files: [
          {
            path: '~/.claude.json',
            content: '{"hasCompletedOnboarding":true}',
            exists: true,
            hash: 'md5:claude-json',
          },
          { path: '~/.claude/settings.json', content: '{}', exists: true, hash: 'md5:settings' },
        ],
      }),
    ),
    updateTemplate: vi.fn((id: string, tpl: SessionTemplate) => ok({ ...tpl, id })),
    updateTemplateConfig: vi.fn(() =>
      ok({
        templateId: 'claude',
        scope: 'global' as const,
        files: [
          {
            path: '~/.claude.json',
            content: '{"hasCompletedOnboarding":true}',
            exists: true,
            hash: 'md5:claude-json',
          },
          {
            path: '~/.claude/settings.json',
            content: '{"theme":"dark"}',
            exists: true,
            hash: 'md5:new-settings',
          },
        ],
      }),
    ),
    deleteTemplate: vi.fn(noop),
    getSessionConfig: vi.fn(noop),
    updateSessionConfig: vi.fn(noop),
    getLocalConfig: vi.fn(() => ok({ workingDir: '/home/agent', env: {} })),
    updateLocalConfig: vi.fn((cfg) => ok({ workingDir: '/home/agent', env: cfg.env || {} })),
    ...overrides,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('TemplateManager', () => {
  it('编辑真实全局配置后点击保存，会先弹出确认框', async () => {
    const apiClient = createMockApi();
    renderTemplateManager(apiClient);

    fireEvent.click(await screen.findByTitle('编辑模板'));

    const configTextarea = await screen.findByDisplayValue('{}');
    fireEvent.change(configTextarea, { target: { value: '{"theme":"dark"}' } });
    fireEvent.click(screen.getByText('保存'));

    expect(await screen.findByText('确认保存真实全局配置')).toBeInTheDocument();
    expect(
      screen.getByText(
        '本次保存会直接修改节点上的真实全局配置文件。若这些文件在你编辑期间已被其他来源修改，系统会拒绝覆盖并要求重新加载。',
      ),
    ).toBeInTheDocument();
  });

  it('真实全局配置保存冲突时，会提示重新加载', async () => {
    const updateTemplateConfig = vi.fn(() =>
      Promise.resolve({
        status: 'error' as const,
        code: 'config_conflict',
        message: '配置已变更，请重新加载后再修改',
        conflicts: [{ path: '~/.claude/settings.json', currentHash: 'md5:conflict' }],
      }),
    );
    const apiClient = createMockApi({ updateTemplateConfig });
    renderTemplateManager(apiClient);

    fireEvent.click(await screen.findByTitle('编辑模板'));

    const configTextarea = await screen.findByDisplayValue('{}');
    fireEvent.change(configTextarea, { target: { value: '{"theme":"dark"}' } });
    fireEvent.click(screen.getByText('保存'));
    fireEvent.click(await screen.findByText('确认保存'));

    await waitFor(() => {
      expect(
        screen.getAllByText('配置已变更，请重新加载后再修改：~/.claude/settings.json').length,
      ).toBeGreaterThan(0);
    });
    expect(updateTemplateConfig).toHaveBeenCalledTimes(1);
  });
});
