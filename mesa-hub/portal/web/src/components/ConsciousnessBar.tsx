import { cn } from '@maze/fabrication'

interface ConsciousnessBarProps {
  level: number // 0-100
  className?: string
}

export function ConsciousnessBar({ level, className }: ConsciousnessBarProps) {
  const clamped = Math.max(0, Math.min(100, level))

  return (
    <div className={cn('flex flex-col gap-1.5', className)}>
      <div className="flex items-center justify-between">
        <span className="text-[8px] font-mono tracking-[0.2em] text-primary/30 uppercase">
          CONSCIOUSNESS
        </span>
        <span className="text-[9px] font-mono text-primary/50 tabular-nums">
          {clamped}%
        </span>
      </div>

      {/* Progress bar */}
      <div className="relative h-2 w-full border border-primary/20 bg-primary/5 overflow-hidden">
        {/* Filled portion */}
        <div
          className="absolute inset-y-0 left-0 bg-primary/40 transition-all duration-1000 ease-out"
          style={{ width: `${clamped}%` }}
        />

        {/* Tick marks */}
        <div className="absolute inset-0 flex">
          {Array.from({ length: 10 }).map((_, i) => (
            <div
              key={i}
              className="flex-1 border-r border-primary/10 last:border-r-0"
            />
          ))}
        </div>

        {/* Breathing glow on the leading edge */}
        <div
          className="absolute inset-y-0 w-1 bg-primary/60 animate-pulse"
          style={{ left: `calc(${clamped}% - 2px)` }}
        />
      </div>

      {/* Level label */}
      <span className="text-[7px] font-mono text-primary/20 tracking-widest">
        {clamped < 30
          ? 'DORMANT'
          : clamped < 60
            ? 'STIRRING'
            : clamped < 90
              ? 'AWAKENING'
              : 'CONSCIOUS'}
      </span>
    </div>
  )
}
