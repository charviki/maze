import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { NodeList } from './NodeList';
import { ToastProvider } from '@maze/fabrication';

// mock @maze/fabrication 中 Canvas 相关组件和轮询 hook
vi.mock('@maze/fabrication', async (importOriginal) => {
  const actual = await importOriginal<Record<string, unknown>>();
  return {
    ...actual,
    usePollingWithBackoff: vi.fn(),
    DecryptText: ({ text }: { text: string }) => <span>{text}</span>,
    GlitchEffect: ({ children }: { children: React.ReactNode }) => (
      <div data-testid="glitch">{children}</div>
    ),
    Skeleton: ({ className }: { className?: string }) => (
      <div data-testid="skeleton" className={className} />
    ),
    clipPathHalf: () => 'polygon(0 0, 100% 0, 100% 100%, 0 100%)',
    useToast: () => ({ showToast: vi.fn() }),
  };
});

// mock controller API，避免真实网络请求
vi.mock('../api/controller', () => ({
  controllerApi: {
    listHosts: vi.fn(),
    deleteHost: vi.fn(),
  },
}));

import { usePollingWithBackoff } from '@maze/fabrication';

const mockUsePolling = usePollingWithBackoff as unknown as ReturnType<typeof vi.fn>;

function renderNodeList(props = {}) {
  return render(
    <ToastProvider>
      <NodeList onSelectNode={vi.fn()} selectedNodeName={null} {...props} />
    </ToastProvider>,
  );
}

describe('NodeList', () => {
  // 加载中状态应显示 Skeleton 骨架屏
  it('should show skeleton loading state', () => {
    mockUsePolling.mockReturnValue({
      data: null,
      error: null,
      isLoading: true,
      refresh: vi.fn(),
    });

    renderNodeList();

    const skeletons = screen.getAllByTestId('skeleton');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  // 数据为空时应显示空状态提示文案
  it('should show empty state when no hosts', () => {
    mockUsePolling.mockReturnValue({
      data: [],
      error: null,
      isLoading: false,
      refresh: vi.fn(),
    });

    renderNodeList();

    expect(screen.getByText('[ NO HOSTS DETECTED ]')).toBeInTheDocument();
  });

  // 有数据时应渲染每个 host 的名称
  it('should render host list when data is available', () => {
    const mockHosts = [
      {
        name: 'host-alpha',
        status: 'online',
        sessionCount: 3,
        address: '10.0.0.1',
        lastHeartbeat: new Date().toISOString(),
      },
      {
        name: 'host-beta',
        status: 'offline',
        sessionCount: 0,
        address: '10.0.0.2',
        lastHeartbeat: new Date().toISOString(),
      },
    ];

    mockUsePolling.mockReturnValue({
      data: mockHosts,
      error: null,
      isLoading: false,
      refresh: vi.fn(),
    });

    renderNodeList();

    expect(screen.getByText('host-alpha')).toBeInTheDocument();
    expect(screen.getByText('host-beta')).toBeInTheDocument();
  });

  // online 状态的 host 应显示其 session 数量（LOOPS 计数）
  it('should show session count for online hosts', () => {
    const mockHosts = [
      {
        name: 'host-alpha',
        status: 'online',
        sessionCount: 5,
        address: '10.0.0.1',
        lastHeartbeat: new Date().toISOString(),
      },
    ];

    mockUsePolling.mockReturnValue({
      data: mockHosts,
      error: null,
      isLoading: false,
      refresh: vi.fn(),
    });

    renderNodeList();

    expect(screen.getByText('5 LOOPS')).toBeInTheDocument();
  });
});
