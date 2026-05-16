import { useState, useEffect, useCallback } from 'react';
import { Button, ConfirmDialog, useToast } from '@maze/fabrication';
import type { V1Skill } from '@maze/fabrication';
import { Plus, Trash2, Pencil, Wrench } from 'lucide-react';
import { SkillEditor } from './SkillEditor';

interface SkillApi {
  list(): Promise<V1Skill[]>;
  create(data: {
    name: string;
    description?: string;
    config?: Record<string, string>;
  }): Promise<V1Skill>;
  update(
    name: string,
    data: { description?: string; config?: Record<string, string> },
  ): Promise<V1Skill>;
  delete(name: string): Promise<void>;
}

export function SkillList({ api }: { api: SkillApi }) {
  const { showToast } = useToast();
  const [items, setItems] = useState<V1Skill[]>([]);
  const [loading, setLoading] = useState(true);
  const [showEditor, setShowEditor] = useState(false);
  const [editingSkill, setEditingSkill] = useState<V1Skill | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  const fetchItems = useCallback(async () => {
    try {
      const list = await api.list();
      setItems(list);
    } catch {
      showToast('error', '加载 Skill 列表失败');
    } finally {
      setLoading(false);
    }
  }, [api, showToast]);

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    void fetchItems();
  }, [fetchItems]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const handleCreate = () => {
    setEditingSkill(null);
    setShowEditor(true);
  };

  const handleEdit = (skill: V1Skill) => {
    setEditingSkill(skill);
    setShowEditor(true);
  };

  const handleSubmit = async (data: {
    name: string;
    description?: string;
    config?: Record<string, string>;
  }) => {
    if (editingSkill) {
      const updated = await api.update(editingSkill.name!, data);
      setItems((prev) => prev.map((s) => (s.name === updated.name ? updated : s)));
      showToast('success', `Skill "${updated.name}" 已更新`);
    } else {
      const created = await api.create(data);
      setItems((prev) => [...prev, created]);
      showToast('success', `Skill "${created.name}" 已创建`);
    }
    setShowEditor(false);
    setEditingSkill(null);
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await api.delete(deleteTarget);
      setItems((prev) => prev.filter((s) => s.name !== deleteTarget));
      showToast('success', `Skill "${deleteTarget}" 已删除`);
    } catch {
      showToast('error', '删除失败');
    }
    setDeleteTarget(null);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full text-primary/40">
        <span className="text-xs uppercase tracking-widest animate-pulse">Loading...</span>
      </div>
    );
  }

  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b border-border/50 flex items-center justify-between">
        <div className="flex items-center gap-2 font-mono uppercase tracking-widest text-primary text-xs">
          <Wrench className="w-4 h-4" />
          Skills
        </div>
        <Button
          variant="ghost"
          size="sm"
          className="text-primary/60 hover:text-primary hover:bg-primary/20 text-xs"
          onClick={handleCreate}
        >
          <Plus className="w-3.5 h-3.5 mr-1" />
          CREATE
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-2">
        {items.length === 0 ? (
          <div className="text-center text-primary/30 text-xs uppercase tracking-widest py-12">
            No skills configured
          </div>
        ) : (
          items.map((skill) => (
            <div
              key={skill.name}
              className="bg-card/50 border border-primary/20 rounded p-3 flex items-center justify-between group hover:border-primary/40 transition-colors"
            >
              <div className="flex-1 min-w-0">
                <div className="font-mono text-sm text-foreground">{skill.name}</div>
                <div className="text-xs text-foreground/50 mt-1 truncate">
                  {skill.description || 'No description'}
                </div>
                {skill.config && Object.keys(skill.config).length > 0 && (
                  <div className="text-[10px] text-primary/40 mt-1">
                    {Object.keys(skill.config).length} config item(s)
                  </div>
                )}
              </div>
              <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 rounded-none text-primary/60 hover:text-primary hover:bg-primary/20"
                  onClick={() => handleEdit(skill)}
                >
                  <Pencil className="w-3 h-3" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 rounded-none text-destructive/60 hover:text-destructive hover:bg-destructive/20"
                  onClick={() => setDeleteTarget(skill.name!)}
                >
                  <Trash2 className="w-3 h-3" />
                </Button>
              </div>
            </div>
          ))
        )}
      </div>

      <SkillEditor
        open={showEditor}
        onOpenChange={setShowEditor}
        skill={editingSkill}
        onSubmit={handleSubmit}
      />

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(v) => !v && setDeleteTarget(null)}
        title="Delete Skill"
        description={`确认删除 Skill "${deleteTarget}"？此操作不可撤销。`}
        onConfirm={handleDelete}
      />
    </div>
  );
}
