import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { NodeConfigPanel } from './NodeConfigPanel';
import type { IAgentApiClient } from '../../api';
import { ToastProvider } from '../ui/Toast';

// mock Toast 的 useToast hook
vi.mock('../ui/Toast', async (importOriginal) => {
  const actual = await importOriginal<Record<string, unknown>>();
  return {
    ...actual,
    useToast: () => ({ showToast: vi.fn() }),
  };
});

function createMockApi(overrides: Partial<IAgentApiClient> = {}): IAgentApiClient {
  const ok = <T,>(data: T) => Promise.resolve({ status: 'ok' as const, data });
  return {
    getLocalConfig: vi.fn(() =>
      ok({ workingDir: '/opt/agent', env: { NODE_ENV: 'production', PORT: '8080' } }),
    ),
    updateLocalConfig: vi.fn(() => ok(undefined)),
    ...overrides,
  } as unknown as IAgentApiClient;
}

function renderPanel(overrides: Partial<IAgentApiClient> = {}) {
  const apiClient = createMockApi(overrides);
  return render(
    <ToastProvider>
      <NodeConfigPanel nodeName="host-alpha" apiClient={apiClient} onClose={vi.fn()} />
    </ToastProvider>,
  );
}

describe('NodeConfigPanel', () => {
  // 挂载后应显示节点名称和配置标题
  it('should render node name in header', async () => {
    renderPanel();

    expect(await screen.findByText(/host-alpha/)).toBeInTheDocument();
  });

  // 应显示从 API 加载的工作目录
  it('should display working directory from API', async () => {
    renderPanel();

    expect(await screen.findByDisplayValue('/opt/agent')).toBeInTheDocument();
  });

  // 应显示从 API 加载的环境变量键值对
  it('should display environment variables from API', async () => {
    renderPanel();

    expect(await screen.findByDisplayValue('NODE_ENV')).toBeInTheDocument();
    expect(await screen.findByDisplayValue('production')).toBeInTheDocument();
    expect(await screen.findByDisplayValue('PORT')).toBeInTheDocument();
    expect(await screen.findByDisplayValue('8080')).toBeInTheDocument();
  });

  // API 加载失败时组件仍应渲染（不会崩溃），loading 消失后显示空配置
  it('should handle API failure gracefully', async () => {
    const mockApi: Partial<IAgentApiClient> = {
      getLocalConfig: vi.fn(() => Promise.reject(new Error('Network error'))),
      updateLocalConfig: vi.fn(),
    };

    render(
      <ToastProvider>
        <NodeConfigPanel
          nodeName="host-alpha"
          apiClient={mockApi as IAgentApiClient}
          onClose={vi.fn()}
        />
      </ToastProvider>,
    );

    // 组件应显示标题（不会因 API 错误而崩溃）
    expect(await screen.findByText(/host-alpha/)).toBeInTheDocument();
    // loading 结束后应显示空的 env 配置
    expect(screen.getByText('暂无环境变量配置')).toBeInTheDocument();
  });

  // 点击新增按钮应添加一行空的环境变量输入框
  it('should add new env row when add button is clicked', async () => {
    const user = userEvent.setup();
    renderPanel();

    await screen.findByDisplayValue('NODE_ENV');

    await user.click(screen.getByText('新增'));

    // 原有 2 个 + 新增 1 个 = 3 个 KEY 输入框
    const keyInputs = screen.getAllByPlaceholderText('KEY');
    expect(keyInputs).toHaveLength(3);
  });

  // 点击取消按钮应触发 onClose 回调
  it('should call onClose when cancel button is clicked', async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();

    const apiClient = createMockApi();
    render(
      <ToastProvider>
        <NodeConfigPanel nodeName="host-alpha" apiClient={apiClient} onClose={onClose} />
      </ToastProvider>,
    );

    await screen.findByDisplayValue('/opt/agent');

    await user.click(screen.getByText('取消'));
    expect(onClose).toHaveBeenCalledOnce();
  });

  // 点击保存按钮应调用 updateLocalConfig API
  it('should call updateLocalConfig when save button is clicked', async () => {
    const onClose = vi.fn();
    const updateSpy = vi.fn().mockResolvedValue({ status: 'ok', data: undefined });
    const apiClient = createMockApi({ updateLocalConfig: updateSpy });
    const user = userEvent.setup();

    render(
      <ToastProvider>
        <NodeConfigPanel nodeName="host-alpha" apiClient={apiClient} onClose={onClose} />
      </ToastProvider>,
    );

    await screen.findByDisplayValue('/opt/agent');

    await user.click(screen.getByText('保存'));
    expect(updateSpy).toHaveBeenCalledOnce();
    expect(onClose).toHaveBeenCalledOnce();
  });
});
