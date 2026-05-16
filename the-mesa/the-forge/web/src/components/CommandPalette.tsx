import { useState, useEffect, useRef, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { Search, FileText, CheckSquare, X, Command } from 'lucide-react';
import { docs, type Doc } from '@/api';
import { highlightMatch } from '@/lib/search';

interface NavItem {
  type: 'nav';
  path: string;
  title: string;
  icon: 'directives' | 'home' | 'new';
}

interface DocItem {
  type: 'memory';
  data: Doc;
}

type SearchItem = NavItem | DocItem;

export default function CommandPalette({ open, onClose }: { open: boolean; onClose: () => void }) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<Doc[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const navigate = useNavigate();
  const inputRef = useRef<HTMLInputElement>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const allItems = useMemo((): SearchItem[] => {
    const items: SearchItem[] = [];
    if (query.toLowerCase().includes('dashboard') || query.toLowerCase().includes('home')) {
      items.push({ type: 'nav', path: '/', title: 'DASHBOARD', icon: 'home' });
    }
    if (query.toLowerCase().includes('new') || query.toLowerCase().includes('create')) {
      items.push({ type: 'nav', path: '/docs', title: 'NEW DOCUMENT', icon: 'new' });
    }
    results.forEach((r) => items.push({ type: 'memory', data: r }));
    return items;
  }, [results, query]);

  useEffect(() => {
    if (!open) return;
    const t = setTimeout(() => {
      setQuery('');
      setResults([]);
      setSelectedIndex(0);
      inputRef.current?.focus();
    }, 0);
    return () => clearTimeout(t);
  }, [open]);

  useEffect(() => {
    if (!query.trim()) {
      return;
    }
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(async () => {
      setLoading(true);
      try {
        const res = await docs.search(query);
        setResults(res.items || []);
      } catch {
        setResults([]);
      } finally {
        setLoading(false);
      }
    }, 200);
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, [query]);

  const handleSelect = (index: number) => {
    const item = allItems[index];
    if (!item) return;
    if (item.type === 'memory') {
      void navigate(`/docs/${item.data.archiveId}/${item.data.id}`);
    } else {
      void navigate(item.path);
    }
    onClose();
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelectedIndex((prev) => Math.min(prev + 1, allItems.length - 1));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelectedIndex((prev) => Math.max(prev - 1, 0));
    } else if (e.key === 'Enter') {
      e.preventDefault();
      handleSelect(selectedIndex);
    } else if (e.key === 'Escape') {
      onClose();
    }
  };

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 bg-foreground/30 backdrop-blur-sm flex items-start justify-center pt-[15vh] z-50"
      onClick={onClose}
    >
      <div
        className="bg-card border border-border rounded-lg shadow-2xl w-[480px] max-h-[400px] flex flex-col"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center gap-3 px-4 py-3 border-b border-border">
          <Search size={16} className="text-muted-foreground shrink-0" />
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              setSelectedIndex(0);
            }}
            onKeyDown={handleKeyDown}
            placeholder="Search documents, navigate..."
            className="flex-1 bg-transparent text-foreground placeholder:text-muted-foreground font-mono text-sm focus:outline-none"
          />
          <button
            onClick={onClose}
            className="text-muted-foreground hover:text-foreground transition"
          >
            <X size={14} />
          </button>
        </div>
        <div className="flex-1 overflow-y-auto">
          {loading && (
            <div className="px-4 py-6 text-center text-muted-foreground font-mono text-sm">
              Searching...
            </div>
          )}
          {!loading && allItems.length === 0 && query.trim() && (
            <div className="px-4 py-6 text-center text-muted-foreground font-mono text-sm">
              No results
            </div>
          )}
          {!loading && allItems.length === 0 && !query.trim() && (
            <div className="px-4 py-6 text-center text-muted-foreground font-mono text-sm">
              <Command size={14} className="inline mr-2" />
              Start typing to search
            </div>
          )}
          {!loading &&
            allItems.map((item, i) => {
              const isSelected = i === selectedIndex;
              if (item.type === 'nav') {
                return (
                  <button
                    key={`nav-${i}`}
                    onClick={() => handleSelect(i)}
                    className={`w-full flex items-center gap-3 px-4 py-2.5 text-sm font-mono transition-colors ${
                      isSelected
                        ? 'bg-primary/10 text-primary'
                        : 'text-foreground hover:bg-border/30'
                    }`}
                  >
                    {item.icon === 'directives' ? (
                      <CheckSquare size={14} />
                    ) : (
                      <FileText size={14} />
                    )}
                    {item.title}
                  </button>
                );
              }
              const docItem = item.data;
              return (
                <button
                  key={docItem.id}
                  onClick={() => handleSelect(i)}
                  className={`w-full flex items-center gap-3 px-4 py-2.5 text-sm font-mono transition-colors ${
                    isSelected ? 'bg-primary/10 text-primary' : 'text-foreground hover:bg-border/30'
                  }`}
                >
                  <FileText size={14} className="shrink-0 text-muted-foreground" />
                  <span className="flex-1 text-left truncate">
                    {highlightMatch(docItem.title || '', query)}
                  </span>
                </button>
              );
            })}
        </div>
        <div className="px-4 py-2 border-t border-border flex items-center gap-4 text-[10px] font-mono text-muted-foreground">
          <span>↑↓ navigate</span>
          <span>↵ select</span>
          <span>ESC close</span>
        </div>
      </div>
    </div>
  );
}
