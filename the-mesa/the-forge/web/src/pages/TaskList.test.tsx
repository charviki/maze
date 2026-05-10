import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import TaskList from './TaskList';

function renderTaskList(initialEntries?: string[]) {
  return render(
    <MemoryRouter initialEntries={initialEntries || ['/tasks']}>
      <TaskList />
    </MemoryRouter>,
  );
}

describe('TaskList', () => {
  it('renders directive list after loading', async () => {
    renderTaskList();

    await waitFor(() => {
      expect(screen.getByText('DIRECTIVES')).toBeInTheDocument();
    });

    expect(screen.getByText('NEW DIRECTIVE')).toBeInTheDocument();
  });

  it('renders task items from API', async () => {
    renderTaskList();

    await waitFor(() => {
      expect(screen.getByText('Implement Host Reconciler')).toBeInTheDocument();
    });

    expect(screen.getByText('Design Neural Link Visualization')).toBeInTheDocument();
  });

  it('renders filter controls', async () => {
    renderTaskList();

    await waitFor(() => {
      expect(screen.getByText('DIRECTIVES')).toBeInTheDocument();
    });

    expect(screen.getByText('ALL STATUS')).toBeInTheDocument();
    expect(screen.getByText('ALL PRIORITY')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Search directives...')).toBeInTheDocument();
  });

  it('renders status and priority badges on tasks', async () => {
    renderTaskList();

    await waitFor(() => {
      expect(screen.getByText('Implement Host Reconciler')).toBeInTheDocument();
    });

    const activeBadges = screen.getAllByText('active');
    expect(activeBadges.length).toBeGreaterThanOrEqual(1);

    const criticalBadges = screen.getAllByText('critical');
    expect(criticalBadges.length).toBeGreaterThanOrEqual(1);
  });
});
