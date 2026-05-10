import '@testing-library/jest-dom/vitest';
import { server } from './mocks/server';
import { resetMockData } from './mocks/handlers';

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => {
  server.resetHandlers();
  resetMockData();
});
afterAll(() => server.close());
