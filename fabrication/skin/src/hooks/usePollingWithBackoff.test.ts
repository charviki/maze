import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { usePollingWithBackoff } from './usePollingWithBackoff';

// 在 fake timers 环境中安全地 flush microtask（让 Promise resolve 生效）
async function waitForFetch() {
  await act(async () => {
    await new Promise<void>((resolve) => queueMicrotask(resolve));
    await new Promise<void>((resolve) => queueMicrotask(resolve));
  });
}

describe('usePollingWithBackoff', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  // hook 挂载后应立即发起首次 fetch（不等定时器）
  it('should call fetchFn immediately on mount', async () => {
    const fetchFn = vi.fn().mockResolvedValue('data');

    renderHook(() => usePollingWithBackoff({ fetchFn }));

    await waitForFetch();

    expect(fetchFn).toHaveBeenCalledOnce();
  });

  // fetch 成功后应将返回值写入 data，并清除 loading / error 状态
  it('should set data after successful fetch', async () => {
    const fetchFn = vi.fn().mockResolvedValue({ items: [1, 2, 3] });

    const { result } = renderHook(() => usePollingWithBackoff({ fetchFn }));

    await waitForFetch();

    expect(result.current.data).toEqual({ items: [1, 2, 3] });
    expect(result.current.error).toBeNull();
    expect(result.current.isLoading).toBe(false);
  });

  // fetch 失败应设置 error 状态，data 保持 null，loading 清除
  it('should set error after failed fetch', async () => {
    const fetchFn = vi.fn().mockRejectedValue(new Error('Network error'));

    const { result } = renderHook(() => usePollingWithBackoff({ fetchFn }));

    await waitForFetch();

    expect(result.current.error).toBe('请求失败');
    expect(result.current.data).toBeNull();
    expect(result.current.isLoading).toBe(false);
  });

  // 成功后应按 baseInterval 间隔持续轮询
  it('should poll at baseInterval after success', async () => {
    const fetchFn = vi.fn().mockResolvedValue('data');

    renderHook(() => usePollingWithBackoff({ fetchFn, baseInterval: 1000 }));

    await waitForFetch();
    expect(fetchFn).toHaveBeenCalledTimes(1);

    act(() => {
      vi.advanceTimersByTime(1000);
    });
    await waitForFetch();

    expect(fetchFn).toHaveBeenCalledTimes(2);
  });

  // 连续失败后轮询间隔应指数递增：1s → 2s → 4s（backoff 策略）
  it('should increase interval with exponential backoff after failures', async () => {
    const fetchFn = vi.fn().mockRejectedValue(new Error('fail'));

    renderHook(() => usePollingWithBackoff({ fetchFn, baseInterval: 1000, maxInterval: 30000 }));

    await waitForFetch();
    expect(fetchFn).toHaveBeenCalledTimes(1);

    // 第二次：间隔 1s（baseInterval）
    act(() => {
      vi.advanceTimersByTime(1000);
    });
    await waitForFetch();
    expect(fetchFn).toHaveBeenCalledTimes(2);

    // 第三次：间隔 4s（1000 * 2^2 = 4000）
    act(() => {
      vi.advanceTimersByTime(4000);
    });
    await waitForFetch();
    expect(fetchFn).toHaveBeenCalledTimes(3);
  });

  // enabled=false 时不应发起任何 fetch，保持 loading 状态
  it('should not poll when enabled is false', async () => {
    const fetchFn = vi.fn().mockResolvedValue('data');

    const { result } = renderHook(() => usePollingWithBackoff({ fetchFn, enabled: false }));

    await waitForFetch();

    expect(fetchFn).not.toHaveBeenCalled();
    expect(result.current.isLoading).toBe(true);
  });

  // 卸载后应停止轮询，不再调用 fetchFn
  it('should stop polling on unmount', async () => {
    const fetchFn = vi.fn().mockResolvedValue('data');

    const { unmount } = renderHook(() => usePollingWithBackoff({ fetchFn, baseInterval: 1000 }));

    await waitForFetch();
    expect(fetchFn).toHaveBeenCalledTimes(1);

    unmount();

    act(() => {
      vi.advanceTimersByTime(5000);
    });

    expect(fetchFn).toHaveBeenCalledTimes(1);
  });

  // refresh() 应立即触发一次额外的 fetch，不受轮询间隔影响
  it('should provide refresh function that triggers a fetch', async () => {
    const fetchFn = vi.fn().mockResolvedValue('data');

    const { result } = renderHook(() => usePollingWithBackoff({ fetchFn }));

    await waitForFetch();
    expect(fetchFn).toHaveBeenCalledTimes(1);

    await act(async () => {
      await result.current.refresh();
    });

    expect(fetchFn).toHaveBeenCalledTimes(2);
  });
});
