import type { Session, CreateSessionRequest, TerminalOutput, SavedSession, SessionTemplate, LocalAgentConfig, IAgentApiClient, SaveConfigRequest, SessionConfigView, TemplateConfigView } from '@maze/fabrication';
import { createRequest } from '@maze/fabrication';

/**
 * 创建通过 Manager 代理的 Agent API 客户端。
 * 所有请求发送到 Manager 的 /api/v1/nodes/:name/sessions/* 路径，
 * Manager 代理到目标 Agent，保持可观测性（前端不再直连 Agent）。
 */
export function createAgentApi(managerBase: string, nodeName: string): IAgentApiClient {
  const nodeBase = `${managerBase}/api/v1/nodes/${encodeURIComponent(nodeName)}`;
  const base = `${nodeBase}/sessions`;
  const request = createRequest();

  return {
    listSessions: () => request<Session[]>(base),

    createSession: (data: CreateSessionRequest) =>
      request<Session>(base, { method: 'POST', body: JSON.stringify(data) }),

    getSession: (id: string) => request<Session>(`${base}/${id}`),

    getSessionConfig: (id: string) => request<SessionConfigView>(`${base}/${id}/config`),

    deleteSession: (id: string) => request<void>(`${base}/${id}`, { method: 'DELETE' }),

    updateSessionConfig: (id: string, req: SaveConfigRequest) =>
      request<SessionConfigView>(`${base}/${id}/config`, { method: 'PUT', body: JSON.stringify(req) }),

    getOutput: (id: string, lines: number = 50) =>
      request<TerminalOutput>(`${base}/${id}/output?lines=${lines}`),

    sendInput: (id: string, command: string) =>
      request<void>(`${base}/${id}/input`, { method: 'POST', body: JSON.stringify({ command }) }),

    sendSignal: (id: string, signal: string) =>
      request<void>(`${base}/${id}/signal`, { method: 'POST', body: JSON.stringify({ signal }) }),

    getSavedSessions: () => request<SavedSession[]>(`${base}/saved`),

    restoreSession: (id: string) => request<void>(`${base}/${id}/restore`, { method: 'POST' }),

    saveSessions: () => request<{ saved_at: string }>(`${base}/save`, { method: 'POST' }),

    buildWsUrl: (sessionId: string) => {
      const loc = window.location;
      const protocol = loc.protocol === 'https:' ? 'wss:' : 'ws:';
      return `${protocol}//${loc.host}/api/v1/nodes/${encodeURIComponent(nodeName)}/sessions/${sessionId}/ws`;
    },

    listTemplates: () => request<SessionTemplate[]>(`${nodeBase}/templates`),
    createTemplate: (tpl: SessionTemplate) => request<SessionTemplate>(`${nodeBase}/templates`, { method: 'POST', body: JSON.stringify(tpl) }),
    getTemplate: (id: string) => request<SessionTemplate>(`${nodeBase}/templates/${id}`),
    getTemplateConfig: (id: string) => request<TemplateConfigView>(`${nodeBase}/templates/${id}/config`),
    updateTemplate: (id: string, tpl: SessionTemplate) => request<SessionTemplate>(`${nodeBase}/templates/${id}`, { method: 'PUT', body: JSON.stringify(tpl) }),
    updateTemplateConfig: (id: string, req: SaveConfigRequest) =>
      request<TemplateConfigView>(`${nodeBase}/templates/${id}/config`, { method: 'PUT', body: JSON.stringify(req) }),
    deleteTemplate: (id: string) => request<void>(`${nodeBase}/templates/${id}`, { method: 'DELETE' }),

    getLocalConfig: () => request<LocalAgentConfig>(`${nodeBase}/local-config`),
    updateLocalConfig: (cfg: Partial<LocalAgentConfig>) => request<LocalAgentConfig>(`${nodeBase}/local-config`, { method: 'PUT', body: JSON.stringify(cfg) }),
  };
}
