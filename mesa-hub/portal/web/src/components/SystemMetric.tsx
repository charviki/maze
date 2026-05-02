import type { LucideIcon } from 'lucide-react'
import { Panel } from '@maze/fabrication'

interface SystemMetricProps {
  label: string
  value: string
  subValue?: string
  icon: LucideIcon
}

export function SystemMetric({ label, value, subValue, icon: Icon }: SystemMetricProps) {
  return (
    <Panel cornerSize={6} showCrosshairs={false} transparent className="p-3">
      <div className="flex items-start gap-3">
        <Icon className="w-4 h-4 text-primary/40 mt-0.5 shrink-0" />
        <div className="flex flex-col gap-0.5">
          <span className="text-[10px] font-mono tracking-widest text-muted-foreground uppercase">
            {label}
          </span>
          <span className="text-2xl font-bold text-primary font-mono tabular-nums">
            {value}
          </span>
          {subValue && (
            <span className="text-[9px] font-mono text-muted-foreground/60">
              {subValue}
            </span>
          )}
        </div>
      </div>
    </Panel>
  )
}
