import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SessionList, type SessionDisplay } from './SessionList';

// mock Canvas 依赖组件和工具函数
vi.mock('../ui/DecryptText', () => ({
  DecryptText: ({ text }: { text: string }) => <span>{text}</span>,
}));
vi.mock('../ui/HostVitalSign', () => ({
  HostVitalSign: ({ status }: { status: string }) => <span data-testid="vital-sign">{status}</span>,
}));
vi.mock('../ui/ReverieEffect', () => ({
  ReverieEffect: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));
vi.mock('../../utils', async (importOriginal) => {
  const actual = await importOriginal<Record<string, unknown>>();
  return {
    ...actual,
    clipPathHalf: () => 'polygon(0 0, 100% 0, 100% 100%, 0 100%)',
  };
});

const baseSession: SessionDisplay = {
  id: 'session-001-abc',
  name: 'test-loop',
  status: 'running',
  createdAt: '2025-01-01T00:00:00Z',
  windowCount: 1,
};

const defaultProps = {
  sessions: [baseSession] as SessionDisplay[],
  search: '',
  onSearchChange: vi.fn(),
  selectedSessionId: null as string | null,
  onSelectSession: vi.fn(),
  nodeName: 'host-alpha',
  onCreateClick: vi.fn(),
  onNodeConfigClick: vi.fn(),
  onKill: vi.fn(),
  onRestore: vi.fn(),
  onViewPipeline: vi.fn(),
  isLoading: false,
};

describe('SessionList', () => {
  // 应显示节点名称和 "Narrative Loops" 标题
  it('should render node name in header', () => {
    render(<SessionList {...defaultProps} />);
    expect(screen.getByText(/host-alpha \/\/ Narrative Loops/)).toBeInTheDocument();
  });

  // 应渲染搜索框，允许用户过滤 session 列表
  it('should render search input', () => {
    render(<SessionList {...defaultProps} />);
    expect(screen.getByPlaceholderText('SEARCH LOOPS...')).toBeInTheDocument();
  });

  // 应显示每个 session 的名称
  it('should render session names', () => {
    render(<SessionList {...defaultProps} />);
    expect(screen.getByText('test-loop')).toBeInTheDocument();
  });

  // 选中状态的 session 应使用 DecryptText 渲染名称（高亮效果）
  it('should render selected session name with DecryptText', () => {
    render(<SessionList {...defaultProps} selectedSessionId="session-001-abc" />);
    expect(screen.getByText('test-loop')).toBeInTheDocument();
  });

  // 点击 session 项应触发 onSelectSession 回调
  it('should call onSelectSession when session item is clicked', async () => {
    const onSelectSession = vi.fn();
    const user = userEvent.setup();

    render(<SessionList {...defaultProps} onSelectSession={onSelectSession} />);

    await user.click(screen.getByText('test-loop'));
    expect(onSelectSession).toHaveBeenCalledWith('session-001-abc');
  });

  // 搜索过滤：输入关键词后应只显示匹配的 session
  it('should filter sessions by search term', () => {
    const sessions = [
      { ...baseSession, id: 'alpha-001', name: 'alpha-loop' },
      { ...baseSession, id: 'beta-002', name: 'beta-loop' },
    ];

    render(<SessionList {...defaultProps} sessions={sessions} search="alpha" />);

    expect(screen.getByText('alpha-loop')).toBeInTheDocument();
    expect(screen.queryByText('beta-loop')).not.toBeInTheDocument();
  });

  // 空列表状态应显示 "NO LOOPS ACTIVE" 提示
  it('should show empty state when no sessions', () => {
    render(<SessionList {...defaultProps} sessions={[]} />);
    expect(screen.getByText('[ NO LOOPS ACTIVE ]')).toBeInTheDocument();
  });

  // 加载中状态应显示 Skeleton 骨架屏
  it('should show skeleton when loading', () => {
    const { container } = render(<SessionList {...defaultProps} sessions={[]} isLoading />);
    const skeletons = container.querySelectorAll('.animate-pulse');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  // saved 状态的 session 应显示 [SAVED] 标签
  it('should show SAVED badge for saved sessions', () => {
    const savedSession = {
      ...baseSession,
      status: 'saved' as const,
      savedAt: '2025-01-01T00:00:00Z',
    };
    render(<SessionList {...defaultProps} sessions={[savedSession]} />);
    expect(screen.getByText('[ SAVED ]')).toBeInTheDocument();
  });

  // 点击 Add Loop 按钮应触发 onCreateClick
  it('should call onCreateClick when Add Loop button is clicked', async () => {
    const onCreateClick = vi.fn();
    const user = userEvent.setup();

    render(<SessionList {...defaultProps} onCreateClick={onCreateClick} />);

    await user.click(screen.getByTitle('Add Loop'));
    expect(onCreateClick).toHaveBeenCalledOnce();
  });

  // 点击 Node Config 按钮应触发 onNodeConfigClick
  it('should call onNodeConfigClick when Node Config button is clicked', async () => {
    const onNodeConfigClick = vi.fn();
    const user = userEvent.setup();

    render(<SessionList {...defaultProps} onNodeConfigClick={onNodeConfigClick} />);

    await user.click(screen.getByTitle('Node Config'));
    expect(onNodeConfigClick).toHaveBeenCalledOnce();
  });
});
