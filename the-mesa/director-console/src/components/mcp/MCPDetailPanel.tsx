import { useState, useEffect } from 'react';
import { Button, Input } from '@maze/fabrication';
import type { V1McpServer } from '@maze/fabrication';
import { Plug, Plus, X } from 'lucide-react';
import { DetailPanelShell } from '../shared/DetailPanelShell';
import { FormLabel } from '../shared/FormLabel';

const MCP_TYPES = ['stdio', 'sse', 'streamable-http'] as const;

interface MCPDetailPanelProps {
  server: V1McpServer | null;
  isCreating: boolean;
  onSubmit: (data: {
    name: string;
    type: string;
    command?: string;
    url?: string;
    args?: string[];
    env?: Record<string, string>;
  }) => Promise<void>;
  onCancel: () => void;
}

export function MCPDetailPanel({ server, isCreating, onSubmit, onCancel }: MCPDetailPanelProps) {
  const [name, setName] = useState('');
  const [type, setType] = useState<string>('stdio');
  const [command, setCommand] = useState('');
  const [url, setUrl] = useState('');
  const [argsText, setArgsText] = useState('');
  const [envEntries, setEnvEntries] = useState<{ key: string; value: string }[]>([]);
  const [submitting, setSubmitting] = useState(false);

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    if (server) {
      setName(server.name ?? '');
      setType(server.type ?? 'stdio');
      setCommand(server.command ?? '');
      setUrl(server.url ?? '');
      setArgsText((server.args ?? []).join(' '));
      const entries = server.env
        ? Object.entries(server.env).map(([key, value]) => ({ key, value: String(value) }))
        : [];
      setEnvEntries(entries.length > 0 ? entries : [{ key: '', value: '' }]);
    } else if (isCreating) {
      setName('');
      setType('stdio');
      setCommand('');
      setUrl('');
      setArgsText('');
      setEnvEntries([{ key: '', value: '' }]);
    }
    setSubmitting(false);
  }, [server, isCreating]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const addEnvEntry = () => setEnvEntries((prev) => [...prev, { key: '', value: '' }]);
  const removeEnvEntry = (index: number) =>
    setEnvEntries((prev) => prev.filter((_, i) => i !== index));
  const updateEnvEntry = (index: number, field: 'key' | 'value', val: string) =>
    setEnvEntries((prev) =>
      prev.map((entry, i) => (i === index ? { ...entry, [field]: val } : entry)),
    );

  const handleSubmit = async () => {
    if (!name.trim()) return;
    if (type === 'stdio' && !command.trim()) return;
    if (type !== 'stdio' && !url.trim()) return;
    setSubmitting(true);

    const args = argsText.trim().split(/\s+/).filter(Boolean);
    const env: Record<string, string> = {};
    for (const entry of envEntries) {
      if (entry.key.trim()) {
        env[entry.key.trim()] = entry.value;
      }
    }

    try {
      await onSubmit({
        name: name.trim(),
        type,
        command: type === 'stdio' ? command : undefined,
        url: type !== 'stdio' ? url : undefined,
        args: args.length > 0 ? args : undefined,
        env,
      });
    } catch {
      setSubmitting(false);
    }
  };

  return (
    <DetailPanelShell
      icon={Plug}
      emptyText="SELECT MCP SERVER TO INSPECT"
      createTitle="FABRICATE NEW MCP SERVER"
      editTitle="MCP SERVER INSPECTION"
      item={server}
      isCreating={isCreating}
      submitting={submitting}
      canSubmit={!!name.trim() && (type === 'stdio' ? !!command.trim() : !!url.trim())}
      onCancel={onCancel}
      onSubmit={handleSubmit}
    >
      <div>
        <FormLabel>SERVER DESIGNATION</FormLabel>
        <Input
          value={name}
          onChange={(e) => setName(e.target.value)}
          disabled={!!server}
          placeholder="mcp-server-name"
          className="font-mono text-sm rounded-none border-primary/20 bg-card/50"
        />
      </div>

      <div>
        <FormLabel>TRANSPORT TYPE</FormLabel>
        <div className="flex gap-2">
          {MCP_TYPES.map((t) => (
            <button
              key={t}
              className={`px-3 py-1.5 text-xs font-mono uppercase tracking-wider border transition-colors rounded-none ${
                type === t
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-primary/20 text-foreground/50 hover:border-primary/40'
              }`}
              onClick={() => setType(t)}
            >
              {t}
            </button>
          ))}
        </div>
      </div>

      {type === 'stdio' ? (
        <div>
          <FormLabel>COMMAND</FormLabel>
          <Input
            value={command}
            onChange={(e) => setCommand(e.target.value)}
            placeholder="/usr/bin/my-mcp-server"
            className="font-mono text-sm rounded-none border-primary/20 bg-card/50"
          />
        </div>
      ) : (
        <div>
          <FormLabel>URL</FormLabel>
          <Input
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="http://localhost:3000/mcp"
            className="font-mono text-sm rounded-none border-primary/20 bg-card/50"
          />
        </div>
      )}

      <div>
        <FormLabel>
          ARGS{' '}
          <span className="normal-case tracking-normal text-foreground/30">(space separated)</span>
        </FormLabel>
        <Input
          value={argsText}
          onChange={(e) => setArgsText(e.target.value)}
          placeholder="--port 8080 --verbose"
          className="font-mono text-sm rounded-none border-primary/20 bg-card/50"
        />
      </div>

      <div>
        <div className="flex items-center justify-between mb-1">
          <FormLabel inline>ENVIRONMENT</FormLabel>
          <Button
            variant="ghost"
            size="sm"
            className="h-5 text-[10px] text-primary/60 hover:text-primary uppercase tracking-widest"
            onClick={addEnvEntry}
          >
            <Plus className="w-3 h-3 mr-0.5" />
            ADD
          </Button>
        </div>
        <div className="space-y-2">
          {envEntries.map((entry, i) => (
            <div key={i} className="flex items-center gap-2">
              <Input
                value={entry.key}
                onChange={(e) => updateEnvEntry(i, 'key', e.target.value)}
                placeholder="KEY"
                className="font-mono text-xs flex-1 rounded-none border-primary/20 bg-card/50"
              />
              <Input
                value={entry.value}
                onChange={(e) => updateEnvEntry(i, 'value', e.target.value)}
                placeholder="value"
                className="font-mono text-xs flex-1 rounded-none border-primary/20 bg-card/50"
              />
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6 shrink-0 text-foreground/30 hover:text-foreground rounded-none"
                onClick={() => removeEnvEntry(i)}
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
