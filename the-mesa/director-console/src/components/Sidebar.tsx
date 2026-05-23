import { useState } from 'react';
import { NodeList } from './NodeList';
import { RadarView, Button } from '@maze/fabrication';
import type { Host, RadarNode } from '@maze/fabrication';
import {
  Server,
  ChevronDown,
  ChevronRight,
  Wrench,
  Plug,
  BookOpen,
  KeyRound,
  Plus,
} from 'lucide-react';

export type FabricationItem = 'skills' | 'mcp-servers' | 'rules' | 'git-keys';

interface SidebarProps {
  selectedHostName: string | null;
  onSelectHost: (host: Host) => void;
  onHostsChange: (hosts: Host[]) => void;
  refreshTrigger: number;
  onViewLog: (hostName: string) => void;
  selectedFabricationItem: FabricationItem | null;
  onSelectFabricationItem: (item: FabricationItem | null) => void;
  onOpenCreateHost: () => void;
  radarNodes: RadarNode[];
}

const fabricationItems: { id: FabricationItem; label: string; icon: typeof Wrench }[] = [
  { id: 'skills', label: 'Skills', icon: Wrench },
  { id: 'mcp-servers', label: 'MCP', icon: Plug },
  { id: 'rules', label: 'Rules', icon: BookOpen },
  { id: 'git-keys', label: 'Git Keys', icon: KeyRound },
];

export function Sidebar({
  selectedHostName,
  onSelectHost,
  onHostsChange,
  refreshTrigger,
  onViewLog,
  selectedFabricationItem,
  onSelectFabricationItem,
  onOpenCreateHost,
  radarNodes,
}: SidebarProps) {
  const [hostsExpanded, setHostsExpanded] = useState(true);
  const [fabricationExpanded, setFabricationExpanded] = useState(false);

  const handleHostClick = (host: Host) => {
    onSelectFabricationItem(null);
    onSelectHost(host);
  };

  const handleFabricationClick = (item: FabricationItem) => {
    if (selectedFabricationItem === item) {
      onSelectFabricationItem(null);
    } else {
      onSelectFabricationItem(item);
    }
  };

  return (
    <div className="w-full border-r border-border/50 flex flex-col bg-background/50 relative z-10 overflow-hidden">
      <div className="absolute right-0 top-0 w-[1px] h-full bg-gradient-to-b from-primary/20 to-transparent" />

      {/* HOSTS Section */}
      <div
        className="p-3 border-b border-border/50 font-bold flex items-center justify-between text-xs uppercase tracking-widest text-primary/80 cursor-pointer select-none"
        onClick={() => setHostsExpanded(!hostsExpanded)}
      >
        <div className="flex items-center gap-2">
          {hostsExpanded ? (
            <ChevronDown className="w-3 h-3" />
          ) : (
            <ChevronRight className="w-3 h-3" />
          )}
          <Server className="w-4 h-4" />
          HOSTS
        </div>
        <Button
          variant="ghost"
          size="icon"
          className="h-5 w-5 rounded-none text-primary/60 hover:text-primary hover:bg-primary/20"
          onClick={(e) => {
            e.stopPropagation();
            onOpenCreateHost();
          }}
          title="Create Host"
        >
          <Plus className="w-3.5 h-3.5" />
        </Button>
      </div>

      {hostsExpanded && (
        <div className="flex-1 min-h-0 overflow-y-auto p-2">
          <NodeList
            onSelectNode={handleHostClick}
            selectedNodeName={selectedHostName}
            onNodesChange={onHostsChange}
            refreshTrigger={refreshTrigger}
            onViewLog={onViewLog}
          />
        </div>
      )}

      {/* FABRICATION Section */}
      <div
        className="p-3 border-t border-border/50 border-b border-border/50 font-bold flex items-center gap-2 text-xs uppercase tracking-widest text-primary/80 cursor-pointer select-none"
        onClick={() => setFabricationExpanded(!fabricationExpanded)}
      >
        {fabricationExpanded ? (
          <ChevronDown className="w-3 h-3" />
        ) : (
          <ChevronRight className="w-3 h-3" />
        )}
        <Wrench className="w-4 h-4" />
        FABRICATION
      </div>

      {fabricationExpanded && (
        <div className="py-1">
          {fabricationItems.map((item) => {
            const Icon = item.icon;
            const isSelected = selectedFabricationItem === item.id;
            return (
              <button
                key={item.id}
                className={`w-full px-4 py-2 flex items-center gap-2 text-xs uppercase tracking-wider transition-colors ${
                  isSelected
                    ? 'bg-primary/10 text-primary border-l-2 border-primary'
                    : 'text-foreground/60 hover:text-foreground hover:bg-card/50'
                }`}
                onClick={() => handleFabricationClick(item.id)}
              >
                <Icon className="w-3.5 h-3.5" />
                {item.label}
              </button>
            );
          })}
        </div>
      )}

      {/* Radar View */}
      <div className="p-4 border-t border-border/50 flex flex-col items-center justify-center relative overflow-hidden bg-background/40">
        <RadarView className="w-24 h-24 opacity-80" nodes={radarNodes} />
        <div className="mt-3 text-[10px] text-primary/50 tracking-[0.2em] font-mono">
          TOPOLOGY SCAN
        </div>
      </div>
    </div>
  );
}
