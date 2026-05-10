export interface Memory {
  id: string;
  archiveId: string;
  parentId?: string;
  kind: string;
  title: string;
  content: string;
  type: string;
  summary?: string;
  tags: string[];
  author: string;
  visibility: string;
  sharedWith?: string[];
  attachments?: { id: string; key: string; name: string; contentType: string; size: string }[];
  createdAt: string;
  updatedAt: string;
}

export const memories: Memory[] = [
  {
    id: 'mem-1',
    archiveId: 'archive-1',
    kind: 'folder',
    title: 'Architecture',
    content: '',
    type: 'shared',
    summary: 'System architecture documents and design decisions',
    tags: ['architecture', 'design'],
    author: 'forge-operator',
    visibility: 'public',
    createdAt: '2025-11-02T08:00:00Z',
    updatedAt: '2025-12-10T10:00:00Z',
  },
  {
    id: 'mem-2',
    archiveId: 'archive-1',
    parentId: 'mem-1',
    kind: 'folder',
    title: 'Host Management',
    content: '',
    type: 'shared',
    summary: 'Host lifecycle, reconciliation, and scheduling',
    tags: ['architecture', 'hosts'],
    author: 'forge-operator',
    visibility: 'public',
    createdAt: '2025-11-03T09:00:00Z',
    updatedAt: '2025-11-20T11:00:00Z',
  },
  {
    id: 'mem-3',
    archiveId: 'archive-1',
    parentId: 'mem-2',
    kind: 'doc',
    title: 'Host Reconciliation Design',
    content:
      '# Host Reconciliation\n\nThe reconciler continuously drives actual state toward desired state...\n\n## Reconciliation Loop\n\n1. Fetch HostSpec from store\n2. Compare with current host state\n3. Apply necessary changes\n4. Report status back',
    type: 'requirement',
    summary:
      'Describes the reconciliation loop that ensures host actual state matches desired state',
    tags: ['architecture', 'reconciliation', 'hosts'],
    author: 'forge-operator',
    visibility: 'public',
    createdAt: '2025-11-04T10:00:00Z',
    updatedAt: '2025-12-05T14:30:00Z',
  },
  {
    id: 'mem-4',
    archiveId: 'archive-1',
    parentId: 'mem-2',
    kind: 'doc',
    title: 'Host Scheduling Strategy',
    content:
      '# Scheduling Strategy\n\nHosts are scheduled based on resource availability and affinity rules...',
    type: 'requirement',
    summary:
      'Defines how hosts are scheduled across the cluster based on resource and affinity constraints',
    tags: ['architecture', 'scheduling'],
    author: 'dr-ford',
    visibility: 'shared',
    sharedWith: ['forge-operator', 'ops-team'],
    createdAt: '2025-11-05T11:00:00Z',
    updatedAt: '2025-11-28T16:00:00Z',
  },
  {
    id: 'mem-5',
    archiveId: 'archive-1',
    parentId: 'mem-1',
    kind: 'doc',
    title: 'Agent Gateway Architecture',
    content:
      '# Agent Gateway\n\nAll frontend requests are proxied through the Manager gateway...\n\n## Request Flow\n\nFrontend → Manager Gateway → Agent Node\n\n## Benefits\n\n- Audit logging\n- Access control\n- Connection pooling',
    type: 'narrative',
    summary: 'Architecture overview of the agent proxy gateway with audit and access control',
    tags: ['architecture', 'gateway', 'proxy'],
    author: 'forge-operator',
    visibility: 'public',
    createdAt: '2025-11-06T08:30:00Z',
    updatedAt: '2025-12-15T14:30:00Z',
  },
  {
    id: 'mem-6',
    archiveId: 'archive-1',
    kind: 'folder',
    title: 'API Design',
    content: '',
    type: 'shared',
    summary: 'API conventions, proto definitions, and HTTP mapping',
    tags: ['api', 'proto'],
    author: 'forge-operator',
    visibility: 'public',
    createdAt: '2025-11-10T09:00:00Z',
    updatedAt: '2025-12-01T10:00:00Z',
  },
  {
    id: 'mem-7',
    archiveId: 'archive-1',
    parentId: 'mem-6',
    kind: 'doc',
    title: 'REST API Conventions',
    content:
      '# REST API Conventions\n\n## Naming\n\n- Use plural resource nouns\n- Use camelCase for JSON fields\n- Proto field `refresh_token` serializes as `refreshToken`\n\n## Response Format\n\nAll responses use `{ status: "ok", data: {...} }` envelope',
    type: 'requirement',
    summary: 'Documents the REST API naming conventions and response format standards',
    tags: ['api', 'conventions'],
    author: 'forge-operator',
    visibility: 'public',
    createdAt: '2025-11-11T10:00:00Z',
    updatedAt: '2025-11-30T12:00:00Z',
  },
  {
    id: 'mem-8',
    archiveId: 'archive-2',
    kind: 'doc',
    title: 'Incident Response Playbook',
    content:
      '# Incident Response\n\n## Severity Levels\n\n- P0: Total system outage\n- P1: Partial outage affecting multiple users\n- P2: Degraded performance\n- P3: Minor issues\n\n## Escalation Path\n\n1. On-call engineer\n2. Team lead\n3. Engineering manager',
    type: 'ops',
    summary:
      'Standard operating procedure for handling production incidents at different severity levels',
    tags: ['ops', 'incident', 'runbook'],
    author: 'ops-team',
    visibility: 'shared',
    sharedWith: ['forge-operator', 'dr-ford'],
    createdAt: '2025-10-21T08:00:00Z',
    updatedAt: '2026-01-05T10:00:00Z',
  },
  {
    id: 'mem-9',
    archiveId: 'archive-2',
    kind: 'doc',
    title: 'Database Migration Guide',
    content:
      '# Database Migrations\n\n## Rules\n\n- Never modify a released migration\n- Always test with production data volume\n- Use explicit transactions\n\n## Commands\n\n```bash\nmake migrate-up\nmake migrate-down\n```',
    type: 'ops',
    summary: 'Guidelines for safely creating and applying database schema migrations',
    tags: ['ops', 'database', 'migration'],
    author: 'ops-team',
    visibility: 'public',
    createdAt: '2025-10-25T14:00:00Z',
    updatedAt: '2025-12-20T09:00:00Z',
  },
  {
    id: 'mem-10',
    archiveId: 'archive-2',
    kind: 'folder',
    title: 'Monitoring & Alerting',
    content: '',
    type: 'shared',
    summary: 'Monitoring dashboards, alert rules, and SLI/SLO definitions',
    tags: ['ops', 'monitoring'],
    author: 'ops-team',
    visibility: 'public',
    createdAt: '2025-10-22T10:00:00Z',
    updatedAt: '2026-01-08T09:15:00Z',
  },
  {
    id: 'mem-11',
    archiveId: 'archive-2',
    parentId: 'mem-10',
    kind: 'doc',
    title: 'SLO Dashboard Configuration',
    content:
      '# SLO Dashboard\n\n## Key Metrics\n\n- Availability: 99.9%\n- Latency P99: < 500ms\n- Error Rate: < 0.1%\n\n## Alert Rules\n\n- Burn rate > 14.4x → P0\n- Burn rate > 6x → P1',
    type: 'ops',
    summary: 'Configuration for SLO dashboards with availability, latency, and error rate targets',
    tags: ['ops', 'monitoring', 'slo'],
    author: 'ops-team',
    visibility: 'shared',
    sharedWith: ['forge-operator'],
    createdAt: '2025-10-23T11:00:00Z',
    updatedAt: '2026-01-02T15:30:00Z',
  },
  {
    id: 'mem-12',
    archiveId: 'archive-3',
    kind: 'doc',
    title: 'RFC: Neural Link Graph',
    content:
      '# RFC: Neural Link Graph\n\n## Motivation\n\nWe need a visual way to explore relationships between knowledge documents...\n\n## Proposal\n\nUse d3-force to render an interactive graph where nodes are documents and edges are neural links.',
    type: 'narrative',
    summary:
      'Proposal for an interactive knowledge graph visualization using force-directed layout',
    tags: ['rfc', 'visualization', 'graph'],
    author: 'dr-ford',
    visibility: 'public',
    createdAt: '2025-09-20T13:00:00Z',
    updatedAt: '2026-02-15T11:00:00Z',
  },
  {
    id: 'mem-13',
    archiveId: 'archive-3',
    kind: 'doc',
    title: 'Memory Store Evaluation',
    content:
      '# Memory Store Evaluation\n\nCompared PostgreSQL JSONB vs dedicated document stores for knowledge persistence...\n\n## Conclusion\n\nPostgreSQL JSONB with GIN indexing provides the best balance of query flexibility and operational simplicity.',
    type: 'memory',
    summary: 'Technical evaluation of storage backends for the knowledge memory system',
    tags: ['research', 'storage', 'evaluation'],
    author: 'dr-ford',
    visibility: 'private',
    createdAt: '2025-10-01T09:00:00Z',
    updatedAt: '2025-10-15T16:00:00Z',
  },
  {
    id: 'mem-14',
    archiveId: 'archive-1',
    kind: 'doc',
    title: 'Quick Start Guide',
    content:
      '# Quick Start\n\n1. Clone the repository\n2. Run `make dev`\n3. Open `http://localhost:5173`\n\n## Environment Variables\n\n- `DATABASE_URL`\n- `REDIS_URL`\n- `JWT_SECRET`',
    type: 'shared',
    summary: 'Getting started guide for new developers joining the project',
    tags: ['onboarding', 'guide'],
    author: 'forge-operator',
    visibility: 'public',
    attachments: [
      {
        id: 'att-1',
        key: 'uploads/env-example',
        name: '.env.example',
        contentType: 'text/plain',
        size: '256',
      },
    ],
    createdAt: '2025-11-15T08:00:00Z',
    updatedAt: '2026-03-01T10:00:00Z',
  },
  {
    id: 'mem-15',
    archiveId: 'archive-1',
    parentId: 'mem-6',
    kind: 'doc',
    title: 'Error Handling Standards',
    content:
      '# Error Handling\n\n## HTTP Status Codes\n\n- 400: Validation errors\n- 401: Unauthorized\n- 403: Forbidden\n- 404: Not found\n- 409: Conflict\n- 500: Internal error\n\n## Error Response Body\n\n```json\n{ "status": "error", "error": { "code": "VALIDATION_ERROR", "message": "..." } }\n```',
    type: 'requirement',
    summary: 'Standard error response format and HTTP status code usage across all API endpoints',
    tags: ['api', 'errors', 'standards'],
    author: 'forge-operator',
    visibility: 'public',
    createdAt: '2025-11-20T14:00:00Z',
    updatedAt: '2025-12-18T09:30:00Z',
  },
];
