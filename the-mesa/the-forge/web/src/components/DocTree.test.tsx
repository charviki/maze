import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import DocTree from './DocTree';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({ docId: undefined }),
  };
});

vi.mock('@maze/fabrication', () => ({
  fetchWithAuth: (_url: RequestInfo | URL, _init?: RequestInit) =>
    Promise.resolve({
      ok: true,
      json: () => Promise.resolve({}),
      text: () => Promise.resolve('{}'),
    }),
}));

describe('DocTree', () => {
  it('renders empty state when tree is empty', async () => {
    render(
      <MemoryRouter>
        <DocTree archiveId="archive-1" width={280} onResize={() => {}} />
      </MemoryRouter>,
    );

    await waitFor(() => {
      expect(screen.getByText('DOCUMENTS')).toBeInTheDocument();
    });

    expect(screen.getByText('No documents yet')).toBeInTheDocument();
  });
});
