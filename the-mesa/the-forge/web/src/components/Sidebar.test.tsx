import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import Sidebar from './Sidebar';

vi.mock('@/hooks/useIdentity', () => ({
  useIdentity: () => ({ username: 'forge-operator', displayName: 'Forge Operator' }),
}));

vi.mock('@maze/fabrication', () => ({
  useToast: () => ({ showToast: () => {} }),
  Skeleton: ({ className }: { className?: string }) => <div className={className} />,
  Dialog: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  DialogContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  DialogHeader: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  DialogTitle: ({ children, className }: { children: React.ReactNode; className?: string }) => (
    <div className={className}>{children}</div>
  ),
  DialogFooter: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Button: ({ children, onClick }: { children: React.ReactNode; onClick?: () => void }) => (
    <button onClick={onClick}>{children}</button>
  ),
}));

function renderSidebar() {
  return render(
    <MemoryRouter>
      <Sidebar />
    </MemoryRouter>,
  );
}

describe('Sidebar', () => {
  it('renders sidebar with navigation links', async () => {
    renderSidebar();

    await waitFor(() => {
      expect(screen.getByText('THE FORGE')).toBeInTheDocument();
    });

    expect(screen.getByText('DASHBOARD')).toBeInTheDocument();
    expect(screen.getByText('DIRECTIVES')).toBeInTheDocument();
  });

  it('renders archive list from API', async () => {
    renderSidebar();

    await waitFor(() => {
      expect(screen.getByText('Project Delos')).toBeInTheDocument();
    });

    expect(screen.getByText('Operations Runbook')).toBeInTheDocument();
    expect(screen.getByText('Research Notes')).toBeInTheDocument();
  });

  it('renders create archive button', async () => {
    renderSidebar();

    await waitFor(() => {
      expect(screen.getByText('THE FORGE')).toBeInTheDocument();
    });

    expect(screen.getByText('ARCHIVES')).toBeInTheDocument();
  });

  it('displays username', async () => {
    renderSidebar();

    await waitFor(() => {
      expect(screen.getByText('forge-operator')).toBeInTheDocument();
    });
  });
});
