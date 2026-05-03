import { useEffect, useCallback, useRef, useState, type ReactNode } from 'react';
import type { Session, V1SessionState, PipelineStep } from '../../types';
import type { IAgentApiClient } from '../../api';
import { SessionList, type SessionDisplay } from './SessionList';
import { TerminalPane } from './TerminalPane';
import { SessionDialogs } from './SessionDialogs';
import { usePollingWithBackoff } from '../../hooks/usePollingWithBackoff';
import { useToast } from '../ui/Toast';

export type { SessionDisplay } from './SessionList';

export interface AgentPanelProps {
  apiClient: IAgentApiClient;
  nodeName?: string;
  renderCreateDialog?: (props: {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onSuccess: (sessionName: string) => void;
  }) => ReactNode;
  headerActions?: ReactNode;
  listHeaderActions?: ReactNode;
  terminalBackground?: ReactNode;
}

function getSnapshotSummary(snapshot?: string, lines = 5): string {
  if (!snapshot) return '(无快照)';
  const allLines = snapshot.split('\n').filter((l) => l.trim() !== '');
  return allLines.slice(-lines).join('\n');
}

export function AgentPanel({
  apiClient,
  nodeName = 'Local',
  renderCreateDialog,
  headerActions,
  listHeaderActions,
  terminalBackground,
}: AgentPanelProps) {
  const { showToast } = useToast();

  // 独立状态：每个字段用 useState 管理，无需 reducer 样板代码
  const [sessions, setSessions] = useState<Session[]>([]);
  const [savedSessions, setSavedSessions] = useState<V1SessionState[]>([]);
  const [search, setSearch] = useState('');
  const [selectedSessionId, setSelectedSessionId] = useState<string | null>(null);
  const [killTarget, setKillTarget] = useState<SessionDisplay | null>(null);
  const [restoreTarget, setRestoreTarget] = useState<SessionDisplay | null>(null);
  const [killing, setKilling] = useState(false);
  const [restoring, setRestoring] = useState(false);
  const [viewPipelineSession, setViewPipelineSession] = useState<SessionDisplay | null>(null);
  const [viewPipelineSteps, setViewPipelineSteps] = useState<PipelineStep[]>([]);
  const [configOpen, setConfigOpen] = useState(false);
  const loadingConfig = false;
  const [saving, setSaving] = useState(false);
  const [saveCooldown, setSaveCooldown] = useState(false);
  const [lastSaveTime, setLastSaveTime] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [showTemplateManager, setShowTemplateManager] = useState(false);
  const [showNodeConfig, setShowNodeConfig] = useState(false);

  // 用于清理错误/冷却状态的定时器，防止组件卸载后仍触发 setState
  const errorTimerRef = useRef<number | undefined>(undefined);

  // 组件卸载时清理所有未完成的定时器
  useEffect(() => {
    return () => {
      if (errorTimerRef.current !== undefined) {
        clearTimeout(errorTimerRef.current);
      }
    };
  }, []);

  // 安全清除错误：先取消旧定时器再设置新定时器
  const scheduleErrorClear = useCallback(() => {
    if (errorTimerRef.current !== undefined) clearTimeout(errorTimerRef.current);
    errorTimerRef.current = window.setTimeout(() => {
      setActionError(null);
    }, 3000);
  }, []);

  const fetchSessions = useCallback(async () => {
    const res = await apiClient.listSessions();
    if (res.status === 'ok' && res.data) setSessions(res.data);
  }, [apiClient]);

  const fetchSavedSessions = useCallback(async () => {
    const res = await apiClient.getSavedSessions();
    if (res.status === 'ok' && res.data) {
      const valid = res.data.filter((ss) => ss.sessionName);
      setSavedSessions(valid);
    }
  }, [apiClient]);

  const pollingFetch = useCallback(async () => {
    await fetchSessions();
    await fetchSavedSessions();
    return null;
  }, [fetchSessions, fetchSavedSessions]);

  const { refresh: refreshAll } = usePollingWithBackoff({
    fetchFn: pollingFetch,
  });

  // 选中会话后刷新最后保存时间
  useEffect(() => {
    if (selectedSessionId) {
      apiClient
        .getSavedSessions()
        .then((res) => {
          if (res.status === 'ok' && res.data) {
            const valid = res.data.filter((ss) => ss.sessionName);
            const found = valid.find((s) => s.sessionName === selectedSessionId);
            if (found) {
              setLastSaveTime(found.savedAt ?? null);
            }
          }
        })
        .catch(() => {
          showToast('error', '已保存会话获取失败');
        });
    }
  }, [selectedSessionId, apiClient, showToast]);

  // 合并运行中和已保存的会话列表
  const mergedSessions: SessionDisplay[] = (() => {
    const runningSet = new Set(sessions.map((s) => s.name));
    const result: SessionDisplay[] = [];

    for (const s of sessions) {
      result.push({
        id: s.id ?? '',
        name: s.name ?? '',
        status: 'running',
        createdAt: s.createdAt ?? '',
        windowCount: s.windowCount ?? 0,
      });
    }

    for (const ss of savedSessions) {
      if (!ss.sessionName) continue;
      if (!runningSet.has(ss.sessionName ?? '') && ss.restoreStrategy !== 'running') {
        result.push({
          id: ss.sessionName ?? '',
          name: ss.sessionName ?? '',
          status: 'saved',
          createdAt: ss.savedAt ?? '',
          windowCount: 0,
          savedAt: ss.savedAt,
          terminalSnapshot: ss.terminalSnapshot,
        });
      }
    }

    return result;
  })();

  const confirmKill = async () => {
    if (!killTarget || !killTarget.name) return;
    const s = killTarget;
    setKilling(true);
    setActionError(null);

    // saved session 也需要调用 delete 以清理后端状态文件
    const res = await apiClient.deleteSession(s.name);

    setKilling(false);
    if (res.status === 'ok') {
      setKillTarget(null);
      if (selectedSessionId === s.id) {
        setSelectedSessionId(null);
      }
      showToast('success', `已删除 Loop: ${s.name}`);
      await refreshAll();
      return;
    }

    const message = res.message || '删除失败';
    setActionError(message);
    showToast('error', `删除 Loop 失败: ${s.name}`);
    scheduleErrorClear();
  };

  const confirmRestore = async () => {
    if (!restoreTarget) return;
    const targetName = restoreTarget.name;
    setRestoring(true);
    const res = await apiClient.restoreSession(targetName);
    setRestoring(false);
    if (res.status !== 'ok') {
      showToast('error', `恢复 Loop 失败: ${res.message || '未知错误'}`);
      setRestoreTarget(null);
      return;
    }
    setRestoreTarget(null);
    void refreshAll();
    setSelectedSessionId(targetName);
  };

  const handleViewPipeline = (e: React.MouseEvent, session: SessionDisplay) => {
    e.stopPropagation();
    const saved = savedSessions.find((s) => s.sessionName === session.id);
    let steps: PipelineStep[] = [];
    // JSON.parse 需要 try-catch 保护，防止后端返回格式异常导致运行时崩溃
    if (saved?.pipeline) {
      try {
        steps = JSON.parse(saved.pipeline) as PipelineStep[];
      } catch {
        steps = [];
      }
    }
    setViewPipelineSession(session);
    setViewPipelineSteps(steps);
  };

  const handleSaveSessions = async () => {
    setSaving(true);
    setActionError(null);
    const res = await apiClient.saveSessions();
    setSaving(false);
    if (res.status === 'ok' && res.data) {
      setLastSaveTime(res.data.savedAt ?? null);
      setSaveCooldown(true);
      // 冷却倒计时使用 ref 跟踪，确保卸载时可清理
      if (errorTimerRef.current !== undefined) clearTimeout(errorTimerRef.current);
      errorTimerRef.current = window.setTimeout(() => {
        setSaveCooldown(false);
      }, 3000);
    } else {
      setActionError(res.message || '保存失败');
      scheduleErrorClear();
    }
  };

  const handleOpenConfig = (open: boolean) => {
    if (!open) {
      setConfigOpen(false);
      return;
    }
    setConfigOpen(true);
  };

  const handleSessionCreated = (sessionName: string) => {
    void refreshAll();
    setSelectedSessionId(sessionName);
  };

  return (
    <>
      <SessionList
        sessions={mergedSessions}
        search={search}
        onSearchChange={setSearch}
        selectedSessionId={selectedSessionId}
        onSelectSession={setSelectedSessionId}
        nodeName={nodeName}
        onCreateClick={() => setShowCreate(true)}
        onNodeConfigClick={() => setShowNodeConfig(true)}
        onKill={setKillTarget}
        onRestore={setRestoreTarget}
        onViewPipeline={handleViewPipeline}
        listHeaderActions={listHeaderActions}
      />

      <TerminalPane
        selectedSessionId={selectedSessionId}
        nodeName={nodeName}
        apiClient={apiClient}
        terminalBackground={terminalBackground}
        lastSaveTime={lastSaveTime}
        saving={saving}
        saveCooldown={saveCooldown}
        actionError={actionError}
        onSave={handleSaveSessions}
        onOpenConfig={handleOpenConfig}
        loadingConfig={loadingConfig}
        headerActions={headerActions}
      />

      <SessionDialogs
        killTarget={killTarget}
        killing={killing}
        restoreTarget={restoreTarget}
        viewPipelineSession={viewPipelineSession}
        viewPipelineSteps={viewPipelineSteps}
        configOpen={configOpen}
        selectedSessionId={selectedSessionId}
        nodeName={nodeName}
        restoring={restoring}
        showCreate={showCreate}
        onShowCreateChange={setShowCreate}
        showTemplateManager={showTemplateManager}
        onShowTemplateManagerChange={setShowTemplateManager}
        showNodeConfig={showNodeConfig}
        onShowNodeConfigClose={() => setShowNodeConfig(false)}
        apiClient={apiClient}
        renderCreateDialog={renderCreateDialog}
        onSessionCreated={handleSessionCreated}
        onKillConfirm={confirmKill}
        onKillCancel={() => setKillTarget(null)}
        onRestoreConfirm={confirmRestore}
        onRestoreCancel={() => setRestoreTarget(null)}
        onPipelineClose={() => {
          setViewPipelineSession(null);
          setViewPipelineSteps([]);
        }}
        onConfigClose={setConfigOpen}
        getSnapshotSummary={getSnapshotSummary}
      />
    </>
  );
}
