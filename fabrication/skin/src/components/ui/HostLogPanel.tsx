import { useState, useEffect, useRef, useCallback } from 'react';
import { Button } from './button';
import { X } from 'lucide-react';

export interface HostLogPanelProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  hostName: string;
  fetchBuildLog: () => Promise<string>;
  fetchRuntimeLog: () => Promise<string>;
}

export function HostLogPanel({
  open,
  onOpenChange,
  hostName,
  fetchBuildLog,
  fetchRuntimeLog,
}: HostLogPanelProps) {
  const [activeTab, setActiveTab] = useState<'build' | 'runtime'>('build');
  const [logs, setLogs] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const logEndRef = useRef<HTMLDivElement>(null);

  const fetchLogs = useCallback(async () => {
    setIsLoading(true);
    try {
      const content = activeTab === 'build' ? await fetchBuildLog() : await fetchRuntimeLog();
      setLogs(content || '(empty)');
    } catch {
      setLogs('(failed to load logs)');
    } finally {
      setIsLoading(false);
    }
  }, [activeTab, fetchBuildLog, fetchRuntimeLog]);

  // fetch-in-effect: 面板打开时从 API 拉取日志并定时刷新写入 state，属于合法的数据同步模式。
  // React Compiler 的 set-state-in-effect 规则对 fetch-in-effect 场景存在已知误报。
  useEffect(() => {
    /* eslint-disable react-hooks/set-state-in-effect */
    if (!open) return;
    void fetchLogs();
    const interval = setInterval(() => {
      void fetchLogs();
    }, 5000);
    /* eslint-enable react-hooks/set-state-in-effect */
    return () => {
      clearInterval(interval);
    };
  }, [open, fetchLogs]);

  useEffect(() => {
    logEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [logs]);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center pointer-events-none">
      <div
        className="w-full max-w-3xl max-h-[60vh] flex flex-col bg-black/95 border border-primary/30 pointer-events-auto"
        style={{
          clipPath:
            'polygon(0 0, calc(100% - 12px) 0, 100% 12px, 100% 100%, 12px 100%, 0 calc(100% - 12px))',
        }}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-2 border-b border-primary/20 bg-black/80">
          <div className="flex items-center gap-3">
            <span className="font-mono text-[10px] uppercase tracking-[0.2em] text-primary/60">
              HOST LOG
            </span>
            <span className="font-mono text-sm text-primary font-bold uppercase tracking-wider">
              {hostName}
            </span>
          </div>
          <div className="flex items-center gap-2">
            {/* Tab buttons */}
            <div className="flex border border-primary/20">
              <button
                onClick={() => {
                  setActiveTab('build');
                }}
                className={`px-3 py-1 font-mono text-[10px] uppercase tracking-widest transition-colors ${
                  activeTab === 'build'
                    ? 'bg-primary/20 text-primary'
                    : 'bg-transparent text-muted-foreground hover:text-primary/60'
                }`}
              >
                BUILD LOG
              </button>
              <button
                onClick={() => {
                  setActiveTab('runtime');
                }}
                className={`px-3 py-1 font-mono text-[10px] uppercase tracking-widest transition-colors ${
                  activeTab === 'runtime'
                    ? 'bg-primary/20 text-primary'
                    : 'bg-transparent text-muted-foreground hover:text-primary/60'
                }`}
              >
                RUNTIME LOG
              </button>
            </div>
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6 rounded-none text-muted-foreground hover:text-primary"
              onClick={() => {
                onOpenChange(false);
              }}
            >
              <X className="w-3 h-3" />
            </Button>
          </div>
        </div>

        {/* Log content */}
        <div className="flex-1 overflow-y-auto p-4 font-mono text-[12px] text-green-400/90 leading-relaxed">
          {isLoading && !logs ? (
            <div className="text-primary/40 animate-pulse">Loading...</div>
          ) : (
            <pre className="whitespace-pre-wrap break-all">{logs}</pre>
          )}
          <div ref={logEndRef} />
        </div>

        {/* Footer */}
        <div className="px-4 py-1.5 border-t border-primary/20 flex justify-between items-center">
          <span className="font-mono text-[9px] text-primary/30 uppercase tracking-widest">
            {activeTab === 'build' ? 'HOST_LOGS/{name}.log' : 'CONTAINER STDOUT'}
          </span>
          <span className="font-mono text-[9px] text-primary/30 uppercase tracking-widest">
            AUTO-REFRESH 5S
          </span>
        </div>
      </div>
    </div>
  );
}
