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

interface GitKeyEditorProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: { name: string; token: string }) => Promise<void>;
}

export function GitKeyEditor({ open, onOpenChange, onSubmit }: GitKeyEditorProps) {
  const [name, setName] = useState('');
  const [token, setToken] = useState('');
  const [submitting, setSubmitting] = useState(false);

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    if (open) {
      setName('');
      setToken('');
      setSubmitting(false);
    }
  }, [open]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const handleSubmit = async () => {
    if (!name.trim() || !token.trim()) return;
    setSubmitting(true);
    await onSubmit({ name: name.trim(), token: token.trim() });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-background border border-primary/20 max-w-md">
        <DialogHeader>
          <DialogTitle className="font-mono uppercase tracking-widest text-primary text-sm">
            Create Git Key
          </DialogTitle>
          <DialogDescription className="text-foreground/50 text-xs">
            添加 Git 凭据，token 将被加密存储
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
              placeholder="github-deploy-key"
              className="font-mono text-sm"
            />
          </div>

          <div>
            <label className="text-xs uppercase tracking-wider text-primary/60 mb-1 block">
              Token
            </label>
            <Input
              type="password"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              placeholder="ghp_xxxxxxxxxxxx"
              className="font-mono text-sm"
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
            disabled={!name.trim() || !token.trim() || submitting}
            onClick={handleSubmit}
          >
            {submitting ? 'Creating...' : 'Create'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
