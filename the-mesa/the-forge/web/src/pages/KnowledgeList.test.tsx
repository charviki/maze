import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import KnowledgeList from './KnowledgeList';

function renderKnowledgeList(initialEntries?: string[]) {
  return render(
    <MemoryRouter initialEntries={initialEntries || ['/knowledge']}>
      <KnowledgeList />
    </MemoryRouter>,
  );
}

describe('KnowledgeList', () => {
  it('renders memory list after loading', async () => {
    renderKnowledgeList();

    await waitFor(() => {
      expect(screen.getByText('MEMORIES')).toBeInTheDocument();
    });

    expect(screen.getByText('NEW')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Search documents...')).toBeInTheDocument();
  });

  it('renders archive-filtered memories when archiveId is provided', async () => {
    renderKnowledgeList(['/knowledge?archiveId=archive-1']);

    await waitFor(() => {
      expect(screen.getByText('Architecture')).toBeInTheDocument();
    });

    expect(screen.getByText('API Design')).toBeInTheDocument();
  });

  it('renders tree toggle button', async () => {
    renderKnowledgeList();

    await waitFor(() => {
      expect(screen.getByText('TREE')).toBeInTheDocument();
    });
  });

  it('renders type filter dropdown', async () => {
    renderKnowledgeList();

    await waitFor(() => {
      expect(screen.getByText('MEMORIES')).toBeInTheDocument();
    });

    expect(screen.getByText('ALL TYPES')).toBeInTheDocument();
  });
});
