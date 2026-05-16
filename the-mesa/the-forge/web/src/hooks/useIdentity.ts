import { getCurrentUser } from '@maze/fabrication';

export function useIdentity() {
  const username = getCurrentUser() ?? 'anonymous';
  return { username, displayName: username };
}
