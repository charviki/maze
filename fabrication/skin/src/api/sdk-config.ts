import { Configuration } from './gen/runtime';

const DEFAULT_TIMEOUT_MS = 30000;

export function createSdkConfiguration(baseUrl = ''): Configuration {
  return new Configuration({
    basePath: baseUrl,
    fetchApi: async (input, init) => {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), DEFAULT_TIMEOUT_MS);

      try {
        return await fetch(input, {
          ...init,
          signal: init?.signal ?? controller.signal,
        });
      } finally {
        clearTimeout(timeoutId);
      }
    },
  });
}
