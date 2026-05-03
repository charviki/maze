import { useState, useEffect, useCallback } from 'react';
import type {
  NormalizedTemplate,
  ConfigFileUpdate,
  SessionTemplate,
  EnvDef,
  FileDef,
} from '../../types';
import type { IAgentApiClient } from '../../api';
import { ConfirmDialog } from '../ui/ConfirmDialog';
import { useToast } from '../ui/Toast';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '../ui/dialog';
import { TemplateList } from './TemplateList';
import { TemplateEditor, type EditableConfigFile } from './TemplateEditor';

interface TemplateManagerProps {
  open: boolean;
  onClose: () => void;
  apiClient: IAgentApiClient;
}

const emptyTemplate = (): NormalizedTemplate => ({
  id: '',
  name: '',
  command: '',
  description: '',
  icon: '📦',
  builtin: false,
  defaults: { env: {}, files: [] },
  sessionSchema: { envDefs: [], fileDefs: [] },
});

function cloneTemplate(tpl: NormalizedTemplate): NormalizedTemplate {
  return {
    ...tpl,
    defaults: {
      env: { ...(tpl.defaults?.env ?? {}) },
      files: (tpl.defaults?.files ?? []).map((file) => ({ ...file })),
    },
    sessionSchema: {
      envDefs: (tpl.sessionSchema?.envDefs ?? []).map((def) => ({ ...def })),
      fileDefs: (tpl.sessionSchema?.fileDefs ?? []).map((def) => ({ ...def })),
    },
  };
}

export function TemplateManager({ open, onClose, apiClient }: TemplateManagerProps) {
  const { showToast } = useToast();
  const [templates, setTemplates] = useState<NormalizedTemplate[]>([]);
  const [editing, setEditing] = useState<NormalizedTemplate | null>(null);
  const [isNew, setIsNew] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const [globalFiles, setGlobalFiles] = useState<EditableConfigFile[]>([]);
  const [loadingEditor, setLoadingEditor] = useState(false);
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState('');
  const [confirmGlobalSaveOpen, setConfirmGlobalSaveOpen] = useState(false);

  const load = useCallback(async () => {
    const res = await apiClient.listTemplates();
    if (res.status === 'ok' && res.data) {
      setTemplates(res.data);
      return;
    }
    showToast('error', res.message || '加载模板失败');
  }, [apiClient, showToast]);

  // fetch-in-effect: Dialog 打开时从 API 加载模板列表
  useEffect(() => {
    /* eslint-disable react-hooks/set-state-in-effect */
    if (open) void load();
    /* eslint-enable react-hooks/set-state-in-effect */
  }, [open, load]);

  const resetEditor = () => {
    setEditing(null);
    setGlobalFiles([]);
    setIsNew(false);
    setLoadingEditor(false);
    setSaving(false);
    setSaveError('');
    setConfirmGlobalSaveOpen(false);
  };

  const startNew = () => {
    setIsNew(true);
    setEditing(emptyTemplate());
    setGlobalFiles([]);
    setSaveError('');
  };

  const startClone = (tpl: NormalizedTemplate) => {
    setIsNew(true);
    setEditing(
      cloneTemplate({
        ...tpl,
        id: (tpl.id ?? '') + '-copy',
        name: (tpl.name ?? '') + ' (Copy)',
        builtin: false,
      }),
    );
    setGlobalFiles([]);
    setSaveError('');
  };

  const startEdit = async (tpl: NormalizedTemplate) => {
    if (!tpl.id) {
      showToast('error', '模板 ID 不能为空');
      return;
    }
    setLoadingEditor(true);
    setSaveError('');
    const configRes = await apiClient.getTemplateConfig(tpl.id);
    if (configRes.status !== 'ok' || !configRes.data) {
      setLoadingEditor(false);
      showToast('error', configRes.message || '加载模板全局配置失败');
      return;
    }

    setIsNew(false);
    setEditing(cloneTemplate(tpl));
    setGlobalFiles(
      (configRes.data.files ?? []).map((file) => ({
        ...file,
        exists: file._exists ?? false,
        content: file.content ?? '',
        original_content: file.content ?? '',
        baseHash: file.hash ?? '',
      })),
    );
    setLoadingEditor(false);
  };

  const handleDelete = (id: string) => {
    setDeleteTarget(id);
  };

  const confirmDelete = async () => {
    if (!deleteTarget) return;
    const res = await apiClient.deleteTemplate(deleteTarget);
    if (res.status !== 'ok') {
      showToast('error', res.message || '删除模板失败');
      return;
    }
    showToast('success', `已删除模板 ${deleteTarget}`);
    setDeleteTarget(null);
    void load();
  };

  const hasConfigChanges = isNew
    ? false
    : globalFiles.some((file) => file.content !== file.original_content);

  const handleSave = async () => {
    if (!editing) return;

    if (!isNew && hasConfigChanges) {
      setConfirmGlobalSaveOpen(true);
      return;
    }

    await performSave(false);
  };

  const performSave = async (allowRealGlobalWrite: boolean) => {
    if (!editing) return;
    setSaving(true);
    setSaveError('');

    try {
      if (isNew) {
        const createRes = await apiClient.createTemplate(editing);
        if (createRes.status !== 'ok') {
          throw new Error(createRes.message || '创建模板失败');
        }
        showToast('success', `已创建模板 ${editing.name || editing.id}`);
        resetEditor();
        await load();
        return;
      }

      const updateRes = await apiClient.updateTemplate(editing.id ?? '', editing);
      if (updateRes.status !== 'ok') {
        throw new Error(updateRes.message || '保存模板失败');
      }

      if (allowRealGlobalWrite && hasConfigChanges) {
        const configReq: ConfigFileUpdate[] = globalFiles.map((file) => ({
          path: file.path,
          content: file.content,
          baseHash: file.baseHash,
        }));
        const configRes = await apiClient.updateTemplateConfig(editing.id ?? '', {
          files: configReq,
        });
        if (configRes.status !== 'ok' || !configRes.data) {
          if (configRes.code === 'config_conflict') {
            const conflictPaths = (configRes.conflicts || []).map((item) => item.path).join(', ');
            throw new Error(
              conflictPaths
                ? `配置已变更，请重新加载后再修改：${conflictPaths}`
                : configRes.message || '配置已变更，请重新加载后再修改',
            );
          }
          throw new Error(configRes.message || '保存真实全局配置失败');
        }
        setGlobalFiles(
          (configRes.data.files ?? []).map((file) => ({
            ...file,
            exists: file._exists ?? false,
            content: file.content ?? '',
            original_content: file.content ?? '',
            baseHash: file.hash ?? '',
          })),
        );
      }

      showToast('success', `已保存模板 ${editing.name || editing.id}`);
      resetEditor();
      await load();
    } catch (err) {
      const message = err instanceof Error ? err.message : '保存模板失败';
      setSaveError(message);
      showToast('error', message);
    } finally {
      setSaving(false);
      setConfirmGlobalSaveOpen(false);
    }
  };

  const builtinTemplates = templates.filter((t) => t.builtin);
  const userTemplates = templates.filter((t) => !t.builtin);

  // 字段更新与 Schema 操作的回调，由容器统一管理状态
  const updateField = <K extends keyof SessionTemplate>(key: K, val: SessionTemplate[K]) => {
    if (!editing) return;
    setEditing((prev) => (prev ? { ...prev, [key]: val } : null));
  };

  const addDefaultFile = () => {
    if (!editing) return;
    setEditing((prev) =>
      prev
        ? {
            ...prev,
            defaults: {
              ...prev.defaults,
              files: [...prev.defaults.files, { path: '', content: '' }],
            },
          }
        : null,
    );
  };

  const updateDefaultFile = (idx: number, field: 'path' | 'content', val: string) => {
    if (!editing) return;
    const files = [...editing.defaults.files];
    files[idx] = { ...files[idx], [field]: val };
    setEditing((prev) => (prev ? { ...prev, defaults: { ...prev.defaults, files } } : null));
  };

  const removeDefaultFile = (idx: number) => {
    if (!editing) return;
    setEditing((prev) =>
      prev
        ? {
            ...prev,
            defaults: { ...prev.defaults, files: prev.defaults.files.filter((_, i) => i !== idx) },
          }
        : null,
    );
  };

  const addEnvDef = () => {
    if (!editing) return;
    setEditing((prev) =>
      prev
        ? {
            ...prev,
            sessionSchema: {
              ...prev.sessionSchema,
              envDefs: [
                ...prev.sessionSchema.envDefs,
                { key: '', label: '', required: false, placeholder: '', sensitive: false },
              ],
            },
          }
        : null,
    );
  };

  const updateEnvDef = (idx: number, field: keyof EnvDef, val: string | boolean) => {
    if (!editing) return;
    const defs = [...editing.sessionSchema.envDefs];
    defs[idx] = { ...defs[idx], [field]: val };
    setEditing((prev) =>
      prev ? { ...prev, sessionSchema: { ...prev.sessionSchema, envDefs: defs } } : null,
    );
  };

  const removeEnvDef = (idx: number) => {
    if (!editing) return;
    setEditing((prev) =>
      prev
        ? {
            ...prev,
            sessionSchema: {
              ...prev.sessionSchema,
              envDefs: prev.sessionSchema.envDefs.filter((_, i) => i !== idx),
            },
          }
        : null,
    );
  };

  const addFileDef = () => {
    if (!editing) return;
    setEditing((prev) =>
      prev
        ? {
            ...prev,
            sessionSchema: {
              ...prev.sessionSchema,
              fileDefs: [
                ...prev.sessionSchema.fileDefs,
                { path: '', label: '', required: false, defaultContent: '' },
              ],
            },
          }
        : null,
    );
  };

  const updateFileDef = (idx: number, field: keyof FileDef, val: string | boolean) => {
    if (!editing) return;
    const defs = [...editing.sessionSchema.fileDefs];
    defs[idx] = { ...defs[idx], [field]: val };
    setEditing((prev) =>
      prev ? { ...prev, sessionSchema: { ...prev.sessionSchema, fileDefs: defs } } : null,
    );
  };

  const removeFileDef = (idx: number) => {
    if (!editing) return;
    setEditing((prev) =>
      prev
        ? {
            ...prev,
            sessionSchema: {
              ...prev.sessionSchema,
              fileDefs: prev.sessionSchema.fileDefs.filter((_, i) => i !== idx),
            },
          }
        : null,
    );
  };

  const updateGlobalFileContent = (path: string, content: string) => {
    setGlobalFiles((prev) =>
      prev.map((file) => (file.path === path ? { ...file, content } : file)),
    );
  };

  const metadataLocked = !isNew && !!editing?.builtin;
  const existingTemplate = !isNew;

  return (
    <>
      <Dialog
        open={open}
        onOpenChange={(v) => {
          if (!v) onClose();
        }}
      >
        <DialogContent className="max-w-5xl max-h-[85vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>模板管理</DialogTitle>
            <DialogDescription>管理 Agent 模板的配置和参数定义</DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            {loadingEditor && (
              <div className="text-sm text-muted-foreground bg-muted rounded px-3 py-2">
                正在读取节点上的真实模板配置...
              </div>
            )}
            {!editing ? (
              <TemplateList
                builtinTemplates={builtinTemplates}
                userTemplates={userTemplates}
                onStartNew={startNew}
                onStartClone={startClone}
                onStartEdit={startEdit}
                onDelete={handleDelete}
              />
            ) : (
              <TemplateEditor
                editing={editing}
                isNew={isNew}
                metadataLocked={metadataLocked}
                existingTemplate={existingTemplate}
                saving={saving}
                loadingEditor={loadingEditor}
                globalFiles={globalFiles}
                saveError={saveError}
                onUpdateField={updateField}
                onAddDefaultFile={addDefaultFile}
                onUpdateDefaultFile={updateDefaultFile}
                onRemoveDefaultFile={removeDefaultFile}
                onUpdateGlobalFileContent={updateGlobalFileContent}
                onAddEnvDef={addEnvDef}
                onRemoveEnvDef={removeEnvDef}
                onUpdateEnvDef={updateEnvDef}
                onAddFileDef={addFileDef}
                onRemoveFileDef={removeFileDef}
                onUpdateFileDef={updateFileDef}
                onSave={handleSave}
                onCancel={resetEditor}
              />
            )}
          </div>
        </DialogContent>
      </Dialog>
      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(v) => {
          if (!v) setDeleteTarget(null);
        }}
        title="确认删除模板"
        description="确认删除此模板？此操作不可恢复。"
        confirmLabel="确认删除"
        onConfirm={confirmDelete}
      />
      <ConfirmDialog
        open={confirmGlobalSaveOpen}
        onOpenChange={setConfirmGlobalSaveOpen}
        title="确认保存真实全局配置"
        description="本次保存会直接修改节点上的真实全局配置文件。若这些文件在你编辑期间已被其他来源修改，系统会拒绝覆盖并要求重新加载。"
        confirmLabel="确认保存"
        variant="warning"
        onConfirm={() => {
          void performSave(true);
        }}
      />
    </>
  );
}
