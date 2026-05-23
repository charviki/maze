import { useState, useEffect, useCallback } from 'react';
import { Button, ConfirmDialog, Panel, DecryptText, useToast } from '@maze/fabrication';
import type { V1GitKey } from '@maze/fabrication';
import { Plus, Trash2, KeyRound } from 'lucide-react';
import { GitKeyDetailPanel } from './GitKeyDetailPanel';
import { clipPathHalf } from '@maze/fabrication';

interface GitKeyApi {
  list(): Promise<V1GitKey[]>;
  create(data: { name: string; token: string }): Promise<V1GitKey>;
  delete(name: string): Promise<void>;
}

export function GitKeyList({ api }: { api: GitKeyApi }) {
  const { showToast } = useToast();
  const [items, setItems] = useState<V1GitKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [isCreating, setIsCreating] = useState(false);
  const [selectedName, setSelectedName] = useState<string | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  const fetchItems = useCallback(async () => {
    try {
      const list = await api.list();
      setItems(list);
    } catch {
      showToast('error', '加载 Git Key 列表失败');
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
    setIsCreating(true);
    setSelectedName(null);
  };

  const handleSubmit = async (data: { name: string; token: string }) => {
    const created = await api.create(data);
    setItems((prev) => [...prev, created]);
    showToast('success', `Git Key "${created.name}" 已创建`);
    setIsCreating(false);
    setSelectedName(created.name!);
  };

  const handleCancelEdit = () => {
    setIsCreating(false);
    setSelectedName(null);
  };

  const handleSelect = (item: V1GitKey) => {
    setIsCreating(false);
    setSelectedName(item.name!);
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await api.delete(deleteTarget);
      setItems((prev) => prev.filter((s) => s.name !== deleteTarget));
      showToast('success', `Git Key "${deleteTarget}" 已删除`);
      if (selectedName === deleteTarget) {
        setSelectedName(null);
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

  const selectedGitKey = items.find((i) => i.name === selectedName) ?? null;

  return (
    <>
      <div className="border-r border-border/50 flex flex-col bg-background/50 relative z-10 overflow-hidden min-w-[320px]">
        <div className="absolute right-0 top-0 w-[1px] h-full bg-gradient-to-b from-primary/20 to-transparent" />
        <Panel className="flex flex-col h-full relative m-2" cornerSize={16}>
          <div className="pb-4 border-b border-primary/20 space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2 text-primary">
                <KeyRound className="w-4 h-4" />
                <h2 className="text-xs font-bold uppercase tracking-widest">
                  <DecryptText text="GIT KEY REGISTRY" />
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
            {items.length === 0 && !isCreating ? (
              <div className="text-center text-primary/30 text-[10px] uppercase tracking-widest py-12 font-mono">
                [ NO GIT KEYS FABRICATED — INITIATE FIRST SEQUENCE ]
              </div>
            ) : (
              items.map((item) => {
                const isSelected = selectedName === item.name;
                return (
                  <div
                    key={item.name}
                    onClick={() => handleSelect(item)}
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
                          {isSelected ? <DecryptText text={item.name!} /> : item.name}
                        </span>
                        <span className="text-[10px] text-primary/40 font-mono mt-0.5 truncate">
                          {item.tokenMask}
                        </span>
                      </div>
                    </div>
                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity pr-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7 rounded-none text-destructive/60 hover:text-destructive hover:bg-destructive/20"
                        onClick={(e) => {
                          e.stopPropagation();
                          setDeleteTarget(item.name!);
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

      <GitKeyDetailPanel
        gitKey={selectedGitKey}
        isCreating={isCreating}
        onSubmit={handleSubmit}
        onCancel={handleCancelEdit}
      />

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(v) => !v && setDeleteTarget(null)}
        title="DELETE GIT KEY"
        description={`确认删除 Git Key "${deleteTarget}"？此操作不可撤销。`}
        onConfirm={handleDelete}
      />
    </>
  );
}
