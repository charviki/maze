import { useState, useEffect, useCallback } from 'react';
import { Button, ConfirmDialog, useToast } from '@maze/fabrication';
import type { V1MCPServer } from '@maze/fabrication';
import { Plus, Trash2, Pencil, Plug } from 'lucide-react';
import { MCPEditor } from './MCPEditor';

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
  const [showEditor, setShowEditor] = useState(false);
  const [editingItem, setEditingItem] = useState<V1MCPServer | null>(null);
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
    setShowEditor(true);
  };

  const handleEdit = (item: V1MCPServer) => {
    setEditingItem(item);
    setShowEditor(true);
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
    setShowEditor(false);
    setEditingItem(null);
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await api.delete(deleteTarget);
      setItems((prev) => prev.filter((s) => s.name !== deleteTarget));
      showToast('success', `MCP Server "${deleteTarget}" 已删除`);
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
          <Plug className="w-4 h-4" />
          MCP Servers
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
            No MCP servers configured
          </div>
        ) : (
          items.map((item) => (
            <div
              key={item.name}
              className="bg-card/50 border border-primary/20 rounded p-3 flex items-center justify-between group hover:border-primary/40 transition-colors"
            >
              <div className="flex-1 min-w-0">
                <div className="font-mono text-sm text-foreground">{item.name}</div>
                <div className="text-xs text-foreground/50 mt-1">
                  <span className="inline-block px-1.5 py-0.5 bg-primary/10 text-primary/60 rounded text-[10px] uppercase">
                    {item.type}
                  </span>
                  <span className="ml-2 text-foreground/40">
                    {item.type === 'stdio' ? item.command : item.url}
                  </span>
                </div>
              </div>
              <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 rounded-none text-primary/60 hover:text-primary hover:bg-primary/20"
                  onClick={() => handleEdit(item)}
                >
                  <Pencil className="w-3 h-3" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 rounded-none text-destructive/60 hover:text-destructive hover:bg-destructive/20"
                  onClick={() => setDeleteTarget(item.name!)}
                >
                  <Trash2 className="w-3 h-3" />
                </Button>
              </div>
            </div>
          ))
        )}
      </div>

      <MCPEditor
        open={showEditor}
        onOpenChange={setShowEditor}
        server={editingItem}
        onSubmit={handleSubmit}
      />

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(v) => !v && setDeleteTarget(null)}
        title="Delete MCP Server"
        description={`确认删除 MCP Server "${deleteTarget}"？此操作不可撤销。`}
        onConfirm={handleDelete}
      />
    </div>
  );
}
