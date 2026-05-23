import { useState, useEffect, useCallback } from 'react';
import { useToast } from '@maze/fabrication';
import type { FabricationEntity } from './types';

export interface FabricationApi<T extends FabricationEntity, C, U> {
  list(): Promise<T[]>;
  create(data: C): Promise<T>;
  update?(name: string, data: U): Promise<T>;
  delete(name: string): Promise<void>;
}

export interface UseFabricationListOptions<T extends FabricationEntity, C, U = C> {
  api: FabricationApi<T, C, U>;
  entityLabel: string;
  loadErrorMessage: string;
  deleteErrorMessage?: string;
  prepareUpdateData?: (createData: C, editingItem: T) => U;
  onAfterCreate?: (created: T) => void;
  onAfterDelete?: (deletedName: string) => void;
}

export interface UseFabricationListReturn<T extends FabricationEntity, C> {
  items: T[];
  loading: boolean;
  editingItem: T | null;
  isCreating: boolean;
  deleteTarget: string | null;
  setDeleteTarget: (name: string | null) => void;
  handleCreate: () => void;
  handleEdit: (item: T) => void;
  handleCancelEdit: () => void;
  handleSubmit: (data: C) => Promise<void>;
  handleDelete: () => Promise<void>;
}

export function useFabricationList<T extends FabricationEntity, C, U = C>(
  options: UseFabricationListOptions<T, C, U>,
): UseFabricationListReturn<T, C> {
  const {
    api,
    entityLabel,
    loadErrorMessage,
    deleteErrorMessage = '删除失败',
    prepareUpdateData,
    onAfterCreate,
    onAfterDelete,
  } = options;
  const { showToast } = useToast();

  const [items, setItems] = useState<T[]>([]);
  const [loading, setLoading] = useState(true);
  const [editingItem, setEditingItem] = useState<T | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  const fetchItems = useCallback(async () => {
    try {
      const list = await api.list();
      setItems(list);
    } catch {
      showToast('error', loadErrorMessage);
    } finally {
      setLoading(false);
    }
  }, [api, showToast, loadErrorMessage]);

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    void fetchItems();
  }, [fetchItems]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const handleCreate = () => {
    setEditingItem(null);
    setIsCreating(true);
  };

  const handleEdit = (item: T) => {
    setIsCreating(false);
    setEditingItem(item);
  };

  const handleCancelEdit = () => {
    setIsCreating(false);
    setEditingItem(null);
  };

  const handleSubmit = async (createData: C) => {
    try {
      if (editingItem) {
        if (!api.update) return;
        const updateData = prepareUpdateData
          ? prepareUpdateData(createData, editingItem)
          : (createData as unknown as U);
        const updated = await api.update(editingItem.name!, updateData);
        setItems((prev) => prev.map((s) => (s.name === updated.name ? updated : s)));
        showToast('success', `${entityLabel} "${updated.name}" 已更新`);
      } else {
        const created = await api.create(createData);
        setItems((prev) => [...prev, created]);
        showToast('success', `${entityLabel} "${created.name}" 已创建`);
        onAfterCreate?.(created);
      }
      setIsCreating(false);
      setEditingItem(null);
    } catch {
      showToast('error', '操作失败');
    }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await api.delete(deleteTarget);
      setItems((prev) => prev.filter((s) => s.name !== deleteTarget));
      showToast('success', `${entityLabel} "${deleteTarget}" 已删除`);
      if (editingItem?.name === deleteTarget) {
        setEditingItem(null);
      }
      onAfterDelete?.(deleteTarget);
    } catch {
      showToast('error', deleteErrorMessage);
    }
    setDeleteTarget(null);
  };

  return {
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
  };
}
