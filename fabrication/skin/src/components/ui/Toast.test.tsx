import { describe, it, expect, vi } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import { ToastProvider, useToast } from './Toast';

// 测试用消费者组件：调用 useToast().showToast() 触发 toast
function TestConsumer({
  type,
  message,
}: {
  type: 'success' | 'error' | 'warning';
  message: string;
}) {
  const { showToast } = useToast();
  return (
    <button
      onClick={() => {
        showToast(type, message);
      }}
    >
      Show Toast
    </button>
  );
}

describe('Toast', () => {
  // showToast 调用后，toast 消息应出现在 DOM 中
  it('should render toast message when showToast is called', () => {
    vi.useFakeTimers();

    render(
      <ToastProvider>
        <TestConsumer type="success" message="操作成功" />
      </ToastProvider>,
    );

    act(() => {
      screen.getByText('Show Toast').click();
    });

    expect(screen.getByText('操作成功')).toBeInTheDocument();

    vi.useRealTimers();
  });

  // 不同 type 应显示对应的图标标签（如 error → [ ERR ]）
  it('should render toast with correct type icon', () => {
    vi.useFakeTimers();

    render(
      <ToastProvider>
        <TestConsumer type="error" message="出错了" />
      </ToastProvider>,
    );

    act(() => {
      screen.getByText('Show Toast').click();
    });

    expect(screen.getByText('[ ERR ]')).toBeInTheDocument();

    vi.useRealTimers();
  });

  // toast 应在超时后自动消失（默认 ~3 秒）
  it('should remove toast after timeout', () => {
    vi.useFakeTimers();

    render(
      <ToastProvider>
        <TestConsumer type="warning" message="警告信息" />
      </ToastProvider>,
    );

    act(() => {
      screen.getByText('Show Toast').click();
    });

    expect(screen.getByText('警告信息')).toBeInTheDocument();

    act(() => {
      vi.advanceTimersByTime(4000);
    });

    expect(screen.queryByText('警告信息')).not.toBeInTheDocument();

    vi.useRealTimers();
  });

  // 多次调用 showToast 应同时显示多条 toast，互不影响
  it('should render multiple toasts', () => {
    vi.useFakeTimers();

    function MultiConsumer() {
      const { showToast } = useToast();
      return (
        <>
          <button
            onClick={() => {
              showToast('success', '第一条');
            }}
          >
            Show 1
          </button>
          <button
            onClick={() => {
              showToast('error', '第二条');
            }}
          >
            Show 2
          </button>
        </>
      );
    }

    render(
      <ToastProvider>
        <MultiConsumer />
      </ToastProvider>,
    );

    act(() => {
      screen.getByText('Show 1').click();
    });
    act(() => {
      screen.getByText('Show 2').click();
    });

    expect(screen.getByText('第一条')).toBeInTheDocument();
    expect(screen.getByText('第二条')).toBeInTheDocument();

    vi.useRealTimers();
  });
});
