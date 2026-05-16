import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { ToastProvider } from '@maze/fabrication';
import Dashboard from './Dashboard';

function renderDashboard() {
  return render(
    <MemoryRouter>
      <ToastProvider>
        <Dashboard />
      </ToastProvider>
    </MemoryRouter>,
  );
}

describe('Dashboard', () => {
  it('renders header', async () => {
    renderDashboard();

    await waitFor(() => {
      expect(screen.getByText('THE FORGE')).toBeInTheDocument();
    });
  });

  it('renders stats cards', async () => {
    renderDashboard();

    await waitFor(() => {
      expect(screen.getByText('ARCHIVES')).toBeInTheDocument();
    });

    expect(screen.getByText('ACTIVE')).toBeInTheDocument();
    expect(screen.getByText('PENDING')).toBeInTheDocument();
    expect(screen.getByText('DONE')).toBeInTheDocument();
  });

  it('renders quick action links', async () => {
    renderDashboard();

    await waitFor(() => {
      expect(screen.getByText('THE FORGE')).toBeInTheDocument();
    });

    expect(screen.getByText('NEW DOCUMENT')).toBeInTheDocument();
    expect(screen.getByText('BROWSE ARCHIVES')).toBeInTheDocument();
  });
});
