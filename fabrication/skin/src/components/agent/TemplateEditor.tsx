import type {
  NormalizedTemplate,
  SessionTemplate,
  ConfigFileSnapshot,
  EnvDef,
  FileDef,
} from '../../types';
import { Plus, Trash2, Save } from 'lucide-react';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { TemplateSchemaEditor } from './TemplateSchemaEditor';

export interface EditableConfigFile extends ConfigFileSnapshot {
  content: string;
  original_content: string;
  baseHash: string;
}

interface TemplateEditorProps {
  editing: NormalizedTemplate;
  isNew: boolean;
  metadataLocked: boolean;
  existingTemplate: boolean;
  saving: boolean;
  loadingEditor: boolean;
  globalFiles: EditableConfigFile[];
  saveError: string;
  onUpdateField: <K extends keyof SessionTemplate>(key: K, val: SessionTemplate[K]) => void;
  onAddDefaultFile: () => void;
  onUpdateDefaultFile: (idx: number, field: 'path' | 'content', val: string) => void;
  onRemoveDefaultFile: (idx: number) => void;
  onUpdateGlobalFileContent: (path: string, content: string) => void;
  onAddEnvDef: () => void;
  onRemoveEnvDef: (idx: number) => void;
  onUpdateEnvDef: (idx: number, field: keyof EnvDef, val: string | boolean) => void;
  onAddFileDef: () => void;
  onRemoveFileDef: (idx: number) => void;
  onUpdateFileDef: (idx: number, field: keyof FileDef, val: string | boolean) => void;
  onSave: () => void;
  onCancel: () => void;
}

export function TemplateEditor({
  editing,
  isNew,
  metadataLocked,
  existingTemplate,
  saving,
  loadingEditor,
  globalFiles,
  saveError,
  onUpdateField,
  onAddDefaultFile,
  onUpdateDefaultFile,
  onRemoveDefaultFile,
  onUpdateGlobalFileContent,
  onAddEnvDef,
  onRemoveEnvDef,
  onUpdateEnvDef,
  onAddFileDef,
  onRemoveFileDef,
  onUpdateFileDef,
  onSave,
  onCancel,
}: TemplateEditorProps) {
  return (
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
                  onUpdateField('id', e.target.value);
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
                onUpdateField('name', e.target.value);
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
                onUpdateField('command', e.target.value);
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
                onUpdateField('icon', e.target.value);
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
                onUpdateField('description', e.target.value);
              }}
              placeholder="简短描述这个模板"
              disabled={metadataLocked}
            />
          </div>
        </div>
      </div>

      {/* 模板全局配置文件 */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <h3 className="text-xs text-muted-foreground font-medium">模板全局配置文件</h3>
          {isNew && (
            <Button size="sm" variant="ghost" onClick={onAddDefaultFile}>
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
                        onUpdateDefaultFile(i, 'path', e.target.value);
                      }}
                      className="text-sm"
                      placeholder="~/.claude/settings.json"
                    />
                  </div>
                  <button
                    onClick={() => {
                      onRemoveDefaultFile(i);
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
                      onUpdateDefaultFile(i, 'content', e.target.value);
                    }}
                    className="w-full h-24 bg-background text-foreground text-xs font-mono p-2 rounded border border-border"
                    placeholder="默认文件内容..."
                  />
                </div>
              </div>
            ))
          : globalFiles.map((f) => (
              <div key={f.path} className="border border-border rounded p-3 mb-2 space-y-2">
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
                      onUpdateGlobalFileContent(f.path ?? '', e.target.value);
                    }}
                    className="w-full h-32 bg-background text-foreground text-xs font-mono p-2 rounded border border-border"
                  />
                </div>
              </div>
            ))}
      </div>

      {/* Schema 编辑 */}
      <TemplateSchemaEditor
        editing={editing}
        isNew={isNew}
        metadataLocked={metadataLocked}
        existingTemplate={existingTemplate}
        onAddEnvDef={onAddEnvDef}
        onRemoveEnvDef={onRemoveEnvDef}
        onUpdateEnvDef={onUpdateEnvDef}
        onAddFileDef={onAddFileDef}
        onRemoveFileDef={onRemoveFileDef}
        onUpdateFileDef={onUpdateFileDef}
      />

      <div className="flex justify-end gap-2 pt-2">
        <Button variant="outline" onClick={onCancel}>
          取消
        </Button>
        <Button onClick={onSave} disabled={saving || loadingEditor}>
          <Save className="w-4 h-4 mr-1" /> {saving ? '保存中...' : '保存'}
        </Button>
      </div>
    </div>
  );
}
