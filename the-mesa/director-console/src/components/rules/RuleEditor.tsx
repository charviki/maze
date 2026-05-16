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
import type { V1Rule } from '@maze/fabrication';

interface RuleEditorProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  rule: V1Rule | null;
  onSubmit: (data: { name: string; content?: string }) => Promise<void>;
}

export function RuleEditor({ open, onOpenChange, rule, onSubmit }: RuleEditorProps) {
  const [name, setName] = useState('');
  const [content, setContent] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const isEditing = !!rule;

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    if (open) {
      if (rule) {
        setName(rule.name ?? '');
        setContent(rule.content ?? '');
      } else {
        setName('');
        setContent('');
      }
      setSubmitting(false);
    }
  }, [open, rule]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const handleSubmit = async () => {
    if (!name.trim()) return;
    setSubmitting(true);
    await onSubmit({ name: name.trim(), content });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-background border border-primary/20 max-w-lg">
        <DialogHeader>
          <DialogTitle className="font-mono uppercase tracking-widest text-primary text-sm">
            {isEditing ? 'Edit Rule' : 'Create Rule'}
          </DialogTitle>
          <DialogDescription className="text-foreground/50 text-xs">
            {isEditing ? '修改 Rule 内容' : '创建新的 Rule'}
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
              placeholder="rule-name"
              className="font-mono text-sm"
            />
          </div>

          <div>
            <label className="text-xs uppercase tracking-wider text-primary/60 mb-1 block">
              Content
            </label>
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="# Rule content&#10;Write your rule in Markdown..."
              rows={12}
              className="w-full bg-card/50 border border-primary/20 rounded px-3 py-2 text-sm text-foreground font-mono resize-y focus:outline-none focus:border-primary/60 min-h-[200px]"
            />
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
