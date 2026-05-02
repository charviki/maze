import type { ReactNode } from 'react';
import { cn } from '../../utils';
import { useAnimationSettings } from './AnimationSettings';

interface GlitchEffectProps {
  children: ReactNode;
  isActive?: boolean;
  className?: string;
}

export function GlitchEffect({ children, isActive = false, className }: GlitchEffectProps) {
  const { settings } = useAnimationSettings();

  if (!isActive || !settings.glitchEffect) {
    return <div className={className}>{children}</div>;
  }

  return (
    <div className={cn('relative group', className)}>
      <div className="glitch-base relative z-10">{children}</div>

      <div className="glitch-layer glitch-layer-1 z-20" aria-hidden="true">
        {children}
      </div>

      <div className="glitch-layer glitch-layer-2 z-30" aria-hidden="true">
        {children}
      </div>
    </div>
  );
}
