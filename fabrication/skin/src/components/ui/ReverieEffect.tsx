import { useState, useEffect, useRef, useCallback, type ReactNode } from 'react';
import { cn } from '../../utils';
import { useAnimationSettings } from './AnimationSettings';

interface ReverieEffectProps {
  children: ReactNode;
  isActive?: boolean;
  className?: string;
}

const REVERIE_MIN_INTERVAL = 8000;
const REVERIE_MAX_INTERVAL = 20000;
const REVERIE_DURATION = 500;

export function ReverieEffect({ children, isActive = false, className }: ReverieEffectProps) {
  const { settings } = useAnimationSettings();
  const [isRevering, setIsRevering] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const scheduleNext = useCallback(() => {
    if (!isActive || !settings.glitchEffect) return;

    const delay = REVERIE_MIN_INTERVAL + Math.random() * (REVERIE_MAX_INTERVAL - REVERIE_MIN_INTERVAL);
    timerRef.current = setTimeout(() => {
      setIsRevering(true);
      timerRef.current = setTimeout(() => {
        setIsRevering(false);
        scheduleNext();
      }, REVERIE_DURATION);
    }, delay);
  }, [isActive, settings.glitchEffect]);

  useEffect(() => {
    if (!isActive || !settings.glitchEffect) {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
        timerRef.current = null;
      }
      setIsRevering(false);
      return;
    }

    scheduleNext();

    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
        timerRef.current = null;
      }
    };
  }, [isActive, settings.glitchEffect, scheduleNext]);

  if (!isActive || !settings.glitchEffect) {
    return <div className={className}>{children}</div>;
  }

  return (
    <div className={cn('relative', className)}>
      <div
        className="relative z-10"
        style={isRevering ? { animation: 'reverie-flicker 500ms ease-out forwards' } : undefined}
      >
        {children}
      </div>
      {isRevering && (
        <div
          className="absolute inset-0 z-20 pointer-events-none opacity-20 text-primary"
          style={{
            transform: 'translateX(1px)',
          }}
          aria-hidden="true"
        >
          {children}
        </div>
      )}
    </div>
  );
}
