import type { ApiResponse } from '../types';
import { ResponseError, FetchError } from './gen/runtime';

export async function unwrapSdkResponse<T>(
  promise: Promise<T>,
): Promise<ApiResponse<T>> {
  try {
    const data = await promise;
    return { status: 'ok', data };
  } catch (e: unknown) {
    return handleError(e);
  }
}

export async function unwrapVoidResponse(
  promise: Promise<unknown>,
): Promise<ApiResponse<void>> {
  try {
    await promise;
    return { status: 'ok', data: undefined };
  } catch (e: unknown) {
    return handleError(e);
  }
}

async function handleError(e: unknown): Promise<ApiResponse<never>> {
  if (e instanceof ResponseError) {
    return extractErrorFromResponse(e.response);
  }
  // FetchError 是 SDK 对 fetch 网络异常的包装，提取原始错误信息
  if (e instanceof FetchError) {
    return { status: 'error', message: e.cause?.message || '网络请求失败' };
  }
  if (e instanceof DOMException && e.name === 'AbortError') {
    return { status: 'error', message: '请求超时' };
  }
  if (e instanceof Error) {
    return { status: 'error', message: e.message };
  }
  return { status: 'error', message: String(e) };
}

async function extractErrorFromResponse(resp: Response): Promise<ApiResponse<never>> {
  const text = await resp.text();
  if (text) {
    try {
      const parsed = JSON.parse(text) as Record<string, unknown>;
      return {
        status: 'error',
        message: (parsed.message as string) || `HTTP ${resp.status}`,
        code: parsed.code as string | undefined,
        conflicts: parsed.conflicts as ApiResponse<never>['conflicts'],
      };
    } catch {
      return { status: 'error', message: `HTTP ${resp.status}` };
    }
  }
  return { status: 'error', message: `HTTP ${resp.status}` };
}
