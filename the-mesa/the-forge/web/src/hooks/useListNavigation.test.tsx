import { describe, it, expect } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { useListNavigation } from './useListNavigation';

function wrapper({ children }: { children: React.ReactNode }) {
  return <MemoryRouter>{children}</MemoryRouter>;
}

describe('useListNavigation', () => {
  it('returns expected values', () => {
    const items = [{ id: '1' }, { id: '2' }, { id: '3' }];

    const { result } = renderHook(() => useListNavigation({ items, basePath: '/docs' }), {
      wrapper,
    });

    expect(result.current.focusedIndex).toBe(-1);
    expect(typeof result.current.handleKeyDown).toBe('function');
    expect(typeof result.current.setFocusedIndex).toBe('function');
  });

  it('returns -1 focusedIndex for empty items', () => {
    const { result } = renderHook(() => useListNavigation({ items: [], basePath: '/docs' }), {
      wrapper,
    });

    expect(result.current.focusedIndex).toBe(-1);
  });

  it('updates focusedIndex on ArrowDown', () => {
    const items = [{ id: '1' }, { id: '2' }];

    const { result } = renderHook(() => useListNavigation({ items, basePath: '/docs' }), {
      wrapper,
    });

    act(() => {
      result.current.handleKeyDown({
        key: 'ArrowDown',
        preventDefault: () => {},
      } as React.KeyboardEvent);
    });

    expect(result.current.focusedIndex).toBe(0);
  });

  it('wraps to 0 on ArrowDown at end of list', () => {
    const items = [{ id: '1' }, { id: '2' }];

    const { result } = renderHook(() => useListNavigation({ items, basePath: '/docs' }), {
      wrapper,
    });

    act(() => {
      result.current.handleKeyDown({
        key: 'ArrowDown',
        preventDefault: () => {},
      } as React.KeyboardEvent);
    });
    act(() => {
      result.current.handleKeyDown({
        key: 'ArrowDown',
        preventDefault: () => {},
      } as React.KeyboardEvent);
    });
    act(() => {
      result.current.handleKeyDown({
        key: 'ArrowDown',
        preventDefault: () => {},
      } as React.KeyboardEvent);
    });

    expect(result.current.focusedIndex).toBe(0);
  });

  it('updates focusedIndex on ArrowUp', () => {
    const items = [{ id: '1' }, { id: '2' }];

    const { result } = renderHook(() => useListNavigation({ items, basePath: '/docs' }), {
      wrapper,
    });

    act(() => {
      result.current.handleKeyDown({
        key: 'ArrowUp',
        preventDefault: () => {},
      } as React.KeyboardEvent);
    });

    expect(result.current.focusedIndex).toBe(1);
  });
});
