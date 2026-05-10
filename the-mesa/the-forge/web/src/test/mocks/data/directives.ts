export interface Directive {
  id: string;
  title: string;
  description: string;
  status: string;
  priority: string;
  assignee: string;
  author: string;
  requireDocIds: string[];
  narrativeId?: string;
  archiveId: string;
  visibility: string;
  createdAt: string;
  updatedAt: string;
}

export const directives: Directive[] = [
  {
    id: 'dir-1',
    title: 'Implement Host Reconciler',
    description:
      'Build the reconciliation engine that continuously drives host actual state toward desired state based on HostSpec',
    status: 'active',
    priority: 'critical',
    assignee: 'forge-operator',
    author: 'dr-ford',
    requireDocIds: ['mem-3'],
    narrativeId: 'mem-5',
    archiveId: 'archive-1',
    visibility: 'public',
    createdAt: '2025-11-07T08:00:00Z',
    updatedAt: '2026-01-10T14:00:00Z',
  },
  {
    id: 'dir-2',
    title: 'Design Neural Link Visualization',
    description: 'Create interactive force-directed graph for exploring document relationships',
    status: 'active',
    priority: 'high',
    assignee: 'dr-ford',
    author: 'dr-ford',
    requireDocIds: ['mem-12'],
    archiveId: 'archive-3',
    visibility: 'public',
    createdAt: '2025-09-25T10:00:00Z',
    updatedAt: '2026-02-18T11:30:00Z',
  },
  {
    id: 'dir-3',
    title: 'Write API Documentation',
    description: 'Document all REST API endpoints with request/response examples and error codes',
    status: 'done',
    priority: 'normal',
    assignee: 'forge-operator',
    author: 'forge-operator',
    requireDocIds: ['mem-7', 'mem-15'],
    archiveId: 'archive-1',
    visibility: 'public',
    createdAt: '2025-11-12T09:00:00Z',
    updatedAt: '2025-12-20T16:00:00Z',
  },
  {
    id: 'dir-4',
    title: 'Set Up Monitoring Dashboards',
    description: 'Configure Grafana dashboards and Prometheus alert rules for SLO tracking',
    status: 'done',
    priority: 'high',
    assignee: 'ops-team',
    author: 'ops-team',
    requireDocIds: ['mem-11'],
    archiveId: 'archive-2',
    visibility: 'shared',
    createdAt: '2025-10-24T08:00:00Z',
    updatedAt: '2026-01-02T15:30:00Z',
  },
  {
    id: 'dir-5',
    title: 'Migrate Storage Backend',
    description:
      'Evaluate and migrate to PostgreSQL JSONB for knowledge persistence based on research findings',
    status: 'pending',
    priority: 'normal',
    assignee: 'dr-ford',
    author: 'dr-ford',
    requireDocIds: ['mem-13'],
    archiveId: 'archive-3',
    visibility: 'private',
    createdAt: '2025-10-16T11:00:00Z',
    updatedAt: '2025-10-16T11:00:00Z',
  },
  {
    id: 'dir-6',
    title: 'Implement Access Control',
    description:
      'Add visibility and sharing controls to knowledge documents with role-based permissions',
    status: 'active',
    priority: 'high',
    assignee: 'forge-operator',
    author: 'dr-ford',
    requireDocIds: ['mem-5', 'mem-7'],
    narrativeId: 'mem-5',
    archiveId: 'archive-1',
    visibility: 'public',
    createdAt: '2025-12-01T08:00:00Z',
    updatedAt: '2026-02-28T10:00:00Z',
  },
  {
    id: 'dir-7',
    title: 'Database Migration Automation',
    description: 'Build CI pipeline to automatically validate and apply database migrations',
    status: 'pending',
    priority: 'low',
    assignee: 'ops-team',
    author: 'ops-team',
    requireDocIds: ['mem-9'],
    archiveId: 'archive-2',
    visibility: 'public',
    createdAt: '2025-12-21T09:00:00Z',
    updatedAt: '2025-12-21T09:00:00Z',
  },
  {
    id: 'dir-8',
    title: 'Host Scheduling Algorithm',
    description: 'Implement the scheduling algorithm with resource and affinity constraints',
    status: 'failed',
    priority: 'critical',
    assignee: 'forge-operator',
    author: 'dr-ford',
    requireDocIds: ['mem-4'],
    narrativeId: 'mem-5',
    archiveId: 'archive-1',
    visibility: 'shared',
    createdAt: '2025-11-06T10:00:00Z',
    updatedAt: '2026-01-15T16:00:00Z',
  },
  {
    id: 'dir-9',
    title: 'Onboarding Guide Update',
    description:
      'Update the quick start guide with current setup instructions and troubleshooting tips',
    status: 'done',
    priority: 'low',
    assignee: 'forge-operator',
    author: 'forge-operator',
    requireDocIds: ['mem-14'],
    archiveId: 'archive-1',
    visibility: 'public',
    createdAt: '2026-02-25T08:00:00Z',
    updatedAt: '2026-03-01T10:00:00Z',
  },
  {
    id: 'dir-10',
    title: 'Incident Response Drill',
    description:
      'Conduct a P1 simulation drill and update the incident response playbook based on findings',
    status: 'pending',
    priority: 'normal',
    assignee: 'ops-team',
    author: 'ops-team',
    requireDocIds: ['mem-8'],
    archiveId: 'archive-2',
    visibility: 'shared',
    createdAt: '2026-03-05T09:00:00Z',
    updatedAt: '2026-03-05T09:00:00Z',
  },
];
