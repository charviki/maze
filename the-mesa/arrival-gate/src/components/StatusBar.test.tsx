import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { act } from '@testing-library/react';
import { StatusBar } from './StatusBar';

// DecryptText 依赖 Canvas，用纯文本 span 替代
vi.mock('@maze/fabrication', () => ({
  DecryptText: ({ text }: { text: string }) => <span>{text}</span>,
}));

describe('StatusBar', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
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

  // 系统时钟应随时间推移实时更新
  it('should update clock over time', () => {
    render(<StatusBar />);
    const before = screen.getByText(/SYS_CLOCK:/).textContent;

    act(() => {
      vi.advanceTimersByTime(2000);
    });

    const after = screen.getByText(/SYS_CLOCK:/).textContent;
    expect(after).not.toBe(before);
  });
});
