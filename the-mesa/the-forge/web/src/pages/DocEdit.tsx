import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Save, Eye } from 'lucide-react';
import { docs, type Doc } from '@/api';
import { useToast } from '@maze/fabrication';
import { Skeleton } from '@maze/fabrication';
import CodeMirrorEditor from '@/components/CodeMirrorEditor';
import EditorToolbar from '@/components/EditorToolbar';
import ReactMarkdown from 'react-markdown';
import { useDocTreeRefresh } from '@/contexts/DocTreeContext';

export default function DocEdit() {
  const { archiveId, docId, parentId } = useParams<{
    archiveId: string;
    docId?: string;
    parentId?: string;
  }>();
  const navigate = useNavigate();
  const { showToast } = useToast();
  const { refreshTree } = useDocTreeRefresh();

  const isNew = !docId || docId === 'new' || parentId === 'new';
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [status, setStatus] = useState('');
  const [priority, setPriority] = useState('');
  const [assignee, setAssignee] = useState('');
  const [tags, setTags] = useState('');
  const [visibility, setVisibility] = useState('public');
  const [summary, setSummary] = useState('');
  const [loading, setLoading] = useState(!isNew);
  const [saving, setSaving] = useState(false);
  const [preview, setPreview] = useState(false);

  useEffect(() => {
    if (isNew || !docId) return;
    let cancelled = false;
    docs
      .get(docId)
      .then((doc: Doc) => {
        if (cancelled) return;
        setTitle(doc.title || '');
        setContent(doc.content || '');
        setStatus(doc.status || '');
        setPriority(doc.priority || '');
        setAssignee(doc.assignee || '');
        setTags(doc.tags?.join(', ') || '');
        setVisibility(doc.visibility || 'public');
        setSummary(doc.summary || '');
      })
      .catch(() => {
        if (cancelled) return;
        showToast('error', 'Failed to load document');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [docId, isNew, showToast]);

  const handleInsert = useCallback((_before: string, _after?: string) => {
    // CodeMirror handles insertion internally
  }, []);

  const handleSave = async () => {
    if (!title.trim()) {
      showToast('error', 'Title is required');
      return;
    }
    if (!archiveId) {
      showToast('error', 'No archive selected');
      return;
    }
    setSaving(true);
    try {
      const data: Partial<Doc> = {
        archiveId,
        title: title.trim(),
        content,
        status: status || undefined,
        priority: priority || undefined,
        assignee: assignee || undefined,
        tags: tags
          .split(',')
          .map((t) => t.trim())
          .filter(Boolean),
        visibility,
        summary,
        parentId: isNew && parentId ? parentId : undefined,
      };
      if (isNew) {
        const result = await docs.create(data);
        refreshTree();
        showToast('success', 'Document created');
        void navigate(`/docs/${archiveId}/${result.id}`);
      } else if (docId) {
        await docs.update(docId, data);
        refreshTree();
        showToast('success', 'Document saved');
        void navigate(`/docs/${archiveId}/${docId}`);
      }
    } catch {
      showToast('error', 'Failed to save document');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-10 w-3/4" />
        <Skeleton className="h-96" />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Actions */}
      <div className="flex items-center justify-between">
        <button
          onClick={() => navigate(isNew ? `/docs/${archiveId}` : `/docs/${archiveId}/${docId}`)}
          className="text-muted-foreground hover:text-foreground font-mono text-sm transition"
        >
          ← BACK
        </button>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setPreview(!preview)}
            className="flex items-center gap-1 px-3 py-1.5 bg-card border border-border text-foreground font-mono text-xs rounded hover:bg-card/80 transition"
          >
            <Eye size={12} />
            {preview ? 'EDITOR' : 'PREVIEW'}
          </button>
          <button
            onClick={handleSave}
            disabled={saving}
            className="flex items-center gap-1 px-3 py-1.5 bg-primary text-primary-foreground font-mono text-xs rounded hover:bg-primary/90 transition disabled:opacity-50"
          >
            <Save size={12} />
            {saving ? 'SAVING...' : 'SAVE'}
          </button>
        </div>
      </div>

      {/* Metadata */}
      <div className="space-y-3">
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Document title..."
          className="w-full bg-transparent text-foreground placeholder:text-muted-foreground font-mono text-xl font-bold focus:outline-none"
        />

        <div className="flex items-center gap-2 flex-wrap">
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className="bg-card border border-border rounded px-2 py-1.5 text-foreground font-mono text-xs focus:outline-none focus:border-primary"
          >
            <option value="">NO STATUS</option>
            <option value="pending">PENDING</option>
            <option value="active">ACTIVE</option>
            <option value="done">DONE</option>
            <option value="failed">FAILED</option>
          </select>
          {status && (
            <>
              <select
                value={priority}
                onChange={(e) => setPriority(e.target.value)}
                className="bg-card border border-border rounded px-2 py-1.5 text-foreground font-mono text-xs focus:outline-none focus:border-primary"
              >
                <option value="">NO PRIORITY</option>
                <option value="low">LOW</option>
                <option value="medium">MEDIUM</option>
                <option value="high">HIGH</option>
                <option value="critical">CRITICAL</option>
              </select>
              <input
                value={assignee}
                onChange={(e) => setAssignee(e.target.value)}
                placeholder="Assignee..."
                className="flex-1 min-w-[120px] bg-card border border-border rounded px-2 py-1.5 text-foreground placeholder:text-muted-foreground font-mono text-xs focus:outline-none focus:border-primary"
              />
            </>
          )}
          <input
            value={tags}
            onChange={(e) => setTags(e.target.value)}
            placeholder="Tags (comma separated)..."
            className="flex-1 min-w-[150px] bg-card border border-border rounded px-2 py-1.5 text-foreground placeholder:text-muted-foreground font-mono text-xs focus:outline-none focus:border-primary"
          />
          <select
            value={visibility}
            onChange={(e) => setVisibility(e.target.value)}
            className="bg-card border border-border rounded px-2 py-1.5 text-foreground font-mono text-xs focus:outline-none focus:border-primary"
          >
            <option value="public">PUBLIC</option>
            <option value="private">PRIVATE</option>
          </select>
        </div>

        <input
          value={summary}
          onChange={(e) => setSummary(e.target.value)}
          placeholder="Summary (optional)..."
          className="w-full bg-card border border-border rounded px-3 py-1.5 text-foreground placeholder:text-muted-foreground font-mono text-xs focus:outline-none focus:border-primary"
        />
      </div>

      {/* Editor */}
      <div
        className="bg-card border border-border rounded overflow-hidden"
        style={{ height: 'calc(100vh - 350px)' }}
      >
        {!preview && <EditorToolbar onInsert={handleInsert} />}
        <div className="h-full">
          {preview ? (
            <div className="p-6 overflow-auto h-full">
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
                <ReactMarkdown>{content}</ReactMarkdown>
              </div>
            </div>
          ) : (
            <CodeMirrorEditor
              value={content}
              onChange={setContent}
              placeholder="Start writing in Markdown..."
              autoFocus
              className="h-full"
            />
          )}
        </div>
      </div>
    </div>
  );
}
