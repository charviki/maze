import type {
  Session,
  CreateSessionRequest,
  TerminalOutput,
  SavedSession,
  SessionTemplate,
  LocalAgentConfig,
  IAgentApiClient,
  SaveConfigRequest,
  SessionConfigView,
  TemplateConfigView,
} from '@maze/fabrication';
import { createRequest } from '@maze/fabrication';

const emptySchema = { envDefs: [], fileDefs: [] };
const emptyDefaults = { env: {}, files: [] };

function normalizeTemplate(tpl: SessionTemplate): SessionTemplate {
  return {
    ...tpl,
    defaults: tpl.defaults || emptyDefaults,
    sessionSchema: tpl.sessionSchema || emptySchema,
  };
}

export function createAgentApi(managerBase: string, nodeName: string): IAgentApiClient {
  const nodeBase = `${managerBase}/api/v1/nodes/${encodeURIComponent(nodeName)}`;
  const base = `${nodeBase}/sessions`;
  const request = createRequest();

  return {
    listSessions: async () => {
      const res = await request<{ sessions: Session[] }>(base);
      return { ...res, data: res.data?.sessions };
    },

    createSession: (data: CreateSessionRequest) =>
      request<Session>(base, { method: 'POST', body: JSON.stringify(data) }),

    getSession: (id: string) => request<Session>(`${base}/${id}`),

    getSessionConfig: (id: string) => request<SessionConfigView>(`${base}/${id}/config`),

    deleteSession: (id: string) => request<void>(`${base}/${id}`, { method: 'DELETE' }),

    updateSessionConfig: (id: string, req: SaveConfigRequest) =>
      request<SessionConfigView>(`${base}/${id}/config`, {
        method: 'PUT',
        body: JSON.stringify(req),
      }),

    getOutput: (id: string, lines = 50) =>
      request<TerminalOutput>(`${base}/${id}/output?lines=${lines}`),

    sendInput: (id: string, command: string) =>
      request<void>(`${base}/${id}/input`, { method: 'POST', body: JSON.stringify({ command }) }),

    sendSignal: (id: string, signal: string) =>
      request<void>(`${base}/${id}/signal`, { method: 'POST', body: JSON.stringify({ signal }) }),

    getSavedSessions: async () => {
      const res = await request<{ sessions: SavedSession[] }>(`${base}/saved`);
      return { ...res, data: res.data?.sessions };
    },

    restoreSession: (id: string) => request<void>(`${base}/${id}/restore`, { method: 'POST' }),

    saveSessions: () => request<{ savedAt: string }>(`${base}/save`, { method: 'POST' }),

    buildWsUrl: (sessionId: string) => {
      const loc = window.location;
      const protocol = loc.protocol === 'https:' ? 'wss:' : 'ws:';
      return `${protocol}//${loc.host}/api/v1/nodes/${encodeURIComponent(nodeName)}/sessions/${sessionId}/ws`;
    },

    listTemplates: async () => {
      const res = await request<{ templates: SessionTemplate[] }>(`${nodeBase}/templates`);
      return { ...res, data: res.data?.templates?.map(normalizeTemplate) };
    },
    createTemplate: (tpl: SessionTemplate) =>
      request<SessionTemplate>(`${nodeBase}/templates`, {
        method: 'POST',
        body: JSON.stringify(tpl),
      }).then((res) => ({ ...res, data: res.data ? normalizeTemplate(res.data) : undefined })),
    getTemplate: (id: string) =>
      request<SessionTemplate>(`${nodeBase}/templates/${id}`).then((res) => ({
        ...res,
        data: res.data ? normalizeTemplate(res.data) : undefined,
      })),
    getTemplateConfig: (id: string) =>
      request<TemplateConfigView>(`${nodeBase}/templates/${id}/config`),
    updateTemplate: (id: string, tpl: SessionTemplate) =>
      request<SessionTemplate>(`${nodeBase}/templates/${id}`, {
        method: 'PUT',
        body: JSON.stringify(tpl),
      }).then((res) => ({ ...res, data: res.data ? normalizeTemplate(res.data) : undefined })),
    updateTemplateConfig: (id: string, req: SaveConfigRequest) =>
      request<TemplateConfigView>(`${nodeBase}/templates/${id}/config`, {
        method: 'PUT',
        body: JSON.stringify(req),
      }),
    deleteTemplate: (id: string) =>
      request<void>(`${nodeBase}/templates/${id}`, { method: 'DELETE' }),

    getLocalConfig: () => request<LocalAgentConfig>(`${nodeBase}/local-config`),
    updateLocalConfig: (cfg: Partial<LocalAgentConfig>) =>
      request<LocalAgentConfig>(`${nodeBase}/local-config`, {
        method: 'PUT',
        body: JSON.stringify(cfg),
      }),
  };
}
