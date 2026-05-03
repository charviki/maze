import type { ApiResponse } from '../types';

const DEFAULT_TIMEOUT_MS = 30000;

/**
 * 创建带 baseUrl 前缀的请求函数。
 * 所有 API client 共享统一的错误处理和响应解析逻辑。
 *
 * 服务端 grpc-gateway 返回标准 proto JSON（成功时直接是业务数据，失败时为 rpcStatus 格式），
 * 此函数在解析层自行包装为 ApiResponse<T>，对调用方完全透明。
 */
export function createRequest(baseUrl = '') {
  return async function request<T>(url: string, options?: RequestInit): Promise<ApiResponse<T>> {
    try {
      const externalSignal = options?.signal;
      let controller: AbortController | null = null;
      let timeoutId: ReturnType<typeof setTimeout> | undefined;

      if (externalSignal) {
        // 外部控制超时，不创建内部 AbortController
      } else {
        controller = new AbortController();
        timeoutId = setTimeout(() => {
          controller!.abort();
        }, DEFAULT_TIMEOUT_MS);
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
        // rpcStatus 格式 {"code": int32, "message": "..."} → 提取 message
        try {
          const parsed = JSON.parse(text) as { code?: number; message?: string };
          return { status: 'error', message: parsed.message || `HTTP ${response.status}` };
        } catch {
          return { status: 'error', message: `HTTP ${response.status}` };
        }
      }
      if (!text) {
        return { status: 'ok', data: undefined };
      }
      // 成功：proto JSON 直接就是业务数据，包装为 ApiResponse
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
