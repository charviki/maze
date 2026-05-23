import { useState, useEffect } from 'react';
import { Input } from '@maze/fabrication';
import type { V1GitKey } from '@maze/fabrication';
import { KeyRound } from 'lucide-react';
import { DetailPanelShell } from '../shared/DetailPanelShell';
import { FormLabel } from '../shared/FormLabel';

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

  return (
    <DetailPanelShell
      icon={KeyRound}
      emptyText="SELECT GIT KEY TO INSPECT"
      createTitle="FABRICATE NEW GIT KEY"
      editTitle="GIT KEY INSPECTION"
      item={gitKey}
      isCreating={isCreating}
      submitting={submitting}
      canSubmit={!!name.trim() && !!token.trim()}
      canEdit={false}
      readOnlyView={
        <>
          <div>
            <FormLabel>KEY DESIGNATION</FormLabel>
            <div className="font-mono text-sm text-foreground px-3 py-2 bg-card/50 border border-primary/20 rounded-none">
              {gitKey?.name}
            </div>
          </div>
          <div>
            <FormLabel>TOKEN (MASKED)</FormLabel>
            <div className="font-mono text-sm text-primary/60 px-3 py-2 bg-card/50 border border-primary/20 rounded-none">
              {gitKey?.tokenMask}
            </div>
          </div>
        </>
      }
      onCancel={onCancel}
      onSubmit={handleSubmit}
    >
      <div>
        <FormLabel>KEY DESIGNATION</FormLabel>
        <Input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="github-deploy-key"
          className="font-mono text-sm rounded-none border-primary/20 bg-card/50"
        />
      </div>
      <div>
        <FormLabel>TOKEN</FormLabel>
        <Input
          type="password"
          value={token}
          onChange={(e) => setToken(e.target.value)}
          placeholder="ghp_xxxxxxxxxxxx"
          className="font-mono text-sm rounded-none border-primary/20 bg-card/50"
        />
      </div>
    </DetailPanelShell>
  );
}
