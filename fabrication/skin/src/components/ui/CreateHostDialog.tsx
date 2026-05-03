import { useState, useEffect, useRef } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from './dialog';
import { Button } from './button';
import { Input } from './input';
import type { Tool, CreateHostRequest, Host, HostStatus } from '../../types';
import { Loader2, Cpu, MemoryStick, Wrench, CheckSquare, Square, X } from 'lucide-react';

export interface CreateHostDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  tools: Tool[];
  onSubmit: (request: CreateHostRequest) => Promise<Host>;
  onWaitOnline: (hostName: string) => Promise<boolean>;
  getHostBuildLog: (name: string) => Promise<string>;
}

const CLIP_PATH =
  'polygon(0 0, calc(100% - 8px) 0, 100% 8px, 100% 100%, 8px 100%, 0 calc(100% - 8px))';

function formatElapsed(seconds: number): string {
  const m = Math.floor(seconds / 60)
    .toString()
    .padStart(2, '0');
  const s = (seconds % 60).toString().padStart(2, '0');
  return `${m}:${s}`;
}

export function CreateHostDialog({
  open,
  onOpenChange,
  tools,
  onSubmit,
  onWaitOnline,
  getHostBuildLog,
}: CreateHostDialogProps) {
  const [name, setName] = useState('');
  const [selectedTools, setSelectedTools] = useState<string[]>([]);
  const [cpuLimit, setCpuLimit] = useState('0.5');
  const [memoryLimit, setMemoryLimit] = useState('256m');
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [phase, setPhase] = useState<'form' | 'building' | 'waiting' | 'online' | 'timeout'>(
    'form',
  );
  const [createdHostName, setCreatedHostName] = useState('');
  const [elapsedSeconds, setElapsedSeconds] = useState(0);
  const [buildLog, setBuildLog] = useState('');
  const [hostStatus, setHostStatus] = useState<HostStatus>('pending');
  const logEndRef = useRef<HTMLDivElement>(null);

  const allSelected = tools.length > 0 && selectedTools.length === tools.length;

  const toggleTool = (id: string) => {
    setSelectedTools((prev) => (prev.includes(id) ? prev.filter((t) => t !== id) : [...prev, id]));
  };

  const toggleAll = () => {
    setSelectedTools(allSelected ? [] : tools.map((t) => t.id));
  };

  const resetForm = () => {
    setName('');
    setSelectedTools([]);
    setCpuLimit('2');
    setMemoryLimit('4g');
    setError(null);
    setPhase('form');
    setCreatedHostName('');
    setElapsedSeconds(0);
    setBuildLog('');
    setHostStatus('pending');
  };

  // 轮询构建日志
  useEffect(() => {
    if (!createdHostName || phase === 'form' || phase === 'online' || phase === 'timeout') return;

    const interval = setInterval(async () => {
      try {
        const res = await getHostBuildLog(createdHostName);
        if (res) setBuildLog(res);
      } catch {
        // 日志可能还没准备好
      }
    }, 2000);
    return () => {
      clearInterval(interval);
    };
  }, [createdHostName, phase, getHostBuildLog]);

  // 自动滚动日志到底部
  useEffect(() => {
    logEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [buildLog]);

  const handleSubmit = async () => {
    if (!name || selectedTools.length === 0) return;
    setIsCreating(true);
    setError(null);
    try {
      const response = await onSubmit({
        name,
        tools: selectedTools,
        resources: {
          cpuLimit: cpuLimit || undefined,
          memoryLimit: memoryLimit || undefined,
        },
      });
      setCreatedHostName(response.name);
      setHostStatus(response.status);
      setPhase('waiting');
      setIsCreating(false);

      const startTime = Date.now();
      const timer = setInterval(() => {
        setElapsedSeconds(Math.floor((Date.now() - startTime) / 1000));
      }, 1000);

      try {
        const online = await onWaitOnline(response.name);
        clearInterval(timer);
        if (online) {
          setPhase('online');
          setTimeout(() => {
            resetForm();
            onOpenChange(false);
          }, 1500);
        } else {
          setPhase('timeout');
        }
      } catch {
        clearInterval(timer);
        setPhase('timeout');
      }
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      setError(msg);
    } finally {
      setIsCreating(false);
    }
  };

  const canSubmit =
    name.trim() !== '' && selectedTools.length > 0 && !isCreating && phase === 'form';

  const statusLabel = (): string => {
    switch (hostStatus) {
      case 'pending':
        return 'FABRICATION QUEUED';
      case 'deploying':
        return 'FABRICATION IN PROGRESS';
      case 'online':
        return 'CONSCIOUSNESS ONLINE';
      case 'offline':
        return 'CONSCIOUSNESS DRIFT';
      case 'failed':
        return 'FABRICATION FAILED';
    }
  };

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

        {/* 表单阶段 */}
        {phase === 'form' && (
          <div className="space-y-5 max-h-[60vh] overflow-y-auto pr-1">
            <div className="space-y-2">
              <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono">
                HOST DESIGNATION
              </label>
              <Input
                value={name}
                onChange={(e) => {
                  setName(e.target.value);
                }}
                placeholder="e.g. host-alpha-01"
                className="font-mono rounded-none border-primary/20 bg-card/50 placeholder:text-muted-foreground/40 focus-visible:ring-primary/50"
              />
            </div>

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
                      onClick={() => {
                        toggleTool(tool.id);
                      }}
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
                    onChange={(e) => {
                      setCpuLimit(e.target.value);
                    }}
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
                    onChange={(e) => {
                      setMemoryLimit(e.target.value);
                    }}
                    placeholder="4g"
                    className="font-mono rounded-none border-primary/20 bg-card/50 placeholder:text-muted-foreground/40 focus-visible:ring-primary/50 h-8 text-xs"
                  />
                </div>
              </div>
            </div>
          </div>
        )}

        {/* 等待上线阶段 — Westworld 觉醒动画 + 实时构建日志 */}
        {phase === 'waiting' && (
          <div className="space-y-3">
            <div
              className="relative overflow-hidden py-6 flex flex-col items-center justify-center"
              style={{ clipPath: CLIP_PATH }}
            >
              <div className="absolute inset-0 bg-black/40" />
              <div
                className="absolute left-0 w-full h-px bg-primary/60"
                style={{ animation: 'host-scanline 2.5s ease-in-out infinite' }}
              />

              <div className="relative z-10 font-mono text-primary text-lg uppercase tracking-[0.3em] mb-2">
                {createdHostName}
              </div>
              <div
                className="relative z-10 font-mono text-sm uppercase tracking-[0.4em] text-primary/80"
                style={{ animation: 'host-breathe 2s ease-in-out infinite' }}
              >
                {statusLabel()}
              </div>
              <div className="absolute bottom-2 z-10 font-mono text-[10px] uppercase tracking-[0.3em] text-primary/40">
                ELAPSED: {formatElapsed(elapsedSeconds)}
              </div>
            </div>

            {/* 构建日志终端 */}
            {buildLog && (
              <div
                className="bg-black/80 border border-primary/20 p-3 max-h-40 overflow-y-auto font-mono text-[11px] text-green-400/90 leading-relaxed"
                style={{ clipPath: CLIP_PATH }}
              >
                <pre className="whitespace-pre-wrap break-all">{buildLog}</pre>
                <div ref={logEndRef} />
              </div>
            )}
          </div>
        )}

        {/* 上线成功阶段 */}
        {phase === 'online' && (
          <div
            className="relative overflow-hidden py-10 flex flex-col items-center justify-center min-h-[280px]"
            style={{
              clipPath: CLIP_PATH,
              animation: 'host-online-glow 0.6s ease-out',
            }}
          >
            <div className="absolute inset-0 bg-green-500/5" />
            <div
              className="absolute inset-0 bg-green-400/20"
              style={{ animation: 'host-flash 0.4s ease-out forwards' }}
            />
            <div className="relative z-10 text-center space-y-3">
              <div className="font-mono text-green-400 text-2xl uppercase tracking-[0.3em] font-bold">
                HOST ONLINE
              </div>
              <div className="font-mono text-green-400/60 text-xs uppercase tracking-[0.3em]">
                // CONSCIOUSNESS ESTABLISHED
              </div>
            </div>
          </div>
        )}

        {/* 超时阶段 */}
        {phase === 'timeout' && (
          <div
            className="relative overflow-hidden py-8 flex flex-col items-center justify-center min-h-[280px]"
            style={{ clipPath: CLIP_PATH }}
          >
            <div className="absolute inset-0 bg-amber-500/5" />
            <div className="relative z-10 text-center space-y-4">
              <div className="font-mono text-amber-400 text-lg uppercase tracking-[0.3em] font-bold">
                CONNECTION TIMEOUT
              </div>
              <div className="font-mono text-amber-400/50 text-xs uppercase tracking-widest max-w-xs leading-relaxed">
                Host may still be initializing in the background
              </div>
              <Button
                className="font-mono uppercase tracking-widest text-xs rounded-none bg-amber-500/20 hover:bg-amber-500/30 text-amber-400 border border-amber-500/30"
                onClick={() => {
                  resetForm();
                  onOpenChange(false);
                }}
              >
                CLOSE
              </Button>
            </div>
          </div>
        )}

        {error && phase === 'form' && (
          <div
            className="mt-3 p-3 border border-red-500/30 bg-red-500/10 text-red-400 text-xs font-mono"
            style={{ clipPath: CLIP_PATH }}
          >
            <div className="text-[10px] uppercase tracking-widest mb-1 text-red-500/70">Error</div>
            <div className="break-all">{error}</div>
          </div>
        )}

        <DialogFooter className="pt-2">
          {phase === 'form' ? (
            <>
              <Button
                variant="ghost"
                className="font-mono uppercase tracking-widest text-xs rounded-none"
                onClick={() => {
                  onOpenChange(false);
                }}
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
                    CREATING...
                  </>
                ) : (
                  'CREATE HOST'
                )}
              </Button>
            </>
          ) : phase === 'waiting' ? (
            <Button
              variant="ghost"
              className="font-mono uppercase tracking-widest text-xs rounded-none text-primary/50 hover:text-primary flex items-center gap-1.5"
              onClick={() => {
                resetForm();
                onOpenChange(false);
              }}
            >
              <X className="w-3 h-3" />
              CLOSE
            </Button>
          ) : null}
        </DialogFooter>
      </DialogContent>

      <style>{`
        @keyframes host-scanline {
          0% { top: 0; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { top: 100%; opacity: 0; }
        }
        @keyframes host-breathe {
          0%, 100% { opacity: 0.4; }
          50% { opacity: 1; }
        }
        @keyframes host-online-glow {
          0% { box-shadow: 0 0 0 rgba(74, 222, 128, 0); }
          50% { box-shadow: 0 0 40px rgba(74, 222, 128, 0.3); }
          100% { box-shadow: 0 0 20px rgba(74, 222, 128, 0.1); }
        }
        @keyframes host-flash {
          0% { opacity: 1; }
          100% { opacity: 0; }
        }
      `}</style>
    </Dialog>
  );
}
