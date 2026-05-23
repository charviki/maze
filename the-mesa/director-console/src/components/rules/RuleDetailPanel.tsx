import { useState, useEffect } from 'react';
import { Input } from '@maze/fabrication';
import type { V1Rule } from '@maze/fabrication';
import { BookOpen } from 'lucide-react';
import { DetailPanelShell } from '../shared/DetailPanelShell';
import { FormLabel } from '../shared/FormLabel';

interface RuleDetailPanelProps {
  rule: V1Rule | null;
  isCreating: boolean;
  onSubmit: (data: { name: string; content?: string }) => Promise<void>;
  onCancel: () => void;
}

export function RuleDetailPanel({ rule, isCreating, onSubmit, onCancel }: RuleDetailPanelProps) {
  const [name, setName] = useState('');
  const [content, setContent] = useState('');
  const [submitting, setSubmitting] = useState(false);

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    if (rule) {
      setName(rule.name ?? '');
      setContent(rule.content ?? '');
    } else if (isCreating) {
      setName('');
      setContent('');
    }
    setSubmitting(false);
  }, [rule, isCreating]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const handleSubmit = async () => {
    if (!name.trim()) return;
    setSubmitting(true);
    try {
      await onSubmit({ name: name.trim(), content });
    } catch {
      setSubmitting(false);
    }
  };

  return (
    <DetailPanelShell
      icon={BookOpen}
      emptyText="SELECT RULE TO INSPECT"
      createTitle="FABRICATE NEW RULE"
      editTitle="RULE INSPECTION"
      item={rule}
      isCreating={isCreating}
      submitting={submitting}
      canSubmit={!!name.trim()}
      onCancel={onCancel}
      onSubmit={handleSubmit}
    >
      <div>
        <FormLabel>RULE DESIGNATION</FormLabel>
        <Input
          value={name}
          onChange={(e) => setName(e.target.value)}
          disabled={!!rule}
          placeholder="rule-name"
          className="font-mono text-sm rounded-none border-primary/20 bg-card/50"
        />
      </div>
      <div className="flex-1 min-h-0">
        <FormLabel>RULE CONTENT</FormLabel>
        <textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="# Rule content&#10;Write your rule in Markdown..."
          rows={12}
          className="w-full bg-card/50 border border-primary/20 rounded-none px-3 py-2 text-sm text-foreground font-mono resize-y focus:outline-none focus:border-primary/60 min-h-[200px]"
        />
      </div>
    </DetailPanelShell>
  );
}
