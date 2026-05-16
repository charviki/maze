import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import Layout from './Layout';

vi.mock('@/hooks/useIdentity', () => ({
  useIdentity: () => ({ username: 'test', displayName: 'Test' }),
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
  fetchWithAuth: (input: RequestInfo | URL, init?: RequestInit) => fetch(input, init),
}));

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return { ...actual };
});

describe('Layout', () => {
  it('renders sidebar on all routes', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <Layout>
          <div>Dashboard</div>
        </Layout>
      </MemoryRouter>,
    );

    expect(screen.getByText('THE FORGE')).toBeInTheDocument();
    expect(screen.getByText('Dashboard')).toBeInTheDocument();
  });

  it('shows DocTree when on /docs/:archiveId route', () => {
    render(
      <MemoryRouter initialEntries={['/docs/archive-1']}>
        <Layout>
          <div>Doc Page</div>
        </Layout>
      </MemoryRouter>,
    );

    // DocTree heading appears when archiveId is extracted from path
    expect(screen.getByText('DOCUMENTS')).toBeInTheDocument();
    expect(screen.getByTitle('New document')).toBeInTheDocument();
  });

  it('does not show DocTree on Dashboard route', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <Layout>
          <div>Dashboard</div>
        </Layout>
      </MemoryRouter>,
    );

    expect(screen.queryByText('DOCUMENTS')).not.toBeInTheDocument();
    expect(screen.queryByTitle('New document')).not.toBeInTheDocument();
  });

  it('shows DocTree on nested doc route /docs/:archiveId/:docId', () => {
    render(
      <MemoryRouter initialEntries={['/docs/archive-1/doc-123']}>
        <Layout>
          <div>Doc Detail</div>
        </Layout>
      </MemoryRouter>,
    );

    expect(screen.getByText('DOCUMENTS')).toBeInTheDocument();
  });
});
