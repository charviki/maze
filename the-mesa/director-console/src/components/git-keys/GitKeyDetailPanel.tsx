import { useState, useEffect } from 'react';
import {
  Input,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@maze/fabrication';
import type { V1GitKey } from '@maze/fabrication';
import { KeyRound } from 'lucide-react';
import { DetailPanelShell } from '../shared/DetailPanelShell';
import { FormLabel } from '../shared/FormLabel';
import { TOKEN_TYPE_DISPLAY, type CreateGitKeyData } from './constants';

const TOKEN_TYPE_OPTIONS = [
  { value: 'PERSONAL_ACCESS_TOKEN', label: 'Personal Access Token' },
  { value: 'SSH_KEY', label: 'SSH Key' },
] as const;

const HOST_SUGGESTIONS = ['github.com', 'gitlab.com', 'bitbucket.org', 'gitea.com'];

interface GitKeyDetailPanelProps {
  gitKey: V1GitKey | null;
  isCreating: boolean;
  onSubmit: (data: CreateGitKeyData) => Promise<void>;
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
  const [tokenType, setTokenType] = useState('PERSONAL_ACCESS_TOKEN');
  const [host, setHost] = useState('');
  const [submitting, setSubmitting] = useState(false);

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    if (isCreating) {
      setName('');
      setToken('');
      setTokenType('PERSONAL_ACCESS_TOKEN');
      setHost('');
    }
    setSubmitting(false);
  }, [gitKey, isCreating]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const handleSubmit = async () => {
    if (!name.trim() || !token.trim()) return;
    setSubmitting(true);
    try {
      await onSubmit({
        name: name.trim(),
        token: token.trim(),
        tokenType,
        host: host.trim() || undefined,
      });
    } catch {
      setSubmitting(false);
    }
  };

  const tokenTypeDisplay = (tt?: string) => TOKEN_TYPE_DISPLAY[tt ?? ''] ?? tt ?? '—';

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
          <div className="flex gap-4">
            <div className="flex-1">
              <FormLabel>TOKEN TYPE</FormLabel>
              <div className="flex items-center gap-2">
                <span className="inline-block px-2 py-1 text-[10px] font-mono uppercase tracking-widest border border-primary/30 bg-primary/10 text-primary">
                  {tokenTypeDisplay(gitKey?.tokenType)}
                </span>
              </div>
            </div>
            <div className="flex-1">
              <FormLabel>HOST</FormLabel>
              <div className="font-mono text-sm text-foreground px-3 py-2 bg-card/50 border border-primary/20 rounded-none">
                {gitKey?.host || '—'}
              </div>
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
        <FormLabel>TOKEN TYPE</FormLabel>
        <Select value={tokenType} onValueChange={setTokenType}>
          <SelectTrigger className="font-mono text-sm rounded-none border-primary/20 bg-card/50">
            <SelectValue />
          </SelectTrigger>
          <SelectContent className="font-mono text-sm rounded-none border-primary/20">
            {TOKEN_TYPE_OPTIONS.map((opt) => (
              <SelectItem key={opt.value} value={opt.value} className="font-mono text-sm">
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <div>
        <FormLabel>HOST</FormLabel>
        <Input
          value={host}
          onChange={(e) => setHost(e.target.value)}
          placeholder="github.com"
          list="git-key-host-suggestions"
          className="font-mono text-sm rounded-none border-primary/20 bg-card/50"
        />
        <datalist id="git-key-host-suggestions">
          {HOST_SUGGESTIONS.map((h) => (
            <option key={h} value={h} />
          ))}
        </datalist>
      </div>
      <div>
        <FormLabel>TOKEN</FormLabel>
        <Input
          type="password"
          value={token}
          onChange={(e) => setToken(e.target.value)}
          placeholder={
            tokenType === 'SSH_KEY' ? '-----BEGIN OPENSSH PRIVATE KEY-----...' : 'ghp_xxxxxxxxxxxx'
          }
          className="font-mono text-sm rounded-none border-primary/20 bg-card/50"
        />
      </div>
    </DetailPanelShell>
  );
}
