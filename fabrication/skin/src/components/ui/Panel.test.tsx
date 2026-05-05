import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Panel } from './Panel';

describe('Panel', () => {
  // 基础渲染：验证 children 内容能正确显示
  it('should render children', () => {
    render(<Panel>Test Content</Panel>);
    expect(screen.getByText('Test Content')).toBeInTheDocument();
  });

  // 默认样式：default variant 应使用 primary 色调的边框
  it('should apply default variant styles', () => {
    const { container } = render(<Panel>Content</Panel>);
    const borderDiv = container.querySelector('.border-primary\\/30');
    expect(borderDiv).toBeInTheDocument();
  });

  // destructive 变体：应使用红色警告色调的边框
  it('should apply destructive variant styles', () => {
    const { container } = render(<Panel variant="destructive">Content</Panel>);
    const borderDiv = container.querySelector('.border-destructive\\/40');
    expect(borderDiv).toBeInTheDocument();
  });

  // 默认显示四角装饰线（crosshairs），营造科技感 UI
  it('should render crosshairs when showCrosshairs is true (default)', () => {
    const { container } = render(<Panel>Content</Panel>);
    const corners = container.querySelectorAll('.border-t.border-l');
    expect(corners.length).toBeGreaterThan(0);
  });

  // 关闭四角装饰线，用于更简洁的展示场景
  it('should not render crosshairs when showCrosshairs is false', () => {
    const { container } = render(<Panel showCrosshairs={false}>Content</Panel>);
    const authLabel = container.querySelector('[class*="tracking-widest"]');
    expect(authLabel).toBeNull();
  });

  // 自定义 className 应透传到最外层容器
  it('should apply custom className', () => {
    const { container } = render(<Panel className="custom-class">Content</Panel>);
    expect(container.firstElementChild?.classList.contains('custom-class')).toBe(true);
  });

  // clip-path 切角尺寸应反映 cornerSize prop 的值
  it('should render with custom cornerSize via clip-path style', () => {
    const { container } = render(<Panel cornerSize={20}>Content</Panel>);
    const innerDiv = container.querySelector('[style*="clip-path"]');
    expect(innerDiv).toBeInTheDocument();
    expect(innerDiv?.getAttribute('style')).toContain('20px');
  });

  // transparent 模式下背景应为透明，用于叠加在其他组件之上
  it('should render transparent background when transparent is true', () => {
    const { container } = render(<Panel transparent>Content</Panel>);
    const bgDiv = container.querySelector('.bg-transparent');
    expect(bgDiv).toBeInTheDocument();
  });
});
