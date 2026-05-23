import { useState, useEffect } from 'react';
import { Panel, Button, Input, DecryptText } from '@maze/fabrication';
import type { V1Skill } from '@maze/fabrication';
import { Sparkles, Plus, X } from 'lucide-react';

interface SkillDetailPanelProps {
  skill: V1Skill | null;
  isCreating: boolean;
  onSubmit: (data: {
    name: string;
    description?: string;
    config?: Record<string, string>;
  }) => Promise<void>;
  onCancel: () => void;
}

export function SkillDetailPanel({ skill, isCreating, onSubmit, onCancel }: SkillDetailPanelProps) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [configEntries, setConfigEntries] = useState<{ key: string; value: string }[]>([]);
  const [submitting, setSubmitting] = useState(false);

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    if (skill) {
      setName(skill.name ?? '');
      setDescription(skill.description ?? '');
      const entries = skill.config
        ? Object.entries(skill.config).map(([key, value]) => ({ key, value }))
        : [];
      setConfigEntries(entries.length > 0 ? entries : [{ key: '', value: '' }]);
    } else if (isCreating) {
      setName('');
      setDescription('');
      setConfigEntries([{ key: '', value: '' }]);
    }
    setSubmitting(false);
  }, [skill, isCreating]);
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
    try {
      await onSubmit({ name: name.trim(), description, config });
    } catch {
      setSubmitting(false);
    }
  };

  if (!skill && !isCreating) {
    return (
      <div className="flex-1 flex items-center justify-center text-primary/40 uppercase tracking-widest text-xs">
        <div className="text-center space-y-4">
          <Sparkles className="w-16 h-16 mx-auto opacity-20" />
          <DecryptText text="SELECT SKILL TO INSPECT" />
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 min-w-0 flex flex-col bg-background relative z-10 overflow-hidden">
      <Panel className="flex flex-col h-full relative m-2" cornerSize={16}>
        <div className="pb-4 border-b border-primary/20">
          <div className="flex items-center gap-2 text-primary">
            <Sparkles className="w-4 h-4" />
            <h2 className="text-xs font-bold uppercase tracking-widest">
              <DecryptText text={isCreating ? 'FABRICATE NEW SKILL' : 'SKILL INSPECTION'} />
            </h2>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto pt-4 space-y-4 px-1">
          <div>
            <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono mb-1 block">
              SKILL DESIGNATION
            </label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              disabled={!!skill}
              placeholder="skill-name"
              className="font-mono text-sm rounded-none border-primary/20 bg-card/50"
            />
          </div>

          <div>
            <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono mb-1 block">
              DESCRIPTION
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Skill description..."
              rows={3}
              className="w-full bg-card/50 border border-primary/20 rounded-none px-3 py-2 text-sm text-foreground font-mono resize-none focus:outline-none focus:border-primary/60"
            />
          </div>

          <div>
            <div className="flex items-center justify-between mb-1">
              <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono">
                CONFIG PARAMETERS
              </label>
              <Button
                variant="ghost"
                size="sm"
                className="h-5 text-[10px] text-primary/60 hover:text-primary uppercase tracking-widest"
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
                    className="font-mono text-xs flex-1 rounded-none border-primary/20 bg-card/50"
                  />
                  <Input
                    value={entry.value}
                    onChange={(e) => updateConfigEntry(i, 'value', e.target.value)}
                    placeholder="value"
                    className="font-mono text-xs flex-1 rounded-none border-primary/20 bg-card/50"
                  />
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6 shrink-0 text-foreground/30 hover:text-foreground rounded-none"
                    onClick={() => removeConfigEntry(i)}
                  >
                    <X className="w-3 h-3" />
                  </Button>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="pt-4 border-t border-primary/20 flex justify-end gap-2">
          <Button
            variant="ghost"
            onClick={onCancel}
            className="font-mono uppercase tracking-widest text-xs rounded-none"
          >
            CANCEL
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={!name.trim() || submitting}
            className="font-mono uppercase tracking-widest text-xs rounded-none bg-primary hover:bg-primary/90 text-primary-foreground"
          >
            {submitting ? 'SAVING...' : isCreating ? 'FABRICATE' : 'COMMIT'}
          </Button>
        </div>
      </Panel>
    </div>
  );
}
