import { useEffect, useReducer, useCallback, type ReactNode } from 'react';
import type { Session, SavedSession, PipelineStep } from '../../types';
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

interface AgentState {
  sessions: Session[];
  savedSessions: SavedSession[];
  search: string;
  selectedSessionId: string | null;
  killTarget: SessionDisplay | null;
  restoreTarget: SessionDisplay | null;
  killing: boolean;
  restoring: boolean;
  viewPipelineSession: SessionDisplay | null;
  viewPipelineSteps: PipelineStep[];
  configOpen: boolean;
  loadingConfig: boolean;
  saving: boolean;
  saveCooldown: boolean;
  lastSaveTime: string | null;
  actionError: string | null;
  showCreate: boolean;
  showTemplateManager: boolean;
  showNodeConfig: boolean;
}

type AgentAction =
  | { type: 'SET_SESSIONS'; payload: Session[] }
  | { type: 'SET_SAVED_SESSIONS'; payload: SavedSession[] }
  | { type: 'SET_SEARCH'; payload: string }
  | { type: 'SELECT_SESSION'; payload: string | null }
  | { type: 'SET_KILL_TARGET'; payload: SessionDisplay | null }
  | { type: 'SET_KILLING'; payload: boolean }
  | { type: 'SET_RESTORE_TARGET'; payload: SessionDisplay | null }
  | { type: 'SET_RESTORING'; payload: boolean }
  | {
      type: 'SET_VIEW_PIPELINE';
      payload: { session: SessionDisplay | null; steps: PipelineStep[] };
    }
  | { type: 'SET_CONFIG_OPEN'; payload: boolean }
  | { type: 'SET_LOADING_CONFIG'; payload: boolean }
  | { type: 'SET_SAVING'; payload: boolean }
  | { type: 'SET_SAVE_COOLDOWN'; payload: boolean }
  | { type: 'SET_LAST_SAVE_TIME'; payload: string | null }
  | { type: 'SET_ACTION_ERROR'; payload: string | null }
  | { type: 'SET_SHOW_CREATE'; payload: boolean }
  | { type: 'SET_SHOW_TEMPLATE_MANAGER'; payload: boolean }
  | { type: 'SET_SHOW_NODE_CONFIG'; payload: boolean };

const initialState: AgentState = {
  sessions: [],
  savedSessions: [],
  search: '',
  selectedSessionId: null,
  killTarget: null,
  killing: false,
  restoreTarget: null,
  restoring: false,
  viewPipelineSession: null,
  viewPipelineSteps: [],
  configOpen: false,
  loadingConfig: false,
  saving: false,
  saveCooldown: false,
  lastSaveTime: null,
  actionError: null,
  showCreate: false,
  showTemplateManager: false,
  showNodeConfig: false,
};

function agentReducer(state: AgentState, action: AgentAction): AgentState {
  switch (action.type) {
    case 'SET_SESSIONS':
      return { ...state, sessions: action.payload };
    case 'SET_SAVED_SESSIONS':
      return { ...state, savedSessions: action.payload };
    case 'SET_SEARCH':
      return { ...state, search: action.payload };
    case 'SELECT_SESSION':
      return { ...state, selectedSessionId: action.payload };
    case 'SET_KILL_TARGET':
      return { ...state, killTarget: action.payload };
    case 'SET_KILLING':
      return { ...state, killing: action.payload };
    case 'SET_RESTORE_TARGET':
      return { ...state, restoreTarget: action.payload };
    case 'SET_RESTORING':
      return { ...state, restoring: action.payload };
    case 'SET_VIEW_PIPELINE':
      return {
        ...state,
        viewPipelineSession: action.payload.session,
        viewPipelineSteps: action.payload.steps,
      };
    case 'SET_CONFIG_OPEN':
      return { ...state, configOpen: action.payload };
    case 'SET_LOADING_CONFIG':
      return { ...state, loadingConfig: action.payload };
    case 'SET_SAVING':
      return { ...state, saving: action.payload };
    case 'SET_SAVE_COOLDOWN':
      return { ...state, saveCooldown: action.payload };
    case 'SET_LAST_SAVE_TIME':
      return { ...state, lastSaveTime: action.payload };
    case 'SET_ACTION_ERROR':
      return { ...state, actionError: action.payload };
    case 'SET_SHOW_CREATE':
      return { ...state, showCreate: action.payload };
    case 'SET_SHOW_TEMPLATE_MANAGER':
      return { ...state, showTemplateManager: action.payload };
    case 'SET_SHOW_NODE_CONFIG':
      return { ...state, showNodeConfig: action.payload };
    default:
      return state;
  }
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
  const [state, dispatch] = useReducer(agentReducer, initialState);
  const { showToast } = useToast();

  const fetchSessions = useCallback(async () => {
    const res = await apiClient.listSessions();
    if (res.status === 'ok' && res.data) dispatch({ type: 'SET_SESSIONS', payload: res.data });
  }, [apiClient]);

  const fetchSavedSessions = useCallback(async () => {
    const res = await apiClient.getSavedSessions();
    if (res.status === 'ok' && res.data) {
      const valid = res.data.filter((ss) => ss.session_name);
      dispatch({ type: 'SET_SAVED_SESSIONS', payload: valid });
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
    if (state.selectedSessionId) {
      apiClient
        .getSavedSessions()
        .then((res) => {
          if (res.status === 'ok' && res.data) {
            const valid = res.data.filter((ss) => ss.session_name);
            const found = valid.find((s) => s.session_name === state.selectedSessionId);
            if (found) {
              dispatch({ type: 'SET_LAST_SAVE_TIME', payload: found.saved_at });
            }
          }
        })
        .catch(() => {});
    }
  }, [state.selectedSessionId, apiClient]);

  // 合并运行中和已保存的会话列表
  const mergedSessions: SessionDisplay[] = (() => {
    const runningSet = new Set(state.sessions.map((s) => s.name));
    const result: SessionDisplay[] = [];

    for (const s of state.sessions) {
      result.push({
        id: s.id,
        name: s.name,
        status: 'running',
        created_at: s.created_at,
        window_count: s.window_count,
      });
    }

    for (const ss of state.savedSessions) {
      if (!ss.session_name) continue;
      if (!runningSet.has(ss.session_name) && ss.restore_strategy !== 'running') {
        result.push({
          id: ss.session_name,
          name: ss.session_name,
          status: 'saved',
          created_at: ss.saved_at,
          window_count: 0,
          saved_at: ss.saved_at,
          terminal_snapshot: ss.terminal_snapshot,
        });
      }
    }

    return result;
  })();

  const confirmKill = async () => {
    if (!state.killTarget || !state.killTarget.name) return;
    const s = state.killTarget;
    dispatch({ type: 'SET_KILLING', payload: true });
    dispatch({ type: 'SET_ACTION_ERROR', payload: null });

    // saved session 也需要调用 delete 以清理后端状态文件
    const res = await apiClient.deleteSession(s.name);

    dispatch({ type: 'SET_KILLING', payload: false });
    if (res.status === 'ok') {
      dispatch({ type: 'SET_KILL_TARGET', payload: null });
      if (state.selectedSessionId === s.id) {
        dispatch({ type: 'SELECT_SESSION', payload: null });
      }
      showToast('success', `已删除 Loop: ${s.name}`);
      await refreshAll();
      return;
    }

    const message = res.message || '删除失败';
    dispatch({ type: 'SET_ACTION_ERROR', payload: message });
    showToast('error', `删除 Loop 失败: ${s.name}`);
    setTimeout(() => {
      dispatch({ type: 'SET_ACTION_ERROR', payload: null });
    }, 3000);
  };

  const confirmRestore = async () => {
    if (!state.restoreTarget) return;
    const targetName = state.restoreTarget.name;
    dispatch({ type: 'SET_RESTORING', payload: true });
    await apiClient.restoreSession(targetName);
    dispatch({ type: 'SET_RESTORING', payload: false });
    dispatch({ type: 'SET_RESTORE_TARGET', payload: null });
    void refreshAll();
    dispatch({ type: 'SELECT_SESSION', payload: targetName });
  };

  const handleViewPipeline = (e: React.MouseEvent, session: SessionDisplay) => {
    e.stopPropagation();
    const saved = state.savedSessions.find((s) => s.session_name === session.id);
    dispatch({
      type: 'SET_VIEW_PIPELINE',
      payload: { session, steps: saved?.pipeline || [] },
    });
  };

  const handleSaveSessions = async () => {
    dispatch({ type: 'SET_SAVING', payload: true });
    dispatch({ type: 'SET_ACTION_ERROR', payload: null });
    const res = await apiClient.saveSessions();
    dispatch({ type: 'SET_SAVING', payload: false });
    if (res.status === 'ok' && res.data) {
      dispatch({ type: 'SET_LAST_SAVE_TIME', payload: res.data.saved_at });
      dispatch({ type: 'SET_SAVE_COOLDOWN', payload: true });
      setTimeout(() => {
        dispatch({ type: 'SET_SAVE_COOLDOWN', payload: false });
      }, 3000);
    } else {
      dispatch({ type: 'SET_ACTION_ERROR', payload: res.message || '保存失败' });
      setTimeout(() => {
        dispatch({ type: 'SET_ACTION_ERROR', payload: null });
      }, 3000);
    }
  };

  const handleOpenConfig = (open: boolean) => {
    if (!open) {
      dispatch({ type: 'SET_CONFIG_OPEN', payload: false });
      return;
    }
    dispatch({ type: 'SET_CONFIG_OPEN', payload: true });
  };

  const handleSessionCreated = (sessionName: string) => {
    void refreshAll();
    dispatch({ type: 'SELECT_SESSION', payload: sessionName });
  };

  return (
    <>
      <SessionList
        sessions={mergedSessions}
        search={state.search}
        onSearchChange={(v) => {
          dispatch({ type: 'SET_SEARCH', payload: v });
        }}
        selectedSessionId={state.selectedSessionId}
        onSelectSession={(id) => {
          dispatch({ type: 'SELECT_SESSION', payload: id });
        }}
        nodeName={nodeName}
        onCreateClick={() => {
          dispatch({ type: 'SET_SHOW_CREATE', payload: true });
        }}
        onNodeConfigClick={() => {
          dispatch({ type: 'SET_SHOW_NODE_CONFIG', payload: true });
        }}
        onKill={(session) => {
          dispatch({ type: 'SET_KILL_TARGET', payload: session });
        }}
        onRestore={(session) => {
          dispatch({ type: 'SET_RESTORE_TARGET', payload: session });
        }}
        onViewPipeline={handleViewPipeline}
        listHeaderActions={listHeaderActions}
      />

      <TerminalPane
        selectedSessionId={state.selectedSessionId}
        nodeName={nodeName}
        apiClient={apiClient}
        terminalBackground={terminalBackground}
        lastSaveTime={state.lastSaveTime}
        saving={state.saving}
        saveCooldown={state.saveCooldown}
        actionError={state.actionError}
        onSave={handleSaveSessions}
        onOpenConfig={handleOpenConfig}
        loadingConfig={state.loadingConfig}
        headerActions={headerActions}
      />

      <SessionDialogs
        killTarget={state.killTarget}
        killing={state.killing}
        restoreTarget={state.restoreTarget}
        viewPipelineSession={state.viewPipelineSession}
        viewPipelineSteps={state.viewPipelineSteps}
        configOpen={state.configOpen}
        selectedSessionId={state.selectedSessionId}
        nodeName={nodeName}
        restoring={state.restoring}
        showCreate={state.showCreate}
        onShowCreateChange={(v) => {
          dispatch({ type: 'SET_SHOW_CREATE', payload: v });
        }}
        showTemplateManager={state.showTemplateManager}
        onShowTemplateManagerChange={(v) => {
          dispatch({ type: 'SET_SHOW_TEMPLATE_MANAGER', payload: v });
        }}
        showNodeConfig={state.showNodeConfig}
        onShowNodeConfigClose={() => {
          dispatch({ type: 'SET_SHOW_NODE_CONFIG', payload: false });
        }}
        apiClient={apiClient}
        renderCreateDialog={renderCreateDialog}
        onSessionCreated={handleSessionCreated}
        onKillConfirm={confirmKill}
        onKillCancel={() => {
          dispatch({ type: 'SET_KILL_TARGET', payload: null });
        }}
        onRestoreConfirm={confirmRestore}
        onRestoreCancel={() => {
          dispatch({ type: 'SET_RESTORE_TARGET', payload: null });
        }}
        onPipelineClose={() => {
          dispatch({ type: 'SET_VIEW_PIPELINE', payload: { session: null, steps: [] } });
        }}
        onConfigClose={(open) => {
          dispatch({ type: 'SET_CONFIG_OPEN', payload: open });
        }}
        getSnapshotSummary={getSnapshotSummary}
      />
    </>
  );
}
