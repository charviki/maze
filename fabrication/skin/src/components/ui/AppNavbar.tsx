import type { ReactNode } from 'react';
import { DecryptText } from './DecryptText';

interface AppNavbarProps {
  title: string;
  icon?: ReactNode;
  rightContent?: ReactNode;
  leftContent?: ReactNode;
}

export function AppNavbar({ title, icon, rightContent, leftContent }: AppNavbarProps) {
  return (
    <div className="col-span-2 border-b border-border/50 flex items-center justify-between px-6 bg-background relative overflow-hidden z-10">
      <div className="absolute top-0 left-0 w-full h-[1px] bg-primary/20" />
      <div className="flex items-center gap-3">
        {leftContent}
        <div className="flex items-center gap-3 font-bold text-lg tracking-wider text-primary">
          {icon}
          <DecryptText text={title} className="uppercase" />
        </div>
      </div>
      {rightContent}
    </div>
  );
}
