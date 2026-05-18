import type { ApiResponse, ErrorDetails } from '../types';
import { fetchWithAuthSession } from '../api/auth-session';

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

      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...(options?.headers as Record<string, string>),
      };

      const response = await fetchWithAuthSession(`${baseUrl}${url}`, {
        ...options,
        signal,
        headers,
      });

      if (timeoutId) clearTimeout(timeoutId);

      const text = await response.text();
      if (!response.ok) {
        if (!text) {
          return { status: 'error', message: `HTTP ${response.status}` };
        }
        try {
          const parsed = JSON.parse(text) as Record<string, unknown>;
          const result: ApiResponse<T> = {
            status: 'error' as const,
            message: (parsed.message as string) || `HTTP ${response.status}`,
          };
          if (parsed.reason) {
            result.reason = parsed.reason as string;
          }
          if (parsed.details) {
            // eslint-disable-next-line @typescript-eslint/no-unnecessary-type-assertion
            result.details = parsed.details as ErrorDetails;
          }
          if (result.details?.preconditionViolations) {
            result.conflicts = result.details.preconditionViolations
              .filter((v) => v.type === 'CONFIG_CONFLICT')
              .map((v) => ({ path: v.subject, currentHash: v.description }));
          }
          if (parsed.code !== undefined) {
            result.code = parsed.code as string;
          }
          return result;
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
