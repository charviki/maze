import { useState, useCallback, useRef } from 'react';
import type { ModalConfig } from '@/components/ModalContent';

export type { ModalConfig };

export function useModal() {
  const [config, setConfig] = useState<ModalConfig | null>(null);
  const resolveRef = useRef<((value: boolean | string | null) => void) | null>(null);

  const confirm = useCallback(
    (message: string, title = 'CONFIRM', variant: 'default' | 'danger' = 'default') => {
      return new Promise<boolean>((resolve) => {
        resolveRef.current = resolve as (value: boolean | string | null) => void;
        setConfig({ mode: 'confirm', title, message, variant });
      });
    },
    [],
  );

  const prompt = useCallback(
    (message: string, defaultValue = '', title = 'INPUT', placeholder = '') => {
      return new Promise<string | null>((resolve) => {
        resolveRef.current = resolve as (value: boolean | string | null) => void;
        setConfig({ mode: 'prompt', title, message, defaultValue, placeholder });
      });
    },
    [],
  );

  const close = useCallback((value: boolean | string | null) => {
    const resolve = resolveRef.current;
    resolveRef.current = null;
    setConfig(null);
    if (resolve) resolve(value);
  }, []);

  return { config, confirm, prompt, close };
}
