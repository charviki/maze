import { useState, useEffect } from 'react';
import { controllerApi } from '../api/controller';
import type { Host, HostStatus } from '@maze/fabrication';
import {
  Button,
  DecryptText,
  GlitchEffect,
  ConfirmDialog,
  Skeleton,
  usePollingWithBackoff,
  clipPathHalf,
  useToast,
} from '@maze/fabrication';
import { Trash2, FileText } from 'lucide-react';

interface NodeListProps {
  onSelectNode: (node: Host) => void;
  selectedNodeName: string | null;
  onNodesChange?: (nodes: Host[]) => void;
  refreshTrigger?: number;
  onViewLog?: (hostName: string) => void;
}

export function NodeList({
  onSelectNode,
  selectedNodeName,
  onNodesChange,
  refreshTrigger,
  onViewLog,
}: NodeListProps) {
  const { showToast } = useToast();
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  const {
    data,
    isLoading,
    refresh: refreshNodes,
  } = usePollingWithBackoff<Host[]>({
    fetchFn: async () => {
      const res = await controllerApi.listHosts();
      if (res.status === 'ok' && res.data) {
        onNodesChange?.(res.data);
        return res.data;
      }
      return [];
    },
  });
  const hosts = data ?? [];

  useEffect(() => {
    if (refreshTrigger !== undefined && refreshTrigger > 0) {
      void refreshNodes();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [refreshTrigger]);

  const handleRemove = async (name: string) => {
    const res = await controllerApi.deleteHost(name);
    if (res.status === 'ok') {
      setDeleteTarget(null);
      void refreshNodes();
    } else {
      showToast('error', res.message || '删除 Host 失败');
    }
  };

  const formatTimeAgo = (dateStr: string) => {
    const diffSec = Math.floor((new Date().getTime() - new Date(dateStr).getTime()) / 1000);
    if (diffSec < 60) return `${diffSec}s ago`;
    return `${Math.floor(diffSec / 60)}m ago`;
  };

  // HostStatus 合法值集合，用于运行时校验后端返回的 status 字段
  const HOST_STATUSES = ['pending', 'deploying', 'online', 'offline', 'failed'] as const;

  // 后端可能返回未知状态，回退到 'offline' 作为安全默认值
  const toHostStatus = (raw: string | undefined): HostStatus => {
    if (raw && (HOST_STATUSES as readonly string[]).includes(raw)) {
      return raw as HostStatus;
    }
    return 'offline';
  };

  const statusConfig = (status: HostStatus) => {
    switch (status) {
      case 'pending':
      case 'deploying':
        return {
          dotClass: 'bg-amber-400 animate-pulse',
          glowClass: 'shadow-[0_0_5px_rgba(251,191,36,0.6)]',
          textClass: 'text-amber-400',
          borderClass: 'border-amber-500/30',
          label: status === 'pending' ? 'FABRICATION QUEUED' : 'FABRICATION IN PROGRESS',
          labelClass: 'text-amber-400/70 border-amber-500/30 bg-amber-500/10',
        };
      case 'online':
        return {
          dotClass: 'bg-primary animate-pulse',
          glowClass: 'shadow-[0_0_5px_rgba(0,255,255,0.8)]',
          textClass: 'text-primary',
          borderClass: 'border-primary/20',
          label: 'CONSCIOUSNESS ONLINE',
          labelClass: 'text-primary border-primary/20 bg-primary/10',
        };
      case 'offline':
        return {
          dotClass: 'bg-amber-500',
          glowClass: '',
          textClass: 'text-amber-500',
          borderClass: 'border-amber-500/30',
          label: 'CONSCIOUSNESS DRIFT',
          labelClass: 'text-amber-500 border-amber-500/30 bg-amber-500/10',
        };
      case 'failed':
        return {
          dotClass: 'bg-red-500',
          glowClass: '',
          textClass: 'text-red-500',
          borderClass: 'border-red-500/30',
          label: 'FABRICATION FAILED',
          labelClass: 'text-red-500 border-red-500/30 bg-red-500/10',
        };
    }
  };

  return (
    <div className="space-y-2">
      {isLoading &&
        Array.from({ length: 3 }).map((_, i) => (
          <div
            key={i}
            className="p-3 space-y-2 bg-card/50 border-l-2 border-primary/20"
            style={{
              clipPath: clipPathHalf(8),
            }}
          >
            <Skeleton className="h-3 w-24" />
            <Skeleton className="h-2 w-32 bg-primary/5" />
          </div>
        ))}
      {!isLoading &&
        hosts.map((host) => {
          const hostName = host.name ?? '';
          const hostStatus = toHostStatus(host.status);
          const isSelected = selectedNodeName === hostName;
          const cfg = statusConfig(hostStatus);
          const isOperational = hostStatus === 'online' || hostStatus === 'offline';
          return (
            <GlitchEffect
              key={hostName}
              isActive={hostStatus === 'offline' || hostStatus === 'failed'}
              className="block"
            >
              <div
                onClick={() => (isOperational ? onSelectNode(host) : undefined)}
                tabIndex={isOperational ? 0 : -1}
                role={isOperational ? 'button' : undefined}
                onKeyDown={(e) => {
                  if (isOperational && (e.key === 'Enter' || e.key === ' ')) {
                    e.preventDefault();
                    onSelectNode(host);
                  }
                }}
                className={`
                group flex flex-col p-3 cursor-pointer border-l-2 transition-all gap-2 relative
                ${
                  isSelected
                    ? 'bg-primary/10 border-primary shadow-[0_0_15px_rgba(0,255,255,0.15)]'
                    : `bg-card/50 ${cfg.borderClass} hover:border-primary/50 hover:bg-primary/5`
                }
                ${!isOperational ? 'cursor-default' : ''}
              `}
                style={{
                  clipPath: clipPathHalf(8),
                }}
              >
                {isSelected && (
                  <div className="absolute right-0 top-0 w-[1px] h-full bg-primary/50"></div>
                )}

                <div className="flex items-center justify-between pl-1">
                  <div className="flex items-center gap-3 font-mono text-sm tracking-wide text-primary">
                    <div className="relative flex h-3 w-1.5 shrink-0">
                      {hostStatus === 'online' && (
                        <span
                          className={`animate-pulse absolute inline-flex h-full w-full ${cfg.dotClass} opacity-75 ${cfg.glowClass}`}
                        ></span>
                      )}
                      {(hostStatus === 'pending' || hostStatus === 'deploying') && (
                        <span
                          className={`animate-pulse absolute inline-flex h-full w-full ${cfg.dotClass} opacity-60`}
                        ></span>
                      )}
                      <span className={`relative inline-flex h-3 w-1.5 ${cfg.dotClass}`}></span>
                    </div>
                    <span className={`truncate uppercase font-bold ${cfg.textClass}`}>
                      {isSelected ? <DecryptText text={hostName} /> : hostName}
                    </span>
                  </div>
                  <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    {onViewLog && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-6 w-6 rounded-none text-muted-foreground hover:text-primary"
                        onClick={(e) => {
                          e.stopPropagation();
                          onViewLog(hostName);
                        }}
                      >
                        <FileText className="w-3 h-3" />
                      </Button>
                    )}
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6 rounded-none text-muted-foreground hover:text-destructive"
                      onClick={(e) => {
                        e.stopPropagation();
                        setDeleteTarget(hostName);
                      }}
                    >
                      <Trash2 className="w-3 h-3" />
                    </Button>
                  </div>
                </div>

                <div className="text-[10px] text-muted-foreground space-y-1.5 font-mono uppercase tracking-widest pl-4">
                  <div className="flex justify-between items-center opacity-80">
                    <span>[ HOST_ID: {hostName.substring(0, 8)} ]</span>
                    <span className={`border px-1.5 text-[9px] ${cfg.labelClass}`}>
                      {isOperational ? `${host.sessionCount} LOOPS` : cfg.label}
                    </span>
                  </div>
                  <div className="flex justify-between items-center opacity-60">
                    {host.address ? (
                      <>
                        <span>ADDR: {host.address}</span>
                        <span>
                          SYS_BEAT: {host.lastHeartbeat ? formatTimeAgo(host.lastHeartbeat) : 'N/A'}
                        </span>
                      </>
                    ) : (
                      <>
                        <span>TOOLS: {(host.tools ?? []).join(', ')}</span>
                        <span>
                          CREATED: {host.createdAt ? formatTimeAgo(host.createdAt) : 'N/A'}
                        </span>
                      </>
                    )}
                  </div>
                  {host.errorMsg && (
                    <div className="text-red-400/80 text-[9px] truncate">ERR: {host.errorMsg}</div>
                  )}
                </div>
              </div>
            </GlitchEffect>
          );
        })}
      {!isLoading && hosts.length === 0 && (
        <div
          className="text-center p-6 text-xs font-mono uppercase tracking-widest text-destructive animate-pulse border border-destructive/30 bg-destructive/5"
          style={{
            clipPath: clipPathHalf(8),
          }}
        >
          [ NO HOSTS DETECTED ]
        </div>
      )}

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(v) => {
          if (!v) setDeleteTarget(null);
        }}
        title="确认销毁 Host"
        description={`确认销毁 Host「${deleteTarget || ''}」？该操作将停止容器并清理所有资源。`}
        confirmLabel="确认销毁"
        onConfirm={() => deleteTarget && handleRemove(deleteTarget)}
      />
    </div>
  );
}
