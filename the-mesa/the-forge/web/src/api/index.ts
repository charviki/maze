import { fetchWithAuth } from '@maze/fabrication';
import type {
  V1Archive,
  V1Doc,
  V1DocLink,
  V1DocTreeNode,
  V1Attachment,
} from '@maze/fabrication/api/gen/models';

// Re-export generated types for convenience
export type Archive = V1Archive;
export type Doc = V1Doc;
export type DocLink = V1DocLink;
export type DocTreeNode = V1DocTreeNode;
export type Attachment = V1Attachment;

export interface ListResponse<T> {
  items: T[];
  total: number;
}

const BASE = '/api/v1';

function qs(params?: Record<string, string | undefined>): string {
  if (!params) return '';
  const entries = Object.entries(params).filter(([, v]) => v != null && v !== '');
  if (entries.length === 0) return '';
  return '?' + new URLSearchParams(entries as [string, string][]).toString();
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 30_000);
  try {
    const res = await fetchWithAuth(`${BASE}${path}`, {
      headers: { 'Content-Type': 'application/json' },
      signal: controller.signal,
      ...init,
    });
    if (!res.ok) {
      const body = (await res.json().catch(() => null)) as { message?: string } | null;
      throw new Error(body?.message ?? `HTTP ${res.status}`);
    }
    const text = await res.text();
    if (!text) return {} as T;
    return JSON.parse(text) as T;
  } finally {
    clearTimeout(timeout);
  }
}

export const archives = {
  async list(): Promise<ListResponse<V1Archive>> {
    const res = await request<{ archives: V1Archive[] }>('/archives');
    return { items: res.archives || [], total: res.archives?.length ?? 0 };
  },
  get(id: string): Promise<V1Archive> {
    return request(`/archives/${id}`);
  },
  create(data: Partial<Pick<V1Archive, 'name' | 'description' | 'icon'>>): Promise<V1Archive> {
    return request('/archives', { method: 'POST', body: JSON.stringify(data) });
  },
  update(
    id: string,
    data: Partial<Pick<V1Archive, 'name' | 'description' | 'icon'>>,
  ): Promise<V1Archive> {
    return request(`/archives/${id}`, { method: 'PUT', body: JSON.stringify(data) });
  },
  async remove(id: string): Promise<void> {
    await request(`/archives/${id}`, { method: 'DELETE' });
  },
};

export const docs = {
  async list(params?: {
    archiveId?: string;
    parentId?: string;
    status?: string;
    visibility?: string;
    author?: string;
  }): Promise<ListResponse<V1Doc>> {
    const res = await request<{ items: V1Doc[]; total: number }>(`/docs${qs(params)}`);
    return { items: res.items || [], total: res.total ?? 0 };
  },
  get(id: string): Promise<V1Doc> {
    return request(`/docs/${id}`);
  },
  create(data: Partial<V1Doc>): Promise<V1Doc> {
    return request('/docs', { method: 'POST', body: JSON.stringify(data) });
  },
  update(id: string, data: Partial<V1Doc>): Promise<V1Doc> {
    return request(`/docs/${id}`, { method: 'PUT', body: JSON.stringify(data) });
  },
  async remove(id: string): Promise<void> {
    await request(`/docs/${id}`, { method: 'DELETE' });
  },
  async search(query: string): Promise<ListResponse<V1Doc>> {
    const res = await request<{ items: V1Doc[] }>(`/docs:search?q=${encodeURIComponent(query)}`);
    return { items: res.items || [], total: res.items?.length ?? 0 };
  },
  async getTree(params?: {
    archiveId?: string;
    parentId?: string;
  }): Promise<ListResponse<V1DocTreeNode>> {
    const res = await request<{ nodes: V1DocTreeNode[] }>(`/docs:tree${qs(params)}`);
    return { items: res.nodes || [], total: res.nodes?.length ?? 0 };
  },
  async getAncestors(id: string): Promise<{ ancestors: V1Doc[] }> {
    return request(`/docs/${id}/ancestors`);
  },
};

export const links = {
  async list(docId: string): Promise<{ links: V1DocLink[] }> {
    return request(`/docs/${docId}/links`);
  },
  create(docId: string, data: { targetId: string; relationType: string }): Promise<V1DocLink> {
    return request(`/docs/${docId}/links`, { method: 'POST', body: JSON.stringify(data) });
  },
  async remove(docId: string, linkId: string): Promise<void> {
    await request(`/docs/${docId}/links/${linkId}`, { method: 'DELETE' });
  },
};
