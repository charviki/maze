import '@testing-library/jest-dom/vitest';
import { server } from './mocks/server';
import { resetMockData } from './mocks/handlers';

vi.mock('@maze/fabrication', async (importOriginal) => {
  const mod = await importOriginal();
  return {
    ...(mod as Record<string, unknown>),
    fetchWithAuth: (input: RequestInfo | URL, init?: RequestInit) => fetch(input, init),
  };
});

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => {
  server.resetHandlers();
  resetMockData();
});
afterAll(() => server.close());
