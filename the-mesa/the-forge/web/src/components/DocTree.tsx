import { useState, useEffect, useCallback } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  ChevronRight,
  ChevronDown,
  FileText,
  Plus,
  Circle,
  CheckCircle2,
  XCircle,
  Clock,
  Loader2,
} from 'lucide-react';
import { docs, type DocTreeNode } from '@/api';
import { getStatusColor } from '@/lib/constants';

interface DocTreeProps {
  archiveId: string;
  width: number;
  onResize: (delta: number) => void;
}

function StatusIcon({ status }: { status: string | undefined }) {
  if (!status) return <FileText size={13} className="shrink-0 text-muted-foreground" />;
  switch (status) {
    case 'active':
      return <Loader2 size={13} className="shrink-0 text-amber-400 animate-spin" />;
    case 'done':
      return <CheckCircle2 size={13} className="shrink-0 text-emerald-400" />;
    case 'failed':
      return <XCircle size={13} className="shrink-0 text-red-400" />;
    case 'pending':
      return <Clock size={13} className="shrink-0 text-slate-400" />;
    default:
      return <Circle size={13} className="shrink-0 text-muted-foreground" />;
  }
}

function TreeNode({
  node,
  archiveId,
  depth,
  selectedId,
  onSelect,
  onAddChild,
}: {
  node: DocTreeNode;
  archiveId: string;
  depth: number;
  selectedId: string | undefined;
  onSelect: (id: string) => void;
  onAddChild: (parentId: string) => void;
}) {
  const [expanded, setExpanded] = useState(true);
  const [hovering, setHovering] = useState(false);
  const d = node.doc;
  if (!d) return null;
  const hasChildren = node.children && node.children.length > 0;
  const isSelected = d.id === selectedId;
  const statusColors = d.status ? getStatusColor(d.status) : null;

  return (
    <div onMouseEnter={() => setHovering(true)} onMouseLeave={() => setHovering(false)}>
      <div
        className={`flex items-center gap-1.5 py-1 cursor-pointer group transition-colors ${
          isSelected
            ? 'bg-primary/10 text-primary border-r-2 border-primary'
            : 'text-muted-foreground hover:text-foreground hover:bg-border/20'
        }`}
        style={{ paddingLeft: `${8 + depth * 14}px`, paddingRight: '8px' }}
        onClick={() => {
          if (hasChildren) {
            setExpanded(!expanded);
          }
          onSelect(d.id || '');
        }}
      >
        {hasChildren ? (
          expanded ? (
            <ChevronDown size={13} className="shrink-0" />
          ) : (
            <ChevronRight size={13} className="shrink-0" />
          )
        ) : null}
        <StatusIcon status={d.status} />
        <span className="text-xs font-mono truncate flex-1">{d.title}</span>
        {d.status && statusColors && (
          <span
            className={`px-1.5 py-0.5 rounded text-[9px] font-mono shrink-0 ${statusColors.bg} ${statusColors.text} border ${statusColors.border}`}
          >
            {d.status}
          </span>
        )}
        {hovering && d.id && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              onAddChild(d.id!);
            }}
            className="shrink-0 text-muted-foreground hover:text-primary transition"
            title="Add sub-document"
          >
            <Plus size={12} />
          </button>
        )}
      </div>
      {hasChildren && expanded && (
        <div>
          {node.children?.map((child) => (
            <TreeNode
              key={child.doc?.id}
              node={child}
              archiveId={archiveId}
              depth={depth + 1}
              selectedId={selectedId}
              onSelect={onSelect}
              onAddChild={onAddChild}
            />
          ))}
        </div>
      )}
    </div>
  );
}

export default function DocTree({ archiveId, width, onResize }: DocTreeProps) {
  const navigate = useNavigate();
  const { docId } = useParams<{ docId?: string }>();
  const [tree, setTree] = useState<DocTreeNode[]>([]);
  const [resizing, setResizing] = useState(false);

  const loadTree = useCallback(async () => {
    if (!archiveId) return;
    try {
      const res = await docs.getTree({ archiveId });
      setTree(res.items || []);
    } catch {
      setTree([]);
    }
  }, [archiveId]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void loadTree();
  }, [loadTree]);

  const handleSelect = (id: string) => {
    void navigate(`/docs/${archiveId}/${id}`);
  };

  const handleNewDoc = () => {
    void navigate(`/docs/${archiveId}/new`);
  };

  const handleAddChild = (parentId: string) => {
    void navigate(`/docs/${archiveId}/${parentId}/new`);
  };

  // Resize handling
  useEffect(() => {
    if (!resizing) return;
    const handleMouseMove = (e: MouseEvent) => {
      onResize(e.movementX);
    };
    const handleMouseUp = () => setResizing(false);
    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
  }, [resizing, onResize]);

  return (
    <div className="flex h-full" style={{ width }}>
      <div className="flex-1 flex flex-col h-full bg-background border-r border-border overflow-hidden">
        <div className="px-3 py-2 border-b border-border flex items-center justify-between">
          <p className="text-[10px] font-mono text-muted-foreground tracking-widest">DOCUMENTS</p>
          <button
            onClick={handleNewDoc}
            className="text-muted-foreground hover:text-primary transition"
            title="New document"
          >
            <Plus size={13} />
          </button>
        </div>
        <div className="flex-1 overflow-y-auto py-1">
          {tree.length === 0 ? (
            <p className="px-3 py-4 text-[10px] font-mono text-muted-foreground text-center">
              No documents yet
            </p>
          ) : (
            tree.map((node) => (
              <TreeNode
                key={node.doc?.id}
                node={node}
                archiveId={archiveId}
                depth={0}
                selectedId={docId}
                onSelect={handleSelect}
                onAddChild={handleAddChild}
              />
            ))
          )}
        </div>
      </div>
      {/* Resize handle */}
      <div
        className="w-1 cursor-col-resize hover:bg-primary/30 transition-colors shrink-0"
        onMouseDown={() => setResizing(true)}
      />
    </div>
  );
}
