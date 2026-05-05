import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/react';
import { Skeleton } from './Skeleton';

describe('Skeleton', () => {
  // 应包含 animate-pulse 动画 class，表示正在加载
  it('should render with animate-pulse class', () => {
    const { container } = render(<Skeleton />);
    const el = container.firstElementChild as HTMLElement;
    expect(el.classList.contains('animate-pulse')).toBe(true);
  });

  // 应使用 primary 色调的半透明背景，匹配项目主题
  it('should render with bg-primary/10 class', () => {
    const { container } = render(<Skeleton />);
    const el = container.firstElementChild as HTMLElement;
    expect(el.classList.contains('bg-primary/10')).toBe(true);
  });

  // 自定义 className 应透传，允许调用方控制尺寸（如 h-4 w-full）
  it('should apply custom className', () => {
    const { container } = render(<Skeleton className="h-4 w-full" />);
    const el = container.firstElementChild as HTMLElement;
    expect(el.classList.contains('h-4')).toBe(true);
    expect(el.classList.contains('w-full')).toBe(true);
  });

  // 应包含 clip-path polygon 样式，与其他 Panel 组件的切角风格保持一致
  it('should have clip-path style for corner-cut effect', () => {
    const { container } = render(<Skeleton />);
    const el = container.firstElementChild as HTMLElement;
    expect(el.style.clipPath).toBeDefined();
    expect(el.style.clipPath).toContain('polygon');
  });
});
