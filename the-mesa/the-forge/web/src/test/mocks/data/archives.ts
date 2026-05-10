export interface Archive {
  id: string;
  name: string;
  description?: string;
  icon?: string;
  author: string;
  createdAt: string;
  updatedAt: string;
}

export const archives: Archive[] = [
  {
    id: 'archive-1',
    name: 'Project Delos',
    description: 'Core knowledge base for the Delos platform architecture and design decisions',
    icon: '🧠',
    author: 'forge-operator',
    createdAt: '2025-11-01T08:00:00Z',
    updatedAt: '2025-12-15T14:30:00Z',
  },
  {
    id: 'archive-2',
    name: 'Operations Runbook',
    description: 'Operational procedures, incident response playbooks, and SRE guides',
    icon: '🔧',
    author: 'forge-operator',
    createdAt: '2025-10-20T10:00:00Z',
    updatedAt: '2026-01-08T09:15:00Z',
  },
  {
    id: 'archive-3',
    name: 'Research Notes',
    description: 'Exploratory research, RFCs, and technical investigations',
    icon: '🔬',
    author: 'dr-ford',
    createdAt: '2025-09-15T12:00:00Z',
    updatedAt: '2026-02-20T16:45:00Z',
  },
];
