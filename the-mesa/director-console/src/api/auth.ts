import {
  fetchWithAuthSession,
  refreshAccessToken as refreshSharedAccessToken,
} from '@maze/fabrication';

export async function refreshTokens(): Promise<boolean> {
  return refreshSharedAccessToken();
}

export async function fetchWithAuth(
  input: RequestInfo | URL,
  init?: RequestInit,
): Promise<Response> {
  return fetchWithAuthSession(input, init);
}
