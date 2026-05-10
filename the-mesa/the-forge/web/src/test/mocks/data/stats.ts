import type { Memory } from './memories.ts';

export interface Stats {
  totalMemories: number;
  totalDirectives: number;
  directivesByStatus: Record<string, number>;
  recentMemories: Memory[];
}

export const stats: Stats = {
  totalMemories: 15,
  totalDirectives: 10,
  directivesByStatus: {
    pending: 3,
    active: 3,
    done: 3,
    failed: 1,
  },
  recentMemories: [
    {
      id: 'mem-14',
      archiveId: 'archive-1',
      kind: 'doc',
      title: 'Quick Start Guide',
      content: '',
      type: 'shared',
      tags: ['onboarding', 'guide'],
      author: 'forge-operator',
      visibility: 'public',
      createdAt: '2025-11-15T08:00:00Z',
      updatedAt: '2026-03-01T10:00:00Z',
    },
    {
      id: 'mem-12',
      archiveId: 'archive-3',
      kind: 'doc',
      title: 'RFC: Neural Link Graph',
      content: '',
      type: 'narrative',
      tags: ['rfc', 'visualization', 'graph'],
      author: 'dr-ford',
      visibility: 'public',
      createdAt: '2025-09-20T13:00:00Z',
      updatedAt: '2026-02-15T11:00:00Z',
    },
    {
      id: 'mem-11',
      archiveId: 'archive-2',
      parentId: 'mem-10',
      kind: 'doc',
      title: 'SLO Dashboard Configuration',
      content: '',
      type: 'ops',
      tags: ['ops', 'monitoring', 'slo'],
      author: 'ops-team',
      visibility: 'shared',
      createdAt: '2025-10-23T11:00:00Z',
      updatedAt: '2026-01-02T15:30:00Z',
    },
  ],
};
