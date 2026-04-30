import { useState, useMemo, useCallback } from 'react';
import { NodeList } from './components/NodeList';
import { createAgentApi } from './api/agent';
import { controllerApi } from './api/controller';
import { AgentPanel, DecryptText, RadarView, BootSequence, TerrainBackground, ErrorBoundary, Button, AnimationSettingsProvider, AnimationSettingsPanel, ToastProvider, CreateHostDialog } from '@maze/fabrication';
import type { Node, RadarNode, Tool, CreateHostRequest } from '@maze/fabrication';
import { Server, Activity, Menu, Settings, Plus } from 'lucide-react';
import './index.css';

function App() {
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);
  const [isBooting, setIsBooting] = useState(true);
  const [radarNodes, setRadarNodes] = useState<RadarNode[]>([]);
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [showAnimSettings, setShowAnimSettings] = useState(false);
  const [showCreateHost, setShowCreateHost] = useState(false);
  const [availableTools, setAvailableTools] = useState<Tool[]>([]);
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  // 选中节点后自动收起侧边栏（仅小屏幕生效）
  const handleSelectNode = useCallback((node: Node) => {
    setSelectedNode(node);
    if (window.innerWidth < 768) {
      setSidebarOpen(false);
    }
  }, []);

  const handleNodesChange = useCallback((nodes: Node[]) => {
    setRadarNodes(nodes.map(n => ({ name: n.name, status: n.status === 'online' ? 'online' : 'offline' })));
  }, []);

  // 缓存 apiClient，避免每次渲染创建新实例导致 XtermTerminal 的 wsUrl effect 重复触发
  const agentApi = useMemo(
    () => selectedNode ? createAgentApi('', selectedNode.name) : null,
    [selectedNode]
  );

  const terminalBg = useMemo(() => <TerrainBackground />, []);

  const handleOpenCreateHost = useCallback(async () => {
    setShowCreateHost(true);
    try {
      const res = await controllerApi.listTools();
      if (res.status === 'ok' && res.data) {
        setAvailableTools(res.data);
      }
    } catch {
      setAvailableTools([]);
    }
  }, []);

  const handleCreateHost = useCallback(async (data: CreateHostRequest) => {
    const res = await controllerApi.createHost(data);
    if (res.status === 'error') {
      throw new Error(res.message || '创建失败');
    }
    setRefreshTrigger((n) => n + 1);
  }, []);

  if (isBooting) {
    return (
      <ErrorBoundary>
        <AnimationSettingsProvider>
          <BootSequence onComplete={() => setIsBooting(false)} division="mesa-hub" />
        </AnimationSettingsProvider>
      </ErrorBoundary>
    );
  }

  return (
    <ErrorBoundary>
    <AnimationSettingsProvider>
    <ToastProvider>
    <div className="h-screen w-screen bg-background text-foreground dark relative overflow-hidden grid grid-rows-[56px_1fr] grid-cols-[288px_1fr]">
      {/* Background calibration marks */}
      <div className="pointer-events-none absolute inset-0">
        <div className="absolute top-4 left-4 w-4 h-4 border-t border-l border-primary/30"></div>
        <div className="absolute top-4 right-4 w-4 h-4 border-t border-r border-primary/30"></div>
        <div className="absolute bottom-4 left-4 w-4 h-4 border-b border-l border-primary/30"></div>
        <div className="absolute bottom-4 right-4 w-4 h-4 border-b border-r border-primary/30"></div>
        
        <div className="absolute top-1/2 left-4 w-2 h-[1px] bg-primary/30"></div>
        <div className="absolute top-1/2 right-4 w-2 h-[1px] bg-primary/30"></div>
        <div className="absolute top-4 left-1/2 w-[1px] h-2 bg-primary/30"></div>
        <div className="absolute bottom-4 left-1/2 w-[1px] h-2 bg-primary/30"></div>
      </div>

      {/* Top Navbar */}
      <div className="col-span-2 border-b border-border/50 flex items-center justify-between px-6 bg-background relative overflow-hidden z-10">
        {/* Decorative sci-fi scanline */}
        <div className="absolute top-0 left-0 w-full h-[1px] bg-primary/20"></div>
        <div className="flex items-center gap-3 font-bold text-lg tracking-wider text-primary">
          <Button variant="ghost" size="icon" className="md:hidden text-primary hover:text-primary hover:bg-primary/20" onClick={() => setSidebarOpen(!sidebarOpen)}>
            <Menu className="w-5 h-5" />
          </Button>
          <Activity className="w-5 h-5 text-primary" />
          <DecryptText text="MESA-HUB // DELOS CONTROL" className="uppercase" />
        </div>
        <Button variant="ghost" size="icon" className="text-primary hover:text-primary hover:bg-primary/20 rounded-none" onClick={() => setShowAnimSettings(true)} title="Visual Effects">
          <Settings className="w-4 h-4" />
        </Button>
      </div>

      {/* Pane 1: Nodes */}
      <div className={`border-r border-border/50 ${!sidebarOpen ? 'hidden md:flex' : 'flex'} flex-col bg-background/50 relative z-10 overflow-hidden`}>
        <div className="absolute right-0 top-0 w-[1px] h-full bg-gradient-to-b from-primary/20 to-transparent"></div>
        <div className="p-3 border-b border-border/50 font-bold flex items-center justify-between text-xs uppercase tracking-widest text-primary/80">
          <div className="flex items-center gap-2">
            <Server className="w-4 h-4" />
            HOSTS
          </div>
          <Button
            variant="ghost"
            size="icon"
            className="h-5 w-5 rounded-none text-primary/60 hover:text-primary hover:bg-primary/20"
            onClick={handleOpenCreateHost}
            title="Create Host"
          >
            <Plus className="w-3.5 h-3.5" />
          </Button>
        </div>
        <div className="flex-1 overflow-y-auto p-2">
          <NodeList 
            onSelectNode={handleSelectNode} 
            selectedNodeName={selectedNode?.name || null}
            onNodesChange={handleNodesChange}
            refreshTrigger={refreshTrigger}
          />
        </div>
        {/* Radar View */}
        <div className="p-4 border-t border-border/50 flex flex-col items-center justify-center relative overflow-hidden bg-background/40">
          <RadarView className="w-24 h-24 opacity-80" nodes={radarNodes} />
          <div className="mt-3 text-[10px] text-primary/50 tracking-[0.2em] font-mono">TOPOLOGY SCAN</div>
        </div>
      </div>

      {/* Pane 2 & 3: AgentPanel */}
      <div className={`flex w-full h-full relative z-10 bg-background overflow-hidden ${!sidebarOpen ? 'col-span-2' : ''}`}>
        {agentApi ? (
          <AgentPanel 
            apiClient={agentApi} 
            nodeName={selectedNode!.name}
            terminalBackground={terminalBg}
          />
        ) : (
          <div className="flex-1 flex items-center justify-center text-primary/40 uppercase tracking-widest text-xs">
            <div className="text-center space-y-4">
              <Activity className="w-16 h-16 mx-auto opacity-20" />
              <DecryptText text="AWAITING HOST SELECTION" />
            </div>
          </div>
        )}
      </div>
    </div>
    </ToastProvider>
    <CreateHostDialog
      open={showCreateHost}
      onOpenChange={setShowCreateHost}
      tools={availableTools}
      onSubmit={handleCreateHost}
    />
    <AnimationSettingsPanel open={showAnimSettings} onOpenChange={setShowAnimSettings} />
    </AnimationSettingsProvider>
    </ErrorBoundary>
  );
}

export default App;
