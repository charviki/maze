import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SessionDialogs, type SessionDialogsProps } from './SessionDialogs';
import type { SessionDisplay } from './SessionList';
import type { IAgentApiClient } from '../../api';
import { ToastProvider } from '../ui/Toast';

// mock Canvas 依赖组件
vi.mock('../ui/DecryptText', () => ({
  DecryptText: ({ text }: { text: string }) => <span>{text}</span>,
}));
vi.mock('../ui/SessionPipeline', () => ({
  SessionPipeline: ({ steps }: { steps: unknown[] }) => (
    <div data-testid="session-pipeline">{steps.length} steps</div>
  ),
}));
vi.mock('./CreateSessionWithTemplateDialog', () => ({
  CreateSessionWithTemplateDialog: ({
    open,
    onSuccess,
  }: {
    open: boolean;
    onSuccess: (name: string) => void;
  }) =>
    open ? (
      <div data-testid="create-dialog">
        <button onClick={() => onSuccess('new-session')}>Create</button>
      </div>
    ) : null,
}));
vi.mock('./TemplateManager', () => ({
  TemplateManager: ({ open, onClose }: { open: boolean; onClose: () => void }) =>
    open ? (
      <div data-testid="template-manager">
        <button onClick={onClose}>Close</button>
      </div>
    ) : null,
}));
vi.mock('./NodeConfigPanel', () => ({
  NodeConfigPanel: ({ nodeName, onClose }: { nodeName: string; onClose: () => void }) => (
    <div data-testid="node-config-panel">
      {nodeName}
      <button onClick={onClose}>Close Config</button>
    </div>
  ),
}));

const mockKillTarget: SessionDisplay = {
  id: 'kill-001',
  name: 'doomed-loop',
  status: 'running',
  createdAt: '2025-01-01T00:00:00Z',
  windowCount: 1,
};

const mockRestoreTarget: SessionDisplay = {
  id: 'restore-001',
  name: 'archived-loop',
  status: 'saved',
  createdAt: '2025-01-01T00:00:00Z',
  windowCount: 0,
  savedAt: '2025-01-02T00:00:00Z',
  terminalSnapshot: 'line1\nline2',
};

function createMockApi() {
  return {
    listSessions: vi.fn(),
    createSession: vi.fn(),
    getSession: vi.fn(),
    deleteSession: vi.fn(),
    saveSessions: vi.fn(),
    getSavedSessions: vi.fn(),
    restoreSession: vi.fn(),
    getSessionConfig: vi.fn(),
    updateSessionConfig: vi.fn(),
    getLocalConfig: vi.fn(),
    updateLocalConfig: vi.fn(),
    getTemplateConfig: vi.fn(),
    updateTemplateConfig: vi.fn(),
  } as unknown as IAgentApiClient;
}

const baseProps: SessionDialogsProps = {
  killTarget: null,
  killing: false,
  restoreTarget: null,
  viewPipelineSession: null,
  viewPipelineSteps: [],
  configOpen: false,
  selectedSessionId: null,
  nodeName: 'host-alpha',
  restoring: false,
  showCreate: false,
  onShowCreateChange: vi.fn(),
  showTemplateManager: false,
  onShowTemplateManagerChange: vi.fn(),
  showNodeConfig: false,
  onShowNodeConfigClose: vi.fn(),
  apiClient: createMockApi(),
  onSessionCreated: vi.fn(),
  onKillConfirm: vi.fn(),
  onKillCancel: vi.fn(),
  onRestoreConfirm: vi.fn(),
  onRestoreCancel: vi.fn(),
  onPipelineClose: vi.fn(),
  onConfigClose: vi.fn(),
  getSnapshotSummary: (s?: string) => s?.slice(0, 50) ?? '',
};

function renderDialogs(props: Partial<SessionDialogsProps> = {}) {
  return render(
    <ToastProvider>
      <SessionDialogs {...baseProps} {...props} />
    </ToastProvider>,
  );
}

describe('SessionDialogs', () => {
  // killTarget 存在时应显示终止确认对话框，包含 session 名称
  it('should show kill confirmation dialog when killTarget is set', () => {
    renderDialogs({ killTarget: mockKillTarget });

    expect(screen.getByText('确认终止 Loop')).toBeInTheDocument();
    expect(screen.getByText(/doomed-loop/)).toBeInTheDocument();
  });

  // 点击确认终止按钮应触发 onKillConfirm 回调
  it('should call onKillConfirm when confirm button is clicked', async () => {
    const onKillConfirm = vi.fn();
    const user = userEvent.setup();

    renderDialogs({ killTarget: mockKillTarget, onKillConfirm });

    await user.click(screen.getByText('确认终止'));
    expect(onKillConfirm).toHaveBeenCalledOnce();
  });

  // killing 状态下按钮应显示"终止中..."并禁用
  it('should disable buttons when killing', () => {
    renderDialogs({ killTarget: mockKillTarget, killing: true });

    expect(screen.getByText('终止中...')).toBeDisabled();
  });

  // restoreTarget 存在时应显示恢复确认对话框
  it('should show restore confirmation dialog when restoreTarget is set', () => {
    renderDialogs({ restoreTarget: mockRestoreTarget });

    expect(screen.getByText('确认恢复 Loop')).toBeInTheDocument();
    expect(screen.getByText(/archived-loop/)).toBeInTheDocument();
  });

  // 恢复对话框应显示终端快照摘要
  it('should show terminal snapshot summary in restore dialog', () => {
    renderDialogs({
      restoreTarget: mockRestoreTarget,
      getSnapshotSummary: (s) => s?.toUpperCase() ?? '',
    });

    // pre 标签内包含换行的文本，使用函数匹配器
    expect(
      screen.getByText((_content, element) => {
        return element?.tagName === 'PRE' && element.textContent === 'LINE1\nLINE2';
      }),
    ).toBeInTheDocument();
  });

  // viewPipelineSession 存在时应显示管线查看对话框
  it('should show pipeline dialog when viewPipelineSession is set', () => {
    renderDialogs({
      viewPipelineSession: mockKillTarget,
      viewPipelineSteps: [
        {
          id: '1',
          type: 'command' as const,
          phase: 'user' as const,
          order: 0,
          value: 'echo hello',
        },
      ],
    });

    expect(screen.getByText(/doomed-loop/)).toBeInTheDocument();
    expect(screen.getByTestId('session-pipeline')).toBeInTheDocument();
  });

  // showNodeConfig 为 true 时应显示节点配置面板
  it('should show node config panel when showNodeConfig is true', () => {
    renderDialogs({ showNodeConfig: true });

    expect(screen.getByTestId('node-config-panel')).toBeInTheDocument();
  });

  // showCreate 为 true 时应显示创建 session 对话框
  it('should show create session dialog when showCreate is true', () => {
    renderDialogs({ showCreate: true });

    expect(screen.getByTestId('create-dialog')).toBeInTheDocument();
  });

  // showTemplateManager 为 true 时应显示模板管理器
  it('should show template manager when showTemplateManager is true', () => {
    renderDialogs({ showTemplateManager: true });

    expect(screen.getByTestId('template-manager')).toBeInTheDocument();
  });
});
