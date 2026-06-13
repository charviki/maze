import { useState } from 'react';
import { Plus, Trash2, Zap } from 'lucide-react';
import { Button } from './button';
import { Input } from './input';

/**
 * 管线步骤配置组件 — 在 CreateSessionWithTemplateDialog 中使用。
 * 允许用户添加 prompt 类型的管线步骤（向交互式 CLI 发送初始 prompt）。
 */

// ConfigItemType 尚未在生成的 API 类型中存在，暂用字符串字面量
export type PipelineConfigStepType = 'CONFIG_ITEM_TYPE_PROMPT';

export interface PipelineConfigStep {
  id: string;
  type: PipelineConfigStepType;
  value: string;
}

interface PipelineStepConfigProps {
  steps: PipelineConfigStep[];
  onChange: (steps: PipelineConfigStep[]) => void;
}

// prompt 步骤的展示样式
const promptStyle = {
  icon: Zap,
  color: 'text-[#00d4aa]',
  bgColor: 'bg-[#0a2a2a] border-[#00d4aa]/30',
  label: 'PROMPT',
};

export function PipelineStepConfig({ steps, onChange }: PipelineStepConfigProps) {
  const [newStepValue, setNewStepValue] = useState('');

  const addStep = () => {
    if (!newStepValue.trim()) return;
    const step: PipelineConfigStep = {
      // 不使用 crypto.randomUUID()：HTTP 非 localhost 域名（如 http://maze.local）
      // 不属于 Secure Context，浏览器会抛出 TypeError。
      id: `step-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`,
      type: 'CONFIG_ITEM_TYPE_PROMPT',
      value: newStepValue.trim(),
    };
    onChange([...steps, step]);
    setNewStepValue('');
  };

  const removeStep = (id: string) => {
    onChange(steps.filter((s) => s.id !== id));
  };

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
          ▸ Pipeline Steps
        </span>
        {steps.length > 0 && (
          <span className="text-[10px] bg-muted px-1.5 py-0.5 rounded font-mono">
            {steps.length}
          </span>
        )}
      </div>

      {steps.length > 0 && (
        <div className="space-y-1">
          {steps.map((step) => {
            const Icon = promptStyle.icon;
            return (
              <div
                key={step.id}
                className={`flex items-center gap-2 px-3 py-2 rounded border text-xs ${promptStyle.bgColor}`}
              >
                <Icon className={`w-3 h-3 shrink-0 ${promptStyle.color}`} />
                <span
                  className={`px-1.5 py-0.5 rounded text-[10px] font-medium shrink-0 ${promptStyle.color} bg-black/20`}
                >
                  {promptStyle.label}
                </span>
                <span className="font-mono flex-1 min-w-0 truncate text-foreground/90">
                  {step.value}
                </span>
                <button
                  onClick={() => removeStep(step.id)}
                  className="p-0.5 hover:bg-red-100 dark:hover:bg-red-900/30 rounded shrink-0"
                >
                  <Trash2 className="w-3 h-3 text-red-400" />
                </button>
              </div>
            );
          })}
        </div>
      )}

      <div className="flex items-center gap-2">
        <Input
          value={newStepValue}
          onChange={(e) => setNewStepValue(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') addStep();
          }}
          placeholder="输入 prompt 内容..."
          className="h-8 text-xs font-mono flex-1"
        />
        <Button
          size="sm"
          variant="outline"
          onClick={addStep}
          disabled={!newStepValue.trim()}
          className="h-8 shrink-0"
        >
          <Plus className="w-3.5 h-3.5 mr-1" />
          添加
        </Button>
      </div>
    </div>
  );
}
