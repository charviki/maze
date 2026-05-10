import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { ArrowLeft, Edit, Trash2, Link2, Clock, User, Tag, Eye, FileText } from 'lucide-react';
import { memories, links, type ParsedMemory, type Link as DocLink } from '@/api';
import { useToast } from '@maze/fabrication';
import { Skeleton } from '@maze/fabrication';
import { useModal } from '@/hooks/useModal';
import ModalPortal from '@/components/ModalContent';
import ReactMarkdown from 'react-markdown';
import KnowledgeGraph from '@/components/KnowledgeGraph';
import EmptyState from '@/components/EmptyState';
import { formatAbsoluteTime } from '@/lib/time';
import { getTypeColor } from '@/lib/constants';

export default function KnowledgeDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { showToast } = useToast();
  const { config: modalConfig, confirm, close } = useModal();
  const [doc, setDoc] = useState<ParsedMemory | null>(null);
  const [docLinks, setDocLinks] = useState<DocLink[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!id) return;
    let cancelled = false;
    void Promise.all([memories.get(id), links.list(id)])
      .then(([docData, linksData]) => {
        if (cancelled) return;
        setDoc(docData);
        setDocLinks(linksData.links || []);
      })
      .catch(() => {
        if (cancelled) return;
        setDoc(null);
        setDocLinks([]);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [id]);

  const handleDelete = () => {
    void confirm('Delete this document? This cannot be undone.', 'DELETE', 'danger').then(
      async (ok) => {
        if (!ok || !id) return;
        await memories.remove(id);
        showToast('success', 'Document deleted');
        void navigate('/knowledge');
      },
    );
  };

  if (loading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-10 w-3/4" />
        <Skeleton className="h-64" />
      </div>
    );
  }

  if (!doc) {
    return (
      <EmptyState
        title="Document not found"
        description="The document may have been deleted or moved"
        action={{
          label: 'BACK TO LIST',
          onClick: () => {
            void navigate('/knowledge');
          },
        }}
      />
    );
  }

  const meta = doc.meta;
  const colors = getTypeColor(meta.type || '');

  return (
    <div className="space-y-6">
      <ModalPortal config={modalConfig} onClose={close} />

      <div className="flex items-center justify-between">
        <Link
          to="/knowledge"
          className="flex items-center gap-2 text-muted-foreground hover:text-foreground font-mono text-sm transition"
        >
          <ArrowLeft size={14} />
          BACK
        </Link>
        <div className="flex items-center gap-2">
          <Link
            to={`/knowledge/${id}/edit`}
            className="flex items-center gap-1 px-3 py-1.5 bg-card border border-border text-foreground font-mono text-xs rounded hover:bg-card/80 transition"
          >
            <Edit size={12} />
            EDIT
          </Link>
          <button
            onClick={handleDelete}
            className="flex items-center gap-1 px-3 py-1.5 bg-destructive text-primary-foreground font-mono text-xs rounded hover:bg-destructive/90 transition"
          >
            <Trash2 size={12} />
            DELETE
          </button>
        </div>
      </div>

      <div>
        <div className="flex items-center gap-3 mb-2">
          <span
            className={`px-2 py-0.5 rounded text-xs font-mono ${colors.bg} ${colors.text} ${colors.border} border`}
          >
            {meta.type}
          </span>
          {meta.tags?.map((tag) => (
            <span
              key={tag}
              className="flex items-center gap-1 px-1.5 py-0.5 bg-secondary text-muted-foreground font-mono text-[10px] rounded"
            >
              <Tag size={8} />
              {tag}
            </span>
          ))}
        </div>
        <h1 className="text-foreground font-mono font-bold text-2xl">{meta.title}</h1>
      </div>

      <div className="flex items-center gap-4 text-muted-foreground font-mono text-xs">
        <span className="flex items-center gap-1">
          <User size={10} />
          {meta.author}
        </span>
        <span className="flex items-center gap-1">
          <Eye size={10} />
          {meta.visibility}
        </span>
        {meta.createdAt && (
          <span className="flex items-center gap-1">
            <Clock size={10} />
            {formatAbsoluteTime(meta.createdAt)}
          </span>
        )}
      </div>

      {doc.summary && (
        <div className="bg-secondary border border-border rounded p-4">
          <p className="text-muted-foreground font-mono text-xs mb-1">SUMMARY</p>
          <p className="text-foreground font-mono text-sm">{doc.summary}</p>
        </div>
      )}

      <div className="bg-card border border-border rounded p-6">
        <div
          className="prose prose-invert prose-sm max-w-none font-mono text-foreground
          prose-headings:text-foreground prose-headings:font-mono
          prose-p:text-foreground prose-p:font-mono
          prose-code:text-primary prose-code:bg-primary/10 prose-code:px-1 prose-code:rounded
          prose-pre:bg-background prose-pre:border prose-pre:border-border
          prose-a:text-primary prose-a:no-underline hover:prose-a:underline
          prose-strong:text-foreground
          prose-blockquote:border-primary prose-blockquote:text-muted-foreground
          prose-li:text-foreground"
        >
          <ReactMarkdown>{doc.content || ''}</ReactMarkdown>
        </div>
      </div>

      {docLinks.length > 0 && (
        <div>
          <div className="flex items-center gap-2 mb-3">
            <Link2 size={14} className="text-primary" />
            <h2 className="text-foreground font-mono text-sm">LINKED DOCUMENTS</h2>
          </div>
          <div className="flex items-start gap-6">
            <div className="flex-1 space-y-2">
              {docLinks.map((link) => (
                <Link
                  key={link.id}
                  to={`/knowledge/${link.sourceId === id ? link.targetId : link.sourceId}`}
                  className="flex items-center gap-3 px-4 py-2 bg-card border border-border rounded hover:bg-card/80 transition group"
                >
                  <FileText size={12} className="text-muted-foreground" />
                  <span className="flex-1 text-foreground font-mono text-sm group-hover:text-primary transition">
                    {link.sourceId === id ? link.targetTitle : link.sourceTitle}
                  </span>
                  <span className="px-1.5 py-0.5 bg-blue-500/15 text-blue-400 text-[10px] font-mono rounded border border-blue-500/30">
                    {link.relationType}
                  </span>
                </Link>
              ))}
            </div>
            <KnowledgeGraph
              links={docLinks}
              currentId={id}
              width={300}
              height={200}
              onNodeClick={(nodeId) => navigate(`/knowledge/${nodeId}`)}
            />
          </div>
        </div>
      )}
    </div>
  );
}
