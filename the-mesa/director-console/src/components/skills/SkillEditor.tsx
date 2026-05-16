import { useState, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  Button,
  Input,
} from '@maze/fabrication';
import type { V1Skill } from '@maze/fabrication';
import { Plus, X } from 'lucide-react';

interface SkillEditorProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  skill: V1Skill | null;
  onSubmit: (data: {
    name: string;
    description?: string;
    config?: Record<string, string>;
  }) => Promise<void>;
}

export function SkillEditor({ open, onOpenChange, skill, onSubmit }: SkillEditorProps) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [configEntries, setConfigEntries] = useState<{ key: string; value: string }[]>([]);
  const [submitting, setSubmitting] = useState(false);

  const isEditing = !!skill;

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    if (open) {
      if (skill) {
        setName(skill.name ?? '');
        setDescription(skill.description ?? '');
        const entries = skill.config
          ? Object.entries(skill.config).map(([key, value]) => ({ key, value }))
          : [];
        setConfigEntries(entries.length > 0 ? entries : [{ key: '', value: '' }]);
      } else {
        setName('');
        setDescription('');
        setConfigEntries([{ key: '', value: '' }]);
      }
      setSubmitting(false);
    }
  }, [open, skill]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const addConfigEntry = () => setConfigEntries((prev) => [...prev, { key: '', value: '' }]);

  const removeConfigEntry = (index: number) =>
    setConfigEntries((prev) => prev.filter((_, i) => i !== index));

  const updateConfigEntry = (index: number, field: 'key' | 'value', val: string) =>
    setConfigEntries((prev) =>
      prev.map((entry, i) => (i === index ? { ...entry, [field]: val } : entry)),
    );

  const handleSubmit = async () => {
    if (!name.trim()) return;
    setSubmitting(true);
    const config: Record<string, string> = {};
    for (const entry of configEntries) {
      if (entry.key.trim()) {
        config[entry.key.trim()] = entry.value;
      }
    }
    await onSubmit({ name: name.trim(), description, config });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-background border border-primary/20 max-w-md">
        <DialogHeader>
          <DialogTitle className="font-mono uppercase tracking-widest text-primary text-sm">
            {isEditing ? 'Edit Skill' : 'Create Skill'}
          </DialogTitle>
          <DialogDescription className="text-foreground/50 text-xs">
            {isEditing ? '修改 Skill 配置' : '创建新的 Skill'}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-2">
          <div>
            <label className="text-xs uppercase tracking-wider text-primary/60 mb-1 block">
              Name
            </label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              disabled={isEditing}
              placeholder="skill-name"
              className="font-mono text-sm"
            />
          </div>

          <div>
            <label className="text-xs uppercase tracking-wider text-primary/60 mb-1 block">
              Description
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Skill description..."
              rows={3}
              className="w-full bg-card/50 border border-primary/20 rounded px-3 py-2 text-sm text-foreground font-mono resize-none focus:outline-none focus:border-primary/60"
            />
          </div>

          <div>
            <div className="flex items-center justify-between mb-1">
              <label className="text-xs uppercase tracking-wider text-primary/60">Config</label>
              <Button
                variant="ghost"
                size="sm"
                className="h-5 text-[10px] text-primary/60 hover:text-primary"
                onClick={addConfigEntry}
              >
                <Plus className="w-3 h-3 mr-0.5" />
                ADD
              </Button>
            </div>
            <div className="space-y-2">
              {configEntries.map((entry, i) => (
                <div key={i} className="flex items-center gap-2">
                  <Input
                    value={entry.key}
                    onChange={(e) => updateConfigEntry(i, 'key', e.target.value)}
                    placeholder="key"
                    className="font-mono text-xs flex-1"
                  />
                  <Input
                    value={entry.value}
                    onChange={(e) => updateConfigEntry(i, 'value', e.target.value)}
                    placeholder="value"
                    className="font-mono text-xs flex-1"
                  />
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6 shrink-0 text-foreground/30 hover:text-foreground"
                    onClick={() => removeConfigEntry(i)}
                  >
                    <X className="w-3 h-3" />
                  </Button>
                </div>
              ))}
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="ghost"
            size="sm"
            className="text-foreground/60"
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button
            variant="default"
            size="sm"
            disabled={!name.trim() || submitting}
            onClick={handleSubmit}
          >
            {submitting ? 'Saving...' : isEditing ? 'Update' : 'Create'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
