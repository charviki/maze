import { useState, useCallback, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

interface UseListNavigationOptions {
  items: Array<{ id: string }>;
  basePath: string;
}

export function useListNavigation({ items, basePath }: UseListNavigationOptions) {
  const navigate = useNavigate();
  const [focusedIndex, setFocusedIndex] = useState(-1);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (!items.length) return;

      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setFocusedIndex((prev) => (prev < items.length - 1 ? prev + 1 : 0));
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        setFocusedIndex((prev) => (prev > 0 ? prev - 1 : items.length - 1));
      } else if (e.key === 'Enter' && focusedIndex >= 0) {
        e.preventDefault();
        void navigate(`${basePath}/${items[focusedIndex].id}`);
      }
    },
    [items, basePath, focusedIndex, navigate],
  );

  useEffect(() => {
    if (focusedIndex >= items.length) {
      const t = setTimeout(() => {
        setFocusedIndex(items.length > 0 ? items.length - 1 : -1);
      }, 0);
      return () => clearTimeout(t);
    }
  }, [items.length, focusedIndex]);

  return { focusedIndex, setFocusedIndex, handleKeyDown };
}
