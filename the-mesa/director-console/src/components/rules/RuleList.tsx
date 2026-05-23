import { ConfirmDialog } from '@maze/fabrication';
import type { V1Rule } from '@maze/fabrication';
import { BookOpen } from 'lucide-react';
import { RuleDetailPanel } from './RuleDetailPanel';
import { useFabricationList } from '../shared/useFabricationList';
import { FabricationListColumn } from '../shared/FabricationListColumn';

interface RuleApi {
  list(): Promise<V1Rule[]>;
  create(data: { name: string; content?: string }): Promise<V1Rule>;
  update(name: string, data: { content?: string }): Promise<V1Rule>;
  delete(name: string): Promise<void>;
}

export function RuleList({ api }: { api: RuleApi }) {
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
  } = useFabricationList({ api, entityLabel: 'Rule', loadErrorMessage: '加载 Rule 列表失败' });

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
        icon={BookOpen}
        label="RULE REGISTRY"
        emptyText="NO RULES FABRICATED — INITIATE FIRST SEQUENCE"
        items={items}
        isCreating={isCreating}
        editingItem={editingItem}
        renderItemSubtitle={(item) => item.content || 'NO CONTENT'}
        onCreate={handleCreate}
        onEdit={handleEdit}
        onDeleteClick={(name) => setDeleteTarget(name)}
      />
      <RuleDetailPanel
        rule={editingItem}
        isCreating={isCreating}
        onSubmit={handleSubmit}
        onCancel={handleCancelEdit}
      />
      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(v) => !v && setDeleteTarget(null)}
        title="DELETE RULE"
        description={`确认删除 Rule "${deleteTarget}"？此操作不可撤销。`}
        onConfirm={handleDelete}
      />
    </>
  );
}
