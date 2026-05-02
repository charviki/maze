import { useAnimationSettings } from './AnimationSettings';
import { cn } from '../../utils';

interface HostVitalSignProps {
  status: 'running' | 'saved' | 'offline';
  size?: 'sm' | 'md';
  className?: string;
}

const SIZE_MAP = {
  sm: { core: 'w-1.5 h-3', ring: 'w-4 h-5' },
  md: { core: 'w-2 h-4', ring: 'w-5 h-6' },
} as const;

export function HostVitalSign({ status, size = 'sm', className }: HostVitalSignProps) {
  const { settings } = useAnimationSettings();
  const s = SIZE_MAP[size];

  if (!settings.glitchEffect) {
    return (
      <div className={cn('relative flex shrink-0', s.ring, className)}>
        <span
          className={cn(
            'relative inline-flex rounded-[1px]',
            s.core,
            status === 'running'
              ? 'bg-primary'
              : status === 'saved'
                ? 'bg-yellow-500'
                : 'bg-gray-500',
          )}
        />
      </div>
    );
  }

  return (
    <div className={cn('relative flex items-center justify-center shrink-0', s.ring, className)}>
      {status === 'running' && (
        <>
          {/* 外圈呼吸光环 */}
          <span
            className="absolute inset-0 rounded-full text-primary"
            style={{ animation: 'vital-ring-pulse 1.25s ease-in-out infinite' }}
          />
          {/* 核心呼吸脉冲 */}
          <span
            className={cn('relative rounded-[1px] bg-primary', s.core)}
            style={{ animation: 'vital-breath 1.25s ease-in-out infinite' }}
          />
        </>
      )}

      {status === 'saved' && (
        <>
          <span
            className="absolute inset-0 rounded-full text-yellow-500 opacity-50"
            style={{ animation: 'vital-ring-pulse 3.33s ease-in-out infinite' }}
          />
          <span
            className={cn('relative rounded-[1px] bg-yellow-500', s.core)}
            style={{ animation: 'vital-breath 3.33s ease-in-out infinite' }}
          />
        </>
      )}

      {status === 'offline' && (
        <span
          className={cn('relative rounded-[1px] bg-gray-500', s.core)}
          style={{ animation: 'vital-irregular 4s steps(1) infinite' }}
        />
      )}
    </div>
  );
}
