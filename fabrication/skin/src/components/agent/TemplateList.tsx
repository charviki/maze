import type { NormalizedTemplate } from '../../types';
import { Plus, Trash2, Edit2, Copy, Lock } from 'lucide-react';
import { Button } from '../ui/button';

interface TemplateListProps {
  builtinTemplates: NormalizedTemplate[];
  userTemplates: NormalizedTemplate[];
  onStartNew: () => void;
  onStartClone: (tpl: NormalizedTemplate) => void;
  onStartEdit: (tpl: NormalizedTemplate) => void;
  onDelete: (id: string) => void;
}

export function TemplateList({
  builtinTemplates,
  userTemplates,
  onStartNew,
  onStartClone,
  onStartEdit,
  onDelete,
}: TemplateListProps) {
  return (
    <>
      <div>
        <div className="flex items-center justify-between mb-2">
          <h3 className="font-medium text-sm text-muted-foreground flex items-center gap-1">
            <Lock className="w-4 h-4" /> 内置模板
          </h3>
          <Button size="sm" variant="outline" onClick={onStartNew}>
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
                  onStartClone(tpl);
                }}
              >
                <Copy className="w-3 h-3" />
              </Button>
              <Button
                size="sm"
                variant="ghost"
                title="编辑模板"
                onClick={() => void onStartEdit(tpl)}
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
                {tpl.icon} {tpl.name} <span className="text-muted-foreground">- {tpl.command}</span>
              </span>
              <div className="flex gap-1 shrink-0">
                <Button
                  size="sm"
                  variant="ghost"
                  title="基于此模板创建副本"
                  onClick={() => {
                    onStartClone(tpl);
                  }}
                >
                  <Copy className="w-3 h-3" />
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  title="编辑模板"
                  onClick={() => void onStartEdit(tpl)}
                >
                  <Edit2 className="w-3 h-3" />
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  className="text-red-500"
                  onClick={() => {
                    onDelete(tpl.id ?? '');
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
  );
}
