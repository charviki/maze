import { client } from './gen/client.gen';
import { fetchWithAuthSession } from './auth-session';

const DEFAULT_TIMEOUT_MS = 30000;

// auth 续期（预刷新）+ 401 refreshAccessToken 重试 + clearTokens 跳转 + 30s 超时
// 全部封装在 fetch 层，对应旧 typescript-fetch createSdkConfiguration 的 fetchApi 语义。
// 不放 hey-api interceptor：fetchWithAuthSession 的「发两次 + clearTokens redirect」
// 逻辑超出 interceptor 单次进出的语义边界，且 hey-api 对 !ok 一律 throw 走 error
// 拦截器（response 拦截器在 !ok 不执行），重试放拦截器易死循环。自定义 fetch 是
// 与旧实现 1:1 语义对齐的最短路径。
export function createAuthFetch(): typeof fetch {
  return async (input, init) => {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), DEFAULT_TIMEOUT_MS);
    try {
      const headers = new Headers(init?.headers);
      return await fetchWithAuthSession(input, {
        ...init,
        headers,
        signal: init?.signal ?? controller.signal,
      });
    } finally {
      clearTimeout(timeoutId);
    }
  };
}

// 配置全局 hey-api client 单例（auth/timeout 经自定义 fetch 注入）。
// 显式 throwOnError:false：失败时返回 {error} 交由 unwrapSdkResponse 解析
// reason/details/conflicts，而非走 reject。固化此默认值，避免未来某调用点单独传
// throwOnError:true、或 hey-api 升级改默认值后，结构化错误退化为 reject —— 届时
// handleError 已无 ResponseError 识别，配置冲突等关键信息将全部丢失。
// 返回 client 供消费层（如 director-console controller）引用；向后兼容旧调用签名。
export function createSdkConfiguration(baseUrl = ''): typeof client {
  client.setConfig({ baseUrl, fetch: createAuthFetch(), throwOnError: false });
  return client;
}
