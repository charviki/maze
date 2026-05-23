import { useState, useEffect } from 'react';
import { Panel, Button, Input, DecryptText } from '@maze/fabrication';
import type { V1MCPServer } from '@maze/fabrication';
import { Plug, Plus, X } from 'lucide-react';

const MCP_TYPES = ['stdio', 'sse', 'streamable-http'] as const;

interface MCPDetailPanelProps {
  server: V1MCPServer | null;
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
        ? Object.entries(server.env).map(([key, value]) => ({ key, value }))
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

  if (!server && !isCreating) {
    return (
      <div className="flex-1 flex items-center justify-center text-primary/40 uppercase tracking-widest text-xs">
        <div className="text-center space-y-4">
          <Plug className="w-16 h-16 mx-auto opacity-20" />
          <DecryptText text="SELECT MCP SERVER TO INSPECT" />
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 min-w-0 flex flex-col bg-background relative z-10 overflow-hidden">
      <Panel className="flex flex-col h-full relative m-2" cornerSize={16}>
        <div className="pb-4 border-b border-primary/20">
          <div className="flex items-center gap-2 text-primary">
            <Plug className="w-4 h-4" />
            <h2 className="text-xs font-bold uppercase tracking-widest">
              <DecryptText
                text={isCreating ? 'FABRICATE NEW MCP SERVER' : 'MCP SERVER INSPECTION'}
              />
            </h2>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto pt-4 space-y-4 px-1">
          <div>
            <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono mb-1 block">
              SERVER DESIGNATION
            </label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              disabled={!!server}
              placeholder="mcp-server-name"
              className="font-mono text-sm rounded-none border-primary/20 bg-card/50"
            />
          </div>

          <div>
            <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono mb-1 block">
              TRANSPORT TYPE
            </label>
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
              <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono mb-1 block">
                COMMAND
              </label>
              <Input
                value={command}
                onChange={(e) => setCommand(e.target.value)}
                placeholder="/usr/bin/my-mcp-server"
                className="font-mono text-sm rounded-none border-primary/20 bg-card/50"
              />
            </div>
          ) : (
            <div>
              <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono mb-1 block">
                URL
              </label>
              <Input
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                placeholder="http://localhost:3000/mcp"
                className="font-mono text-sm rounded-none border-primary/20 bg-card/50"
              />
            </div>
          )}

          <div>
            <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono mb-1 block">
              ARGS{' '}
              <span className="normal-case tracking-normal text-foreground/30">
                (space separated)
              </span>
            </label>
            <Input
              value={argsText}
              onChange={(e) => setArgsText(e.target.value)}
              placeholder="--port 8080 --verbose"
              className="font-mono text-sm rounded-none border-primary/20 bg-card/50"
            />
          </div>

          <div>
            <div className="flex items-center justify-between mb-1">
              <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono">
                ENVIRONMENT
              </label>
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
            disabled={
              !name.trim() || submitting || (type === 'stdio' ? !command.trim() : !url.trim())
            }
            className="font-mono uppercase tracking-widest text-xs rounded-none bg-primary hover:bg-primary/90 text-primary-foreground"
          >
            {submitting ? 'SAVING...' : isCreating ? 'FABRICATE' : 'COMMIT'}
          </Button>
        </div>
      </Panel>
    </div>
  );
}
