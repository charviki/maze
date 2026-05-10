import { useEffect, type ReactNode } from 'react';
import { isAuthenticated } from '../../api/token-store';

interface AuthGuardProps {
  children: ReactNode;
  loginUrl?: string;
}

export function AuthGuard({ children, loginUrl = '/arrival-gate/' }: AuthGuardProps) {
  const authenticated = isAuthenticated();

  useEffect(() => {
    if (!authenticated) {
      window.location.href = loginUrl;
      return;
    }

    const handler = (e: StorageEvent) => {
      if (e.key === 'maze:access_token' || e.key === 'maze:refresh_token') {
        if (!isAuthenticated()) {
          window.location.href = loginUrl;
        }
      }
    };
    window.addEventListener('storage', handler);
    return () => window.removeEventListener('storage', handler);
  }, [authenticated, loginUrl]);

  if (!authenticated) {
    return <div className="flex items-center justify-center h-screen">Loading...</div>;
  }

  return <>{children}</>;
}
