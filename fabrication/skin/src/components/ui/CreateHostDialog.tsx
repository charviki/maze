import { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from './dialog';
import { Button } from './button';
import { Input } from './input';
import type { Tool, CreateHostRequest, CreateHostResponse } from '../../types';
import { Loader2, Cpu, MemoryStick, Wrench, CheckSquare, Square } from 'lucide-react';

export interface CreateHostDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  tools: Tool[];
  onSubmit: (request: CreateHostRequest) => Promise<CreateHostResponse>;
  // 轮询直到 Host 上线或超时，返回是否成功
  onWaitOnline: (hostName: string) => Promise<boolean>;
}

const CLIP_PATH =
  'polygon(0 0, calc(100% - 8px) 0, 100% 8px, 100% 100%, 8px 100%, 0 calc(100% - 8px))';

function formatElapsed(seconds: number): string {
  const m = Math.floor(seconds / 60).toString().padStart(2, '0');
  const s = (seconds % 60).toString().padStart(2, '0');
  return `${m}:${s}`;
}

export function CreateHostDialog({
  open,
  onOpenChange,
  tools,
  onSubmit,
  onWaitOnline,
}: CreateHostDialogProps) {
  const [name, setName] = useState('');
  const [selectedTools, setSelectedTools] = useState<string[]>([]);
  const [cpuLimit, setCpuLimit] = useState('2');
  const [memoryLimit, setMemoryLimit] = useState('4g');
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [phase, setPhase] = useState<'form' | 'building' | 'waiting' | 'online' | 'timeout'>('form');
  const [createdHostName, setCreatedHostName] = useState('');
  const [elapsedSeconds, setElapsedSeconds] = useState(0);

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
    setPhase('form');
    setCreatedHostName('');
    setElapsedSeconds(0);
  };

  const handleSubmit = async () => {
    if (!name || selectedTools.length === 0) return;
    setIsCreating(true);
    setError(null);
    try {
      const response = await onSubmit({
        name,
        tools: selectedTools,
        resources: {
          cpu_limit: cpuLimit || undefined,
          memory_limit: memoryLimit || undefined,
        },
      });
      // 创建成功，进入等待上线状态
      setCreatedHostName(response.name);
      setPhase('waiting');
      setIsCreating(false);

      // 启动计时器
      const startTime = Date.now();
      const timer = setInterval(() => {
        setElapsedSeconds(Math.floor((Date.now() - startTime) / 1000));
      }, 1000);

      // 轮询等待 Host 上线
      try {
        const online = await onWaitOnline(response.name);
        clearInterval(timer);
        if (online) {
          setPhase('online');
          // 1.5 秒后自动关闭
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

  const canSubmit = name.trim() !== '' && selectedTools.length > 0 && !isCreating && phase === 'form';

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
        )}

        {/* 等待上线阶段 — Westworld 觉醒动画 */}
        {phase === 'waiting' && (
        <div
          className="relative overflow-hidden py-10 flex flex-col items-center justify-center min-h-[280px]"
          style={{ clipPath: CLIP_PATH }}
        >
          {/* 深色背景 + 微弱脉冲 */}
          <div className="absolute inset-0 bg-black/60 animate-pulse" />

          {/* 扫描线动画：从顶部到底部循环扫过 */}
          <div
            className="absolute left-0 w-full h-px bg-primary/60"
            style={{
              animation: 'host-scanline 2.5s ease-in-out infinite',
            }}
          />

          {/* Host 名称 */}
          <div className="relative z-10 font-mono text-primary text-lg uppercase tracking-[0.3em] mb-4">
            {createdHostName}
          </div>

          {/* 呼吸闪烁的文字 */}
          <div
            className="relative z-10 font-mono text-sm uppercase tracking-[0.4em] text-primary/80"
            style={{
              animation: 'host-breathe 2s ease-in-out infinite',
            }}
          >
            CONSCIOUSNESS INITIALIZING...
          </div>

          {/* 已用时间 */}
          <div className="absolute bottom-4 z-10 font-mono text-[10px] uppercase tracking-[0.3em] text-primary/40">
            ELAPSED: {formatElapsed(elapsedSeconds)}
          </div>
        </div>
        )}

        {/* 上线成功阶段 — 绿色觉醒闪光 */}
        {phase === 'online' && (
        <div
          className="relative overflow-hidden py-10 flex flex-col items-center justify-center min-h-[280px]"
          style={{
            clipPath: CLIP_PATH,
            animation: 'host-online-glow 0.6s ease-out',
          }}
        >
          {/* 绿色脉冲背景 */}
          <div className="absolute inset-0 bg-green-500/5" />
          {/* 绿色闪光 */}
          <div
            className="absolute inset-0 bg-green-400/20"
            style={{
              animation: 'host-flash 0.4s ease-out forwards',
            }}
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
          <div className="mt-3 p-3 border border-red-500/30 bg-red-500/10 text-red-400 text-xs font-mono" style={{ clipPath: CLIP_PATH }}>
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
            </>
          ) : phase === 'waiting' ? (
            // 等待中允许用户关闭（取消等待）
            <Button
              variant="ghost"
              className="font-mono uppercase tracking-widest text-xs rounded-none text-primary/50 hover:text-primary"
              onClick={() => {
                resetForm();
                onOpenChange(false);
              }}
            >
              DISMISS
            </Button>
          ) : null}
        </DialogFooter>
      </DialogContent>

      {/* 自定义 CSS keyframes — 扫描线、呼吸、上线闪光 */}
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
