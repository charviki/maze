import { useState } from 'react';
import { isAuthenticated } from './auth';
import { LoginPage } from './LoginPage';

interface AuthGateProps {
  children: React.ReactNode;
}

export function AuthGate({ children }: AuthGateProps) {
  const [authed, setAuthed] = useState(isAuthenticated());

  if (!authed) {
    return (
      <LoginPage
        onSuccess={() => {
          setAuthed(true);
        }}
      />
    );
  }

  return <>{children}</>;
}
