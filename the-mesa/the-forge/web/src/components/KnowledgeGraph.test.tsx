import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import KnowledgeGraph from './KnowledgeGraph';

function createMockSimulation() {
  const sim = {
    force: vi.fn().mockReturnThis(),
    stop: vi.fn(),
    on: vi.fn().mockReturnThis(),
  };
  return sim;
}

vi.mock('d3-force', () => ({
  forceSimulation: () => createMockSimulation(),
  forceLink: () => ({ id: () => ({ distance: () => {} }) }),
  forceManyBody: () => ({ strength: () => {} }),
  forceCenter: () => ({}),
  forceCollide: () => ({}),
}));

describe('KnowledgeGraph', () => {
  it('renders empty state when no links provided', () => {
    render(<KnowledgeGraph links={[]} />);

    expect(screen.getByText('No document links')).toBeInTheDocument();
  });

  it('renders svg container when links are provided', () => {
    const links = [
      {
        id: 'link-1',
        sourceId: 'mem-1',
        targetId: 'mem-2',
        relationType: 'reference',
        sourceTitle: 'Source',
        targetTitle: 'Target',
        createdAt: '2026-01-01T00:00:00Z',
      },
    ];

    const { container } = render(<KnowledgeGraph links={links} />);

    const svg = container.querySelector('svg');
    expect(svg).toBeInTheDocument();
  });

  it('renders with custom width and height', () => {
    const links = [
      {
        id: 'link-1',
        sourceId: 'mem-1',
        targetId: 'mem-2',
        relationType: 'reference',
        sourceTitle: 'Source',
        targetTitle: 'Target',
        createdAt: '2026-01-01T00:00:00Z',
      },
    ];

    const { container } = render(<KnowledgeGraph links={links} width={800} height={600} />);

    const svg = container.querySelector('svg');
    expect(svg).toHaveAttribute('width', '800');
    expect(svg).toHaveAttribute('height', '600');
  });
});
