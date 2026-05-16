import { Link } from 'react-router-dom';
import { FileText, CheckSquare, Database } from 'lucide-react';
import { useState, useEffect } from 'react';
import { archives, docs, type Archive, type Doc } from '@/api';
import { useToast } from '@maze/fabrication';

function WestworldCard({
  icon,
  title,
  count,
  description,
  to,
}: {
  icon: React.ReactNode;
  title: string;
  count: number;
  description: string;
  to: string;
}) {
  return (
    <Link
      to={to}
      className="bg-card border border-border rounded p-5 hover:border-primary/30 transition group"
    >
      <div className="flex items-center justify-between mb-3">
        <div className="text-primary group-hover:text-primary/80 transition">{icon}</div>
        <span className="text-foreground font-mono text-3xl font-bold">{count}</span>
      </div>
      <h3 className="text-foreground font-mono text-sm mb-1">{title}</h3>
      <p className="text-muted-foreground font-mono text-[10px]">{description}</p>
    </Link>
  );
}

export default function Dashboard() {
  const [archiveList, setArchiveList] = useState<Archive[]>([]);
  const [allDocs, setAllDocs] = useState<Doc[]>([]);
  const { showToast } = useToast();

  useEffect(() => {
    let cancelled = false;
    void (async () => {
      try {
        const aRes = await archives.list();
        if (!cancelled) setArchiveList(aRes.items || []);
      } catch {
        if (!cancelled) showToast('error', 'Failed to load archives');
      }
      try {
        const dRes = await docs.list({});
        if (!cancelled) setAllDocs(dRes.items || []);
      } catch {
        if (!cancelled) showToast('error', 'Failed to load documents');
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const activeDocs = allDocs.filter((d) => d.status === 'active');
  const pendingDocs = allDocs.filter((d) => d.status === 'pending');
  const doneDocs = allDocs.filter((d) => d.status === 'done');

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-foreground font-mono font-bold text-2xl">THE FORGE</h1>
        <p className="text-muted-foreground font-mono text-xs mt-1">
          KNOWLEDGE BASE // HOST MEMORY MANAGEMENT
        </p>
      </div>

      {/* Stats cards */}
      <div className="grid grid-cols-4 gap-3">
        <WestworldCard
          icon={<Database size={18} />}
          title="ARCHIVES"
          count={archiveList.length}
          description="Project workspaces"
          to={archiveList.length > 0 ? `/docs/${archiveList[0].id}` : '/docs'}
        />
        <WestworldCard
          icon={<CheckSquare size={18} />}
          title="ACTIVE"
          count={activeDocs.length}
          description="Documents in progress"
          to="/docs"
        />
        <WestworldCard
          icon={<FileText size={18} />}
          title="PENDING"
          count={pendingDocs.length}
          description="Awaiting action"
          to="/docs"
        />
        <WestworldCard
          icon={<CheckSquare size={18} />}
          title="DONE"
          count={doneDocs.length}
          description="Completed items"
          to="/docs"
        />
      </div>

      {/* Quick actions */}
      <div className="grid grid-cols-2 gap-3">
        {archiveList.length > 0 ? (
          <Link
            to={`/docs/${archiveList[0].id}/new`}
            className="bg-card border border-border rounded p-4 hover:border-primary/30 transition flex items-center gap-3 group"
          >
            <div className="w-8 h-8 rounded bg-primary/10 flex items-center justify-center">
              <FileText size={16} className="text-primary" />
            </div>
            <div>
              <p className="text-foreground font-mono text-sm group-hover:text-primary transition">
                NEW DOCUMENT
              </p>
              <p className="text-muted-foreground font-mono text-[10px]">
                Create a knowledge document or task
              </p>
            </div>
          </Link>
        ) : (
          <div className="bg-card border border-border rounded p-4 flex items-center gap-3">
            <div className="w-8 h-8 rounded bg-amber-500/10 flex items-center justify-center">
              <Database size={16} className="text-amber-400" />
            </div>
            <div>
              <p className="text-amber-400 font-mono text-sm">CREATE AN ARCHIVE FIRST</p>
              <p className="text-muted-foreground font-mono text-[10px]">
                Use the + button in the sidebar to create a workspace
              </p>
            </div>
          </div>
        )}
        <Link
          to={archiveList.length > 0 ? `/docs/${archiveList[0].id}` : '/docs'}
          className="bg-card border border-border rounded p-4 hover:border-primary/30 transition flex items-center gap-3 group"
        >
          <div className="w-8 h-8 rounded bg-primary/10 flex items-center justify-center">
            <Database size={16} className="text-primary" />
          </div>
          <div>
            <p className="text-foreground font-mono text-sm group-hover:text-primary transition">
              BROWSE ARCHIVES
            </p>
            <p className="text-muted-foreground font-mono text-[10px]">
              Explore your knowledge base
            </p>
          </div>
        </Link>
      </div>
    </div>
  );
}
