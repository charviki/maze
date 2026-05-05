import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ConfirmDialog } from './ConfirmDialog';

describe('ConfirmDialog', () => {
  const defaultProps = {
    open: true,
    onOpenChange: vi.fn(),
    title: '确认删除',
    description: '确定要删除这个项目吗？',
    onConfirm: vi.fn(),
  };

  // 打开状态下应显示标题和描述文本
  it('should render title and description when open', () => {
    render(<ConfirmDialog {...defaultProps} />);
    expect(screen.getByText('确认删除')).toBeInTheDocument();
    expect(screen.getByText('确定要删除这个项目吗？')).toBeInTheDocument();
  });

  // 自定义按钮文案：confirmLabel / cancelLabel 应覆盖默认的「确认」「取消」
  it('should render custom confirm and cancel labels', () => {
    render(<ConfirmDialog {...defaultProps} confirmLabel="删除" cancelLabel="返回" />);
    expect(screen.getByRole('button', { name: '删除' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '返回' })).toBeInTheDocument();
  });

  // 点击确认按钮应触发 onConfirm 回调，并关闭对话框
  it('should call onConfirm and close dialog when confirm button is clicked', async () => {
    const onConfirm = vi.fn();
    const onOpenChange = vi.fn();
    const user = userEvent.setup();

    render(<ConfirmDialog {...defaultProps} onConfirm={onConfirm} onOpenChange={onOpenChange} />);

    await user.click(screen.getByRole('button', { name: '确认' }));
    expect(onConfirm).toHaveBeenCalledOnce();
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  // 点击取消按钮应仅关闭对话框，不触发确认回调
  it('should close dialog when cancel button is clicked', async () => {
    const onOpenChange = vi.fn();
    const user = userEvent.setup();

    render(<ConfirmDialog {...defaultProps} onOpenChange={onOpenChange} />);

    await user.click(screen.getByRole('button', { name: '取消' }));
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  // 关闭状态下不应渲染任何内容（由 Radix Dialog 控制可见性）
  it('should not render content when open is false', () => {
    render(<ConfirmDialog {...defaultProps} open={false} />);
    expect(screen.queryByText('确认删除')).not.toBeInTheDocument();
  });
});
