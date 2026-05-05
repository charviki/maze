import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import App from './App';

// mock @maze/fabrication 中依赖 Canvas/WebGL 的组件，替换为可测试的占位符
vi.mock('@maze/fabrication', () => ({
  AgentPanel: ({ nodeName }: { nodeName: string }) => (
    <div data-testid="agent-panel">{nodeName}</div>
  ),
  DecryptText: ({ text }: { text: string }) => <span>{text}</span>,
  BootSequence: ({ onComplete }: { onComplete: () => void }) => (
    <button data-testid="boot-sequence" onClick={onComplete}>
      Boot
    </button>
  ),
  TerrainBackground: () => <div data-testid="terrain-bg" />,
  ErrorBoundary: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  Button: ({
    children,
    onClick,
    ...props
  }: {
    children: React.ReactNode;
    onClick?: React.MouseEventHandler;
    [key: string]: unknown;
  }) => (
    <button onClick={onClick} {...props}>
      {children}
    </button>
  ),
  AnimationSettingsProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  AnimationSettingsPanel: () => <div data-testid="anim-settings" />,
  ToastProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

// mock API client 模块，避免真实的网络请求
vi.mock('./api/client', () => ({
  api: {},
}));

describe('App', () => {
  // 应用启动时应先显示 BootSequence 启动动画
  it('should show boot sequence initially', () => {
    render(<App />);
    expect(screen.getByTestId('boot-sequence')).toBeInTheDocument();
  });

  // 启动动画完成后应切换到主界面，显示应用标题
  it('should show main interface after boot completes', async () => {
    const user = userEvent.setup();
    render(<App />);

    await user.click(screen.getByTestId('boot-sequence'));

    expect(screen.getByText('SWEETWATER // FIELD DIAGNOSTIC UNIT')).toBeInTheDocument();
  });

  // 主界面应包含视觉设置按钮，允许用户调整动画效果
  it('should render settings button after boot', async () => {
    const user = userEvent.setup();
    render(<App />);

    await user.click(screen.getByTestId('boot-sequence'));

    expect(screen.getByTitle('Visual Effects')).toBeInTheDocument();
  });
});
