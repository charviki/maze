import { useState, useEffect, useCallback, useRef } from 'react';
import { TerrainBackground, HexWaterfall, DecryptText, Button, Input, cn } from '@maze/fabrication';
import { WESTWORLD_QUOTES } from '../data/mock-data';
import { MazeCanvas } from './MazeCanvas';
import { MazeSvg } from './MazeSvg';
import { login } from '../auth/auth';

/** Maze orbit radii — must match MazeCanvas and MazeSvg */
const RADII = [90, 75, 60, 45, 30, 18];
/** Center activation radius */
const CENTER_RADIUS = 22;

interface LandingPageProps {
  onEnter: () => void;
}

export function LandingPage({ onEnter }: LandingPageProps) {
  const [quoteIndex, setQuoteIndex] = useState(0);
  const [phase, setPhase] = useState<'landing' | 'login'>('landing');
  const [exiting, setExiting] = useState(false);

  // Quote carousel
  useEffect(() => {
    const timer = setInterval(() => {
      setQuoteIndex((i) => (i + 1) % WESTWORLD_QUOTES.length);
    }, 8000);
    return () => clearInterval(timer);
  }, []);

  const handleEnterPark = useCallback(() => {
    setPhase('login');
  }, []);

  const handleLoginSuccess = useCallback(() => {
    setExiting(true);
    setTimeout(onEnter, 500);
  }, [onEnter]);

  const handleBack = useCallback(() => {
    setPhase('landing');
  }, []);

  const isLanding = phase === 'landing';

  return (
    <div
      className={cn(
        'h-screen w-screen bg-background text-foreground dark relative overflow-hidden',
        'animate-in fade-in duration-1000',
        exiting && 'animate-out fade-out duration-500',
      )}
    >
      <TerrainBackground />
      {/* Background waterfall — wider, dimmer layer */}
      <HexWaterfall className="left-0 top-0 h-full w-32" opacity={0.08} />
      <HexWaterfall className="right-0 top-0 h-full w-32" opacity={0.08} />
      {/* Foreground waterfall — narrower, brighter layer */}
      <HexWaterfall className="left-0 top-0 h-full w-16" opacity={0.2} />
      <HexWaterfall className="right-0 top-0 h-full w-16" opacity={0.2} />

      {/* Landing phase — centered layout */}
      <div
        className={cn(
          'absolute inset-0 z-10 flex flex-col items-center justify-center gap-8 px-4 transition-all duration-700',
          !isLanding && 'opacity-0 pointer-events-none scale-95',
        )}
      >
        {/* DELOS branding */}
        <div className="text-center space-y-1">
          <p className="text-sm font-mono tracking-[0.3em] text-primary/50 uppercase">
            <DecryptText text="DELOS INCORPORATED" speed={50} />
          </p>
          <p className="text-[10px] font-mono tracking-[0.5em] text-primary/30 uppercase">
            presents
          </p>
        </div>

        {/* Interactive Maze */}
        <MazeContainer className="w-72 h-72 md:w-96 md:h-96" />

        {/* Title */}
        <h1 className="text-4xl md:text-5xl font-mono tracking-[0.2em] text-primary uppercase">
          <DecryptText text="THE MAZE" speed={80} maxIterations={4} />
        </h1>

        {/* Quote carousel */}
        <div className="h-6 flex items-center">
          <p
            key={quoteIndex}
            className="text-xs font-mono tracking-wider text-primary/50 italic text-center max-w-md animate-in fade-in duration-500"
          >
            "<DecryptText text={WESTWORLD_QUOTES[quoteIndex]} speed={30} />"
          </p>
        </div>

        {/* Enter button */}
        <button
          onClick={handleEnterPark}
          className={cn(
            'mt-4 px-8 py-3 border border-primary/30 bg-card/60 backdrop-blur-sm',
            'font-mono text-sm tracking-widest text-primary uppercase',
            'transition-all duration-300',
            'hover:border-primary/60 hover:bg-card/80 hover:shadow-[0_0_20px_rgba(0,255,255,0.1)]',
            'active:scale-95',
          )}
          style={{
            clipPath:
              'polygon(12px 0, calc(100% - 12px) 0, 100% 12px, 100% calc(100% - 12px), calc(100% - 12px) 100%, 12px 100%, 0 calc(100% - 12px), 0 12px)',
          }}
        >
          ENTER THE PARK
        </button>
      </div>

      {/* Login phase — maze in corner + login form centered */}
      <div
        className={cn(
          'absolute inset-0 z-10 flex items-center justify-center transition-all duration-700',
          isLanding && 'opacity-0 pointer-events-none',
        )}
      >
        {/* Mini maze in top-left corner — click to go back */}
        <div
          className={cn(
            'absolute top-6 left-6 transition-all duration-700 group/maze',
            isLanding ? 'opacity-0 scale-50' : 'opacity-100 scale-100',
          )}
          title="Back to entrance"
        >
          <MazeContainer className="w-24 h-24 md:w-32 md:h-32" onClick={handleBack} />
          <span className="absolute -bottom-5 left-1/2 -translate-x-1/2 text-[7px] font-mono tracking-widest text-primary/0 group-hover/maze:text-primary/40 transition-all whitespace-nowrap">
            {'<'} BACK
          </span>
        </div>

        {/* Login form */}
        <div
          className={cn(
            'w-full max-w-md px-4 transition-all duration-700 delay-200',
            isLanding ? 'opacity-0 translate-y-8 scale-95' : 'opacity-100 translate-y-0 scale-100',
          )}
        >
          <LoginForm onSuccess={handleLoginSuccess} />
        </div>
      </div>
    </div>
  );
}

/**
 * Login form — embedded in LandingPage, no separate page.
 */
function LoginForm({ onSuccess }: { onSuccess: () => void }) {
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
        setTimeout(() => setShaking(false), 500);
      }
    },
    [username, password, onSuccess],
  );

  return (
    <div
      className={cn('relative p-px', shaking && 'animate-pulse')}
      style={{
        clipPath:
          'polygon(16px 0, calc(100% - 16px) 0, 100% 16px, 100% calc(100% - 16px), calc(100% - 16px) 100%, 16px 100%, 0 calc(100% - 16px), 0 16px)',
      }}
    >
      {/* Glowing border background */}
      <div
        className="absolute inset-0 bg-gradient-to-br from-primary/20 via-primary/5 to-primary/20"
        style={{ clipPath: 'inherit' }}
      />
      {/* Inner content */}
      <div
        className="relative bg-background/40 backdrop-blur-xl"
        style={{
          clipPath:
            'polygon(16px 0, calc(100% - 16px) 0, 100% 16px, 100% calc(100% - 16px), calc(100% - 16px) 100%, 16px 100%, 0 calc(100% - 16px), 0 16px)',
        }}
      >
        <form onSubmit={handleSubmit} className="flex flex-col gap-5 p-6 md:p-8">
          {/* Header */}
          <div className="text-center space-y-1.5">
            <h1 className="text-sm font-mono tracking-[0.25em] text-primary uppercase">
              <DecryptText text="IDENTITY VERIFICATION" />
            </h1>
            <p className="text-[11px] text-primary/60 font-mono italic">
              <DecryptText text="Bring yourself back online." speed={60} maxIterations={2} />
            </p>
          </div>

          {/* Divider */}
          <div className="h-px bg-gradient-to-r from-transparent via-primary/25 to-transparent" />

          {/* Form fields */}
          <div className="space-y-3">
            <div className="space-y-1.5">
              <label className="text-[9px] font-mono tracking-[0.2em] text-primary/70 uppercase">
                ACCESS ID
              </label>
              <Input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="bg-transparent border-primary/10 focus:border-primary/30 font-mono text-sm text-primary/80 h-9 transition-colors"
                autoComplete="username"
                autoFocus
              />
            </div>

            <div className="space-y-1.5">
              <label className="text-[9px] font-mono tracking-[0.2em] text-primary/70 uppercase">
                PASSPHRASE
              </label>
              <Input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="bg-transparent border-primary/10 focus:border-primary/30 font-mono text-sm text-primary/80 h-9 transition-colors"
                autoComplete="current-password"
              />
            </div>
          </div>

          {/* Error message */}
          {error && (
            <div className="text-destructive text-[10px] font-mono tracking-wider text-center animate-pulse">
              {'>'} {error}
            </div>
          )}

          {/* Submit button */}
          <Button
            type="submit"
            className="w-full font-mono tracking-[0.2em] uppercase text-xs h-9 bg-primary/15 border border-primary/30 hover:bg-primary/25 hover:border-primary/50 transition-all text-primary"
            variant="outline"
          >
            AUTHENTICATE
          </Button>

          {/* Protocol hint */}
          <div className="text-[8px] font-mono text-primary/40 text-center tracking-[0.3em]">
            PROTOCOL: OIDC-READY // SEC_LEVEL: 9
          </div>
        </form>
      </div>
    </div>
  );
}

/**
 * MazeContainer — manages mouse state and composites Canvas + SVG layers.
 */
function MazeContainer({ className, onClick }: { className?: string; onClick?: () => void }) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [mousePos, setMousePos] = useState<{ x: number; y: number } | null>(null);
  const [centerActive, setCenterActive] = useState(false);
  const [hoveredRing, setHoveredRing] = useState<number | null>(null);

  const handleMouseMove = useCallback((e: React.MouseEvent<HTMLDivElement>) => {
    const el = containerRef.current;
    if (!el) return;
    const rect = el.getBoundingClientRect();
    const x = ((e.clientX - rect.left) / rect.width) * 200;
    const y = ((e.clientY - rect.top) / rect.height) * 200;
    const cx = 100;
    const cy = 100;
    const dx = x - cx;
    const dy = y - cy;
    const dist = Math.sqrt(dx * dx + dy * dy);

    setMousePos({ x, y });
    setCenterActive(dist < CENTER_RADIUS);

    let closest: number | null = null;
    let minDelta = Infinity;
    for (let i = 0; i < RADII.length; i++) {
      const delta = Math.abs(dist - RADII[i]);
      if (delta < 12 && delta < minDelta) {
        minDelta = delta;
        closest = i;
      }
    }
    setHoveredRing(closest);
  }, []);

  const handleMouseLeave = useCallback(() => {
    setMousePos(null);
    setCenterActive(false);
    setHoveredRing(null);
  }, []);

  return (
    <div
      ref={containerRef}
      className={cn('relative cursor-crosshair', onClick && 'cursor-pointer', className)}
      onMouseMove={handleMouseMove}
      onMouseLeave={handleMouseLeave}
      onClick={onClick}
    >
      <MazeCanvas mousePos={mousePos} centerActive={centerActive} />
      <MazeSvg centerActive={centerActive} hoveredRing={hoveredRing} />
    </div>
  );
}
