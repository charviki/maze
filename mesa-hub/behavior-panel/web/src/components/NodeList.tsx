import { useState } from 'react';
import { controllerApi } from '../api/controller';
import type { Node } from '@maze/fabrication';
import { Button, DecryptText, GlitchEffect, ConfirmDialog, Skeleton, usePollingWithBackoff } from '@maze/fabrication';
import { Trash2 } from 'lucide-react';

interface NodeListProps {
  onSelectNode: (node: Node) => void;
  selectedNodeName: string | null;
  onNodesChange?: (nodes: Node[]) => void;
}

export function NodeList({ onSelectNode, selectedNodeName, onNodesChange }: NodeListProps) {
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  const { data, isLoading, refresh: refreshNodes } = usePollingWithBackoff<Node[]>({
    fetchFn: async () => {
      const res = await controllerApi.listNodes();
      if (res.status === 'ok' && res.data) {
        onNodesChange?.(res.data);
        return res.data;
      }
      return [];
    },
  });
  const nodes = data ?? [];

  const handleRemove = async (name: string) => {
    await controllerApi.deleteNode(name);
    setDeleteTarget(null);
    refreshNodes();
  };

  const formatTimeAgo = (dateStr: string) => {
    const diffSec = Math.floor((new Date().getTime() - new Date(dateStr).getTime()) / 1000);
    if (diffSec < 60) return `${diffSec}s ago`;
    return `${Math.floor(diffSec / 60)}m ago`;
  };

  return (
    <div className="space-y-2">
      {isLoading && Array.from({ length: 3 }).map((_, i) => (
        <div key={i} className="p-3 space-y-2 bg-card/50 border-l-2 border-primary/20" style={{ clipPath: 'polygon(0 0, calc(100% - 8px) 0, 100% 8px, 100% 100%, 8px 100%, 0 calc(100% - 8px))' }}>
          <Skeleton className="h-3 w-24" />
          <Skeleton className="h-2 w-32 bg-primary/5" />
        </div>
      ))}
      {!isLoading && nodes.map((node) => {
        const isSelected = selectedNodeName === node.name;
        const isOnline = node.status === 'online';
        return (
          <GlitchEffect key={node.name} isActive={!isOnline} className="block">
            <div 
              onClick={() => onSelectNode(node)}
              className={`
                group flex flex-col p-3 cursor-pointer border-l-2 transition-all gap-2 relative
                ${isSelected 
                  ? 'bg-primary/10 border-primary shadow-[0_0_15px_rgba(0,255,255,0.15)]' 
                  : 'bg-card/50 border-primary/20 hover:border-primary/50 hover:bg-primary/5'}
              `}
              style={{
                clipPath: 'polygon(0 0, calc(100% - 8px) 0, 100% 8px, 100% 100%, 8px 100%, 0 calc(100% - 8px))'
              }}
            >
              {isSelected && <div className="absolute right-0 top-0 w-[1px] h-full bg-primary/50"></div>}
              
              <div className="flex items-center justify-between pl-1">
                <div className="flex items-center gap-3 font-mono text-sm tracking-wide text-primary">
                  <div className="relative flex h-3 w-1.5 shrink-0">
                    {isOnline && <span className="animate-pulse absolute inline-flex h-full w-full bg-primary opacity-75 shadow-[0_0_5px_rgba(0,255,255,0.8)]"></span>}
                    <span className={`relative inline-flex h-3 w-1.5 ${isOnline ? 'bg-primary' : 'bg-muted-foreground'}`}></span>
                  </div>
                  <span className={`truncate uppercase font-bold ${!isOnline ? 'text-destructive' : ''}`}>
                    {isSelected ? <DecryptText text={node.name} /> : node.name}
                  </span>
                </div>
                {!isOnline && (
                  <Button variant="ghost" size="icon" className="h-6 w-6 rounded-none text-muted-foreground hover:text-destructive opacity-0 group-hover:opacity-100 transition-opacity" onClick={(e) => { e.stopPropagation(); setDeleteTarget(node.name); }}>
                    <Trash2 className="w-3 h-3" />
                  </Button>
                )}
              </div>
              
              <div className="text-[10px] text-muted-foreground space-y-1.5 font-mono uppercase tracking-widest pl-4">
                <div className="flex justify-between items-center opacity-80">
                  <span>[ HOST_ID: {node.name.substring(0, 8)} ]</span>
                  <span className={`bg-primary/10 border px-1.5 text-[9px] ${!isOnline ? 'text-destructive border-destructive/30' : 'text-primary border-primary/20'}`}>
                    {node.session_count} LOOPS
                  </span>
                </div>
                <div className="flex justify-between items-center opacity-60">
                  <span>ADDR: {node.external_addr}</span>
                  <span>SYS_BEAT: {formatTimeAgo(node.last_heartbeat)}</span>
                </div>
              </div>
            </div>
          </GlitchEffect>
        );
      })}
      {!isLoading && nodes.length === 0 && (
        <div className="text-center p-6 text-xs font-mono uppercase tracking-widest text-destructive animate-pulse border border-destructive/30 bg-destructive/5" style={{ clipPath: 'polygon(0 0, calc(100% - 8px) 0, 100% 8px, 100% 100%, 8px 100%, 0 calc(100% - 8px))' }}>
          [ NO HOSTS DETECTED ]
        </div>
      )}

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(v) => { if (!v) setDeleteTarget(null); }}
        title="确认移除 Host"
        description={`确认移除 Host「${deleteTarget || ''}」？该操作将从监控列表中移除此节点。`}
        confirmLabel="确认移除"
        onConfirm={() => deleteTarget && handleRemove(deleteTarget)}
      />
    </div>
  );
}
