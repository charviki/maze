import { useState, useEffect, useCallback } from 'react';
import type {
  SessionTemplate,
  EnvDef,
  FileDef,
  ConfigLayer,
  ConfigFileSnapshot,
  ConfigFileUpdate,
} from '../../types';
import type { IAgentApiClient } from '../../api';
import { Plus, Trash2, Edit2, Save, Lock, Copy } from 'lucide-react';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { ConfirmDialog } from '../ui/ConfirmDialog';
import { useToast } from '../ui/Toast';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '../ui/dialog';

interface TemplateManagerProps {
  open: boolean;
  onClose: () => void;
  apiClient: IAgentApiClient;
}

interface EditableConfigFile extends ConfigFileSnapshot {
  content: string;
  original_content: string;
  base_hash: string;
}

const emptyConfigLayer = (): ConfigLayer => ({ env: {}, files: [] });

const emptyTemplate = (): SessionTemplate => ({
  id: '',
  name: '',
  command: '',
  description: '',
  icon: '📦',
  builtin: false,
  defaults: emptyConfigLayer(),
  session_schema: { env_defs: [], file_defs: [] },
});

export function TemplateManager({ open, onClose, apiClient }: TemplateManagerProps) {
  const { showToast } = useToast();
  const [templates, setTemplates] = useState<SessionTemplate[]>([]);
  const [editing, setEditing] = useState<SessionTemplate | null>(null);
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

  // fetch-in-effect: Dialog 打开时从 API 加载模板列表并写入 state，属于合法的数据同步模式。
  // React Compiler 的 set-state-in-effect 规则对 fetch-in-effect 场景存在已知误报。
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

  const startClone = (tpl: SessionTemplate) => {
    setIsNew(true);
    setEditing(
      cloneTemplate({
        ...tpl,
        id: tpl.id + '-copy',
        name: tpl.name + ' (Copy)',
        builtin: false,
      }),
    );
    setGlobalFiles([]);
    setSaveError('');
  };

  const startEdit = async (tpl: SessionTemplate) => {
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
      configRes.data.files.map((file) => ({
        ...file,
        content: file.content,
        original_content: file.content,
        base_hash: file.hash,
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

      const updateRes = await apiClient.updateTemplate(editing.id, editing);
      if (updateRes.status !== 'ok') {
        throw new Error(updateRes.message || '保存模板失败');
      }

      if (allowRealGlobalWrite && hasConfigChanges) {
        const configReq: ConfigFileUpdate[] = globalFiles.map((file) => ({
          path: file.path,
          content: file.content,
          base_hash: file.base_hash,
        }));
        const configRes = await apiClient.updateTemplateConfig(editing.id, { files: configReq });
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
          configRes.data.files.map((file) => ({
            ...file,
            content: file.content,
            original_content: file.content,
            base_hash: file.hash,
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

  const updateField = <K extends keyof SessionTemplate>(key: K, val: SessionTemplate[K]) => {
    if (!editing) return;
    setEditing((prev) => (prev ? { ...prev, [key]: val } : null));
  };

  // Defaults files 操作
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

  // EnvDef 操作
  const addEnvDef = () => {
    if (!editing) return;
    setEditing((prev) =>
      prev
        ? {
            ...prev,
            session_schema: {
              ...prev.session_schema,
              env_defs: [
                ...prev.session_schema.env_defs,
                { key: '', label: '', required: false, placeholder: '', sensitive: false },
              ],
            },
          }
        : null,
    );
  };

  const updateEnvDef = (idx: number, field: keyof EnvDef, val: string | boolean) => {
    if (!editing) return;
    const defs = [...editing.session_schema.env_defs];
    defs[idx] = { ...defs[idx], [field]: val };
    setEditing((prev) =>
      prev ? { ...prev, session_schema: { ...prev.session_schema, env_defs: defs } } : null,
    );
  };

  const removeEnvDef = (idx: number) => {
    if (!editing) return;
    setEditing((prev) =>
      prev
        ? {
            ...prev,
            session_schema: {
              ...prev.session_schema,
              env_defs: prev.session_schema.env_defs.filter((_, i) => i !== idx),
            },
          }
        : null,
    );
  };

  // FileDef 操作
  const addFileDef = () => {
    if (!editing) return;
    setEditing((prev) =>
      prev
        ? {
            ...prev,
            session_schema: {
              ...prev.session_schema,
              file_defs: [
                ...prev.session_schema.file_defs,
                { path: '', label: '', required: false, default_content: '' },
              ],
            },
          }
        : null,
    );
  };

  const updateFileDef = (idx: number, field: keyof FileDef, val: string | boolean) => {
    if (!editing) return;
    const defs = [...editing.session_schema.file_defs];
    defs[idx] = { ...defs[idx], [field]: val };
    setEditing((prev) =>
      prev ? { ...prev, session_schema: { ...prev.session_schema, file_defs: defs } } : null,
    );
  };

  const removeFileDef = (idx: number) => {
    if (!editing) return;
    setEditing((prev) =>
      prev
        ? {
            ...prev,
            session_schema: {
              ...prev.session_schema,
              file_defs: prev.session_schema.file_defs.filter((_, i) => i !== idx),
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

  const hasConfigChanges = isNew
    ? false
    : globalFiles.some((file) => file.content !== file.original_content);

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
              <>
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="font-medium text-sm text-muted-foreground flex items-center gap-1">
                      <Lock className="w-4 h-4" /> 内置模板
                    </h3>
                    <Button size="sm" variant="outline" onClick={startNew}>
                      <Plus className="w-4 h-4 mr-1" /> 新建
                    </Button>
                  </div>
                  {builtinTemplates.map((tpl) => (
                    <div
                      key={tpl.id}
                      className="flex items-center justify-between py-1.5 px-2 hover:bg-muted rounded text-sm gap-2"
                    >
                      <span className="truncate min-w-0">
                        {tpl.icon} {tpl.name}{' '}
                        <span className="text-muted-foreground">- {tpl.command || '(empty)'}</span>
                      </span>
                      <div className="flex gap-1 shrink-0">
                        <Button
                          size="sm"
                          variant="ghost"
                          title="基于此模板创建副本"
                          onClick={() => {
                            startClone(tpl);
                          }}
                        >
                          <Copy className="w-3 h-3" />
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          title="编辑模板"
                          onClick={() => void startEdit(tpl)}
                        >
                          <Edit2 className="w-3 h-3" />
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>

                {userTemplates.length > 0 && (
                  <div>
                    <h3 className="font-medium text-sm text-muted-foreground mb-2">自定义模板</h3>
                    {userTemplates.map((tpl) => (
                      <div
                        key={tpl.id}
                        className="flex items-center justify-between py-1.5 px-2 hover:bg-muted rounded text-sm gap-2"
                      >
                        <span className="truncate min-w-0">
                          {tpl.icon} {tpl.name}{' '}
                          <span className="text-muted-foreground">- {tpl.command}</span>
                        </span>
                        <div className="flex gap-1 shrink-0">
                          <Button
                            size="sm"
                            variant="ghost"
                            title="基于此模板创建副本"
                            onClick={() => {
                              startClone(tpl);
                            }}
                          >
                            <Copy className="w-3 h-3" />
                          </Button>
                          <Button
                            size="sm"
                            variant="ghost"
                            title="编辑模板"
                            onClick={() => void startEdit(tpl)}
                          >
                            <Edit2 className="w-3 h-3" />
                          </Button>
                          <Button
                            size="sm"
                            variant="ghost"
                            className="text-red-500"
                            onClick={() => {
                              handleDelete(tpl.id);
                            }}
                          >
                            <Trash2 className="w-3 h-3" />
                          </Button>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </>
            ) : (
              <div className="space-y-5">
                {isNew && (
                  <div className="text-xs text-muted-foreground bg-muted px-3 py-2 rounded">
                    正在创建新模板。固定路径集合会作为模板定义保存，后续在配置页中只允许编辑内容。
                  </div>
                )}
                {editing.builtin && !isNew && (
                  <div className="text-xs text-yellow-600 bg-yellow-50 dark:bg-yellow-950/30 px-3 py-2 rounded">
                    内置模板的固定路径由系统维护。这里读取并编辑的是节点上的真实全局配置内容。
                  </div>
                )}
                {saveError && (
                  <div className="text-xs text-destructive bg-destructive/10 border border-destructive/20 px-3 py-2 rounded">
                    {saveError}
                  </div>
                )}

                {/* 基础信息 */}
                <div>
                  <h3 className="text-xs text-muted-foreground font-medium mb-2">基础信息</h3>
                  <div className="grid grid-cols-2 gap-3">
                    {isNew && (
                      <div>
                        <label className="text-xs text-muted-foreground mb-1 block">模板 ID</label>
                        <Input
                          value={editing.id}
                          onChange={(e) => {
                            updateField('id', e.target.value);
                          }}
                          placeholder="例如 my-agent"
                        />
                      </div>
                    )}
                    <div>
                      <label className="text-xs text-muted-foreground mb-1 block">模板名称</label>
                      <Input
                        value={editing.name}
                        onChange={(e) => {
                          updateField('name', e.target.value);
                        }}
                        placeholder="例如 Claude Code"
                        disabled={metadataLocked}
                      />
                    </div>
                    <div>
                      <label className="text-xs text-muted-foreground mb-1 block">启动命令</label>
                      <Input
                        value={editing.command}
                        onChange={(e) => {
                          updateField('command', e.target.value);
                        }}
                        placeholder="例如 claude"
                        disabled={metadataLocked}
                      />
                    </div>
                    <div>
                      <label className="text-xs text-muted-foreground mb-1 block">图标</label>
                      <Input
                        value={editing.icon}
                        onChange={(e) => {
                          updateField('icon', e.target.value);
                        }}
                        placeholder="🤖"
                        className="w-20"
                        disabled={metadataLocked}
                      />
                    </div>
                    <div className="col-span-2">
                      <label className="text-xs text-muted-foreground mb-1 block">描述</label>
                      <Input
                        value={editing.description}
                        onChange={(e) => {
                          updateField('description', e.target.value);
                        }}
                        placeholder="简短描述这个模板"
                        disabled={metadataLocked}
                      />
                    </div>
                  </div>
                </div>

                {/* 模板默认配置 */}
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="text-xs text-muted-foreground font-medium">模板全局配置文件</h3>
                    {isNew && (
                      <Button size="sm" variant="ghost" onClick={addDefaultFile}>
                        <Plus className="w-3 h-3" />
                      </Button>
                    )}
                  </div>
                  {isNew
                    ? editing.defaults.files.map((f, i) => (
                        <div key={i} className="border border-border rounded p-2 mb-2 space-y-1">
                          <div className="flex items-center gap-2">
                            <div className="flex-1">
                              <label className="text-xs text-muted-foreground">文件路径</label>
                              <Input
                                value={f.path}
                                onChange={(e) => {
                                  updateDefaultFile(i, 'path', e.target.value);
                                }}
                                className="text-sm"
                                placeholder="~/.claude/settings.json"
                              />
                            </div>
                            <button
                              onClick={() => {
                                removeDefaultFile(i);
                              }}
                              className="text-red-500 mt-4"
                            >
                              <Trash2 className="w-3 h-3" />
                            </button>
                          </div>
                          <div>
                            <label className="text-xs text-muted-foreground">默认内容</label>
                            <textarea
                              value={f.content}
                              onChange={(e) => {
                                updateDefaultFile(i, 'content', e.target.value);
                              }}
                              className="w-full h-24 bg-background text-foreground text-xs font-mono p-2 rounded border border-border"
                              placeholder="默认文件内容..."
                            />
                          </div>
                        </div>
                      ))
                    : globalFiles.map((f) => (
                        <div
                          key={f.path}
                          className="border border-border rounded p-3 mb-2 space-y-2"
                        >
                          <div className="flex items-center justify-between gap-2">
                            <div className="flex-1">
                              <label className="text-xs text-muted-foreground">固定路径</label>
                              <Input value={f.path} readOnly className="text-sm font-mono" />
                            </div>
                            <div className="mt-5 text-[11px] text-muted-foreground shrink-0">
                              {f.exists ? '真实文件已存在' : '真实文件不存在，当前按空内容处理'}
                            </div>
                          </div>
                          <div>
                            <label className="text-xs text-muted-foreground">真实内容</label>
                            <textarea
                              value={f.content}
                              onChange={(e) => {
                                updateGlobalFileContent(f.path, e.target.value);
                              }}
                              className="w-full h-32 bg-background text-foreground text-xs font-mono p-2 rounded border border-border"
                            />
                          </div>
                        </div>
                      ))}
                </div>

                {/* Loop Schema: Env Defs */}
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="text-xs text-muted-foreground font-medium">Loop 环境变量定义</h3>
                    <Button size="sm" variant="ghost" onClick={addEnvDef}>
                      <Plus className="w-3 h-3" />
                    </Button>
                  </div>
                  {editing.session_schema.env_defs.map((d, i) => (
                    <div key={i} className="border border-border rounded p-3 mb-3 space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="text-xs font-medium">Env #{i + 1}</span>
                        <button
                          onClick={() => {
                            removeEnvDef(i);
                          }}
                          className="text-red-500"
                        >
                          <Trash2 className="w-3 h-3" />
                        </button>
                      </div>
                      <div className="grid grid-cols-2 gap-2">
                        <div>
                          <label className="text-xs text-muted-foreground">Key</label>
                          <Input
                            value={d.key}
                            onChange={(e) => {
                              updateEnvDef(i, 'key', e.target.value);
                            }}
                            className="text-sm"
                            placeholder="ANTHROPIC_API_KEY"
                          />
                        </div>
                        <div>
                          <label className="text-xs text-muted-foreground">显示名称</label>
                          <Input
                            value={d.label}
                            onChange={(e) => {
                              updateEnvDef(i, 'label', e.target.value);
                            }}
                            className="text-sm"
                            placeholder="API Key"
                          />
                        </div>
                      </div>
                      <div className="grid grid-cols-2 gap-2">
                        <div>
                          <label className="text-xs text-muted-foreground">占位提示</label>
                          <Input
                            value={d.placeholder}
                            onChange={(e) => {
                              updateEnvDef(i, 'placeholder', e.target.value);
                            }}
                            className="text-sm"
                            placeholder="sk-ant-..."
                          />
                        </div>
                        <div className="flex items-end gap-3 pb-1">
                          <label className="flex items-center gap-1.5 text-xs text-muted-foreground cursor-pointer">
                            <input
                              type="checkbox"
                              checked={d.required}
                              onChange={(e) => {
                                updateEnvDef(i, 'required', e.target.checked);
                              }}
                              className="rounded"
                            />
                            必填
                          </label>
                          <label className="flex items-center gap-1.5 text-xs text-muted-foreground cursor-pointer">
                            <input
                              type="checkbox"
                              checked={d.sensitive}
                              onChange={(e) => {
                                updateEnvDef(i, 'sensitive', e.target.checked);
                              }}
                              className="rounded"
                            />
                            敏感
                          </label>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>

                {/* Loop Schema: File Defs */}
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="text-xs text-muted-foreground font-medium">Loop 配置文件定义</h3>
                    {isNew && (
                      <Button size="sm" variant="ghost" onClick={addFileDef}>
                        <Plus className="w-3 h-3" />
                      </Button>
                    )}
                  </div>
                  {editing.session_schema.file_defs.map((d, i) => (
                    <div key={i} className="border border-border rounded p-3 mb-3 space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="text-xs font-medium">File #{i + 1}</span>
                        {isNew && (
                          <button
                            onClick={() => {
                              removeFileDef(i);
                            }}
                            className="text-red-500"
                          >
                            <Trash2 className="w-3 h-3" />
                          </button>
                        )}
                      </div>
                      <div className="grid grid-cols-2 gap-2">
                        <div>
                          <label className="text-xs text-muted-foreground">文件路径</label>
                          <Input
                            value={d.path}
                            onChange={(e) => {
                              updateFileDef(i, 'path', e.target.value);
                            }}
                            className="text-sm"
                            placeholder="CLAUDE.md"
                            readOnly={existingTemplate}
                          />
                        </div>
                        <div>
                          <label className="text-xs text-muted-foreground">显示名称</label>
                          <Input
                            value={d.label}
                            onChange={(e) => {
                              updateFileDef(i, 'label', e.target.value);
                            }}
                            className="text-sm"
                            placeholder="项目记忆文件"
                            disabled={metadataLocked}
                          />
                        </div>
                      </div>
                      <div>
                        <div className="flex items-center gap-3 mb-1">
                          <label className="text-xs text-muted-foreground">默认内容</label>
                          <label className="flex items-center gap-1.5 text-xs text-muted-foreground cursor-pointer">
                            <input
                              type="checkbox"
                              checked={d.required}
                              onChange={(e) => {
                                updateFileDef(i, 'required', e.target.checked);
                              }}
                              className="rounded"
                              disabled={metadataLocked}
                            />
                            必填
                          </label>
                        </div>
                        <textarea
                          value={d.default_content}
                          onChange={(e) => {
                            updateFileDef(i, 'default_content', e.target.value);
                          }}
                          className="w-full h-20 bg-background text-foreground text-xs font-mono p-1 rounded border border-border"
                          placeholder="默认文件内容..."
                          disabled={metadataLocked}
                        />
                      </div>
                    </div>
                  ))}
                </div>

                <div className="flex justify-end gap-2 pt-2">
                  <Button variant="outline" onClick={resetEditor}>
                    取消
                  </Button>
                  <Button onClick={handleSave} disabled={saving || loadingEditor}>
                    <Save className="w-4 h-4 mr-1" /> {saving ? '保存中...' : '保存'}
                  </Button>
                </div>
              </div>
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

function cloneTemplate(tpl: SessionTemplate): SessionTemplate {
  return {
    ...tpl,
    defaults: {
      env: { ...tpl.defaults.env },
      files: tpl.defaults.files.map((file) => ({ ...file })),
    },
    session_schema: {
      env_defs: tpl.session_schema.env_defs.map((def) => ({ ...def })),
      file_defs: tpl.session_schema.file_defs.map((def) => ({ ...def })),
    },
  };
}
