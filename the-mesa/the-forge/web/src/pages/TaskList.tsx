import { useState, useEffect, useMemo } from 'react';
import { Link, useSearchParams, useNavigate } from 'react-router-dom';
import { CheckSquare, Plus, Search } from 'lucide-react';
import { directives, type Directive } from '@/api';
import { Skeleton } from '@maze/fabrication';
import { useListNavigation } from '@/hooks/useListNavigation';
import { formatRelativeTime } from '@/lib/time';
import { statusColors, priorityColors } from '@/lib/constants';
import EmptyState from '@/components/EmptyState';

export default function TaskList() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [tasks, setTasks] = useState<Directive[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState(searchParams.get('status') || '');
  const [priorityFilter, setPriorityFilter] = useState(searchParams.get('priority') || '');

  useEffect(() => {
    const params: Record<string, string> = {};
    if (statusFilter) params.status = statusFilter;
    if (priorityFilter) params.priority = priorityFilter;
    directives
      .list(params)
      .then((res) => setTasks(res.items || []))
      .catch(() => setTasks([]))
      .finally(() => setLoading(false));
  }, [statusFilter, priorityFilter]);

  const filteredTasks = useMemo(() => {
    if (!searchQuery.trim()) return tasks;
    const q = searchQuery.toLowerCase();
    return tasks.filter(
      (t) => t.title.toLowerCase().includes(q) || t.description?.toLowerCase().includes(q),
    );
  }, [tasks, searchQuery]);

  const navItems = useMemo(() => filteredTasks.map((t) => ({ id: t.id })), [filteredTasks]);
  const { focusedIndex, handleKeyDown } = useListNavigation({
    items: navItems,
    basePath: '/tasks',
  });

  if (loading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-40" />
        <div className="flex gap-3">
          <Skeleton className="h-9 flex-1" />
          <Skeleton className="h-9 w-24" />
          <Skeleton className="h-9 w-24" />
        </div>
        {[1, 2, 3, 4, 5].map((i) => (
          <Skeleton key={i} className="h-20" />
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-4" onKeyDown={handleKeyDown} tabIndex={0}>
      <div className="flex items-center justify-between">
        <h1 className="text-foreground font-mono font-bold text-xl">DIRECTIVES</h1>
        <Link
          to="/tasks/new"
          className="flex items-center gap-1 px-3 py-1.5 bg-primary text-primary-foreground font-mono text-xs rounded hover:bg-primary/90 transition"
        >
          <Plus size={12} />
          NEW DIRECTIVE
        </Link>
      </div>

      <div className="flex items-center gap-3">
        <div className="flex-1 flex items-center gap-2 bg-card border border-border rounded px-3 py-2">
          <Search size={14} className="text-muted-foreground shrink-0" />
          <input
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search directives..."
            className="flex-1 bg-transparent text-foreground placeholder:text-muted-foreground font-mono text-sm focus:outline-none"
          />
        </div>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="bg-card border border-border rounded px-2 py-2 text-foreground font-mono text-xs focus:outline-none focus:border-primary"
        >
          <option value="">ALL STATUS</option>
          <option value="pending">PENDING</option>
          <option value="in_progress">IN PROGRESS</option>
          <option value="completed">COMPLETED</option>
          <option value="cancelled">CANCELLED</option>
        </select>
        <select
          value={priorityFilter}
          onChange={(e) => setPriorityFilter(e.target.value)}
          className="bg-card border border-border rounded px-2 py-2 text-foreground font-mono text-xs focus:outline-none focus:border-primary"
        >
          <option value="">ALL PRIORITY</option>
          <option value="high">HIGH</option>
          <option value="medium">MEDIUM</option>
          <option value="low">LOW</option>
        </select>
      </div>

      {filteredTasks.length === 0 ? (
        <EmptyState
          icon={<CheckSquare size={20} className="text-primary" />}
          title="No directives found"
          description={
            searchQuery || statusFilter || priorityFilter
              ? 'Try adjusting your filters'
              : 'Create your first directive'
          }
          action={
            !searchQuery && !statusFilter && !priorityFilter
              ? {
                  label: 'CREATE DIRECTIVE',
                  onClick: () => {
                    void navigate('/tasks/new');
                  },
                }
              : undefined
          }
        />
      ) : (
        <div className="space-y-1">
          {filteredTasks.map((task, i) => {
            const isFocused = i === focusedIndex;
            const sc = statusColors[task.status] || statusColors.pending;
            const pc = priorityColors[task.priority] || priorityColors.medium;

            return (
              <Link
                key={task.id}
                to={`/tasks/${task.id}`}
                className={`flex items-center gap-4 px-4 py-3 rounded transition group ${
                  isFocused
                    ? 'bg-primary/10 border border-primary/30'
                    : 'bg-card border border-border hover:bg-card/80'
                }`}
              >
                <CheckSquare
                  size={16}
                  className={`shrink-0 ${task.status === 'completed' ? 'text-green-500' : 'text-muted-foreground'}`}
                />
                <div className="flex-1 min-w-0">
                  <p
                    className={`font-mono text-sm truncate ${isFocused ? 'text-primary' : 'text-foreground group-hover:text-primary'} transition ${task.status === 'completed' ? 'line-through text-muted-foreground' : ''}`}
                  >
                    {task.title}
                  </p>
                  {task.description && (
                    <p className="text-muted-foreground font-mono text-xs truncate mt-0.5">
                      {task.description}
                    </p>
                  )}
                </div>
                <span className={`px-1.5 py-0.5 rounded text-[10px] font-mono ${sc.bg} ${sc.text}`}>
                  {task.status}
                </span>
                <span className={`px-1.5 py-0.5 rounded text-[10px] font-mono ${pc.bg} ${pc.text}`}>
                  {task.priority}
                </span>
                {task.updatedAt && (
                  <span className="text-muted-foreground font-mono text-[10px] shrink-0">
                    {formatRelativeTime(task.updatedAt)}
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
