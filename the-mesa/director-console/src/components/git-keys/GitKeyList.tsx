import { useState } from 'react';
import { ConfirmDialog } from '@maze/fabrication';
import type { V1GitKey } from '@maze/fabrication';
import { KeyRound } from 'lucide-react';
import { GitKeyDetailPanel } from './GitKeyDetailPanel';
import { useFabricationList } from '../shared/useFabricationList';
import { FabricationListColumn } from '../shared/FabricationListColumn';
import { TOKEN_TYPE_DISPLAY, type CreateGitKeyData } from './constants';

interface GitKeyApi {
  list(): Promise<V1GitKey[]>;
  create(data: CreateGitKeyData): Promise<V1GitKey>;
  delete(name: string): Promise<void>;
}

export function GitKeyList({ api }: { api: GitKeyApi }) {
  const [selectedName, setSelectedName] = useState<string | null>(null);

  const {
    items,
    loading,
    isCreating,
    deleteTarget,
    setDeleteTarget,
    handleCreate,
    handleCancelEdit,
    handleSubmit,
    handleDelete,
  } = useFabricationList({
    api,
    entityLabel: 'Git Key',
    loadErrorMessage: '加载 Git Key 列表失败',
    onAfterCreate: (created) => setSelectedName(created.name!),
    onAfterDelete: (name) => {
      if (selectedName === name) setSelectedName(null);
    },
  });

  const selectedGitKey = items.find((i) => i.name === selectedName) ?? null;

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
        icon={KeyRound}
        label="GIT KEY REGISTRY"
        emptyText="NO GIT KEYS FABRICATED — INITIATE FIRST SEQUENCE"
        items={items}
        isCreating={isCreating}
        editingItem={null}
        selectedName={selectedName}
        renderItemSubtitle={(item) => item.tokenMask}
        renderItemExtra={(item) => (
          <div className="flex items-center gap-2 mt-1">
            {item.tokenType && (
              <span className="inline-block px-1.5 py-0.5 text-[9px] font-mono uppercase tracking-widest border border-primary/30 bg-primary/10 text-primary">
                {TOKEN_TYPE_DISPLAY[item.tokenType] ?? item.tokenType}
              </span>
            )}
            {item.host && (
              <span className="text-[9px] font-mono text-primary/40 uppercase tracking-widest truncate">
                {item.host}
              </span>
            )}
          </div>
        )}
        onCreate={() => {
          handleCreate();
          setSelectedName(null);
        }}
        onEdit={(item) => {
          handleCancelEdit();
          setSelectedName(item.name!);
        }}
        onDeleteClick={(name) => setDeleteTarget(name)}
      />
      <GitKeyDetailPanel
        gitKey={selectedGitKey}
        isCreating={isCreating}
        onSubmit={handleSubmit}
        onCancel={() => {
          handleCancelEdit();
          setSelectedName(null);
        }}
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
