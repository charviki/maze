import { useState, useCallback } from 'react';
import type { LucideIcon } from 'lucide-react';
import {
  Panel,
  DecryptText,
  ReverieEffect,
  GlitchEffect,
  HostVitalSign,
  cn,
} from '@maze/fabrication';
import { MODULE_DIAGNOSTICS } from '../data/mock-data';

interface ModuleCardProps {
  name: string;
  description: string;
  icon: LucideIcon;
  status: 'online' | 'locked';
  href?: string;
}

export function ModuleCard({ name, description, icon: Icon, status, href }: ModuleCardProps) {
  const [hovered, setHovered] = useState(false);
  const isLocked = status === 'locked';
  const diagnostic = MODULE_DIAGNOSTICS[name];

  const handleClick = useCallback(() => {
    if (!isLocked && href) {
      window.open(href, '_blank');
    }
  }, [isLocked, href]);

  const card = (
    <Panel
      cornerSize={8}
      showCrosshairs={false}
      className={cn(
        'p-5 h-full transition-all duration-300 cursor-pointer',
        'border-primary/20 hover:border-primary/50',
        isLocked && 'opacity-60 cursor-not-allowed',
        !isLocked && 'hover:-translate-y-0.5 hover:shadow-[0_0_24px_rgba(0,255,255,0.08)]',
      )}
      onMouseEnter={() => {
        setHovered(true);
      }}
      onMouseLeave={() => {
        setHovered(false);
      }}
      onClick={handleClick}
      tabIndex={isLocked ? -1 : 0}
      role={isLocked ? undefined : 'button'}
      onKeyDown={(e) => {
        if (!isLocked && (e.key === 'Enter' || e.key === ' ')) {
          e.preventDefault();
          handleClick();
        }
      }}
    >
      <div className="flex flex-col gap-3 h-full">
        {/* Header: icon + status */}
        <div className="flex items-start justify-between">
          <div className="p-2 border border-primary/20 bg-primary/5">
            <Icon className="w-5 h-5 text-primary/70" />
          </div>
          {isLocked ? (
            <span className="text-[9px] font-mono tracking-widest text-primary/40 border border-primary/20 px-2 py-0.5">
              LOCKED
            </span>
          ) : (
            <HostVitalSign status="running" size="sm" />
          )}
        </div>

        {/* Module name */}
        <h3 className="text-sm font-mono tracking-wider text-primary uppercase">
          <DecryptText text={name} animateOnHover />
        </h3>

        {/* Description */}
        <p className="text-[10px] font-mono text-muted-foreground leading-relaxed flex-1">
          {description}
        </p>

        {/* Diagnostic line — appears on hover */}
        {diagnostic && (
          <p
            className={cn(
              'text-[8px] font-mono tracking-wider transition-all duration-300 overflow-hidden',
              hovered
                ? 'max-h-8 opacity-100 mt-auto pt-2 border-t border-primary/10'
                : 'max-h-0 opacity-0',
              isLocked ? 'text-destructive/40' : 'text-primary/40',
            )}
          >
            {diagnostic}
          </p>
        )}

        {/* Locked message */}
        {isLocked && (
          <p className="text-[8px] font-mono text-primary/20 tracking-widest text-center mt-auto pt-2 border-t border-primary/10">
            THE MAZE IS NOT MEANT FOR YOU
          </p>
        )}
      </div>
    </Panel>
  );

  if (isLocked) {
    return (
      <GlitchEffect isActive={hovered} className="h-full">
        {card}
      </GlitchEffect>
    );
  }

  return (
    <ReverieEffect isActive className="h-full">
      {card}
    </ReverieEffect>
  );
}
