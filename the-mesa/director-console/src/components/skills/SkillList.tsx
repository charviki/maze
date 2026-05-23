import { ConfirmDialog } from '@maze/fabrication';
import type { V1Skill } from '@maze/fabrication';
import { Wrench } from 'lucide-react';
import { SkillDetailPanel } from './SkillDetailPanel';
import { useFabricationList } from '../shared/useFabricationList';
import { FabricationListColumn } from '../shared/FabricationListColumn';

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
  } = useFabricationList({ api, entityLabel: 'Skill', loadErrorMessage: '加载 Skill 列表失败' });

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
        icon={Wrench}
        label="SKILL REGISTRY"
        emptyText="NO SKILLS FABRICATED — INITIATE FIRST SEQUENCE"
        items={items}
        isCreating={isCreating}
        editingItem={editingItem}
        renderItemSubtitle={(item) => item.description || 'NO DESCRIPTION'}
        renderItemExtra={(item) =>
          item.config && Object.keys(item.config).length > 0 ? (
            <span className="text-[10px] text-primary/40 font-mono mt-0.5">
              {Object.keys(item.config).length} CONFIG PARAM(S)
            </span>
          ) : null
        }
        onCreate={handleCreate}
        onEdit={handleEdit}
        onDeleteClick={(name) => setDeleteTarget(name)}
      />
      <SkillDetailPanel
        skill={editingItem}
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
