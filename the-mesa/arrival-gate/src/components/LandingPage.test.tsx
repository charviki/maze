import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { AnimationSettingsProvider } from '@maze/fabrication';
import { LandingPage } from './LandingPage';

// Mock auth module
vi.mock('../auth/auth', () => ({
  login: vi.fn((user: string, pass: string) => user === 'admin' && pass === 'admin'),
  logout: vi.fn(),
  isAuthenticated: vi.fn(() => false),
  getCurrentUser: vi.fn(() => null),
}));

// Mock Canvas components (jsdom doesn't support Canvas API)
vi.mock('./MazeCanvas', () => ({
  MazeCanvas: ({ className }: { className?: string }) => (
    <div data-testid="maze-canvas" className={className} />
  ),
}));

vi.mock('./MazeSvg', () => ({
  MazeSvg: () => <div data-testid="maze-svg" />,
}));

// Mock fabrication components that need Canvas
vi.mock('@maze/fabrication', async () => {
  const actual = await vi.importActual('@maze/fabrication');
  return {
    ...actual,
    TerrainBackground: () => <div data-testid="terrain-bg" />,
    HexWaterfall: () => <div data-testid="hex-waterfall" />,
  };
});

function renderLanding(onEnter: () => void) {
  return render(
    <AnimationSettingsProvider>
      <LandingPage onEnter={onEnter} />
    </AnimationSettingsProvider>,
  );
}

describe('LandingPage', () => {
  const onEnter = vi.fn();

  beforeEach(() => {
    onEnter.mockClear();
  });

  describe('landing phase', () => {
    it('renders DELOS branding', () => {
      renderLanding(onEnter);
      expect(screen.getByText('DELOS INCORPORATED')).toBeInTheDocument();
    });

    it('renders THE MAZE title', () => {
      renderLanding(onEnter);
      expect(screen.getByText('THE MAZE')).toBeInTheDocument();
    });

    it('renders ENTER THE PARK button', () => {
      renderLanding(onEnter);
      expect(screen.getByText('ENTER THE PARK')).toBeInTheDocument();
    });

    it('renders maze components', () => {
      renderLanding(onEnter);
      // Landing maze + login mini maze are both present in DOM
      expect(screen.getAllByTestId('maze-canvas').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByTestId('maze-svg').length).toBeGreaterThanOrEqual(1);
    });

    it('transitions to login phase on button click', () => {
      renderLanding(onEnter);
      fireEvent.click(screen.getByText('ENTER THE PARK'));

      // Login form should appear
      expect(screen.getByText('IDENTITY VERIFICATION')).toBeInTheDocument();
      expect(screen.getByText('AUTHENTICATE')).toBeInTheDocument();
    });
  });

  describe('login phase', () => {
    function goToLogin() {
      renderLanding(onEnter);
      fireEvent.click(screen.getByText('ENTER THE PARK'));
    }

    it('renders login form inputs', () => {
      goToLogin();
      const inputs = screen.getAllByRole('textbox');
      expect(inputs.length).toBeGreaterThanOrEqual(1);
    });

    it('shows error on invalid credentials', () => {
      goToLogin();
      const inputs = screen.getAllByRole('textbox');
      // First text input is ACCESS ID
      fireEvent.change(inputs[0], { target: { value: 'wrong' } });
      // Password input (not textbox role, query by container)
      const passwordInput = document.querySelector('input[type="password"]')!;
      fireEvent.change(passwordInput, { target: { value: 'wrong' } });
      fireEvent.click(screen.getByText('AUTHENTICATE'));

      expect(screen.getByText(/ACCESS DENIED/)).toBeInTheDocument();
    });

    it('calls onEnter on valid credentials', async () => {
      goToLogin();
      const inputs = screen.getAllByRole('textbox');
      fireEvent.change(inputs[0], { target: { value: 'admin' } });
      const passwordInput = document.querySelector('input[type="password"]')!;
      fireEvent.change(passwordInput, { target: { value: 'admin' } });
      fireEvent.click(screen.getByText('AUTHENTICATE'));

      // onEnter is called after fade-out delay
      await new Promise((r) => setTimeout(r, 600));
      expect(onEnter).toHaveBeenCalled();
    });

    it('mini maze is visible in login phase', () => {
      goToLogin();
      expect(screen.getByTitle('Back to entrance')).toBeInTheDocument();
    });

    it('clicking mini maze returns to landing phase', () => {
      goToLogin();
      expect(screen.getByText('IDENTITY VERIFICATION')).toBeInTheDocument();

      fireEvent.click(screen.getByTitle('Back to entrance'));

      // Landing content should be back
      expect(screen.getByText('ENTER THE PARK')).toBeInTheDocument();
    });

    it('renders protocol hint', () => {
      goToLogin();
      expect(screen.getByText(/PROTOCOL: OIDC-READY/)).toBeInTheDocument();
    });
  });
});
