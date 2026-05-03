import * as React from 'react';
import { cn } from '../../utils';

type PanelVariant = 'default' | 'destructive' | 'warning' | 'success';

// variant 对应的边框颜色和装饰文本样式映射
const variantStyles: Record<
  PanelVariant,
  { border: string; corner: string; label: string; hash: string }
> = {
  default: {
    border: 'border-primary/30 group-hover:border-primary/50',
    corner: 'border-primary/70',
    label: 'text-primary/40',
    hash: 'text-primary/40',
  },
  destructive: {
    border: 'border-destructive/40 group-hover:border-destructive/60',
    corner: 'border-destructive/70',
    label: 'text-destructive/50',
    hash: 'text-destructive/50',
  },
  warning: {
    border: 'border-yellow-500/40 group-hover:border-yellow-500/60',
    corner: 'border-yellow-500/70',
    label: 'text-yellow-500/50',
    hash: 'text-yellow-500/50',
  },
  success: {
    border: 'border-green-500/40 group-hover:border-green-500/60',
    corner: 'border-green-500/70',
    label: 'text-green-500/50',
    hash: 'text-green-500/50',
  },
};

const variantLabels: Record<PanelVariant, { auth: string; hash: string }> = {
  default: { auth: '[SYS_AUTH: SEC_LEVEL_9]', hash: '// HASH: 0x8F9B2C4A //' },
  destructive: { auth: '[ALERT: CRITICAL_FAULT]', hash: '// ERR: 0xDEAD4A21 //' },
  warning: { auth: '[WARN: ELEVATED_RISK]', hash: '// WARN: 0xCAFE0000 //' },
  success: { auth: '[STATUS: NOMINAL]', hash: '// OK: 0x00FF00FF //' },
};

interface PanelProps extends Omit<React.HTMLAttributes<HTMLDivElement>, 'title'> {
  children: React.ReactNode;
  className?: string;
  cornerSize?: number;
  showCrosshairs?: boolean;
  variant?: 'default' | 'destructive' | 'warning' | 'success';
  transparent?: boolean;
}

export function Panel({
  children,
  className = '',
  cornerSize = 12,
  showCrosshairs = true,
  variant = 'default',
  transparent = false,
  ...props
}: PanelProps) {
  // 使用 clip-path 生成切角形状
  const clipPathStyle = {
    clipPath: `polygon(
      0 ${cornerSize}px,
      ${cornerSize}px 0,
      calc(100% - ${cornerSize}px) 0,
      100% ${cornerSize}px,
      100% calc(100% - ${cornerSize}px),
      calc(100% - ${cornerSize}px) 100%,
      ${cornerSize}px 100%,
      0 calc(100% - ${cornerSize}px)
    )`,
  };

  const vs = variantStyles[variant] || variantStyles.default;
  const vl = variantLabels[variant] || variantLabels.default;

  return (
    <div className={cn('relative group', className)} {...props}>
      <div
        className={cn(
          'w-full h-full border transition-colors',
          transparent ? 'bg-transparent' : 'bg-card/80 backdrop-blur-sm',
          vs.border,
        )}
        style={clipPathStyle}
      >
        <div className="p-4 h-full">{children}</div>
      </div>

      {showCrosshairs && (
        <>
          <div
            className={cn(
              'absolute top-0 left-0 w-3 h-3 border-t border-l -translate-x-1 -translate-y-1 pointer-events-none',
              vs.corner,
            )}
          />
          <div
            className={cn(
              'absolute top-0 right-0 w-3 h-3 border-t border-r translate-x-1 -translate-y-1 pointer-events-none',
              vs.corner,
            )}
          />
          <div
            className={cn(
              'absolute bottom-0 left-0 w-3 h-3 border-b border-l -translate-x-1 translate-y-1 pointer-events-none',
              vs.corner,
            )}
          />
          <div
            className={cn(
              'absolute bottom-0 right-0 w-3 h-3 border-b border-r translate-x-1 translate-y-1 pointer-events-none',
              vs.corner,
            )}
          />

          <div
            className={cn(
              'absolute -top-3 left-4 text-[8px] font-mono tracking-widest whitespace-nowrap pointer-events-none',
              vs.label,
            )}
          >
            {vl.auth}
          </div>

          <div
            className={cn(
              'absolute -bottom-3 right-4 text-[8px] font-mono tracking-widest whitespace-nowrap pointer-events-none',
              vs.hash,
            )}
          >
            {vl.hash}
          </div>

          <div className="absolute top-1/2 left-0 -translate-x-2 -translate-y-1/2 flex flex-col gap-1 opacity-40 pointer-events-none">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className={`h-[1px] bg-primary ${i === 2 ? 'w-2' : 'w-1'}`}></div>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
