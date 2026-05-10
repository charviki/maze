export interface Link {
  id: string;
  sourceId: string;
  targetId: string;
  relationType: string;
  sourceTitle: string;
  targetTitle: string;
  createdAt: string;
}

export const links: Link[] = [
  {
    id: 'link-1',
    sourceId: 'mem-3',
    targetId: 'mem-4',
    relationType: 'relates_to',
    sourceTitle: 'Host Reconciliation Design',
    targetTitle: 'Host Scheduling Strategy',
    createdAt: '2025-11-06T12:00:00Z',
  },
  {
    id: 'link-2',
    sourceId: 'mem-3',
    targetId: 'mem-5',
    relationType: 'implements',
    sourceTitle: 'Host Reconciliation Design',
    targetTitle: 'Agent Gateway Architecture',
    createdAt: '2025-11-08T09:00:00Z',
  },
  {
    id: 'link-3',
    sourceId: 'mem-7',
    targetId: 'mem-15',
    relationType: 'reference',
    sourceTitle: 'REST API Conventions',
    targetTitle: 'Error Handling Standards',
    createdAt: '2025-11-22T10:00:00Z',
  },
  {
    id: 'link-4',
    sourceId: 'mem-4',
    targetId: 'mem-12',
    relationType: 'depends_on',
    sourceTitle: 'Host Scheduling Strategy',
    targetTitle: 'RFC: Neural Link Graph',
    createdAt: '2025-11-10T14:00:00Z',
  },
  {
    id: 'link-5',
    sourceId: 'mem-9',
    targetId: 'mem-13',
    relationType: 'reference',
    sourceTitle: 'Database Migration Guide',
    targetTitle: 'Memory Store Evaluation',
    createdAt: '2025-10-26T08:00:00Z',
  },
];
