import type { ApiResponse } from '../types';

/**
 * 默认请求超时时间（毫秒）。
 * 超过此时间未收到响应，AbortController 会中断 fetch 并返回超时错误。
 */
const DEFAULT_TIMEOUT_MS = 30000;

/**
 * 创建带 baseUrl 前缀的请求函数。
 * 所有 API client 共享统一的错误处理和响应解析逻辑。
 */
export function createRequest(baseUrl: string = '') {
  return async function request<T>(url: string, options?: RequestInit): Promise<ApiResponse<T>> {
    try {
      // 如果调用方提供了 signal（自定义超时控制），直接使用；否则用默认超时
      const externalSignal = options?.signal;
      let controller: AbortController | null = null;
      let timeoutId: ReturnType<typeof setTimeout> | undefined;

      if (externalSignal) {
        // 外部控制超时，不创建内部 AbortController
      } else {
        controller = new AbortController();
        timeoutId = setTimeout(() => controller!.abort(), DEFAULT_TIMEOUT_MS);
      }

      const signal = externalSignal || controller!.signal;

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
          return JSON.parse(text);
        } catch {
          return { status: 'error', message: `HTTP ${response.status}` };
        }
      }
      if (!text) {
        return { status: 'ok', data: undefined } as ApiResponse<T>;
      }
      try {
        return JSON.parse(text);
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
