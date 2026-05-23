import { useState, useEffect, useCallback } from 'react';
import { Button, ConfirmDialog, Panel, DecryptText, useToast } from '@maze/fabrication';
import type { V1MCPServer } from '@maze/fabrication';
import { Plus, Trash2, Plug } from 'lucide-react';
import { MCPDetailPanel } from './MCPDetailPanel';
import { clipPathHalf } from '@maze/fabrication';

interface MCPServerApi {
  list(): Promise<V1MCPServer[]>;
  create(data: {
    name: string;
    type: string;
    command?: string;
    url?: string;
    args?: string[];
    env?: Record<string, string>;
  }): Promise<V1MCPServer>;
  update(
    name: string,
    data: {
      type: string;
      command?: string;
      url?: string;
      args?: string[];
      env?: Record<string, string>;
    },
  ): Promise<V1MCPServer>;
  delete(name: string): Promise<void>;
}

export function MCPList({ api }: { api: MCPServerApi }) {
  const { showToast } = useToast();
  const [items, setItems] = useState<V1MCPServer[]>([]);
  const [loading, setLoading] = useState(true);
  const [editingItem, setEditingItem] = useState<V1MCPServer | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  const fetchItems = useCallback(async () => {
    try {
      const list = await api.list();
      setItems(list);
    } catch {
      showToast('error', '加载 MCP Server 列表失败');
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
    setEditingItem(null);
    setIsCreating(true);
  };

  const handleEdit = (item: V1MCPServer) => {
    setIsCreating(false);
    setEditingItem(item);
  };

  const handleSubmit = async (data: {
    name: string;
    type: string;
    command?: string;
    url?: string;
    args?: string[];
    env?: Record<string, string>;
  }) => {
    if (editingItem) {
      const updateData = {
        type: data.type,
        command: data.command,
        url: data.url,
        args: data.args,
        env: data.env,
      };
      const updated = await api.update(editingItem.name!, updateData);
      setItems((prev) => prev.map((s) => (s.name === updated.name ? updated : s)));
      showToast('success', `MCP Server "${updated.name}" 已更新`);
    } else {
      const created = await api.create(data);
      setItems((prev) => [...prev, created]);
      showToast('success', `MCP Server "${created.name}" 已创建`);
    }
    setIsCreating(false);
    setEditingItem(null);
  };

  const handleCancelEdit = () => {
    setIsCreating(false);
    setEditingItem(null);
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await api.delete(deleteTarget);
      setItems((prev) => prev.filter((s) => s.name !== deleteTarget));
      showToast('success', `MCP Server "${deleteTarget}" 已删除`);
      if (editingItem?.name === deleteTarget) {
        setEditingItem(null);
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
                <Plug className="w-4 h-4" />
                <h2 className="text-xs font-bold uppercase tracking-widest">
                  <DecryptText text="MCP SERVER REGISTRY" />
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
                [ NO MCP SERVERS FABRICATED — INITIATE FIRST SEQUENCE ]
              </div>
            ) : (
              items.map((item) => {
                const isSelected = editingItem?.name === item.name;
                return (
                  <div
                    key={item.name}
                    onClick={() => handleEdit(item)}
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
                        <span className="text-[10px] text-primary/50 font-mono uppercase tracking-widest truncate mt-0.5 flex items-center gap-2">
                          <span className="inline-block px-1.5 py-0.5 bg-primary/10 text-primary/60 rounded-none text-[10px] uppercase border border-primary/20">
                            {item.type}
                          </span>
                          <span className="text-primary/40 truncate">
                            {item.type === 'stdio' ? item.command : item.url}
                          </span>
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

      <MCPDetailPanel
        server={editingItem}
        isCreating={isCreating}
        onSubmit={handleSubmit}
        onCancel={handleCancelEdit}
      />

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(v) => !v && setDeleteTarget(null)}
        title="DELETE MCP SERVER"
        description={`确认删除 MCP Server "${deleteTarget}"？此操作不可撤销。`}
        onConfirm={handleDelete}
      />
    </>
  );
}
