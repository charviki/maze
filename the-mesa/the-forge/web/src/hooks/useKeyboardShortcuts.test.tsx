import { describe, it, expect } from 'vitest';
import { renderHook } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { useKeyboardShortcuts } from './useKeyboardShortcuts';

vi.mock('@/hooks/useSidebar', () => ({
  useSidebar: () => ({ collapsed: false, toggleCollapsed: () => {} }),
}));

function wrapper({ children }: { children: React.ReactNode }) {
  return <MemoryRouter>{children}</MemoryRouter>;
}

describe('useKeyboardShortcuts', () => {
  it('initializes without errors', () => {
    const { result } = renderHook(() => useKeyboardShortcuts({}), { wrapper });

    expect(result.current).toBeUndefined();
  });

  it('initializes with callbacks', () => {
    const callbacks = {
      onSave: () => {},
      onToggleCommandPalette: () => {},
      onToggleSidebar: () => {},
      onEscape: () => {},
    };

    const { result } = renderHook(() => useKeyboardShortcuts(callbacks), { wrapper });

    expect(result.current).toBeUndefined();
  });
});
