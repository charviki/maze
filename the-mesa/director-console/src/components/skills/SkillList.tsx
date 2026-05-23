import { useState, useEffect, useCallback } from 'react';
import { Button, ConfirmDialog, Panel, DecryptText, useToast } from '@maze/fabrication';
import type { V1Skill } from '@maze/fabrication';
import { Plus, Trash2, Wrench } from 'lucide-react';
import { SkillDetailPanel } from './SkillDetailPanel';
import { clipPathHalf } from '@maze/fabrication';

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
  const [editingSkill, setEditingSkill] = useState<V1Skill | null>(null);
  const [isCreating, setIsCreating] = useState(false);
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
    setIsCreating(true);
  };

  const handleEdit = (skill: V1Skill) => {
    setIsCreating(false);
    setEditingSkill(skill);
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
    setIsCreating(false);
    setEditingSkill(null);
  };

  const handleCancelEdit = () => {
    setIsCreating(false);
    setEditingSkill(null);
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await api.delete(deleteTarget);
      setItems((prev) => prev.filter((s) => s.name !== deleteTarget));
      showToast('success', `Skill "${deleteTarget}" 已删除`);
      if (editingSkill?.name === deleteTarget) {
        setEditingSkill(null);
      }
    } catch {
      showToast('error', '删除失败');
    }
    setDeleteTarget(null);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full w-full text-primary/40">
        <span className="text-xs uppercase tracking-widest animate-pulse">Loading...</span>
      </div>
    );
  }

  return (
    <>
      <div className="border-r border-border/50 flex flex-col bg-background/50 relative z-10 overflow-hidden min-w-[320px]">
        <div className="absolute right-0 top-0 w-[1px] h-full bg-gradient-to-b from-primary/20 to-transparent" />
        <Panel className="flex flex-col h-full relative m-2" cornerSize={16}>
          <div className="pb-4 border-b border-primary/20 space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2 text-primary">
                <Wrench className="w-4 h-4" />
                <h2 className="text-xs font-bold uppercase tracking-widest">
                  <DecryptText text="SKILL REGISTRY" />
                </h2>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="text-primary/60 hover:text-primary hover:bg-primary/20 rounded-none"
                onClick={handleCreate}
              >
                <Plus className="w-4 h-4" />
              </Button>
            </div>
          </div>

          <div className="flex-1 overflow-y-auto pt-4 space-y-2">
            {items.length === 0 ? (
              <div className="text-center text-primary/30 text-[10px] uppercase tracking-widest py-12 font-mono">
                [ NO SKILLS FABRICATED — INITIATE FIRST SEQUENCE ]
              </div>
            ) : (
              items.map((skill) => {
                const isSelected = editingSkill?.name === skill.name;
                return (
                  <div
                    key={skill.name}
                    onClick={() => handleEdit(skill)}
                    className={`group flex items-center justify-between p-3 border-l-2
                      transition-all cursor-pointer backdrop-blur-sm
                      ${
                        isSelected
                          ? 'bg-primary/20 border-primary shadow-[0_0_15px_rgba(0,255,255,0.2)]'
                          : 'bg-black/40 border-primary/30 hover:border-primary/60 hover:bg-primary/10'
                      }`}
                    style={{ clipPath: clipPathHalf(8) }}
                  >
                    <div className="flex-1 min-w-0 pl-1">
                      <div className="flex flex-col overflow-hidden">
                        <span className="text-sm font-mono font-bold tracking-wide uppercase truncate text-primary">
                          {isSelected ? <DecryptText text={skill.name!} /> : skill.name}
                        </span>
                        <span className="text-[10px] text-primary/50 font-mono uppercase tracking-widest truncate mt-0.5">
                          {skill.description || 'NO DESCRIPTION'}
                        </span>
                        {skill.config && Object.keys(skill.config).length > 0 && (
                          <span className="text-[10px] text-primary/40 font-mono mt-0.5">
                            {Object.keys(skill.config).length} CONFIG PARAM(S)
                          </span>
                        )}
                      </div>
                    </div>
                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity pr-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7 rounded-none text-destructive/60 hover:text-destructive hover:bg-destructive/20"
                        onClick={(e) => {
                          e.stopPropagation();
                          setDeleteTarget(skill.name!);
                        }}
                      >
                        <Trash2 className="w-3 h-3" />
                      </Button>
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </Panel>
      </div>

      <SkillDetailPanel
        skill={editingSkill}
        isCreating={isCreating}
        onSubmit={handleSubmit}
        onCancel={handleCancelEdit}
      />

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(v) => !v && setDeleteTarget(null)}
        title="DELETE SKILL"
        description={`确认删除 Skill "${deleteTarget}"？此操作不可撤销。`}
        onConfirm={handleDelete}
      />
    </>
  );
}
