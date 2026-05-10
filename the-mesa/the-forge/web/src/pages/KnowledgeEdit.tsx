import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate, useSearchParams, Link } from 'react-router-dom';
import { ArrowLeft, Save, Eye } from 'lucide-react';
import { memories, type ParsedMemory } from '@/api';
import { useToast } from '@maze/fabrication';
import { Skeleton } from '@maze/fabrication';
import CodeMirrorEditor from '@/components/CodeMirrorEditor';
import EditorToolbar from '@/components/EditorToolbar';
import ReactMarkdown from 'react-markdown';

export default function KnowledgeEdit() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const archiveId = searchParams.get('archiveId') || '';
  const { showToast } = useToast();

  const isNew = !id || id === 'new';
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [docType, setDocType] = useState('requirement');
  const [tags, setTags] = useState('');
  const [visibility, setVisibility] = useState('private');
  const [summary, setSummary] = useState('');
  const [loading, setLoading] = useState(!isNew);
  const [saving, setSaving] = useState(false);
  const [preview, setPreview] = useState(false);

  useEffect(() => {
    if (isNew || !id) return;
    let cancelled = false;
    memories
      .get(id)
      .then((doc: ParsedMemory) => {
        if (cancelled) return;
        setTitle(doc.meta.title);
        setContent(doc.content || '');
        setDocType(doc.meta.type || 'requirement');
        setTags(doc.meta.tags?.join(', ') || '');
        setVisibility(doc.meta.visibility || 'private');
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
  }, [id, isNew, showToast]);

  const handleInsert = useCallback((_before: string, _after?: string) => {
    // CodeMirror handles insertion internally
  }, []);

  const handleSave = async () => {
    if (!title.trim()) {
      showToast('error', 'Title is required');
      return;
    }
    setSaving(true);
    try {
      const data = {
        title: title.trim(),
        content,
        type: docType,
        tags: tags
          .split(',')
          .map((t) => t.trim())
          .filter(Boolean),
        visibility,
        summary,
        archiveId: archiveId || undefined,
      };
      if (isNew) {
        const result = await memories.create(data);
        showToast('success', 'Document created');
        void navigate(`/knowledge/${result.id}`);
      } else if (id) {
        await memories.update(id, data);
        showToast('success', 'Document saved');
        void navigate(`/knowledge/${id}`);
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
      <div className="flex items-center justify-between">
        <Link
          to={isNew ? '/knowledge' : `/knowledge/${id}`}
          className="flex items-center gap-2 text-muted-foreground hover:text-foreground font-mono text-sm transition"
        >
          <ArrowLeft size={14} />
          BACK
        </Link>
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

      <div className="space-y-3">
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Document title..."
          className="w-full bg-transparent text-foreground placeholder:text-muted-foreground font-mono text-xl font-bold focus:outline-none"
        />

        <div className="flex items-center gap-3">
          <select
            value={docType}
            onChange={(e) => setDocType(e.target.value)}
            className="bg-card border border-border rounded px-2 py-1.5 text-foreground font-mono text-xs focus:outline-none focus:border-primary"
          >
            <option value="requirement">REQUIREMENT</option>
            <option value="shared">SHARED</option>
            <option value="ops">OPS</option>
            <option value="narrative">NARRATIVE</option>
            <option value="memory">MEMORY</option>
          </select>
          <input
            value={tags}
            onChange={(e) => setTags(e.target.value)}
            placeholder="Tags (comma separated)..."
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

        <input
          value={summary}
          onChange={(e) => setSummary(e.target.value)}
          placeholder="Summary (optional)..."
          className="w-full bg-card border border-border rounded px-3 py-1.5 text-foreground placeholder:text-muted-foreground font-mono text-xs focus:outline-none focus:border-primary"
        />
      </div>

      <div
        className="bg-card border border-border rounded overflow-hidden"
        style={{ height: 'calc(100vh - 320px)' }}
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
