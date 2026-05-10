import { http, HttpResponse } from 'msw';
import { archives } from './data/archives.ts';
import { memories, type Memory } from './data/memories.ts';
import { directives } from './data/directives.ts';
import { links } from './data/links.ts';
import { stats } from './data/stats.ts';

function ok(data: unknown) {
  return HttpResponse.json({ status: 'ok', data });
}

function notFound(message: string) {
  return HttpResponse.json(
    { status: 'error', error: { code: 'NOT_FOUND', message } },
    { status: 404 },
  );
}

let memoryStore = [...memories];
let archiveStore = [...archives];
let directiveStore = [...directives];
let linkStore = [...links];

export function resetMockData() {
  memoryStore = [...memories];
  archiveStore = [...archives];
  directiveStore = [...directives];
  linkStore = [...links];
}

interface TreeNode {
  memory: Memory;
  children: TreeNode[];
}

function buildTree(items: Memory[], parentId?: string): TreeNode[] {
  return items
    .filter((m) => (parentId ? m.parentId === parentId : !m.parentId))
    .map((m) => ({
      memory: m,
      children: buildTree(items, m.id),
    }));
}

function getAncestors(items: Memory[], id: string): Memory[] {
  const chain: Memory[] = [];
  let current = items.find((m) => m.id === id);
  while (current) {
    chain.unshift(current);
    if (!current.parentId) break;
    current = items.find((m) => m.id === current!.parentId);
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
  http.get('/api/v1/stats', () => {
    return ok({
      ...stats,
      totalMemories: memoryStore.length,
      totalDirectives: directiveStore.length,
    });
  }),

  http.get('/api/v1/archives', () => {
    return ok({ items: archiveStore });
  }),

  http.get('/api/v1/archives/:id', ({ params }) => {
    const archive = archiveStore.find((a) => a.id === params.id);
    if (!archive) return notFound('Archive not found');
    return ok(archive);
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
    return ok(archive);
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
    return ok(updated);
  }),

  http.delete('/api/v1/archives/:id', ({ params }) => {
    const idx = archiveStore.findIndex((a) => a.id === params.id);
    if (idx === -1) return notFound('Archive not found');
    archiveStore.splice(idx, 1);
    return ok(null);
  }),

  http.get('/api/v1/memories', ({ request }) => {
    const url = new URL(request.url);
    const archiveId = url.searchParams.get('archiveId');
    const parentId = url.searchParams.get('parentId');
    const kind = url.searchParams.get('kind');
    const type = url.searchParams.get('type');

    let filtered = [...memoryStore];
    if (archiveId) filtered = filtered.filter((m) => m.archiveId === archiveId);
    if (parentId) filtered = filtered.filter((m) => m.parentId === parentId);
    if (kind) filtered = filtered.filter((m) => m.kind === kind);
    if (type) filtered = filtered.filter((m) => m.type === type);

    return ok({ items: filtered, total: filtered.length });
  }),

  http.get('/api/v1/memories:search', ({ request }) => {
    const url = new URL(request.url);
    const q = url.searchParams.get('q')?.toLowerCase() ?? '';
    const results = memoryStore.filter(
      (m) =>
        m.kind === 'doc' &&
        (m.title.toLowerCase().includes(q) ||
          (m.summary ?? '').toLowerCase().includes(q) ||
          m.tags.some((t) => t.toLowerCase().includes(q))),
    );
    const items = results.map((m) => ({
      meta: {
        id: m.id,
        archiveId: m.archiveId,
        parentId: m.parentId,
        kind: m.kind,
        title: m.title,
        type: m.type,
        summary: m.summary,
        tags: m.tags,
        author: m.author,
        visibility: m.visibility,
        createdAt: m.createdAt,
        updatedAt: m.updatedAt,
      },
      summary: m.summary,
      content: m.content,
    }));
    return ok({ items, total: items.length });
  }),

  http.get('/api/v1/memories/tree', ({ request }) => {
    const url = new URL(request.url);
    const archiveId = url.searchParams.get('archiveId');
    const parentId = url.searchParams.get('parentId');
    let filtered = [...memoryStore];
    if (archiveId) filtered = filtered.filter((m) => m.archiveId === archiveId);
    const tree = buildTree(filtered, parentId ?? undefined);
    return ok({ items: tree });
  }),

  http.get('/api/v1/memories/:id', ({ params }) => {
    const memory = memoryStore.find((m) => m.id === params.id);
    if (!memory) return notFound('Memory not found');
    return ok({
      meta: {
        id: memory.id,
        archiveId: memory.archiveId,
        parentId: memory.parentId,
        kind: memory.kind,
        title: memory.title,
        type: memory.type,
        summary: memory.summary,
        tags: memory.tags,
        author: memory.author,
        visibility: memory.visibility,
        sharedWith: memory.sharedWith,
        attachments: memory.attachments,
        createdAt: memory.createdAt,
        updatedAt: memory.updatedAt,
      },
      summary: memory.summary,
      content: memory.content,
    });
  }),

  http.post('/api/v1/memories', async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    const now = new Date().toISOString();
    const memory: Memory = {
      id: `mem-${Date.now()}`,
      archiveId: str(body.archiveId),
      parentId: strOrUndef(body.parentId),
      kind: str(body.kind, 'doc'),
      title: str(body.title),
      content: str(body.content),
      type: str(body.type, 'shared'),
      summary: strOrUndef(body.summary),
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
    memoryStore.push(memory);
    return ok(memory);
  }),

  http.put('/api/v1/memories/:id', async ({ params, request }) => {
    const idx = memoryStore.findIndex((m) => m.id === params.id);
    if (idx === -1) return notFound('Memory not found');
    const body = (await request.json()) as Record<string, unknown>;
    const updated: Memory = {
      ...memoryStore[idx],
      ...(body.title != null ? { title: str(body.title) } : {}),
      ...(body.content != null ? { content: str(body.content) } : {}),
      ...(body.type != null ? { type: str(body.type) } : {}),
      ...(body.summary != null ? { summary: strOrUndef(body.summary) } : {}),
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
      ...(body.parentId != null ? { parentId: strOrUndef(body.parentId) } : {}),
      updatedAt: new Date().toISOString(),
    };
    memoryStore[idx] = updated;
    return ok(updated);
  }),

  http.delete('/api/v1/memories/:id', ({ params }) => {
    const idx = memoryStore.findIndex((m) => m.id === params.id);
    if (idx === -1) return notFound('Memory not found');
    memoryStore.splice(idx, 1);
    return ok(null);
  }),

  http.get('/api/v1/memories/:id/ancestors', ({ params }) => {
    const chain = getAncestors(memoryStore, String(params.id));
    return ok({ ancestors: chain });
  }),

  http.get('/api/v1/memories/:id/links', ({ params }) => {
    const id = String(params.id);
    const memLinks = linkStore.filter((l) => l.sourceId === id || l.targetId === id);
    return ok({ links: memLinks });
  }),

  http.post('/api/v1/memories/:id/links', async ({ params, request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    const sourceId = String(params.id);
    const targetId = str(body.targetId);
    const targetMemory = memoryStore.find((m) => m.id === targetId);
    const sourceMemory = memoryStore.find((m) => m.id === sourceId);
    const link = {
      id: `link-${Date.now()}`,
      sourceId,
      targetId,
      relationType: str(body.relationType, 'relates_to'),
      sourceTitle: sourceMemory?.title ?? '',
      targetTitle: targetMemory?.title ?? '',
      createdAt: new Date().toISOString(),
    };
    linkStore.push(link);
    return ok(link);
  }),

  http.delete('/api/v1/memories/:id/links/:linkId', ({ params }) => {
    const idx = linkStore.findIndex((l) => l.id === params.linkId);
    if (idx === -1) return notFound('Link not found');
    linkStore.splice(idx, 1);
    return ok(null);
  }),

  http.get('/api/v1/directives', ({ request }) => {
    const url = new URL(request.url);
    const status = url.searchParams.get('status');
    const priority = url.searchParams.get('priority');
    const archiveId = url.searchParams.get('archiveId');

    let filtered = [...directiveStore];
    if (status) filtered = filtered.filter((d) => d.status === status);
    if (priority) filtered = filtered.filter((d) => d.priority === priority);
    if (archiveId) filtered = filtered.filter((d) => d.archiveId === archiveId);

    return ok({ items: filtered, total: filtered.length });
  }),

  http.get('/api/v1/directives/:id', ({ params }) => {
    const directive = directiveStore.find((d) => d.id === params.id);
    if (!directive) return notFound('Directive not found');
    return ok(directive);
  }),

  http.post('/api/v1/directives', async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    const now = new Date().toISOString();
    const directive = {
      id: `dir-${Date.now()}`,
      title: str(body.title),
      description: str(body.description),
      status: str(body.status, 'pending'),
      priority: str(body.priority, 'normal'),
      assignee: str(body.assignee, 'forge-operator'),
      author: 'forge-operator',
      requireDocIds: strArr(body.requireDocIds),
      narrativeId: strOrUndef(body.narrativeId),
      archiveId: str(body.archiveId),
      visibility: str(body.visibility, 'public'),
      createdAt: now,
      updatedAt: now,
    };
    directiveStore.push(directive);
    return ok(directive);
  }),

  http.put('/api/v1/directives/:id', async ({ params, request }) => {
    const idx = directiveStore.findIndex((d) => d.id === params.id);
    if (idx === -1) return notFound('Directive not found');
    const body = (await request.json()) as Record<string, unknown>;
    const updated = {
      ...directiveStore[idx],
      ...(body.title != null ? { title: str(body.title) } : {}),
      ...(body.description != null ? { description: str(body.description) } : {}),
      ...(body.status != null ? { status: str(body.status) } : {}),
      ...(body.priority != null ? { priority: str(body.priority) } : {}),
      ...(body.assignee != null ? { assignee: str(body.assignee) } : {}),
      ...(body.requireDocIds != null ? { requireDocIds: strArr(body.requireDocIds) } : {}),
      ...(body.narrativeId != null ? { narrativeId: strOrUndef(body.narrativeId) } : {}),
      ...(body.visibility != null ? { visibility: str(body.visibility) } : {}),
      updatedAt: new Date().toISOString(),
    };
    directiveStore[idx] = updated;
    return ok(updated);
  }),

  http.delete('/api/v1/directives/:id', ({ params }) => {
    const idx = directiveStore.findIndex((d) => d.id === params.id);
    if (idx === -1) return notFound('Directive not found');
    directiveStore.splice(idx, 1);
    return ok(null);
  }),
];
