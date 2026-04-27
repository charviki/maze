import { useEffect, useState } from 'react';
import type { SessionTemplate, ConfigItem, ConfigLayer, PipelineStep, LocalAgentConfig } from '../../types';
import type { IAgentApiClient } from '../../api';
import { Plus, Trash2, CheckCircle, Variable } from 'lucide-react';
import { maskEnvValue } from '../../utils/mask';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { SessionPipeline } from '../ui/SessionPipeline';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '../ui/dialog';
import { Skeleton } from '../ui/Skeleton';

interface CreateSessionWithTemplateDialogProps {
  apiClient: IAgentApiClient;
  nodeName: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: (sessionName: string) => void;
  onOpenTemplateManager: () => void;
}

export function CreateSessionWithTemplateDialog({ apiClient, nodeName: _nodeName, open, onOpenChange, onSuccess, onOpenTemplateManager }: CreateSessionWithTemplateDialogProps) {
  const [templates, setTemplates] = useState<SessionTemplate[]>([]);
  const [selectedTemplate, setSelectedTemplate] = useState<SessionTemplate | null>(null);
  const [createName, setCreateName] = useState('');
  const [relativeDir, setRelativeDir] = useState('');
  const [relativeDirTouched, setRelativeDirTouched] = useState(false);
  const [sessionEnvValues, setSessionEnvValues] = useState<Record<string, string>>({});
  const [sessionFileContents, setSessionFileContents] = useState<Record<string, string>>({});
  const [creating, setCreating] = useState(false);
  const [createError, setCreateError] = useState('');
  const [nodeConfig, setNodeConfig] = useState<LocalAgentConfig | null>(null);
  const [pipelineSteps, setPipelineSteps] = useState<PipelineStep[]>([]);
  const [customEnv, setCustomEnv] = useState<Record<string, string>>({});
  const [newEnvKey, setNewEnvKey] = useState('');
  const [newEnvValue, setNewEnvValue] = useState('');

  useEffect(() => {
    if (!open) return;
    apiClient.listTemplates().then(res => {
      if (res.status === 'ok' && res.data) setTemplates(res.data);
    }).catch(() => {});
    
    apiClient.getLocalConfig().then(res => {
      if (res.status === 'ok' && res.data) setNodeConfig(res.data);
    }).catch(() => {});
  }, [open, apiClient]);

  const baseWorkingDir = nodeConfig?.working_dir || '/home/agent';

  const buildAbsoluteWorkingDir = (baseDir: string, relativePath: string): string => {
    const trimmedBaseDir = baseDir.replace(/\/+$/, '') || '/home/agent';
    const trimmedRelativePath = relativePath.trim().replace(/^\/+/, '').replace(/\/+$/, '');
    return trimmedRelativePath ? `${trimmedBaseDir}/${trimmedRelativePath}` : trimmedBaseDir;
  };

  const currentWorkingDir = buildAbsoluteWorkingDir(baseWorkingDir, relativeDir);

  const resetDialogState = () => {
    setSelectedTemplate(null);
    setCreateName('');
    setRelativeDir('');
    setRelativeDirTouched(false);
    setSessionEnvValues({});
    setSessionFileContents({});
    setCustomEnv({});
    setPipelineSteps([]);
    setCreateError('');
  };

  useEffect(() => {
    if (!selectedTemplate || !nodeConfig) return;
    const schemaKeys = new Set(selectedTemplate.session_schema.env_defs.map(d => d.key));
    setSessionEnvValues(prev => {
      const next = { ...prev };
      for (const d of selectedTemplate.session_schema.env_defs) {
        if (nodeConfig.env?.[d.key] && !next[d.key]) {
          next[d.key] = nodeConfig.env[d.key];
        }
      }
      return next;
    });
    setCustomEnv(prev => {
      const next = { ...prev };
      if (nodeConfig.env) {
        for (const [key, value] of Object.entries(nodeConfig.env)) {
          if (!schemaKeys.has(key) && value && !next[key]) {
            next[key] = value;
          }
        }
      }
      return next;
    });
  }, [nodeConfig, selectedTemplate]);

  const toConfigItems = (configs: ConfigLayer): ConfigItem[] => {
    const items: ConfigItem[] = [];
    for (const [key, value] of Object.entries(configs.env)) {
      items.push({ type: 'env', key, value: String(value) });
    }
    for (const file of configs.files) {
      items.push({ type: 'file', key: file.path, value: String(file.content) });
    }
    return items;
  };

  const buildPipelineSteps = (workingDir: string, command: string, configs: ConfigItem[]): PipelineStep[] => {
    const steps: PipelineStep[] = [];
    let order = 0;

    if (workingDir) {
      steps.push({ id: 'sys-cd', type: 'cd', phase: 'system', order: order++, key: workingDir, value: '' });
    }
    for (const cfg of configs) {
      if (cfg.type === 'env') {
        steps.push({ id: `sys-env-${cfg.key}`, type: 'env', phase: 'system', order: order++, key: cfg.key, value: cfg.value });
      }
    }
    for (const cfg of configs) {
      if (cfg.type === 'file') {
        steps.push({ id: `sys-file-${cfg.key}`, type: 'file', phase: 'system', order: order++, key: cfg.key, value: cfg.value });
      }
    }
    if (command) {
      steps.push({ id: 'tpl-command', type: 'command', phase: 'template', order: order++, key: '', value: command });
    }

    return steps;
  };

  const selectTemplate = (tpl: SessionTemplate) => {
    setSelectedTemplate(tpl);
    const initialName = tpl.id + '-' + Date.now().toString(36);
    setCreateName(initialName);
    setRelativeDir(initialName);
    setRelativeDirTouched(false);
    const fileDefaults: Record<string, string> = {};
    // 新建 session 不主动生成项目级默认副本，不存在的项目文件按空内容处理。
    // 这里仅保留固定路径的空编辑态，避免把模板默认内容误写进工作目录。
    (tpl.session_schema.file_defs).forEach(d => { fileDefaults[d.path] = ''; });
    setSessionFileContents(fileDefaults);

    const envDefaults: Record<string, string> = {};
    const schemaKeys = new Set(tpl.session_schema.env_defs.map(d => d.key));
    (tpl.session_schema.env_defs).forEach(d => {
      envDefaults[d.key] = nodeConfig?.env?.[d.key] || '';
    });
    setSessionEnvValues(envDefaults);

    const extraEnv: Record<string, string> = {};
    if (nodeConfig?.env) {
      for (const [key, value] of Object.entries(nodeConfig.env)) {
        if (!schemaKeys.has(key) && value) {
          extraEnv[key] = String(value);
        }
      }
    }
    setCustomEnv(extraEnv);
    setCreateError('');
  };

  const validate = (): string | null => {
    if (!createName.trim()) return 'Loop 名称不能为空';
    if (!selectedTemplate) return '请选择模板';
    if (!relativeDir.trim()) return '相对目录不能为空';
    if (relativeDir.trim().startsWith('/')) return '请输入相对目录，不要以 / 开头';
    if (relativeDir.trim() === '.') return '相对目录不能是根目录';
    for (const def of selectedTemplate.session_schema.env_defs) {
      if (def.required && !sessionEnvValues[def.key]?.trim()) {
        return `${def.label || def.key} 为必填项`;
      }
    }
    return null;
  };

  const buildFinalConfigs = (): ConfigLayer => {
    const configs: ConfigLayer = { env: {}, files: [] };

    for (const [path, content] of Object.entries(sessionFileContents)) {
      if (content) {
        configs.files.push({ path, content: String(content) });
      }
    }

    for (const [key, value] of Object.entries(sessionEnvValues)) {
      if (value) configs.env[key] = String(value);
    }

    for (const [key, value] of Object.entries(customEnv)) {
      if (value) configs.env[key] = String(value);
    }

    return configs;
  };

  useEffect(() => {
    if (!selectedTemplate) return;
    if (!relativeDirTouched) {
      setRelativeDir(createName.trim());
    }
  }, [createName, relativeDirTouched, selectedTemplate]);

  useEffect(() => {
    if (!selectedTemplate) return;

    const configItems = toConfigItems(buildFinalConfigs());
    setPipelineSteps(prev => {
      const userCommands = prev
        .filter(step => step.phase === 'user' && step.type === 'command')
        .map(step => ({ ...step }));
      const baseSteps = buildPipelineSteps(currentWorkingDir, selectedTemplate.command || '', configItems);
      const userSteps = userCommands.map((step, index) => ({
        ...step,
        order: baseSteps.length + index,
      }));
      return [...baseSteps, ...userSteps];
    });
  }, [selectedTemplate, currentWorkingDir, sessionEnvValues, sessionFileContents, customEnv]);

  const getDefaultRestoreStrategy = (template: SessionTemplate): string => {
    if (template.command.includes('claude')) return 'auto';
    return 'manual';
  };

  const handleCreate = async () => {
    const err = validate();
    if (err) { setCreateError(err); return; }
    setCreateError('');
    setCreating(true);

    try {
      const configs = buildFinalConfigs();

      const configItems: ConfigItem[] = [];
      for (const [key, value] of Object.entries(configs.env)) {
        configItems.push({ type: 'env', key, value });
      }
      for (const file of configs.files) {
        configItems.push({ type: 'file', key: file.path, value: file.content });
      }

      const restoreStrategy = getDefaultRestoreStrategy(selectedTemplate!);

      await apiClient.createSession({
        name: createName.trim(),
        command: selectedTemplate!.command || undefined,
        working_dir: relativeDir.trim() || undefined,
        session_confs: configItems,
        restore_strategy: restoreStrategy,
        template_id: selectedTemplate!.id,
      });

      onSuccess(createName.trim());
      onOpenChange(false);
      resetDialogState();
    } catch {
      setCreateError('创建 Loop 失败');
    } finally {
      setCreating(false);
    }
  };

  const addCustomEnv = () => {
    if (!newEnvKey.trim()) return;
    setCustomEnv(prev => ({ ...prev, [newEnvKey.trim()]: newEnvValue }));
    setNewEnvKey('');
    setNewEnvValue('');
  };

  const removeCustomEnv = (key: string) => {
    setCustomEnv(prev => {
      const next = { ...prev };
      delete next[key];
      return next;
    });
  };

  const getEnvLabel = (key: string): string | null => {
    if (!selectedTemplate) return null;
    const def = selectedTemplate.session_schema.env_defs.find(d => d.key === key);
    return def ? def.label : null;
  };

  const handleRelativeDirChange = (value: string) => {
    setRelativeDirTouched(true);
    setRelativeDir(value.replace(/^\/+/, ''));
  };

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) { onOpenChange(false); resetDialogState(); } }}>
      <DialogContent className="max-w-4xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>创建新 Loop</DialogTitle>
          <DialogDescription>选择模板并配置参数来创建新的 Narrative Loop</DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {!selectedTemplate && (
            <div>
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium text-muted-foreground">选择模板</span>
                <button onClick={onOpenTemplateManager} className="text-xs text-primary hover:underline">管理模板</button>
              </div>
              <div className="grid grid-cols-3 gap-2">
                {templates.length === 0 && Array.from({ length: 3 }).map((_, i) => (
                  <div key={i} className="p-3 rounded-lg border border-border flex flex-col items-center gap-2">
                    <Skeleton className="h-8 w-8 rounded" />
                    <Skeleton className="h-3 w-20" />
                    <Skeleton className="h-2 w-28 bg-primary/5" />
                  </div>
                ))}
                {templates.map(tpl => (
                  <button
                    key={tpl.id}
                    onClick={() => selectTemplate(tpl)}
                    className="p-3 rounded-lg border text-center transition-all text-sm border-border hover:border-primary/50"
                  >
                    <div className="text-xl">{tpl.icon}</div>
                    <div className="font-medium truncate mt-1">{tpl.name}</div>
                    <div className="text-xs text-muted-foreground truncate">{tpl.description}</div>
                  </button>
                ))}
              </div>
            </div>
          )}

          {selectedTemplate && (
            <>
              <div className="flex items-center gap-2 p-3 bg-muted/50 rounded-lg">
                <span className="text-xl">{selectedTemplate.icon}</span>
                <div className="flex-1">
                  <div className="text-sm font-medium">{selectedTemplate.name}</div>
                  <div className="text-xs text-muted-foreground">{selectedTemplate.description}</div>
                </div>
                <Button size="sm" variant="ghost" onClick={() => { setSelectedTemplate(null); setCreateError(''); }}>
                  切换
                </Button>
              </div>

              {(nodeConfig?.env && Object.keys(nodeConfig.env).length > 0) || nodeConfig?.working_dir ? (
                <div className="bg-muted/50 rounded-lg p-3 space-y-1">
                  <span className="text-xs text-muted-foreground font-medium">上次使用的配置（已自动填入）</span>
                  {baseWorkingDir && (
                    <div className="flex items-center gap-1 text-xs min-w-0">
                      <CheckCircle className="w-3 h-3 text-green-500 shrink-0" />
                      <span className="shrink-0">基础工作目录: </span>
                      <span className="font-mono truncate">{baseWorkingDir}</span>
                    </div>
                  )}
                  {nodeConfig?.env && Object.entries(nodeConfig.env).map(([key, value]) => {
                    const label = getEnvLabel(key);
                    return (
                      <div key={key} className="flex items-center gap-1 text-xs min-w-0">
                        <CheckCircle className="w-3 h-3 text-green-500 shrink-0" />
                        <span className="shrink-0">{label ? `${label}` : ''}</span>
                        <span className="font-mono shrink-0">({key})</span>
                        <span className="text-muted-foreground truncate">= {maskEnvValue(key, String(value))}</span>
                      </div>
                    );
                  })}
                </div>
              ) : null}

              <div className="space-y-3">
                <span className="text-sm font-medium text-muted-foreground">Loop 配置</span>

                <div>
                  <label className="text-xs text-muted-foreground mb-1 block">Loop 名称</label>
                  <Input value={createName} onChange={e => setCreateName(e.target.value)} className="h-9 text-sm" placeholder="my-loop" />
                </div>

                <div>
                  <label className="text-xs text-muted-foreground mb-1 block">相对工作目录</label>
                  <Input value={relativeDir} onChange={e => handleRelativeDirChange(e.target.value)} className="h-9 text-sm font-mono" placeholder={createName || 'my-loop'} />
                  <div className="mt-1 text-xs text-muted-foreground">
                    完整目录: <span className="font-mono">{currentWorkingDir}</span>
                  </div>
                </div>

                {selectedTemplate.session_schema.env_defs.map(def => (
                  <div key={def.key}>
                    <label className="text-xs text-muted-foreground flex items-center gap-1 mb-1">
                      {def.label || def.key}
                      {def.required && <span className="text-red-500">*</span>}
                      {def.placeholder && <span className="text-muted-foreground/60">— {def.placeholder}</span>}
                    </label>
                    <Input
                      placeholder={def.placeholder || ''}
                      value={sessionEnvValues[def.key] || ''}
                      onChange={e => setSessionEnvValues(prev => ({ ...prev, [def.key]: e.target.value }))}
                      className="h-9 text-sm"
                    />
                  </div>
                ))}

                <div>
                  <h4 className="text-xs text-muted-foreground font-medium mb-2 flex items-center gap-1">
                    <Variable className="w-3 h-3" /> 额外环境变量
                  </h4>
                  {Object.keys(customEnv).length > 0 && (
                    <div className="space-y-2 mb-2">
                      {Object.entries(customEnv).map(([key, value]) => (
                        <div key={key} className="flex items-center gap-2">
                          <span className="text-sm font-mono w-40 shrink-0 truncate">{key}</span>
                          <div className="flex-1">
                            <Input value={String(value)} onChange={e => setCustomEnv(prev => ({ ...prev, [key]: e.target.value }))} className="text-sm h-8" placeholder="value" />
                          </div>
                          <button onClick={() => removeCustomEnv(key)} className="p-1 hover:bg-muted rounded text-red-500">
                            <Trash2 className="w-3.5 h-3.5" />
                          </button>
                        </div>
                      ))}
                    </div>
                  )}
                  <div className="flex items-center gap-2">
                    <div className="w-40">
                      <Input value={newEnvKey} onChange={e => setNewEnvKey(e.target.value)} onKeyDown={e => { if (e.key === 'Enter') addCustomEnv(); }} placeholder="变量名" className="text-sm h-8" />
                    </div>
                    <div className="flex-1">
                      <Input value={newEnvValue} onChange={e => setNewEnvValue(e.target.value)} onKeyDown={e => { if (e.key === 'Enter') addCustomEnv(); }} placeholder="值" className="text-sm h-8" />
                    </div>
                    <Button size="sm" variant="outline" onClick={addCustomEnv} disabled={!newEnvKey.trim()} className="h-8">
                      <Plus className="w-3.5 h-3.5" />
                    </Button>
                  </div>
                </div>

                {selectedTemplate.session_schema.file_defs.map(def => (
                  <div key={def.path}>
                    <label className="text-xs text-muted-foreground flex items-center gap-1 mb-1">
                      {def.label || def.path}
                      {def.required && <span className="text-red-500">*</span>}
                    </label>
                    <textarea
                      value={sessionFileContents[def.path] || ''}
                      onChange={e => setSessionFileContents(prev => ({ ...prev, [def.path]: e.target.value }))}
                      className="w-full h-40 bg-background text-foreground text-xs font-mono p-2 rounded border border-border"
                      placeholder={def.default_content || '留空表示创建后该文件暂不存在'}
                    />
                    <div className="mt-1 text-[11px] text-muted-foreground">
                      路径固定为 <span className="font-mono">{def.path}</span>；留空时按不存在处理。
                    </div>
                  </div>
                ))}

                <div>
                  <span className="text-sm font-medium text-muted-foreground">命令管线预览</span>
                  <div className="mt-2">
                    <SessionPipeline steps={pipelineSteps} onChange={setPipelineSteps} readOnly />
                  </div>
                </div>
              </div>

              {createError && (
                <div className="text-sm text-red-500 bg-red-50 dark:bg-red-950/30 px-3 py-2 rounded">
                  {createError}
                </div>
              )}

              <div className="flex justify-end gap-2">
                <Button variant="ghost" size="sm" onClick={() => { onOpenChange(false); setSelectedTemplate(null); setCreateError(''); }}>取消</Button>
                <Button size="sm" onClick={handleCreate} disabled={!createName.trim() || creating}>
                  {creating ? '创建中...' : '创建'}
                </Button>
              </div>
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
