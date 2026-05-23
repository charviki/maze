import { useState, useEffect } from 'react';
import { Panel, Button, Input, DecryptText } from '@maze/fabrication';
import type { V1Rule } from '@maze/fabrication';
import { BookOpen } from 'lucide-react';

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

  if (!rule && !isCreating) {
    return (
      <div className="flex-1 flex items-center justify-center text-primary/40 uppercase tracking-widest text-xs">
        <div className="text-center space-y-4">
          <BookOpen className="w-16 h-16 mx-auto opacity-20" />
          <DecryptText text="SELECT RULE TO INSPECT" />
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 min-w-0 flex flex-col bg-background relative z-10 overflow-hidden">
      <Panel className="flex flex-col h-full relative m-2" cornerSize={16}>
        <div className="pb-4 border-b border-primary/20">
          <div className="flex items-center gap-2 text-primary">
            <BookOpen className="w-4 h-4" />
            <h2 className="text-xs font-bold uppercase tracking-widest">
              <DecryptText text={isCreating ? 'FABRICATE NEW RULE' : 'RULE INSPECTION'} />
            </h2>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto pt-4 space-y-4 px-1">
          <div>
            <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono mb-1 block">
              RULE DESIGNATION
            </label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              disabled={!!rule}
              placeholder="rule-name"
              className="font-mono text-sm rounded-none border-primary/20 bg-card/50"
            />
          </div>

          <div className="flex-1 min-h-0">
            <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono mb-1 block">
              RULE CONTENT
            </label>
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="# Rule content&#10;Write your rule in Markdown..."
              rows={12}
              className="w-full bg-card/50 border border-primary/20 rounded-none px-3 py-2 text-sm text-foreground font-mono resize-y focus:outline-none focus:border-primary/60 min-h-[200px]"
            />
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
