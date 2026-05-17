import { useState, useMemo, useCallback, useRef, useEffect } from 'react';
import { createAgentApi } from './api/agent';
import { controllerApi } from './api/controller';
import { Sidebar, type FabricationItem } from './components/Sidebar';
import { FabricationPanel } from './components/FabricationPanel';
import {
  AgentPanel,
  DecryptText,
  BootSequence,
  TerrainBackground,
  Button,
  AnimationSettingsPanel,
  CreateHostDialog,
  HostLogPanel,
  useToast,
  AppShell,
  AppNavbar,
  CalibrationMarks,
} from '@maze/fabrication';
import type {
  Host,
  RadarNode,
  Tool,
  CreateHostRequest,
  HostSpec,
  Skill,
  MCPServer,
} from '@maze/fabrication';
import { Activity, Menu, Settings } from 'lucide-react';
import './index.css';

// Toast hook 依赖 Provider，上层壳组件负责先装配 Provider，避免应用启动即白屏。
function AppContent() {
  const { showToast } = useToast();
  const [selectedHost, setSelectedHost] = useState<Host | null>(null);
  const [isBooting, setIsBooting] = useState(true);
  const [radarNodes, setRadarNodes] = useState<RadarNode[]>([]);
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [showAnimSettings, setShowAnimSettings] = useState(false);
  const [showCreateHost, setShowCreateHost] = useState(false);
  const [availableTools, setAvailableTools] = useState<Tool[]>([]);
  const [availableSkills, setAvailableSkills] = useState<Skill[]>([]);
  const [availableMcpServers, setAvailableMcpServers] = useState<MCPServer[]>([]);
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [logPanelHost, setLogPanelHost] = useState<string | null>(null);
  const [selectedFabricationItem, setSelectedFabricationItem] = useState<FabricationItem | null>(
    null,
  );
  const abortRef = useRef<AbortController | null>(null);

  useEffect(() => {
    return () => {
      abortRef.current?.abort();
    };
  }, []);

  const handleSelectNode = useCallback((node: Host) => {
    setSelectedHost(node);
    setSelectedFabricationItem(null);
    if (window.innerWidth < 768) {
      setSidebarOpen(false);
    }
  }, []);

  const handleNodesChange = useCallback((nodes: Host[]) => {
    setRadarNodes(
      nodes.map((n) => ({
        name: n.name ?? '',
        status: n.status === 'online' ? 'online' : 'offline',
      })),
    );
    // Host 列表刷新后要同步校正当前选中项；否则已删除/下线的旧节点仍会驱动 AgentPanel 继续轮询会话接口。
    setSelectedHost((current) => {
      if (!current?.name) {
        return current;
      }
      return nodes.find((node) => node.name === current.name) ?? null;
    });
  }, []);

  const agentApi = useMemo(
    () => (selectedHost?.name ? createAgentApi('', selectedHost.name) : null),
    [selectedHost],
  );

  const terminalBg = useMemo(() => <TerrainBackground />, []);

  const handleOpenCreateHost = useCallback(async () => {
    setShowCreateHost(true);
    const [toolsResult, skillsResult, mcpServersResult] = await Promise.allSettled([
      controllerApi.listTools(),
      controllerApi.listSkills(),
      controllerApi.listMCPServers(),
    ]);

    if (toolsResult.status === 'fulfilled') {
      const res = toolsResult.value;
      if (res.status === 'ok' && res.data) {
        setAvailableTools(res.data);
      }
    } else {
      setAvailableTools([]);
      showToast('error', '工具列表获取失败');
    }

    if (skillsResult.status === 'fulfilled') {
      setAvailableSkills(skillsResult.value);
    } else {
      setAvailableSkills([]);
      showToast('error', 'Skill 列表获取失败');
    }

    if (mcpServersResult.status === 'fulfilled') {
      setAvailableMcpServers(mcpServersResult.value);
    } else {
      setAvailableMcpServers([]);
      showToast('error', 'MCP Server 列表获取失败');
    }
  }, [showToast]);

  const handleCreateHost = useCallback(async (data: CreateHostRequest): Promise<HostSpec> => {
    const res = await controllerApi.createHost(data);
    if (res.status === 'error') {
      throw new Error(res.message || '创建失败');
    }
    setRefreshTrigger((n) => n + 1);
    return res.data!;
  }, []);

  // 轮询等待 Host 上线，3 分钟超时，每 2 秒轮询一次
  const handleWaitOnline = useCallback(async (hostName: string): Promise<boolean> => {
    abortRef.current?.abort();
    const ac = new AbortController();
    abortRef.current = ac;

    const maxWaitMs = 3 * 60 * 1000;
    const pollIntervalMs = 2000;
    const startTime = Date.now();

    while (Date.now() - startTime < maxWaitMs) {
      if (ac.signal.aborted) return false;
      try {
        const res = await controllerApi.getHost(hostName);
        if (res.status === 'ok' && res.data?.status === 'online') {
          setRefreshTrigger((n) => n + 1);
          return true;
        }
      } catch {
        // 节点可能还未注册，继续等待
      }
      await new Promise<void>((resolve) => {
        const timer = setTimeout(resolve, pollIntervalMs);
        ac.signal.addEventListener(
          'abort',
          () => {
            clearTimeout(timer);
            resolve();
          },
          { once: true },
        );
      });
    }

    setRefreshTrigger((n) => n + 1);
    return false;
  }, []);

  const handleViewLog = useCallback((hostName: string) => {
    setLogPanelHost(hostName);
  }, []);

  const handleFetchBuildLog = useCallback(async (): Promise<string> => {
    if (!logPanelHost) return '';
    const res = await controllerApi.getHostBuildLog(logPanelHost);
    return res.status === 'ok' && res.data ? res.data : '';
  }, [logPanelHost]);

  const handleFetchRuntimeLog = useCallback(async (): Promise<string> => {
    if (!logPanelHost) return '';
    const res = await controllerApi.getHostRuntimeLog(logPanelHost);
    return res.status === 'ok' && res.data ? res.data : '';
  }, [logPanelHost]);

  if (isBooting) {
    return (
      <BootSequence
        onComplete={() => {
          setIsBooting(false);
        }}
        division="the-mesa"
      />
    );
  }

  return (
    <>
      <div className="h-screen w-screen bg-background text-foreground dark relative overflow-hidden grid grid-rows-[56px_1fr] grid-cols-[288px_1fr]">
        <CalibrationMarks />

        <AppNavbar
          title="THE MESA // DELOS CONTROL"
          icon={<Activity className="w-5 h-5 text-primary" />}
          leftContent={
            <Button
              variant="ghost"
              size="icon"
              className="md:hidden text-primary hover:text-primary hover:bg-primary/20"
              onClick={() => {
                setSidebarOpen(!sidebarOpen);
              }}
            >
              <Menu className="w-5 h-5" />
            </Button>
          }
          rightContent={
            <Button
              variant="ghost"
              size="icon"
              className="text-primary hover:text-primary hover:bg-primary/20 rounded-none"
              onClick={() => {
                setShowAnimSettings(true);
              }}
              title="Visual Effects"
            >
              <Settings className="w-4 h-4" />
            </Button>
          }
        />

        {/* Pane 1: Sidebar with HOSTS + FABRICATION */}
        <div className={`row-start-2 ${!sidebarOpen ? 'hidden md:flex' : 'flex'}`}>
          <Sidebar
            selectedHostName={selectedHost?.name || null}
            onSelectHost={handleSelectNode}
            onHostsChange={handleNodesChange}
            refreshTrigger={refreshTrigger}
            onViewLog={handleViewLog}
            selectedFabricationItem={selectedFabricationItem}
            onSelectFabricationItem={setSelectedFabricationItem}
            onOpenCreateHost={handleOpenCreateHost}
            radarNodes={radarNodes}
          />
        </div>

        {/* Pane 2 & 3: Main Content */}
        <div
          className={`row-start-2 flex w-full h-full relative z-10 bg-background overflow-hidden ${!sidebarOpen ? 'col-span-2' : ''}`}
        >
          {selectedFabricationItem ? (
            <FabricationPanel item={selectedFabricationItem} />
          ) : agentApi ? (
            <AgentPanel
              apiClient={agentApi}
              nodeName={selectedHost!.name}
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
      <CreateHostDialog
        open={showCreateHost}
        onOpenChange={setShowCreateHost}
        tools={availableTools}
        skills={availableSkills}
        mcpServers={availableMcpServers}
        onSubmit={handleCreateHost}
        onWaitOnline={handleWaitOnline}
        getHostBuildLog={async (name: string) => {
          const res = await controllerApi.getHostBuildLog(name);
          return res.status === 'ok' && res.data ? res.data : '';
        }}
      />
      <HostLogPanel
        open={!!logPanelHost}
        onOpenChange={(v) => {
          if (!v) setLogPanelHost(null);
        }}
        hostName={logPanelHost || ''}
        fetchBuildLog={handleFetchBuildLog}
        fetchRuntimeLog={handleFetchRuntimeLog}
      />
      <AnimationSettingsPanel open={showAnimSettings} onOpenChange={setShowAnimSettings} />
    </>
  );
}

function App() {
  return (
    <AppShell requireAuth>
      <AppContent />
    </AppShell>
  );
}

export default App;
