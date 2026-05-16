import { http, HttpResponse } from 'msw';
import { archives } from './data/archives.ts';
import { docs } from './data/docs.ts';
import { docLinks } from './data/doc_links.ts';
import type { Doc, DocTreeNode } from '@/api';

function notFound(message: string) {
  return HttpResponse.json({ message }, { status: 404 });
}

let docStore = [...docs];
let archiveStore = [...archives];
let linkStore = [...docLinks];

export function resetMockData() {
  docStore = [...docs];
  archiveStore = [...archives];
  linkStore = [...docLinks];
}

function buildTree(items: Doc[], parentId?: string): DocTreeNode[] {
  return items
    .filter((d) => (parentId ? d.parentId === parentId : !d.parentId))
    .map((d) => ({
      doc: d,
      children: buildTree(items, d.id),
    }));
}

function getAncestors(items: Doc[], id: string): Doc[] {
  const chain: Doc[] = [];
  let current = items.find((d) => d.id === id);
  while (current) {
    chain.unshift(current);
    if (!current.parentId) break;
    current = items.find((d) => d.id === current!.parentId);
  }
  return chain;
}

function str(val: unknown, fallback = ''): string {
  if (val == null) return fallback;
  return typeof val === 'string' ? val : fallback;
}

function strArr(val: unknown): string[] {
  return Array.isArray(val) ? (val as string[]) : [];
}

function strOrUndef(val: unknown): string | undefined {
  if (typeof val === 'string' && val.length > 0) return val;
  return undefined;
}

export const handlers = [
  // ── Archives ──────────────────────────────────────────────

  http.get('/api/v1/archives', () => {
    return HttpResponse.json({ archives: archiveStore });
  }),

  http.get('/api/v1/archives/:id', ({ params }) => {
    const archive = archiveStore.find((a) => a.id === params.id);
    if (!archive) return notFound('Archive not found');
    return HttpResponse.json(archive);
  }),

  http.post('/api/v1/archives', async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    const now = new Date().toISOString();
    const archive = {
      id: `archive-${Date.now()}`,
      name: str(body.name),
      description: strOrUndef(body.description),
      icon: strOrUndef(body.icon),
      author: 'forge-operator',
      createdAt: now,
      updatedAt: now,
    };
    archiveStore.push(archive);
    return HttpResponse.json(archive);
  }),

  http.put('/api/v1/archives/:id', async ({ params, request }) => {
    const idx = archiveStore.findIndex((a) => a.id === params.id);
    if (idx === -1) return notFound('Archive not found');
    const body = (await request.json()) as Record<string, unknown>;
    const updated = {
      ...archiveStore[idx],
      ...(body.name != null ? { name: str(body.name) } : {}),
      ...(body.description != null ? { description: strOrUndef(body.description) } : {}),
      ...(body.icon != null ? { icon: strOrUndef(body.icon) } : {}),
      updatedAt: new Date().toISOString(),
    };
    archiveStore[idx] = updated;
    return HttpResponse.json(updated);
  }),

  http.delete('/api/v1/archives/:id', ({ params }) => {
    const idx = archiveStore.findIndex((a) => a.id === params.id);
    if (idx === -1) return notFound('Archive not found');
    archiveStore.splice(idx, 1);
    return HttpResponse.json({});
  }),

  // ── Docs ──────────────────────────────────────────────────

  http.get('/api/v1/docs', ({ request }) => {
    const url = new URL(request.url);
    const archiveId = url.searchParams.get('archiveId');
    const parentId = url.searchParams.get('parentId');
    const status = url.searchParams.get('status');
    const visibility = url.searchParams.get('visibility');
    const author = url.searchParams.get('author');

    let filtered = [...docStore];
    if (archiveId) filtered = filtered.filter((d) => d.archiveId === archiveId);
    if (parentId) filtered = filtered.filter((d) => d.parentId === parentId);
    if (status) filtered = filtered.filter((d) => d.status === status);
    if (visibility) filtered = filtered.filter((d) => d.visibility === visibility);
    if (author) filtered = filtered.filter((d) => d.author === author);

    return HttpResponse.json({ items: filtered, total: filtered.length });
  }),

  http.get('/api/v1/docs:search', ({ request }) => {
    const url = new URL(request.url);
    const q = url.searchParams.get('q')?.toLowerCase() ?? '';
    const results = docStore.filter(
      (d) =>
        (d.title ?? '').toLowerCase().includes(q) ||
        (d.summary ?? '').toLowerCase().includes(q) ||
        (d.tags ?? []).some((t) => t.toLowerCase().includes(q)),
    );
    return HttpResponse.json({ items: results });
  }),

  http.get('/api/v1/docs:tree', ({ request }) => {
    const url = new URL(request.url);
    const archiveId = url.searchParams.get('archiveId');
    const parentId = url.searchParams.get('parentId');
    let filtered = [...docStore];
    if (archiveId) filtered = filtered.filter((d) => d.archiveId === archiveId);
    const nodes = buildTree(filtered, parentId ?? undefined);
    return HttpResponse.json({ nodes });
  }),

  http.get('/api/v1/docs/:id', ({ params }) => {
    const doc = docStore.find((d) => d.id === params.id);
    if (!doc) return notFound('Doc not found');
    return HttpResponse.json(doc);
  }),

  http.post('/api/v1/docs', async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    const now = new Date().toISOString();
    const doc: Doc = {
      id: `doc-${Date.now()}`,
      archiveId: str(body.archiveId),
      parentId: strOrUndef(body.parentId),
      title: str(body.title),
      content: str(body.content),
      summary: strOrUndef(body.summary),
      status: strOrUndef(body.status),
      priority: strOrUndef(body.priority),
      assignee: strOrUndef(body.assignee),
      tags: strArr(body.tags),
      author: 'forge-operator',
      visibility: str(body.visibility, 'public'),
      sharedWith:
        Array.isArray(body.sharedWith) && body.sharedWith.length > 0
          ? (body.sharedWith as string[])
          : undefined,
      createdAt: now,
      updatedAt: now,
    };
    docStore.push(doc);
    return HttpResponse.json(doc);
  }),

  http.put('/api/v1/docs/:id', async ({ params, request }) => {
    const idx = docStore.findIndex((d) => d.id === params.id);
    if (idx === -1) return notFound('Doc not found');
    const body = (await request.json()) as Record<string, unknown>;
    const updated: Doc = {
      ...docStore[idx],
      ...(body.title != null ? { title: str(body.title) } : {}),
      ...(body.content != null ? { content: str(body.content) } : {}),
      ...(body.summary != null ? { summary: strOrUndef(body.summary) } : {}),
      ...(body.status !== undefined ? { status: strOrUndef(body.status) } : {}),
      ...(body.priority !== undefined ? { priority: strOrUndef(body.priority) } : {}),
      ...(body.assignee !== undefined ? { assignee: strOrUndef(body.assignee) } : {}),
      ...(body.tags != null ? { tags: strArr(body.tags) } : {}),
      ...(body.visibility != null ? { visibility: str(body.visibility) } : {}),
      ...(body.sharedWith != null
        ? {
            sharedWith:
              Array.isArray(body.sharedWith) && body.sharedWith.length > 0
                ? (body.sharedWith as string[])
                : undefined,
          }
        : {}),
      ...(body.parentId !== undefined ? { parentId: strOrUndef(body.parentId) } : {}),
      updatedAt: new Date().toISOString(),
    };
    docStore[idx] = updated;
    return HttpResponse.json(updated);
  }),

  http.delete('/api/v1/docs/:id', ({ params }) => {
    const idx = docStore.findIndex((d) => d.id === params.id);
    if (idx === -1) return notFound('Doc not found');
    docStore.splice(idx, 1);
    return HttpResponse.json({});
  }),

  http.get('/api/v1/docs/:id/ancestors', ({ params }) => {
    const chain = getAncestors(docStore, String(params.id));
    return HttpResponse.json({ ancestors: chain });
  }),

  // ── Links ─────────────────────────────────────────────────

  http.get('/api/v1/docs/:id/links', ({ params }) => {
    const id = String(params.id);
    const links = linkStore.filter((l) => l.sourceId === id || l.targetId === id);
    return HttpResponse.json({ links });
  }),

  http.post('/api/v1/docs/:id/links', async ({ params, request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    const sourceId = String(params.id);
    const targetId = str(body.targetId);
    const targetDoc = docStore.find((d) => d.id === targetId);
    const sourceDoc = docStore.find((d) => d.id === sourceId);
    const link = {
      id: `link-${Date.now()}`,
      sourceId,
      targetId,
      relationType: str(body.relationType, 'relates_to'),
      sourceTitle: sourceDoc?.title ?? '',
      targetTitle: targetDoc?.title ?? '',
      createdAt: new Date().toISOString(),
    };
    linkStore.push(link);
    return HttpResponse.json(link);
  }),

  http.delete('/api/v1/docs/:id/links/:linkId', ({ params }) => {
    const idx = linkStore.findIndex((l) => l.id === params.linkId);
    if (idx === -1) return notFound('Link not found');
    linkStore.splice(idx, 1);
    return HttpResponse.json({});
  }),
];
