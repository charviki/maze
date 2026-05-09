import { useEffect, type ReactNode } from 'react';
import { isTokenAuthenticated } from '@maze/fabrication';

interface AuthGuardProps {
  children: ReactNode;
}

export function AuthGuard({ children }: AuthGuardProps) {
  const authenticated = isTokenAuthenticated();

  useEffect(() => {
    if (!authenticated) {
      window.location.href = '/arrival-gate/';
      return;
    }

    const handler = (e: StorageEvent) => {
      if (e.key === 'maze:access_token' || e.key === 'maze:refresh_token') {
        if (!isTokenAuthenticated()) {
          window.location.href = '/arrival-gate/';
        }
      }
    };
    window.addEventListener('storage', handler);
    return () => window.removeEventListener('storage', handler);
  }, [authenticated]);

  if (!authenticated) {
    return <div className="flex items-center justify-center h-screen">Loading...</div>;
  }

  return <>{children}</>;
}
