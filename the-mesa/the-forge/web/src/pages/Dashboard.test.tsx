import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import Dashboard from './Dashboard';

function renderDashboard() {
  return render(
    <MemoryRouter>
      <Dashboard />
    </MemoryRouter>,
  );
}

describe('Dashboard', () => {
  it('renders stats cards after loading', async () => {
    renderDashboard();

    await waitFor(() => {
      expect(screen.getByText('TOTAL MEMORIES')).toBeInTheDocument();
    });
    expect(screen.getByText('TOTAL DIRECTIVES')).toBeInTheDocument();
    expect(screen.getByText('DIRECTIVE STATUS')).toBeInTheDocument();
  });

  it('displays memory and directive counts from API', async () => {
    renderDashboard();

    await waitFor(() => {
      expect(screen.getByText('TOTAL MEMORIES')).toBeInTheDocument();
    });

    expect(screen.getByText('15')).toBeInTheDocument();
    expect(screen.getByText('10')).toBeInTheDocument();
  });

  it('renders recent memories list', async () => {
    renderDashboard();

    await waitFor(() => {
      expect(screen.getByText('RECENT MEMORIES')).toBeInTheDocument();
    });

    expect(screen.getByText('Quick Start Guide')).toBeInTheDocument();
  });

  it('renders directive status badges', async () => {
    renderDashboard();

    await waitFor(() => {
      expect(screen.getByText('DIRECTIVE STATUS')).toBeInTheDocument();
    });

    expect(screen.getByText(/pending:/)).toBeInTheDocument();
    expect(screen.getByText(/active:/)).toBeInTheDocument();
  });
});
