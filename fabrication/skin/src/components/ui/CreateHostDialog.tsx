import { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from './dialog';
import { Button } from './button';
import { Input } from './input';
import type { Tool, CreateHostRequest } from '../../types';
import { Loader2, Cpu, MemoryStick, Wrench, CheckSquare, Square } from 'lucide-react';

export interface CreateHostDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  tools: Tool[];
  onSubmit: (request: CreateHostRequest) => Promise<void>;
}

const CLIP_PATH =
  'polygon(0 0, calc(100% - 8px) 0, 100% 8px, 100% 100%, 8px 100%, 0 calc(100% - 8px))';

export function CreateHostDialog({
  open,
  onOpenChange,
  tools,
  onSubmit,
}: CreateHostDialogProps) {
  const [name, setName] = useState('');
  const [selectedTools, setSelectedTools] = useState<string[]>([]);
  const [cpuLimit, setCpuLimit] = useState('2');
  const [memoryLimit, setMemoryLimit] = useState('4g');
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const allSelected = tools.length > 0 && selectedTools.length === tools.length;

  const toggleTool = (id: string) => {
    setSelectedTools((prev) =>
      prev.includes(id) ? prev.filter((t) => t !== id) : [...prev, id]
    );
  };

  // 全选 / 取消全选切换
  const toggleAll = () => {
    setSelectedTools(allSelected ? [] : tools.map((t) => t.id));
  };

  const resetForm = () => {
    setName('');
    setSelectedTools([]);
    setCpuLimit('2');
    setMemoryLimit('4g');
    setError(null);
  };

  const handleSubmit = async () => {
    if (!name || selectedTools.length === 0) return;
    setIsCreating(true);
    setError(null);
    try {
      await onSubmit({
        name,
        tools: selectedTools,
        resources: {
          cpu_limit: cpuLimit || undefined,
          memory_limit: memoryLimit || undefined,
        },
      });
      resetForm();
      onOpenChange(false);
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      setError(msg);
    } finally {
      setIsCreating(false);
    }
  };

  const canSubmit = name.trim() !== '' && selectedTools.length > 0 && !isCreating;

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        if (!v) resetForm();
        onOpenChange(v);
      }}
    >
      <DialogContent className="max-w-xl border-primary/20 bg-background/95 backdrop-blur-sm">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 font-mono uppercase tracking-wider text-primary">
            <Cpu className="w-5 h-5" />
            CREATE HOST
          </DialogTitle>
          <DialogDescription className="font-mono text-muted-foreground text-xs uppercase tracking-widest">
            配置新 Host 的名称、工具集和资源限制
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-5 max-h-[60vh] overflow-y-auto pr-1">
          {/* Host 名称 */}
          <div className="space-y-2">
            <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono">
              HOST DESIGNATION
            </label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. host-alpha-01"
              className="font-mono rounded-none border-primary/20 bg-card/50 placeholder:text-muted-foreground/40 focus-visible:ring-primary/50"
            />
          </div>

          {/* 工具选择 */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono">
                TOOL SELECTION [{selectedTools.length}/{tools.length}]
              </label>
              <button
                type="button"
                onClick={toggleAll}
                className="text-[10px] text-primary/60 hover:text-primary uppercase tracking-widest font-mono transition-colors"
              >
                {allSelected ? 'DESELECT ALL' : 'SELECT ALL'}
              </button>
            </div>
            <div
              className="border border-primary/20 bg-card/50 p-3 space-y-1 max-h-48 overflow-y-auto"
              style={{ clipPath: CLIP_PATH }}
            >
              {tools.length === 0 && (
                <div className="text-center text-[10px] text-muted-foreground font-mono uppercase tracking-widest py-4 animate-pulse">
                  [ LOADING TOOLS... ]
                </div>
              )}
              {tools.map((tool) => {
                const isSelected = selectedTools.includes(tool.id);
                return (
                  <button
                    key={tool.id}
                    type="button"
                    onClick={() => toggleTool(tool.id)}
                    className={`
                      w-full flex items-start gap-3 p-2 transition-all text-left
                      ${isSelected ? 'bg-primary/10 border-l-2 border-primary' : 'border-l-2 border-transparent hover:bg-primary/5'}
                    `}
                  >
                    {isSelected ? (
                      <CheckSquare className="w-4 h-4 mt-0.5 text-primary shrink-0" />
                    ) : (
                      <Square className="w-4 h-4 mt-0.5 text-muted-foreground shrink-0" />
                    )}
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <Wrench className="w-3 h-3 text-primary/60 shrink-0" />
                        <span className="font-mono text-sm text-foreground truncate">
                          {tool.id}
                        </span>
                        <span className="text-[9px] text-muted-foreground/60 font-mono uppercase tracking-wider shrink-0">
                          [{tool.category}]
                        </span>
                      </div>
                      <p className="text-[11px] text-muted-foreground font-mono mt-0.5 leading-tight truncate">
                        {tool.description}
                      </p>
                    </div>
                  </button>
                );
              })}
            </div>
          </div>

          {/* 资源配置 */}
          <div className="space-y-2">
            <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono">
              RESOURCE LIMITS
            </label>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <div className="flex items-center gap-1.5 text-[10px] text-muted-foreground font-mono">
                  <Cpu className="w-3 h-3" />
                  CPU LIMIT
                </div>
                <Input
                  value={cpuLimit}
                  onChange={(e) => setCpuLimit(e.target.value)}
                  placeholder="2"
                  className="font-mono rounded-none border-primary/20 bg-card/50 placeholder:text-muted-foreground/40 focus-visible:ring-primary/50 h-8 text-xs"
                />
              </div>
              <div className="space-y-1.5">
                <div className="flex items-center gap-1.5 text-[10px] text-muted-foreground font-mono">
                  <MemoryStick className="w-3 h-3" />
                  MEMORY LIMIT
                </div>
                <Input
                  value={memoryLimit}
                  onChange={(e) => setMemoryLimit(e.target.value)}
                  placeholder="4g"
                  className="font-mono rounded-none border-primary/20 bg-card/50 placeholder:text-muted-foreground/40 focus-visible:ring-primary/50 h-8 text-xs"
                />
              </div>
            </div>
          </div>
        </div>

        {error && (
          <div className="mt-3 p-3 border border-red-500/30 bg-red-500/10 text-red-400 text-xs font-mono" style={{ clipPath: CLIP_PATH }}>
            <div className="text-[10px] uppercase tracking-widest mb-1 text-red-500/70">Error</div>
            <div className="break-all">{error}</div>
          </div>
        )}

        <DialogFooter className="pt-2">
          <Button
            variant="ghost"
            className="font-mono uppercase tracking-widest text-xs rounded-none"
            onClick={() => onOpenChange(false)}
          >
            CANCEL
          </Button>
          <Button
            className="font-mono uppercase tracking-widest text-xs rounded-none bg-primary hover:bg-primary/90"
            disabled={!canSubmit}
            onClick={handleSubmit}
          >
            {isCreating ? (
              <>
                <Loader2 className="w-3.5 h-3.5 animate-spin" />
                BUILDING IMAGE...
              </>
            ) : (
              'CREATE HOST'
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
