import { useState, useEffect } from 'react';
import { Button, Input } from '@maze/fabrication';
import type { V1Skill } from '@maze/fabrication';
import { Sparkles, Plus, X } from 'lucide-react';
import { DetailPanelShell } from '../shared/DetailPanelShell';
import { FormLabel } from '../shared/FormLabel';

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

  return (
    <DetailPanelShell
      icon={Sparkles}
      emptyText="SELECT SKILL TO INSPECT"
      createTitle="FABRICATE NEW SKILL"
      editTitle="SKILL INSPECTION"
      item={skill}
      isCreating={isCreating}
      submitting={submitting}
      canSubmit={!!name.trim()}
      onCancel={onCancel}
      onSubmit={handleSubmit}
    >
      <div>
        <FormLabel>SKILL DESIGNATION</FormLabel>
        <Input
          value={name}
          onChange={(e) => setName(e.target.value)}
          disabled={!!skill}
          placeholder="skill-name"
          className="font-mono text-sm rounded-none border-primary/20 bg-card/50"
        />
      </div>
      <div>
        <FormLabel>DESCRIPTION</FormLabel>
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
          <FormLabel inline>CONFIG PARAMETERS</FormLabel>
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
    </DetailPanelShell>
  );
}
