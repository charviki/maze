import { useState, useEffect, useMemo, useCallback } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { FileText, Folder, Search, Plus, Filter, ChevronRight, ChevronDown } from 'lucide-react';
import {
  memories,
  type Memory,
  type MemoryMeta,
  type MemoryTreeNode,
  type ParsedMemory,
} from '@/api';
import { Skeleton } from '@maze/fabrication';
import { useListNavigation } from '@/hooks/useListNavigation';
import { formatRelativeTime } from '@/lib/time';
import { getTypeColor, folderColorMap } from '@/lib/constants';
import { highlightMatch } from '@/lib/search';
import EmptyState from '@/components/EmptyState';

function memoryToMeta(m: Memory): MemoryMeta {
  return {
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
  };
}

export default function KnowledgeList() {
  const [searchParams] = useSearchParams();
  const archiveId = searchParams.get('archiveId') || '';
  const parentId = searchParams.get('parentId') || '';

  const [memList, setMemList] = useState<Memory[]>([]);
  const [searchResults, setSearchResults] = useState<ParsedMemory[]>([]);
  const [tree, setTree] = useState<MemoryTreeNode[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [expandedFolders, setExpandedFolders] = useState<Set<string>>(new Set());
  const [treeMode, setTreeMode] = useState(false);

  const isSearchMode = searchQuery.trim().length > 0;

  const loadTree = useCallback(async () => {
    if (!archiveId) return;
    try {
      const params: { archiveId: string; parentId?: string } = { archiveId };
      if (parentId) params.parentId = parentId;
      const res = await memories.getTree(params);
      setTree(res.items || []);
    } catch {
      setTree([]);
    }
  }, [archiveId, parentId]);

  useEffect(() => {
    const params: Record<string, string | undefined> = {};
    if (archiveId) params.archiveId = archiveId;
    if (parentId) params.parentId = parentId;
    if (typeFilter) params.type = typeFilter;
    memories
      .list(params)
      .then((res) => setMemList(res.items || []))
      .catch(() => setMemList([]))
      .finally(() => setLoading(false));
  }, [archiveId, parentId, typeFilter]);

  useEffect(() => {
    if (!archiveId) return;
    memories
      .getTree({ archiveId, ...(parentId ? { parentId } : {}) })
      .then((res) => setTree(res.items || []))
      .catch(() => setTree([]));
  }, [archiveId, parentId]);

  useEffect(() => {
    if (!searchQuery.trim()) return;
    const timer = setTimeout(async () => {
      try {
        const res = await memories.search(searchQuery);
        setSearchResults(res.items || []);
      } catch {
        setSearchResults([]);
      }
    }, 200);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  const flatMetas = useMemo(() => {
    if (isSearchMode) {
      return searchResults.map((r) => r.meta);
    }
    if (!treeMode || !tree.length) {
      return memList.map(memoryToMeta);
    }
    const metas: MemoryMeta[] = [];
    const walk = (nodes: MemoryTreeNode[]) => {
      for (const node of nodes) {
        metas.push(memoryToMeta(node.memory));
        if (node.memory.kind === 'folder' && expandedFolders.has(node.memory.id) && node.children) {
          walk(node.children);
        }
      }
    };
    walk(tree);
    return metas;
  }, [isSearchMode, searchResults, treeMode, tree, memList, expandedFolders]);

  const navItems = useMemo(() => flatMetas.map((m) => ({ id: m.id })), [flatMetas]);
  const { focusedIndex, handleKeyDown } = useListNavigation({
    items: navItems,
    basePath: '/knowledge',
  });

  const toggleFolder = (id: string) => {
    setExpandedFolders((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const breadcrumbs = useMemo(() => {
    const parts: { label: string; path: string }[] = [{ label: 'MEMORIES', path: '/knowledge' }];
    if (archiveId) parts.push({ label: archiveId, path: `/knowledge?archiveId=${archiveId}` });
    if (parentId)
      parts.push({
        label: parentId,
        path: `/knowledge?archiveId=${archiveId}&parentId=${parentId}`,
      });
    return parts;
  }, [archiveId, parentId]);

  if (loading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-60" />
        <div className="flex gap-3">
          <Skeleton className="h-9 flex-1" />
          <Skeleton className="h-9 w-24" />
        </div>
        {[1, 2, 3, 4, 5].map((i) => (
          <Skeleton key={i} className="h-16" />
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-4" onKeyDown={handleKeyDown} tabIndex={0}>
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-sm font-mono">
          {breadcrumbs.map((bc, i) => (
            <span key={i} className="flex items-center gap-2">
              {i > 0 && <ChevronRight size={12} className="text-muted-foreground" />}
              <Link
                to={bc.path}
                className={
                  i === breadcrumbs.length - 1
                    ? 'text-foreground'
                    : 'text-muted-foreground hover:text-foreground transition'
                }
              >
                {bc.label}
              </Link>
            </span>
          ))}
        </div>
        <Link
          to={`/knowledge/new${archiveId ? `?archiveId=${archiveId}` : ''}`}
          className="flex items-center gap-1 px-3 py-1.5 bg-primary text-primary-foreground font-mono text-xs rounded hover:bg-primary/90 transition"
        >
          <Plus size={12} />
          NEW
        </Link>
      </div>

      <div className="flex items-center gap-3">
        <div className="flex-1 flex items-center gap-2 bg-card border border-border rounded px-3 py-2">
          <Search size={14} className="text-muted-foreground shrink-0" />
          <input
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search documents..."
            className="flex-1 bg-transparent text-foreground placeholder:text-muted-foreground font-mono text-sm focus:outline-none"
          />
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => {
              setTreeMode(!treeMode);
              void loadTree();
            }}
            className={`px-3 py-2 font-mono text-xs rounded border transition ${
              treeMode
                ? 'bg-primary/10 text-primary border-primary/30'
                : 'bg-card text-muted-foreground border-border hover:text-foreground'
            }`}
          >
            <Filter size={12} className="inline mr-1" />
            TREE
          </button>
          <select
            value={typeFilter}
            onChange={(e) => setTypeFilter(e.target.value)}
            className="bg-card border border-border rounded px-2 py-2 text-foreground font-mono text-xs focus:outline-none focus:border-primary"
          >
            <option value="">ALL TYPES</option>
            <option value="requirement">REQUIREMENT</option>
            <option value="shared">SHARED</option>
            <option value="ops">OPS</option>
            <option value="narrative">NARRATIVE</option>
            <option value="memory">MEMORY</option>
          </select>
        </div>
      </div>

      {flatMetas.length === 0 ? (
        <EmptyState
          title="No documents found"
          description={searchQuery ? 'Try a different search term' : 'Create your first document'}
          action={!searchQuery ? { label: 'CREATE DOCUMENT', onClick: () => {} } : undefined}
        />
      ) : (
        <div className="space-y-1">
          {flatMetas.map((meta, i) => {
            const isFolder = meta.kind === 'folder';
            const isFocused = i === focusedIndex;
            const colors = isFolder ? folderColorMap : getTypeColor(meta.type || '');

            return (
              <Link
                key={meta.id}
                to={
                  isFolder
                    ? `/knowledge?archiveId=${archiveId}&parentId=${meta.id}`
                    : `/knowledge/${meta.id}`
                }
                className={`flex items-center gap-3 px-4 py-3 rounded transition group ${
                  isFocused
                    ? 'bg-primary/10 border border-primary/30'
                    : 'bg-card border border-border hover:bg-card/80'
                }`}
              >
                {isFolder ? (
                  <button
                    onClick={(e) => {
                      e.preventDefault();
                      toggleFolder(meta.id);
                    }}
                    className="text-muted-foreground hover:text-primary transition"
                  >
                    {expandedFolders.has(meta.id) ? (
                      <ChevronDown size={14} />
                    ) : (
                      <ChevronRight size={14} />
                    )}
                  </button>
                ) : null}
                {isFolder ? (
                  <Folder size={14} className="text-primary shrink-0" />
                ) : (
                  <FileText size={14} className="text-muted-foreground shrink-0" />
                )}
                <div className="flex-1 min-w-0">
                  <p
                    className={`font-mono text-sm truncate ${isFocused ? 'text-primary' : 'text-foreground group-hover:text-primary'} transition`}
                  >
                    {searchQuery ? highlightMatch(meta.title, searchQuery) : meta.title}
                  </p>
                  {meta.tags && meta.tags.length > 0 && (
                    <div className="flex gap-1 mt-1">
                      {meta.tags.slice(0, 3).map((tag) => (
                        <span
                          key={tag}
                          className="px-1.5 py-0.5 bg-secondary text-muted-foreground font-mono text-[10px] rounded"
                        >
                          {tag}
                        </span>
                      ))}
                    </div>
                  )}
                </div>
                <span
                  className={`px-1.5 py-0.5 rounded text-[10px] font-mono ${colors.bg} ${colors.text} ${colors.border} border`}
                >
                  {isFolder ? 'FOLDER' : meta.type}
                </span>
                {meta.updatedAt && (
                  <span className="text-muted-foreground font-mono text-[10px] shrink-0">
                    {formatRelativeTime(meta.updatedAt)}
                  </span>
                )}
              </Link>
            );
          })}
        </div>
      )}
    </div>
  );
}
