import { useState, useEffect } from 'react';
import { Panel, Button, Input, DecryptText } from '@maze/fabrication';
import type { V1GitKey } from '@maze/fabrication';
import { KeyRound } from 'lucide-react';

interface GitKeyDetailPanelProps {
  gitKey: V1GitKey | null;
  isCreating: boolean;
  onSubmit: (data: { name: string; token: string }) => Promise<void>;
  onCancel: () => void;
}

export function GitKeyDetailPanel({
  gitKey,
  isCreating,
  onSubmit,
  onCancel,
}: GitKeyDetailPanelProps) {
  const [name, setName] = useState('');
  const [token, setToken] = useState('');
  const [submitting, setSubmitting] = useState(false);

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    if (isCreating) {
      setName('');
      setToken('');
    }
    setSubmitting(false);
  }, [gitKey, isCreating]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const handleSubmit = async () => {
    if (!name.trim() || !token.trim()) return;
    setSubmitting(true);
    try {
      await onSubmit({ name: name.trim(), token: token.trim() });
    } catch {
      setSubmitting(false);
    }
  };

  if (!gitKey && !isCreating) {
    return (
      <div className="flex-1 flex items-center justify-center text-primary/40 uppercase tracking-widest text-xs">
        <div className="text-center space-y-4">
          <KeyRound className="w-16 h-16 mx-auto opacity-20" />
          <DecryptText text="SELECT GIT KEY TO INSPECT" />
        </div>
      </div>
    );
  }

  if (gitKey && !isCreating) {
    return (
      <div className="flex-1 min-w-0 flex flex-col bg-background relative z-10 overflow-hidden">
        <Panel className="flex flex-col h-full relative m-2" cornerSize={16}>
          <div className="pb-4 border-b border-primary/20">
            <div className="flex items-center gap-2 text-primary">
              <KeyRound className="w-4 h-4" />
              <h2 className="text-xs font-bold uppercase tracking-widest">
                <DecryptText text="GIT KEY INSPECTION" />
              </h2>
            </div>
          </div>

          <div className="flex-1 overflow-y-auto pt-4 space-y-4 px-1">
            <div>
              <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono mb-1 block">
                KEY DESIGNATION
              </label>
              <div className="font-mono text-sm text-foreground px-3 py-2 bg-card/50 border border-primary/20 rounded-none">
                {gitKey.name}
              </div>
            </div>

            <div>
              <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono mb-1 block">
                TOKEN (MASKED)
              </label>
              <div className="font-mono text-sm text-primary/60 px-3 py-2 bg-card/50 border border-primary/20 rounded-none">
                {gitKey.tokenMask}
              </div>
            </div>
          </div>
        </Panel>
      </div>
    );
  }

  return (
    <div className="flex-1 min-w-0 flex flex-col bg-background relative z-10 overflow-hidden">
      <Panel className="flex flex-col h-full relative m-2" cornerSize={16}>
        <div className="pb-4 border-b border-primary/20">
          <div className="flex items-center gap-2 text-primary">
            <KeyRound className="w-4 h-4" />
            <h2 className="text-xs font-bold uppercase tracking-widest">
              <DecryptText text="FABRICATE NEW GIT KEY" />
            </h2>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto pt-4 space-y-4 px-1">
          <div>
            <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono mb-1 block">
              KEY DESIGNATION
            </label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="github-deploy-key"
              className="font-mono text-sm rounded-none border-primary/20 bg-card/50"
            />
          </div>

          <div>
            <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono mb-1 block">
              TOKEN
            </label>
            <Input
              type="password"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              placeholder="ghp_xxxxxxxxxxxx"
              className="font-mono text-sm rounded-none border-primary/20 bg-card/50"
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
            disabled={!name.trim() || !token.trim() || submitting}
            className="font-mono uppercase tracking-widest text-xs rounded-none bg-primary hover:bg-primary/90 text-primary-foreground"
          >
            {submitting ? 'SAVING...' : 'FABRICATE'}
          </Button>
        </div>
      </Panel>
    </div>
  );
}
