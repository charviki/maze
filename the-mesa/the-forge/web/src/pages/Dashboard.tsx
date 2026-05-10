import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { FileText, CheckSquare, Clock, ArrowRight, Plus } from 'lucide-react';
import { stats, type Stats } from '@/api';
import { Skeleton } from '@maze/fabrication';
import ChatPanel from '@/components/ChatPanel';
import { formatRelativeTime } from '@/lib/time';
import { getTypeColor } from '@/lib/constants';

export default function Dashboard() {
  const [data, setData] = useState<Stats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    stats
      .get()
      .then((res) => {
        if (!cancelled) setData(res);
      })
      .catch(() => {
        if (!cancelled) setData(null);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <Skeleton className="h-8 w-40" />
        </div>
        <div className="grid grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-24" />
          ))}
        </div>
        <div className="grid grid-cols-2 gap-4">
          <Skeleton className="h-64" />
          <Skeleton className="h-64" />
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-foreground font-mono font-bold text-xl">DASHBOARD</h1>
        <div className="flex items-center gap-2">
          <Link
            to="/knowledge/new"
            className="flex items-center gap-1 px-3 py-1.5 bg-primary text-primary-foreground font-mono text-xs rounded hover:bg-primary/90 transition"
          >
            <Plus size={12} />
            NEW DOC
          </Link>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-4">
        <div className="bg-card border border-border rounded p-4">
          <div className="flex items-center gap-2 text-muted-foreground font-mono text-xs mb-2">
            <FileText size={12} />
            TOTAL MEMORIES
          </div>
          <p className="text-foreground font-mono text-2xl font-bold">{data?.totalMemories || 0}</p>
        </div>
        <div className="bg-card border border-border rounded p-4">
          <div className="flex items-center gap-2 text-muted-foreground font-mono text-xs mb-2">
            <CheckSquare size={12} />
            TOTAL DIRECTIVES
          </div>
          <p className="text-foreground font-mono text-2xl font-bold">
            {data?.totalDirectives || 0}
          </p>
        </div>
        <div className="bg-card border border-border rounded p-4">
          <div className="flex items-center gap-2 text-muted-foreground font-mono text-xs mb-2">
            <Clock size={12} />
            DIRECTIVE STATUS
          </div>
          <div className="flex flex-wrap gap-2 mt-1">
            {data?.directivesByStatus &&
              Object.entries(data.directivesByStatus).map(([status, count]) => (
                <span
                  key={status}
                  className="px-2 py-0.5 bg-secondary text-foreground font-mono text-xs rounded"
                >
                  {status}: {count}
                </span>
              ))}
            {(!data?.directivesByStatus || Object.keys(data.directivesByStatus).length === 0) && (
              <span className="text-muted-foreground font-mono text-xs">No directives</span>
            )}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="bg-card border border-border rounded">
          <div className="flex items-center justify-between px-4 py-3 border-b border-border">
            <h2 className="text-foreground font-mono text-sm">RECENT MEMORIES</h2>
            <Link
              to="/knowledge"
              className="text-primary font-mono text-xs hover:underline flex items-center gap-1"
            >
              VIEW ALL <ArrowRight size={10} />
            </Link>
          </div>
          <div className="divide-y divide-border">
            {data?.recentMemories && data.recentMemories.length > 0 ? (
              data.recentMemories.slice(0, 5).map((mem) => {
                const colors = getTypeColor(mem.type);
                return (
                  <Link
                    key={mem.id}
                    to={`/knowledge/${mem.id}`}
                    className="flex items-center gap-3 px-4 py-3 hover:bg-card/80 transition group"
                  >
                    <FileText size={14} className="text-muted-foreground shrink-0" />
                    <div className="flex-1 min-w-0">
                      <p className="text-foreground font-mono text-sm truncate group-hover:text-primary transition">
                        {mem.title}
                      </p>
                      <p className="text-muted-foreground font-mono text-[10px]">
                        {formatRelativeTime(mem.createdAt)}
                      </p>
                    </div>
                    <span
                      className={`px-1.5 py-0.5 rounded text-[10px] font-mono ${colors.bg} ${colors.text} ${colors.border} border`}
                    >
                      {mem.type}
                    </span>
                  </Link>
                );
              })
            ) : (
              <div className="px-4 py-6 text-center text-muted-foreground font-mono text-sm">
                No recent memories
              </div>
            )}
          </div>
        </div>

        <div className="bg-card border border-border rounded h-[400px]">
          <ChatPanel />
        </div>
      </div>
    </div>
  );
}
