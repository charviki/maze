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
import type { V1MCPServer } from '@maze/fabrication';
import { Plus, X } from 'lucide-react';

const MCP_TYPES = ['stdio', 'sse', 'streamable-http'] as const;

interface MCPEditorProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  server: V1MCPServer | null;
  onSubmit: (data: {
    name: string;
    type: string;
    command?: string;
    url?: string;
    args?: string[];
    env?: Record<string, string>;
  }) => Promise<void>;
}

export function MCPEditor({ open, onOpenChange, server, onSubmit }: MCPEditorProps) {
  const [name, setName] = useState('');
  const [type, setType] = useState<string>('stdio');
  const [command, setCommand] = useState('');
  const [url, setUrl] = useState('');
  const [argsText, setArgsText] = useState('');
  const [envEntries, setEnvEntries] = useState<{ key: string; value: string }[]>([]);
  const [submitting, setSubmitting] = useState(false);

  const isEditing = !!server;

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    if (open) {
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
      } else {
        setName('');
        setType('stdio');
        setCommand('');
        setUrl('');
        setArgsText('');
        setEnvEntries([{ key: '', value: '' }]);
      }
      setSubmitting(false);
    }
  }, [open, server]);
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
    if (type === 'stdio' && !command.trim()) {
      alert('Command is required for stdio type');
      return;
    }
    if (type !== 'stdio' && !url.trim()) {
      alert('URL is required for ' + type + ' type');
      return;
    }
    setSubmitting(true);

    const args = argsText.trim().split(/\s+/).filter(Boolean);
    const env: Record<string, string> = {};
    for (const entry of envEntries) {
      if (entry.key.trim()) {
        env[entry.key.trim()] = entry.value;
      }
    }

    await onSubmit({
      name: name.trim(),
      type,
      command: type === 'stdio' ? command : undefined,
      url: type !== 'stdio' ? url : undefined,
      args: args.length > 0 ? args : undefined,
      env,
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-background border border-primary/20 max-w-md">
        <DialogHeader>
          <DialogTitle className="font-mono uppercase tracking-widest text-primary text-sm">
            {isEditing ? 'Edit MCP Server' : 'Create MCP Server'}
          </DialogTitle>
          <DialogDescription className="text-foreground/50 text-xs">
            {isEditing ? '修改 MCP Server 配置' : '创建新的 MCP Server'}
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
              placeholder="mcp-server-name"
              className="font-mono text-sm"
            />
          </div>

          <div>
            <label className="text-xs uppercase tracking-wider text-primary/60 mb-1 block">
              Type
            </label>
            <div className="flex gap-2">
              {MCP_TYPES.map((t) => (
                <button
                  key={t}
                  className={`px-3 py-1.5 text-xs font-mono uppercase tracking-wider border transition-colors ${
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
              <label className="text-xs uppercase tracking-wider text-primary/60 mb-1 block">
                Command
              </label>
              <Input
                value={command}
                onChange={(e) => setCommand(e.target.value)}
                placeholder="/usr/bin/my-mcp-server"
                className="font-mono text-sm"
              />
            </div>
          ) : (
            <div>
              <label className="text-xs uppercase tracking-wider text-primary/60 mb-1 block">
                URL
              </label>
              <Input
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                placeholder="http://localhost:3000/mcp"
                className="font-mono text-sm"
              />
            </div>
          )}

          <div>
            <label className="text-xs uppercase tracking-wider text-primary/60 mb-1 block">
              Args{' '}
              <span className="normal-case tracking-normal text-foreground/30">
                (space separated)
              </span>
            </label>
            <Input
              value={argsText}
              onChange={(e) => setArgsText(e.target.value)}
              placeholder="--port 8080 --verbose"
              className="font-mono text-sm"
            />
          </div>

          <div>
            <div className="flex items-center justify-between mb-1">
              <label className="text-xs uppercase tracking-wider text-primary/60">
                Environment
              </label>
              <Button
                variant="ghost"
                size="sm"
                className="h-5 text-[10px] text-primary/60 hover:text-primary"
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
                    className="font-mono text-xs flex-1"
                  />
                  <Input
                    value={entry.value}
                    onChange={(e) => updateEnvEntry(i, 'value', e.target.value)}
                    placeholder="value"
                    className="font-mono text-xs flex-1"
                  />
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6 shrink-0 text-foreground/30 hover:text-foreground"
                    onClick={() => removeEnvEntry(i)}
                  >
                    <X className="w-3 h-3" />
                  </Button>
                </div>
              ))}
            </div>
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
            disabled={
              !name.trim() || submitting || (type === 'stdio' ? !command.trim() : !url.trim())
            }
            onClick={handleSubmit}
          >
            {submitting ? 'Saving...' : isEditing ? 'Update' : 'Create'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
