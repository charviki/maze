const BASE = '/api/v1';

export interface Archive {
  id: string;
  name: string;
  description?: string;
  icon?: string;
  author: string;
  createdAt: string;
  updatedAt: string;
}

export interface Attachment {
  id: string;
  key: string;
  name: string;
  contentType: string;
  size: string;
}

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
  attachments?: Attachment[];
  createdAt: string;
  updatedAt: string;
}

export interface MemoryMeta {
  id: string;
  archiveId?: string;
  parentId?: string;
  kind?: string;
  title: string;
  type?: string;
  summary?: string;
  tags?: string[];
  author?: string;
  visibility?: string;
  sharedWith?: string[];
  attachments?: Attachment[];
  createdAt?: string;
  updatedAt?: string;
}

export interface ParsedMemory {
  meta: MemoryMeta;
  summary?: string;
  content?: string;
}

export interface Link {
  id: string;
  sourceId: string;
  targetId: string;
  relationType: string;
  sourceTitle?: string;
  targetTitle?: string;
  createdAt: string;
}

export interface Directive {
  id: string;
  title: string;
  description: string;
  status: string;
  priority: string;
  assignee: string;
  author: string;
  requireDocIds?: string[];
  narrativeId?: string;
  archiveId?: string;
  visibility: string;
  createdAt: string;
  updatedAt: string;
}

export interface MemoryTreeNode {
  memory: Memory;
  children?: MemoryTreeNode[];
}

export interface Stats {
  totalMemories: number;
  totalDirectives: number;
  directivesByStatus: Record<string, number>;
  recentMemories: Memory[];
}

export interface ListResponse<T> {
  items: T[];
  total: number;
}

function qs(params?: Record<string, string | undefined>): string {
  if (!params) return '';
  const entries = Object.entries(params).filter(([, v]) => v != null);
  if (entries.length === 0) return '';
  return '?' + new URLSearchParams(entries as [string, string][]).toString();
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 30_000);
  try {
    const res = await fetch(`${BASE}${path}`, {
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
  async list(): Promise<ListResponse<Archive>> {
    const res = await request<{ archives: Archive[] }>('/archives');
    return { items: res.archives || [], total: res.archives?.length ?? 0 };
  },
  get(id: string): Promise<Archive> {
    return request(`/archives/${id}`);
  },
  create(data: Partial<Pick<Archive, 'name' | 'description' | 'icon'>>): Promise<Archive> {
    return request('/archives', { method: 'POST', body: JSON.stringify(data) });
  },
  update(
    id: string,
    data: Partial<Pick<Archive, 'name' | 'description' | 'icon'>>,
  ): Promise<Archive> {
    return request(`/archives/${id}`, { method: 'PUT', body: JSON.stringify(data) });
  },
  async remove(id: string): Promise<void> {
    await request(`/archives/${id}`, { method: 'DELETE' });
  },
};

export const memories = {
  async list(params?: {
    archiveId?: string;
    parentId?: string;
    kind?: string;
    type?: string;
  }): Promise<ListResponse<Memory>> {
    const res = await request<{ items: Memory[]; total: number }>(`/memories${qs(params)}`);
    return { items: res.items || [], total: res.total ?? 0 };
  },
  get(id: string): Promise<ParsedMemory> {
    return request(`/memories/${id}`);
  },
  create(data: Partial<Memory>): Promise<Memory> {
    return request('/memories', { method: 'POST', body: JSON.stringify(data) });
  },
  update(id: string, data: Partial<Memory>): Promise<Memory> {
    return request(`/memories/${id}`, { method: 'PUT', body: JSON.stringify(data) });
  },
  async remove(id: string): Promise<void> {
    await request(`/memories/${id}`, { method: 'DELETE' });
  },
  async search(query: string): Promise<ListResponse<ParsedMemory>> {
    const res = await request<{ items: ParsedMemory[] }>(
      `/memories:search?q=${encodeURIComponent(query)}`,
    );
    return { items: res.items || [], total: res.items?.length ?? 0 };
  },
  async getTree(params?: {
    archiveId?: string;
    parentId?: string;
  }): Promise<ListResponse<MemoryTreeNode>> {
    const res = await request<{ nodes: MemoryTreeNode[] }>(`/memories:tree${qs(params)}`);
    return { items: res.nodes || [], total: res.nodes?.length ?? 0 };
  },
  async getAncestors(id: string): Promise<{ ancestors: Memory[] }> {
    return request(`/memories/${id}/ancestors`);
  },
};

export const links = {
  async list(memoryId: string): Promise<{ links: Link[] }> {
    return request(`/memories/${memoryId}/links`);
  },
  create(memoryId: string, data: { targetId: string; relationType: string }): Promise<Link> {
    return request(`/memories/${memoryId}/links`, { method: 'POST', body: JSON.stringify(data) });
  },
  async remove(memoryId: string, linkId: string): Promise<void> {
    await request(`/memories/${memoryId}/links/${linkId}`, { method: 'DELETE' });
  },
};

export const directives = {
  async list(params?: {
    status?: string;
    priority?: string;
    archiveId?: string;
  }): Promise<ListResponse<Directive>> {
    const res = await request<{ items: Directive[]; total: number }>(`/directives${qs(params)}`);
    return { items: res.items || [], total: res.total ?? 0 };
  },
  get(id: string): Promise<Directive> {
    return request(`/directives/${id}`);
  },
  create(data: Partial<Directive>): Promise<Directive> {
    return request('/directives', { method: 'POST', body: JSON.stringify(data) });
  },
  update(id: string, data: Partial<Directive>): Promise<Directive> {
    return request(`/directives/${id}`, { method: 'PUT', body: JSON.stringify(data) });
  },
  async remove(id: string): Promise<void> {
    await request(`/directives/${id}`, { method: 'DELETE' });
  },
};

export const stats = {
  async get(): Promise<Stats> {
    const res = await request<{ stats: Stats }>('/stats');
    return res.stats;
  },
};

export interface ChatMessage {
  role: 'user' | 'assistant' | 'system';
  content: string;
}

export interface ToolCallInfo {
  id?: string;
  name: string;
  input?: string;
}

export interface ToolResultInfo {
  name: string;
  result?: string;
}

export interface DocContentInfo {
  title: string;
  content: string;
}

export interface ChatCallbacks {
  onThinking?: (content: string) => void;
  onText?: (text: string) => void;
  onToolUse?: (data: ToolCallInfo) => void;
  onToolResult?: (data: ToolResultInfo) => void;
  onDocContent?: (data: DocContentInfo) => void;
  onDone?: (fullText: string) => void;
  onError?: (error: string) => void;
}

interface SseEvent {
  type: string;
  id?: string;
  name?: string;
  title?: string;
  data?: string;
}

export const oracle = {
  async chat(prompt: string, callbacks: ChatCallbacks): Promise<void> {
    const res = await fetch(`${BASE}/oracle/chat`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ message: prompt }),
    });

    if (!res.ok) {
      callbacks.onError?.(`HTTP ${res.status}`);
      return;
    }

    const reader = res.body?.getReader();
    if (!reader) {
      callbacks.onError?.('No response body');
      return;
    }

    const decoder = new TextDecoder();
    let buffer = '';
    let fullText = '';

    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() ?? '';

        for (const line of lines) {
          if (!line.startsWith('data: ')) continue;
          const json = line.slice(6);
          let event: SseEvent;
          try {
            event = JSON.parse(json) as SseEvent;
          } catch {
            continue;
          }
          switch (event.type) {
            case 'thinking':
              callbacks.onThinking?.(event.data ?? '');
              break;
            case 'text':
              fullText += event.data;
              callbacks.onText?.(event.data ?? '');
              break;
            case 'tool_use':
              callbacks.onToolUse?.({ id: event.id, name: event.name ?? '', input: event.data });
              break;
            case 'tool_result':
              callbacks.onToolResult?.({ name: event.name ?? '', result: event.data });
              break;
            case 'doc_content':
              callbacks.onDocContent?.({ title: event.title ?? '', content: event.data ?? '' });
              break;
            case 'done':
              callbacks.onDone?.(fullText);
              return;
            case 'error':
              callbacks.onError?.(event.data ?? 'Unknown error');
              return;
          }
        }
      }
    } catch (err) {
      callbacks.onError?.(err instanceof Error ? err.message : 'Stream error');
      return;
    }

    callbacks.onDone?.(fullText);
  },
};

export const files = {
  async upload(file: File): Promise<{ key: string }> {
    const form = new FormData();
    form.append('file', file);
    const res = await fetch(`${BASE}/files/upload`, {
      method: 'POST',
      body: form,
    });
    if (!res.ok) {
      throw new Error(`Upload failed: HTTP ${res.status}`);
    }
    return (await res.json()) as { key: string };
  },
  download(fileId: string): string {
    return `${BASE}/files/${fileId}`;
  },
};
