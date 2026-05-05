import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { LucideIcon } from 'lucide-react';
import { ModuleCard } from './ModuleCard';

// mock @maze/fabrication 中的复杂 UI 组件，只保留可测试的结构
vi.mock('@maze/fabrication', () => ({
  Panel: ({
    children,
    onClick,
    tabIndex,
    role,
    onKeyDown,
  }: {
    children: React.ReactNode;
    onClick?: React.MouseEventHandler;
    tabIndex?: number;
    role?: string;
    onKeyDown?: React.KeyboardEventHandler;
  }) => (
    <div
      data-testid="panel"
      onClick={onClick}
      onKeyDown={onKeyDown}
      tabIndex={tabIndex}
      role={role}
    >
      {children}
    </div>
  ),
  DecryptText: ({ text }: { text: string }) => <span>{text}</span>,
  ReverieEffect: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="reverie">{children}</div>
  ),
  GlitchEffect: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="glitch">{children}</div>
  ),
  HostVitalSign: () => <span data-testid="vital-sign">ONLINE</span>,
  cn: (...args: (string | undefined | null | false)[]) => args.filter(Boolean).join(' '),
}));

describe('ModuleCard', () => {
  const defaultProps = {
    name: 'Director Console',
    description: '测试描述',
    icon: vi.fn(({ className }: { className?: string }) => (
      <svg data-testid="icon" className={className} />
    )) as unknown as LucideIcon,
    status: 'online' as const,
  };

  // 基础渲染：应显示模块名称和描述文本
  it('should render module name and description', () => {
    render(<ModuleCard {...defaultProps} />);
    expect(screen.getByText('Director Console')).toBeInTheDocument();
    expect(screen.getByText('测试描述')).toBeInTheDocument();
  });

  // online 状态下应渲染 HostVitalSign 组件
  it('should show vital sign when status is online', () => {
    render(<ModuleCard {...defaultProps} status="online" />);
    expect(screen.getByTestId('vital-sign')).toBeInTheDocument();
  });

  // locked 状态下应显示 LOCKED 标签，表示模块不可用
  it('should show LOCKED label when status is locked', () => {
    render(<ModuleCard {...defaultProps} status="locked" />);
    expect(screen.getByText('LOCKED')).toBeInTheDocument();
  });

  // locked 状态下应显示拒绝访问的提示文案
  it('should show locked message when status is locked', () => {
    render(<ModuleCard {...defaultProps} status="locked" />);
    expect(screen.getByText('THE MAZE IS NOT MEANT FOR YOU')).toBeInTheDocument();
  });

  // online 状态下点击卡片应在新的浏览器标签中打开 href 链接
  it('should open href in new tab when clicked and not locked', async () => {
    const openSpy = vi.spyOn(window, 'open').mockImplementation(() => null);
    const user = userEvent.setup();

    render(<ModuleCard {...defaultProps} href="http://example.com" />);

    await user.click(screen.getByTestId('panel'));
    expect(openSpy).toHaveBeenCalledWith('http://example.com', '_blank');

    openSpy.mockRestore();
  });

  // locked 状态下点击卡片不应打开任何链接
  it('should not open href when locked', async () => {
    const openSpy = vi.spyOn(window, 'open').mockImplementation(() => null);
    const user = userEvent.setup();

    render(<ModuleCard {...defaultProps} status="locked" href="http://example.com" />);

    await user.click(screen.getByTestId('panel'));
    expect(openSpy).not.toHaveBeenCalled();

    openSpy.mockRestore();
  });

  // locked 状态应使用 GlitchEffect 包裹（故障视觉效果）
  it('should render with GlitchEffect when locked', () => {
    render(<ModuleCard {...defaultProps} status="locked" />);
    expect(screen.getByTestId('glitch')).toBeInTheDocument();
  });

  // online 状态应使用 ReverieEffect 包裹（呼吸灯视觉效果）
  it('should render with ReverieEffect when online', () => {
    render(<ModuleCard {...defaultProps} status="online" />);
    expect(screen.getByTestId('reverie')).toBeInTheDocument();
  });
});
