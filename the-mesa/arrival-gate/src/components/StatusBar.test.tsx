import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { StatusBar } from './StatusBar';

const mockClockFn = vi.fn();

// DecryptText 依赖 Canvas，useClock 用 mock 替代
vi.mock('@maze/fabrication', () => ({
  DecryptText: ({ text }: { text: string }) => <span>{text}</span>,
  useClock: () => mockClockFn(),
}));

describe('StatusBar', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    mockClockFn.mockReturnValue('12:00:00');
  });

  afterEach(() => {
    vi.useRealTimers();
    mockClockFn.mockReset();
  });

  // 状态栏应显示 NODES 计数指标
  it('should render NODES count', () => {
    render(<StatusBar />);
    expect(screen.getByText(/NODES:/)).toBeInTheDocument();
  });

  // 状态栏应显示 HOSTS 计数指标
  it('should render HOSTS count', () => {
    render(<StatusBar />);
    expect(screen.getByText(/HOSTS:/)).toBeInTheDocument();
  });

  // 状态栏应显示当前构建版本号
  it('should render BUILD_VERSION', () => {
    render(<StatusBar />);
    expect(screen.getByText('v0.1.0')).toBeInTheDocument();
  });

  // 状态栏应显示系统时钟（SYS_CLOCK）标签
  it('should render SYS_CLOCK with time', () => {
    render(<StatusBar />);
    expect(screen.getByText(/SYS_CLOCK:/)).toBeInTheDocument();
  });

  // 系统时钟由 useClock 驱动显示
  it('should display clock value from useClock', () => {
    render(<StatusBar />);
    expect(screen.getByText(/SYS_CLOCK: 12:00:00/)).toBeInTheDocument();
  });
});
