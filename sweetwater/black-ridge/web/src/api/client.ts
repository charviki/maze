import type {
  Session,
  CreateSessionRequest,
  TerminalOutput,
  SavedSession,
  SessionTemplate,
  LocalAgentConfig,
  IAgentApiClient,
  SaveConfigRequest,
  TemplateConfigView,
  SessionConfigView,
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

const request = createRequest();

export function createAgentApi(): IAgentApiClient {
  const base = `/api/v1/sessions`;

  return {
    listSessions: async () => {
      const res = await request<{ sessions: Session[] }>(base);
      return { ...res, data: res.data?.sessions };
    },

    createSession: (data: CreateSessionRequest) =>
      request<Session>(base, {
        method: 'POST',
        body: JSON.stringify(data),
      }),

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
      request<void>(`${base}/${id}/input`, {
        method: 'POST',
        body: JSON.stringify({ command }),
      }),

    sendSignal: (id: string, signal: string) =>
      request<void>(`${base}/${id}/signal`, {
        method: 'POST',
        body: JSON.stringify({ signal }),
      }),

    getSavedSessions: async () => {
      const res = await request<{ sessions: SavedSession[] }>(`${base}/saved`);
      return { ...res, data: res.data?.sessions };
    },

    restoreSession: (id: string) => request<void>(`${base}/${id}/restore`, { method: 'POST' }),

    saveSessions: () => request<{ savedAt: string }>(`${base}/save`, { method: 'POST' }),

    buildWsUrl: (sessionId: string) => {
      const loc = window.location;
      const protocol = loc.protocol === 'https:' ? 'wss:' : 'ws:';
      return `${protocol}//${loc.host}/api/v1/sessions/${sessionId}/ws`;
    },

    listTemplates: async () => {
      const res = await request<{ templates: SessionTemplate[] }>('/api/v1/templates');
      return { ...res, data: res.data?.templates?.map(normalizeTemplate) };
    },
    createTemplate: (tpl: SessionTemplate) =>
      request<SessionTemplate>('/api/v1/templates', {
        method: 'POST',
        body: JSON.stringify(tpl),
      }).then((res) => ({ ...res, data: res.data ? normalizeTemplate(res.data) : undefined })),
    getTemplate: (id: string) =>
      request<SessionTemplate>(`/api/v1/templates/${id}`).then((res) => ({
        ...res,
        data: res.data ? normalizeTemplate(res.data) : undefined,
      })),
    getTemplateConfig: (id: string) =>
      request<TemplateConfigView>(`/api/v1/templates/${id}/config`),
    updateTemplate: (id: string, tpl: SessionTemplate) =>
      request<SessionTemplate>(`/api/v1/templates/${id}`, {
        method: 'PUT',
        body: JSON.stringify(tpl),
      }).then((res) => ({ ...res, data: res.data ? normalizeTemplate(res.data) : undefined })),
    updateTemplateConfig: (id: string, req: SaveConfigRequest) =>
      request<TemplateConfigView>(`/api/v1/templates/${id}/config`, {
        method: 'PUT',
        body: JSON.stringify(req),
      }),
    deleteTemplate: (id: string) => request<void>(`/api/v1/templates/${id}`, { method: 'DELETE' }),

    getLocalConfig: () => request<LocalAgentConfig>('/api/v1/local-config'),
    updateLocalConfig: (cfg: Partial<LocalAgentConfig>) =>
      request<LocalAgentConfig>('/api/v1/local-config', {
        method: 'PUT',
        body: JSON.stringify(cfg),
      }),
  };
}

export const api = createAgentApi();
