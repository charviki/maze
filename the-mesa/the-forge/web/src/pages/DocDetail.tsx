import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Edit, Trash2, Link2, Clock, User, Tag, Plus } from 'lucide-react';
import { docs, links, type Doc, type DocLink } from '@/api';
import { useToast } from '@maze/fabrication';
import { Skeleton } from '@maze/fabrication';
import { useModal } from '@/hooks/useModal';
import ModalPortal from '@/components/ModalContent';
import ReactMarkdown from 'react-markdown';
import rehypeSanitize from 'rehype-sanitize';
import KnowledgeGraph from '@/components/KnowledgeGraph';
import EmptyState from '@/components/EmptyState';
import { formatRelativeTime } from '@/lib/time';
import { getStatusColor, priorityColors } from '@/lib/constants';
import { useDocTreeRefresh } from '@/contexts/DocTreeContext';

export default function DocDetail() {
  const { archiveId, docId } = useParams<{ archiveId: string; docId?: string }>();
  const navigate = useNavigate();
  const { showToast } = useToast();
  const { refreshTree } = useDocTreeRefresh();
  const { config: modalConfig, confirm, close } = useModal();
  const [doc, setDoc] = useState<Doc | null>(null);
  const [docLinks, setDocLinks] = useState<DocLink[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!docId) return;
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setLoading(true);
    let cancelled = false;
    void Promise.all([docs.get(docId), links.list(docId)])
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
  }, [docId]);

  const handleDelete = () => {
    if (!docId) return;

    void confirm('Delete this document? This cannot be undone.', 'DELETE', 'danger').then(
      async (ok) => {
        if (!ok) return;
        await docs.remove(docId);
        refreshTree();
        showToast('success', 'Document deleted');
        void navigate(`/docs/${archiveId}`);
      },
    );
  };

  // No document selected
  if (!docId) {
    return (
      <EmptyState
        title="Select a document"
        description="Choose a document from the tree to view its content"
      />
    );
  }

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
          label: 'BACK',
          onClick: () => {
            void navigate(`/docs/${archiveId}`);
          },
        }}
      />
    );
  }

  const statusColors = doc.status ? getStatusColor(doc.status) : null;
  const pc = doc.priority ? priorityColors[doc.priority] || priorityColors.medium : null;

  return (
    <div className="space-y-6">
      <ModalPortal config={modalConfig} onClose={close} />

      {/* Actions */}
      <div className="flex items-center gap-2">
        <button
          onClick={() => navigate(`/docs/${archiveId}/${docId}/edit`)}
          className="flex items-center gap-1 px-3 py-1.5 bg-card border border-border text-foreground font-mono text-xs rounded hover:bg-card/80 transition"
        >
          <Edit size={12} />
          EDIT
        </button>
        <button
          onClick={() => navigate(`/docs/${archiveId}/${docId}/new`)}
          className="flex items-center gap-1 px-3 py-1.5 bg-card border border-border text-foreground font-mono text-xs rounded hover:bg-card/80 transition"
        >
          <Plus size={12} />
          SUB-DOC
        </button>
        <button
          onClick={handleDelete}
          className="flex items-center gap-1 px-3 py-1.5 bg-destructive text-primary-foreground font-mono text-xs rounded hover:bg-destructive/90 transition"
        >
          <Trash2 size={12} />
          DELETE
        </button>
      </div>

      {/* Status & Priority badges */}
      <div className="flex items-center gap-2">
        {doc.status && statusColors && (
          <span
            className={`px-2 py-0.5 rounded text-xs font-mono ${statusColors.bg} ${statusColors.text} border ${statusColors.border}`}
          >
            {doc.status}
          </span>
        )}
        {doc.priority && pc && (
          <span className={`px-2 py-0.5 rounded text-xs font-mono ${pc.bg} ${pc.text}`}>
            {doc.priority}
          </span>
        )}
        {doc.assignee && (
          <span className="text-muted-foreground font-mono text-xs flex items-center gap-1">
            <User size={10} />
            {doc.assignee}
          </span>
        )}
      </div>

      {/* Title */}
      <h1 className="text-foreground font-mono font-bold text-2xl">{doc.title}</h1>

      {/* Meta */}
      <div className="flex items-center gap-4 text-muted-foreground font-mono text-xs">
        <span className="flex items-center gap-1">
          <User size={10} />
          {doc.author}
        </span>
        {doc.tags?.map((tag) => (
          <span key={tag} className="flex items-center gap-1">
            <Tag size={9} />
            {tag}
          </span>
        ))}
        {doc.updatedAt && (
          <span className="flex items-center gap-1">
            <Clock size={10} />
            {formatRelativeTime(doc.updatedAt)}
          </span>
        )}
      </div>

      {/* Summary */}
      {doc.summary && (
        <div className="bg-secondary/50 border border-border rounded p-4">
          <p className="text-muted-foreground font-mono text-[10px] mb-1">SUMMARY</p>
          <p className="text-foreground font-mono text-sm">{doc.summary}</p>
        </div>
      )}

      {/* Content */}
      {doc.content && (
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
            <ReactMarkdown rehypePlugins={[rehypeSanitize]}>{doc.content}</ReactMarkdown>
          </div>
        </div>
      )}

      {/* Links */}
      {docLinks.length > 0 && (
        <div>
          <div className="flex items-center gap-2 mb-3">
            <Link2 size={14} className="text-primary" />
            <h2 className="text-foreground font-mono text-sm">LINKED DOCUMENTS</h2>
          </div>
          <div className="flex items-start gap-6">
            <div className="flex-1 space-y-2">
              {docLinks.map((link) => (
                <button
                  key={link.id}
                  onClick={() => {
                    const targetId = link.sourceId === docId ? link.targetId : link.sourceId;
                    void navigate(`/docs/${archiveId}/${targetId}`);
                  }}
                  className="w-full flex items-center gap-3 px-4 py-2 bg-card border border-border rounded hover:bg-card/80 transition group text-left"
                >
                  <span className="flex-1 text-foreground font-mono text-sm group-hover:text-primary transition">
                    {link.sourceId === docId ? link.targetTitle : link.sourceTitle}
                  </span>
                  <span className="px-1.5 py-0.5 bg-blue-500/15 text-blue-400 text-[10px] font-mono rounded border border-blue-500/30">
                    {link.relationType}
                  </span>
                </button>
              ))}
            </div>
            <KnowledgeGraph
              links={docLinks}
              currentId={docId}
              width={300}
              height={200}
              onNodeClick={(nodeId) => navigate(`/docs/${archiveId}/${nodeId}`)}
            />
          </div>
        </div>
      )}
    </div>
  );
}
