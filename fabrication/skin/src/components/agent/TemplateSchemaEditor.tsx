import type { NormalizedTemplate, EnvDef, FileDef } from '../../types';
import { Plus, Trash2 } from 'lucide-react';
import { Button } from '../ui/button';
import { Input } from '../ui/input';

interface TemplateSchemaEditorProps {
  editing: NormalizedTemplate;
  isNew: boolean;
  metadataLocked: boolean;
  existingTemplate: boolean;
  onAddEnvDef: () => void;
  onRemoveEnvDef: (idx: number) => void;
  onUpdateEnvDef: (idx: number, field: keyof EnvDef, val: string | boolean) => void;
  onAddFileDef: () => void;
  onRemoveFileDef: (idx: number) => void;
  onUpdateFileDef: (idx: number, field: keyof FileDef, val: string | boolean) => void;
}

export function TemplateSchemaEditor({
  editing,
  isNew,
  metadataLocked,
  existingTemplate,
  onAddEnvDef,
  onRemoveEnvDef,
  onUpdateEnvDef,
  onAddFileDef,
  onRemoveFileDef,
  onUpdateFileDef,
}: TemplateSchemaEditorProps) {
  return (
    <>
      {/* Loop Schema: Env Defs */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <h3 className="text-xs text-muted-foreground font-medium">Loop 环境变量定义</h3>
          <Button size="sm" variant="ghost" onClick={onAddEnvDef}>
            <Plus className="w-3 h-3" />
          </Button>
        </div>
        {editing.sessionSchema.envDefs.map((d, i) => (
          <div key={i} className="border border-border rounded p-3 mb-3 space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium">Env #{i + 1}</span>
              <button
                onClick={() => {
                  onRemoveEnvDef(i);
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
                    onUpdateEnvDef(i, 'key', e.target.value);
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
                    onUpdateEnvDef(i, 'label', e.target.value);
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
                    onUpdateEnvDef(i, 'placeholder', e.target.value);
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
                      onUpdateEnvDef(i, 'required', e.target.checked);
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
                      onUpdateEnvDef(i, 'sensitive', e.target.checked);
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
            <Button size="sm" variant="ghost" onClick={onAddFileDef}>
              <Plus className="w-3 h-3" />
            </Button>
          )}
        </div>
        {editing.sessionSchema.fileDefs.map((d, i) => (
          <div key={i} className="border border-border rounded p-3 mb-3 space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium">File #{i + 1}</span>
              {isNew && (
                <button
                  onClick={() => {
                    onRemoveFileDef(i);
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
                    onUpdateFileDef(i, 'path', e.target.value);
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
                    onUpdateFileDef(i, 'label', e.target.value);
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
                      onUpdateFileDef(i, 'required', e.target.checked);
                    }}
                    className="rounded"
                    disabled={metadataLocked}
                  />
                  必填
                </label>
              </div>
              <textarea
                value={d.defaultContent}
                onChange={(e) => {
                  onUpdateFileDef(i, 'defaultContent', e.target.value);
                }}
                className="w-full h-20 bg-background text-foreground text-xs font-mono p-1 rounded border border-border"
                placeholder="默认文件内容..."
                disabled={metadataLocked}
              />
            </div>
          </div>
        ))}
      </div>
    </>
  );
}
