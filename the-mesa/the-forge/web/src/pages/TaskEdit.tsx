import { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { ArrowLeft, Save } from 'lucide-react';
import { directives, type Directive } from '@/api';
import { useToast } from '@maze/fabrication';
import { Skeleton } from '@maze/fabrication';

export default function TaskEdit() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { showToast } = useToast();

  const isNew = !id || id === 'new';
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [status, setStatus] = useState('pending');
  const [priority, setPriority] = useState('medium');
  const [assignee, setAssignee] = useState('');
  const [visibility, setVisibility] = useState('private');
  const [loading, setLoading] = useState(!isNew);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (isNew || !id) return;
    let cancelled = false;
    directives
      .get(id)
      .then((task: Directive) => {
        if (cancelled) return;
        setTitle(task.title);
        setDescription(task.description || '');
        setStatus(task.status || 'pending');
        setPriority(task.priority || 'medium');
        setAssignee(task.assignee || '');
        setVisibility(task.visibility || 'private');
      })
      .catch(() => {
        if (cancelled) return;
        showToast('error', 'Failed to load directive');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [id, isNew, showToast]);

  const handleSave = async () => {
    if (!title.trim()) {
      showToast('error', 'Title is required');
      return;
    }
    setSaving(true);
    try {
      const data = {
        title: title.trim(),
        description,
        status,
        priority,
        assignee: assignee || undefined,
        visibility,
      };
      if (isNew) {
        const result = await directives.create(data);
        showToast('success', 'Directive created');
        void navigate(`/tasks/${result.id}`);
      } else if (id) {
        await directives.update(id, data);
        showToast('success', 'Directive saved');
        void navigate(`/tasks/${id}`);
      }
    } catch {
      showToast('error', 'Failed to save directive');
    } finally {
      setSaving(false);
    }
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

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <Link
          to={isNew ? '/tasks' : `/tasks/${id}`}
          className="flex items-center gap-2 text-muted-foreground hover:text-foreground font-mono text-sm transition"
        >
          <ArrowLeft size={14} />
          BACK
        </Link>
        <button
          onClick={handleSave}
          disabled={saving}
          className="flex items-center gap-1 px-3 py-1.5 bg-primary text-primary-foreground font-mono text-xs rounded hover:bg-primary/90 transition disabled:opacity-50"
        >
          <Save size={12} />
          {saving ? 'SAVING...' : 'SAVE'}
        </button>
      </div>

      <div className="space-y-3">
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Directive title..."
          className="w-full bg-transparent text-foreground placeholder:text-muted-foreground font-mono text-xl font-bold focus:outline-none"
        />

        <div className="flex items-center gap-3">
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className="bg-card border border-border rounded px-2 py-1.5 text-foreground font-mono text-xs focus:outline-none focus:border-primary"
          >
            <option value="pending">PENDING</option>
            <option value="in_progress">IN PROGRESS</option>
            <option value="completed">COMPLETED</option>
            <option value="cancelled">CANCELLED</option>
          </select>
          <select
            value={priority}
            onChange={(e) => setPriority(e.target.value)}
            className="bg-card border border-border rounded px-2 py-1.5 text-foreground font-mono text-xs focus:outline-none focus:border-primary"
          >
            <option value="high">HIGH</option>
            <option value="medium">MEDIUM</option>
            <option value="low">LOW</option>
          </select>
          <input
            value={assignee}
            onChange={(e) => setAssignee(e.target.value)}
            placeholder="Assignee..."
            className="flex-1 bg-card border border-border rounded px-3 py-1.5 text-foreground placeholder:text-muted-foreground font-mono text-xs focus:outline-none focus:border-primary"
          />
          <select
            value={visibility}
            onChange={(e) => setVisibility(e.target.value)}
            className="bg-card border border-border rounded px-2 py-1.5 text-foreground font-mono text-xs focus:outline-none focus:border-primary"
          >
            <option value="private">PRIVATE</option>
            <option value="team">TEAM</option>
            <option value="public">PUBLIC</option>
          </select>
        </div>

        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Description..."
          className="w-full bg-card border border-border rounded px-3 py-2 text-foreground placeholder:text-muted-foreground font-mono text-sm focus:outline-none focus:border-primary resize-y"
          rows={12}
        />
      </div>
    </div>
  );
}
