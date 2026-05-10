import { useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useSidebar } from '@/hooks/useSidebar';

interface ShortcutCallbacks {
  onSave?: () => void;
  onToggleCommandPalette?: () => void;
  onToggleSidebar?: () => void;
  onToggleFocusMode?: () => void;
  onEscape?: () => void;
}

export function useKeyboardShortcuts(callbacks: ShortcutCallbacks = {}) {
  const navigate = useNavigate();
  const { toggleCollapsed } = useSidebar();
  const callbacksRef = useRef(callbacks);
  useEffect(() => {
    callbacksRef.current = callbacks;
  });

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const mod = e.metaKey || e.ctrlKey;
      const cbs = callbacksRef.current;

      if (e.key === 'Escape') {
        cbs.onEscape?.();
        return;
      }

      if (mod && e.key === 'k') {
        e.preventDefault();
        cbs.onToggleCommandPalette?.();
        return;
      }

      if (mod && e.key === 's') {
        e.preventDefault();
        cbs.onSave?.();
        return;
      }

      if (mod && e.shiftKey && e.key === 'N') {
        e.preventDefault();
        void navigate('/tasks/new');
        return;
      }

      if (mod && !e.shiftKey && e.key === 'n') {
        e.preventDefault();
        void navigate('/knowledge/new');
        return;
      }

      if (mod && e.key === '/') {
        e.preventDefault();
        toggleCollapsed();
        return;
      }

      if (mod && e.key === '.') {
        e.preventDefault();
        cbs.onToggleFocusMode?.();
        return;
      }
    };

    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [navigate, toggleCollapsed]);
}
