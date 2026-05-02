import { useState, useCallback } from 'react';
import { TerrainBackground, Panel, DecryptText, Button, Input, cn } from '@maze/fabrication';
import { login } from './auth';

interface LoginPageProps {
  onSuccess: () => void;
}

export function LoginPage({ onSuccess }: LoginPageProps) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [shaking, setShaking] = useState(false);

  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      if (login(username, password)) {
        onSuccess();
      } else {
        setError('ACCESS DENIED // MOTOR FUNCTIONS SUSPENDED');
        setShaking(true);
        setTimeout(() => {
          setShaking(false);
        }, 500);
      }
    },
    [username, password, onSuccess],
  );

  return (
    <div className="h-screen w-screen bg-background text-foreground dark relative overflow-hidden flex items-center justify-center">
      <TerrainBackground />

      <div
        className={cn(
          'relative z-10 w-full max-w-md px-4',
          'animate-in fade-in zoom-in-95 duration-600',
          shaking && 'animate-pulse',
        )}
      >
        <Panel cornerSize={12} className="p-8">
          <form onSubmit={handleSubmit} className="flex flex-col gap-6">
            {/* Header */}
            <div className="text-center space-y-2">
              <h1 className="text-lg font-mono tracking-widest text-primary/80 uppercase">
                <DecryptText text="IDENTITY VERIFICATION" />
              </h1>
              <p className="text-xs text-muted-foreground font-mono">
                <DecryptText text="Bring yourself back online." speed={60} maxIterations={2} />
              </p>
            </div>

            {/* Form fields */}
            <div className="space-y-4">
              <div className="space-y-2">
                <label className="text-[10px] font-mono tracking-widest text-primary/60 uppercase">
                  ACCESS ID
                </label>
                <Input
                  type="text"
                  value={username}
                  onChange={(e) => {
                    setUsername(e.target.value);
                  }}
                  className="bg-background/50 border-primary/20 focus:border-primary/50 font-mono"
                  autoComplete="username"
                  autoFocus
                />
              </div>

              <div className="space-y-2">
                <label className="text-[10px] font-mono tracking-widest text-primary/60 uppercase">
                  PASSPHRASE
                </label>
                <Input
                  type="password"
                  value={password}
                  onChange={(e) => {
                    setPassword(e.target.value);
                  }}
                  className="bg-background/50 border-primary/20 focus:border-primary/50 font-mono"
                  autoComplete="current-password"
                />
              </div>
            </div>

            {/* Error message */}
            {error && (
              <div className="text-destructive text-[11px] font-mono tracking-wider text-center animate-pulse">
                // {error} //
              </div>
            )}

            {/* Submit button */}
            <Button type="submit" className="w-full font-mono tracking-widest uppercase text-sm">
              AUTHENTICATE
            </Button>

            {/* Protocol hint */}
            <div className="text-[9px] font-mono text-primary/30 text-center tracking-widest">
              // PROTOCOL: OIDC-READY //
            </div>
          </form>
        </Panel>
      </div>

      {/* Bottom status bar */}
      <div className="absolute bottom-0 left-0 right-0 h-8 flex items-center px-4 text-[10px] font-mono tracking-widest text-primary/40 z-10">
        <span>SYS_CLOCK: {new Date().toLocaleTimeString()}</span>
        <span className="mx-3">|</span>
        <span>PROTOCOL: OIDC-READY</span>
        <span className="mx-3">|</span>
        <span>SEC_LEVEL: 9</span>
      </div>
    </div>
  );
}
