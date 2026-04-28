import { useEffect, useRef, useState, useCallback } from 'react';

interface PollingOptions<T> {
  fetchFn: () => Promise<T>;
  baseInterval?: number;
  maxInterval?: number;
  enabled?: boolean;
}

export function usePollingWithBackoff<T>({
  fetchFn,
  baseInterval = 5000,
  maxInterval = 30000,
  enabled = true,
}: PollingOptions<T>) {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const consecutiveFailures = useRef(0);
  const mountedRef = useRef(true);
  const isFetchingRef = useRef(false);

  const fetchFnRef = useRef(fetchFn);
  fetchFnRef.current = fetchFn;

  const executeFetch = useCallback(async () => {
    if (isFetchingRef.current) return;
    isFetchingRef.current = true;
    try {
      const result = await fetchFnRef.current();
      if (!mountedRef.current) return;
      setData(result);
      setError(null);
      consecutiveFailures.current = 0;
      setIsLoading(false);
    } catch {
      if (!mountedRef.current) return;
      consecutiveFailures.current++;
      setError('请求失败');
      setIsLoading(false);
    } finally {
      isFetchingRef.current = false;
    }
  }, []);

  useEffect(() => {
    mountedRef.current = true;
    if (!enabled) return;

    executeFetch();

    let timeoutId: ReturnType<typeof setTimeout>;

    const scheduleNext = () => {
      const backoff = Math.min(
        baseInterval * Math.pow(2, consecutiveFailures.current),
        maxInterval
      );
      timeoutId = setTimeout(async () => {
        await executeFetch();
        if (mountedRef.current) scheduleNext();
      }, backoff);
    };

    scheduleNext();

    const handleVisibility = () => {
      if (document.visibilityState === 'visible') {
        clearTimeout(timeoutId);
        executeFetch().then(() => {
          if (mountedRef.current) scheduleNext();
        });
      } else {
        clearTimeout(timeoutId);
      }
    };
    document.addEventListener('visibilitychange', handleVisibility);

    return () => {
      mountedRef.current = false;
      clearTimeout(timeoutId);
      document.removeEventListener('visibilitychange', handleVisibility);
    };
  }, [executeFetch, baseInterval, maxInterval, enabled]);

  const refresh = useCallback(async () => {
    await executeFetch();
  }, [executeFetch]);

  return { data, error, isLoading, refresh };
}
