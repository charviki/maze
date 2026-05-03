import type { ApiResponse } from '../types';

const DEFAULT_TIMEOUT_MS = 30000;

export function createRequest(baseUrl = '') {
  return async function request<T>(url: string, options?: RequestInit): Promise<ApiResponse<T>> {
    try {
      const externalSignal = options?.signal;
      let controller: AbortController | null = null;
      let timeoutId: ReturnType<typeof setTimeout> | undefined;

      if (!externalSignal) {
        controller = new AbortController();
        timeoutId = setTimeout(() => {
          controller!.abort();
        }, DEFAULT_TIMEOUT_MS);
      }

      const signal = externalSignal ?? controller!.signal;

      const response = await fetch(`${baseUrl}${url}`, {
        ...options,
        signal,
        headers: {
          'Content-Type': 'application/json',
          ...options?.headers,
        },
      });

      if (timeoutId) clearTimeout(timeoutId);

      const text = await response.text();
      if (!response.ok) {
        if (!text) {
          return { status: 'error', message: `HTTP ${response.status}` };
        }
        try {
          const parsed = JSON.parse(text) as Record<string, unknown>;
          return {
            status: 'error' as const,
            message: (parsed.message as string) || `HTTP ${response.status}`,
            code: parsed.code as string | undefined,
            conflicts: parsed.conflicts as ApiResponse<T>['conflicts'],
          };
        } catch {
          return { status: 'error', message: `HTTP ${response.status}` };
        }
      }
      if (!text) {
        return { status: 'ok', data: undefined };
      }
      try {
        const parsed = JSON.parse(text) as T;
        return { status: 'ok', data: parsed };
      } catch {
        return { status: 'error', message: '响应解析失败' };
      }
    } catch (e) {
      if (e instanceof DOMException && e.name === 'AbortError') {
        return { status: 'error', message: '请求超时' };
      }
      return { status: 'error', message: String(e) };
    }
  };
}
