import type { ApiResponse, ErrorDetails } from '../types';

// hey-api RequestResult（throwOnError=false, responseStyle='fields'）的宽松形态：
// 成功 { data; error: undefined; request?; response? } / 失败 { data: undefined; error; ... }。
// 精确 union 还含 throwOnError=true 等多分支（{data; request; response} 无 error），类型复杂，
// 这里用宽松类型接收：controller 传入的 hey-api RequestResult 结构兼容可赋值，测试构造也无需 as。
type HeyApiResult<T> = Promise<{
  data?: T;
  error?: unknown;
  request?: Request;
  response?: Response;
}>;

// 保留 unwrapSdkResponse/unwrapVoidResponse 函数名与返回类型 ApiResponse<T>
// （controller.ts 与 index.ts re-export 都依赖该契约），仅改参数与内部解析。
export async function unwrapSdkResponse<T>(result: HeyApiResult<T>): Promise<ApiResponse<T>> {
  try {
    const res = await result;
    if (res.error !== undefined) {
      return extractErrorFromHeyApi(res.error, res.response);
    }
    return { status: 'ok', data: res.data as T };
  } catch (e: unknown) {
    // fetchWithAuthSession 仍可能 reject 网络错误/AbortError（自定义 fetch 层抛出）
    return handleError(e);
  }
}

export async function unwrapVoidResponse(
  result: HeyApiResult<unknown>,
): Promise<ApiResponse<void>> {
  const res = await unwrapSdkResponse(result);
  return res.status === 'ok' ? { status: 'ok', data: undefined } : { ...res, data: undefined };
}

function handleError(e: unknown): ApiResponse<never> {
  if (e instanceof DOMException && e.name === 'AbortError') {
    return { status: 'error', message: '请求超时' };
  }
  if (e instanceof Error) {
    return { status: 'error', message: e.message };
  }
  return { status: 'error', message: String(e) };
}

// 复用旧 extractErrorFromResponse 的 reason/details/conflicts/code 解析逻辑。
// hey-api 已把 response body 解析成 error 对象（JSON 对象或字符串），无需再 text()+JSON.parse
// （RequestResult.response 的 body 已被 hey-api 消费，不能再 read）。
function extractErrorFromHeyApi(
  error: unknown,
  response: Response | undefined,
): ApiResponse<never> {
  const statusText = response ? `HTTP ${response.status}` : '请求失败';
  if (error && typeof error === 'object') {
    const parsed = error as Record<string, unknown>;
    const result: ApiResponse<never> = {
      status: 'error',
      message: (parsed.message as string) || statusText,
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
  }
  if (typeof error === 'string' && error) {
    return { status: 'error', message: error };
  }
  return { status: 'error', message: statusText };
}
