import type { ReactNode } from 'react';
import { ErrorBoundary } from './ErrorBoundary';
import { AnimationSettingsProvider } from './AnimationSettings';
import { ToastProvider } from './Toast';
import { AuthGuard } from './AuthGuard';

interface AppShellProps {
  children: ReactNode;
  requireAuth?: boolean;
  loginUrl?: string;
}

export function AppShell({ children, requireAuth = false, loginUrl }: AppShellProps) {
  const inner = (
    <AnimationSettingsProvider>
      <ToastProvider>{children}</ToastProvider>
    </AnimationSettingsProvider>
  );

  return (
    <ErrorBoundary>
      {requireAuth ? <AuthGuard loginUrl={loginUrl}>{inner}</AuthGuard> : inner}
    </ErrorBoundary>
  );
}
