import type { ReactNode } from 'react';

interface FormLabelProps {
  children: ReactNode;
  inline?: boolean;
}

export function FormLabel({ children, inline }: FormLabelProps) {
  return (
    <label
      className={`text-[10px] text-muted-foreground uppercase tracking-widest font-mono${inline ? '' : ' mb-1 block'}`}
    >
      {children}
    </label>
  );
}
