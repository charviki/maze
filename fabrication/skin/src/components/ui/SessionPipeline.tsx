import { useState } from 'react';
import type { PipelineStep, PipelineStepType, PipelinePhase } from '../../types';
import { Plus, Trash2, Lock, Edit3 } from 'lucide-react';
import { Button } from './button';
import { Input } from './input';

interface SessionPipelineProps {
  steps: PipelineStep[];
  onChange: (steps: PipelineStep[]) => void;
  readOnly?: boolean;
}

// 步骤类型到中文标签的映射
const stepTypeLabel: Record<PipelineStepType, string> = {
  cd: '切换目录',
  env: '环境变量',
  file: '写入文件',
  command: '执行命令',
};

// 层级对应的样式
const phaseStyles: Record<
  PipelinePhase,
  { bg: string; border: string; badge: string; label: string }
> = {
  system: {
    bg: 'bg-gray-50 dark:bg-gray-900/30',
    border: 'border-gray-200 dark:border-gray-700',
    badge: 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400',
    label: '系统步骤',
  },
  template: {
    bg: 'bg-blue-50 dark:bg-blue-900/20',
    border: 'border-blue-200 dark:border-blue-700',
    badge: 'bg-blue-100 text-blue-600 dark:bg-blue-800 dark:text-blue-400',
    label: '模板步骤',
  },
  user: {
    bg: 'bg-green-50 dark:bg-green-900/20',
    border: 'border-green-200 dark:border-green-700',
    badge: 'bg-green-100 text-green-600 dark:bg-green-800 dark:text-green-400',
    label: '用户步骤',
  },
};

export function SessionPipeline({ steps, onChange, readOnly = false }: SessionPipelineProps) {
  const [newCommand, setNewCommand] = useState('');
  const [editingStepId, setEditingStepId] = useState<string | null>(null);
  const [editingValue, setEditingValue] = useState('');

  const sorted = [...steps].sort((a, b) => a.order - b.order);
  const systemSteps = sorted.filter((s) => s.phase === 'system');
  const templateSteps = sorted.filter((s) => s.phase === 'template');
  const userSteps = sorted.filter((s) => s.phase === 'user');

  // 添加用户自定义命令
  const addUserCommand = () => {
    if (!newCommand.trim()) return;
    const maxOrder = steps.reduce((max, s) => Math.max(max, s.order), -1);
    const newStep: PipelineStep = {
      id: `usr-cmd-${Date.now()}`,
      type: 'command',
      phase: 'user',
      order: maxOrder + 1,
      key: '',
      value: newCommand.trim(),
    };
    onChange([...steps, newStep]);
    setNewCommand('');
  };

  // 删除用户步骤
  const removeUserStep = (stepId: string) => {
    onChange(steps.filter((s) => s.id !== stepId));
  };

  // 开始编辑模板步骤
  const startEdit = (step: PipelineStep) => {
    if (readOnly || step.phase === 'system') return;
    setEditingStepId(step.id);
    setEditingValue(step.value);
  };

  // 保存编辑
  const saveEdit = () => {
    if (!editingStepId) return;
    onChange(steps.map((s) => (s.id === editingStepId ? { ...s, value: editingValue } : s)));
    setEditingStepId(null);
    setEditingValue('');
  };

  // 渲染单个步骤
  const renderStep = (step: PipelineStep) => {
    const style = phaseStyles[step.phase];
    const isEditing = editingStepId === step.id;
    const isSystem = step.phase === 'system';
    const canEdit = !readOnly && !isSystem;

    return (
      <div
        key={step.id}
        className={`flex items-center gap-2 px-3 py-2 rounded border ${style.bg} ${style.border} text-xs overflow-x-auto`}
      >
        {/* 步骤类型标签 */}
        <span className={`px-1.5 py-0.5 rounded text-[10px] font-medium shrink-0 ${style.badge}`}>
          {stepTypeLabel[step.type]}
        </span>

        {/* 步骤内容 */}
        {step.type === 'cd' && (
          <span className="font-mono flex-1 min-w-0 whitespace-nowrap">
            cd {step.key || step.value}
          </span>
        )}
        {step.type === 'env' && (
          <span className="font-mono flex-1 min-w-0 whitespace-nowrap">
            {step.key}={step.value}
          </span>
        )}
        {step.type === 'file' && (
          <span className="font-mono flex-1 min-w-0 whitespace-nowrap">{step.key}</span>
        )}
        {step.type === 'command' && isEditing ? (
          <div className="flex-1 flex items-center gap-1 min-w-0">
            <Input
              value={editingValue}
              onChange={(e) => {
                setEditingValue(e.target.value);
              }}
              onKeyDown={(e) => {
                if (e.key === 'Enter') saveEdit();
                if (e.key === 'Escape') setEditingStepId(null);
              }}
              className="h-6 text-xs font-mono min-w-0"
              autoFocus
            />
            <Button
              size="sm"
              variant="ghost"
              onClick={saveEdit}
              className="h-6 px-2 text-xs shrink-0"
            >
              确认
            </Button>
          </div>
        ) : step.type === 'command' ? (
          <span
            className={`font-mono flex-1 min-w-0 whitespace-nowrap ${canEdit ? 'cursor-pointer hover:text-primary' : ''}`}
            onClick={() => canEdit && startEdit(step)}
            title={canEdit ? '点击编辑' : undefined}
          >
            {step.value}
          </span>
        ) : null}

        {/* 操作按钮 */}
        {isSystem && !readOnly && <Lock className="w-3 h-3 text-gray-400 shrink-0" />}
        {canEdit && step.type === 'command' && !isEditing && (
          <button
            onClick={() => {
              startEdit(step);
            }}
            className="p-0.5 hover:bg-muted rounded shrink-0"
          >
            <Edit3 className="w-3 h-3 text-muted-foreground" />
          </button>
        )}
        {step.phase === 'user' && !readOnly && (
          <button
            onClick={() => {
              removeUserStep(step.id);
            }}
            className="p-0.5 hover:bg-red-100 rounded shrink-0"
          >
            <Trash2 className="w-3 h-3 text-red-400" />
          </button>
        )}
      </div>
    );
  };

  // 渲染一个层段
  const renderPhase = (label: string, phaseSteps: PipelineStep[], phase: PipelinePhase) => {
    if (phaseSteps.length === 0 && phase === 'user' && readOnly) return null;
    const style = phaseStyles[phase];
    return (
      <div className="space-y-1.5">
        <div className="flex items-center gap-2">
          <span
            className={`text-[10px] font-semibold uppercase tracking-wider ${style.badge} px-2 py-0.5 rounded`}
          >
            {label}
          </span>
          {phase === 'system' && <span className="text-[10px] text-muted-foreground">只读</span>}
        </div>
        {phaseSteps.map(renderStep)}
        {phaseSteps.length === 0 && (
          <div className="text-xs text-muted-foreground italic px-3 py-1">暂无{label}</div>
        )}
      </div>
    );
  };

  return (
    <div className="space-y-4">
      {/* 管线预览 */}
      {renderPhase('系统步骤', systemSteps, 'system')}
      {renderPhase('模板步骤', templateSteps, 'template')}
      {renderPhase('用户步骤', userSteps, 'user')}

      {/* 添加用户命令 */}
      {!readOnly && (
        <div className="flex items-center gap-2">
          <div className="flex-1">
            <Input
              value={newCommand}
              onChange={(e) => {
                setNewCommand(e.target.value);
              }}
              onKeyDown={(e) => {
                if (e.key === 'Enter') addUserCommand();
              }}
              placeholder="输入自定义 shell 命令..."
              className="h-8 text-xs font-mono"
            />
          </div>
          <Button
            size="sm"
            variant="outline"
            onClick={addUserCommand}
            disabled={!newCommand.trim()}
            className="h-8"
          >
            <Plus className="w-3.5 h-3.5 mr-1" />
            添加命令
          </Button>
        </div>
      )}
    </div>
  );
}
