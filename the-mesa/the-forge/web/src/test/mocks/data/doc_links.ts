import type { DocLink } from '@/api';

export const docLinks: DocLink[] = [
  {
    id: 'link-1',
    sourceId: 'doc-task-1',
    targetId: 'doc-recon',
    relationType: 'implements',
    sourceTitle: 'Implement Host Reconciler',
    targetTitle: 'Host Reconciliation Design',
    createdAt: '2025-11-08T09:00:00Z',
  },
  {
    id: 'link-2',
    sourceId: 'doc-task-1',
    targetId: 'doc-rest',
    relationType: 'reference',
    sourceTitle: 'Implement Host Reconciler',
    targetTitle: 'REST API Conventions',
    createdAt: '2025-11-09T10:00:00Z',
  },
];
