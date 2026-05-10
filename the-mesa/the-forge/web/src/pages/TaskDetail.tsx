import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { ArrowLeft, Edit, Trash2, Clock, User, Eye, CheckSquare } from 'lucide-react';
import { directives, type Directive } from '@/api';
import { useToast } from '@maze/fabrication';
import { Skeleton } from '@maze/fabrication';
import { useModal } from '@/hooks/useModal';
import ModalPortal from '@/components/ModalContent';
import EmptyState from '@/components/EmptyState';
import { formatAbsoluteTime } from '@/lib/time';
import { statusColors, priorityColors } from '@/lib/constants';

export default function TaskDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { showToast } = useToast();
  const { config: modalConfig, confirm, close } = useModal();
  const [task, setTask] = useState<Directive | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!id) return;
    let cancelled = false;
    directives
      .get(id)
      .then((task) => {
        if (!cancelled) setTask(task);
      })
      .catch(() => {
        if (!cancelled) setTask(null);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [id]);

  const handleDelete = () => {
    if (!id) return;
    void confirm('Delete this directive? This cannot be undone.', 'DELETE', 'danger').then(
      async (ok) => {
        if (!ok) return;
        await directives.remove(id);
        showToast('success', 'Directive deleted');
        void navigate('/tasks');
      },
    );
  };

  if (loading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-10 w-3/4" />
        <Skeleton className="h-40" />
      </div>
    );
  }

  if (!task) {
    return (
      <EmptyState
        icon={<CheckSquare size={20} className="text-primary" />}
        title="Directive not found"
        description="The directive may have been deleted"
        action={{
          label: 'BACK TO LIST',
          onClick: () => {
            void navigate('/tasks');
          },
        }}
      />
    );
  }

  const sc = statusColors[task.status] || statusColors.pending;
  const pc = priorityColors[task.priority] || priorityColors.medium;

  return (
    <div className="space-y-6">
      <ModalPortal config={modalConfig} onClose={close} />

      <div className="flex items-center justify-between">
        <Link
          to="/tasks"
          className="flex items-center gap-2 text-muted-foreground hover:text-foreground font-mono text-sm transition"
        >
          <ArrowLeft size={14} />
          BACK
        </Link>
        <div className="flex items-center gap-2">
          <Link
            to={`/tasks/${id}/edit`}
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
          <span className={`px-2 py-0.5 rounded text-xs font-mono ${sc.bg} ${sc.text}`}>
            {task.status}
          </span>
          <span className={`px-2 py-0.5 rounded text-xs font-mono ${pc.bg} ${pc.text}`}>
            {task.priority}
          </span>
        </div>
        <h1 className="text-foreground font-mono font-bold text-2xl">{task.title}</h1>
      </div>

      <div className="flex items-center gap-4 text-muted-foreground font-mono text-xs">
        <span className="flex items-center gap-1">
          <User size={10} />
          {task.author}
        </span>
        {task.assignee && (
          <span className="flex items-center gap-1">
            <User size={10} />
            {task.assignee}
          </span>
        )}
        <span className="flex items-center gap-1">
          <Eye size={10} />
          {task.visibility}
        </span>
        {task.createdAt && (
          <span className="flex items-center gap-1">
            <Clock size={10} />
            {formatAbsoluteTime(task.createdAt)}
          </span>
        )}
      </div>

      {task.description && (
        <div className="bg-card border border-border rounded p-6">
          <p className="text-foreground font-mono text-sm whitespace-pre-wrap">
            {task.description}
          </p>
        </div>
      )}
    </div>
  );
}
