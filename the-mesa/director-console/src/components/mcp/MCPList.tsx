import { ConfirmDialog } from '@maze/fabrication';
import type { V1MCPServer } from '@maze/fabrication';
import { Plug } from 'lucide-react';
import { MCPDetailPanel } from './MCPDetailPanel';
import { useFabricationList } from '../shared/useFabricationList';
import { FabricationListColumn } from '../shared/FabricationListColumn';

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
  const {
    items,
    loading,
    editingItem,
    isCreating,
    deleteTarget,
    setDeleteTarget,
    handleCreate,
    handleEdit,
    handleCancelEdit,
    handleSubmit,
    handleDelete,
  } = useFabricationList({
    api,
    entityLabel: 'MCP Server',
    loadErrorMessage: '加载 MCP Server 列表失败',
    prepareUpdateData: (data) => ({
      type: data.type,
      command: data.command,
      url: data.url,
      args: data.args,
      env: data.env,
    }),
  });

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full w-full text-primary/40">
        <span className="text-xs uppercase tracking-widest animate-pulse">Loading...</span>
      </div>
    );
  }

  return (
    <>
      <FabricationListColumn
        icon={Plug}
        label="MCP SERVER REGISTRY"
        emptyText="NO MCP SERVERS FABRICATED — INITIATE FIRST SEQUENCE"
        items={items}
        isCreating={isCreating}
        editingItem={editingItem}
        renderItemSubtitle={(item) => (
          <span className="flex items-center gap-2">
            <span className="inline-block px-1.5 py-0.5 bg-primary/10 text-primary/60 rounded-none text-[10px] uppercase border border-primary/20">
              {item.type}
            </span>
            <span className="text-primary/40 truncate">
              {item.type === 'stdio' ? item.command : item.url}
            </span>
          </span>
        )}
        onCreate={handleCreate}
        onEdit={handleEdit}
        onDeleteClick={(name) => setDeleteTarget(name)}
      />
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
