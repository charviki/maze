import type { ApiResponse, SaveConfigRequest } from './types';
import type {
  V1Session,
  V1CreateSessionRequest,
  V1TerminalOutput,
  V1SessionState,
  V1SaveSessionsResponse,
  V1SessionTemplate,
  V1TemplateConfigView,
  V1SessionConfigView,
  V1LocalAgentConfig,
} from './api/gen/models';
import type { NormalizedTemplate } from './api/normalize';
import { createRequest } from './utils/request';
import { normalizeTemplate } from './api/normalize';

export interface ISessionApi {
  listSessions(): Promise<ApiResponse<V1Session[]>>;
  createSession(data: V1CreateSessionRequest): Promise<ApiResponse<V1Session>>;
  getSession(id: string): Promise<ApiResponse<V1Session>>;
  deleteSession(id: string): Promise<ApiResponse<void>>;
  getOutput(id: string, lines?: number): Promise<ApiResponse<V1TerminalOutput>>;
  sendInput(id: string, command: string): Promise<ApiResponse<void>>;
  sendSignal(id: string, signal: string): Promise<ApiResponse<void>>;
  getSavedSessions(): Promise<ApiResponse<V1SessionState[]>>;
  restoreSession(id: string): Promise<ApiResponse<void>>;
  saveSessions(): Promise<ApiResponse<V1SaveSessionsResponse>>;
  buildWsUrl(sessionId: string): string;
}

export interface ITemplateApi {
  listTemplates(): Promise<ApiResponse<NormalizedTemplate[]>>;
  createTemplate(tpl: V1SessionTemplate): Promise<ApiResponse<NormalizedTemplate>>;
  getTemplate(id: string): Promise<ApiResponse<NormalizedTemplate>>;
  updateTemplate(id: string, tpl: V1SessionTemplate): Promise<ApiResponse<NormalizedTemplate>>;
  deleteTemplate(id: string): Promise<ApiResponse<void>>;
  getTemplateConfig(id: string): Promise<ApiResponse<V1TemplateConfigView>>;
  updateTemplateConfig(
    id: string,
    req: SaveConfigRequest,
  ): Promise<ApiResponse<V1TemplateConfigView>>;
}

export interface IConfigApi {
  getSessionConfig(id: string): Promise<ApiResponse<V1SessionConfigView>>;
  updateSessionConfig(
    id: string,
    req: SaveConfigRequest,
  ): Promise<ApiResponse<V1SessionConfigView>>;
}

export interface ILocalConfigApi {
  getLocalConfig(): Promise<ApiResponse<V1LocalAgentConfig>>;
  updateLocalConfig(cfg: Partial<V1LocalAgentConfig>): Promise<ApiResponse<V1LocalAgentConfig>>;
}

export type IAgentApiClient = ISessionApi & ITemplateApi & IConfigApi & ILocalConfigApi;

function normalizeList(result: V1SessionTemplate[] | undefined): NormalizedTemplate[] {
  return (result || []).map(normalizeTemplate);
}

function normalizeSingle(res: ApiResponse<V1SessionTemplate>): ApiResponse<NormalizedTemplate> {
  if (res.status === 'ok' && res.data) {
    return { ...res, data: normalizeTemplate(res.data) };
  }
  return res as ApiResponse<NormalizedTemplate>;
}

export function createAgentApiClient(urlPrefix: string): IAgentApiClient {
  // urlPrefix 决定所有 API 路径的起始部分，例如：
  // - 直连模式: '/api/v1/'
  // - 代理模式: '/api/v1/nodes/{nodeName}/'
  const prefix = urlPrefix;
  const request = createRequest();

  return {
    listSessions: async () => {
      const res = await request<{ sessions?: V1Session[] }>(`${prefix}sessions`);
      return { ...res, data: res.data?.sessions };
    },

    createSession: (data) =>
      request<V1Session>(`${prefix}sessions`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),

    getSession: (id) => request<V1Session>(`${prefix}sessions/${encodeURIComponent(id)}`),

    deleteSession: (id) =>
      request<void>(`${prefix}sessions/${encodeURIComponent(id)}`, { method: 'DELETE' }),

    getOutput: (id, lines = 50) =>
      request<V1TerminalOutput>(
        `${prefix}sessions/${encodeURIComponent(id)}/output?lines=${lines}`,
      ),

    sendInput: (id, command) =>
      request<void>(`${prefix}sessions/${encodeURIComponent(id)}/input`, {
        method: 'POST',
        body: JSON.stringify({ command }),
      }),

    sendSignal: (id, signal) =>
      request<void>(`${prefix}sessions/${encodeURIComponent(id)}/signal`, {
        method: 'POST',
        body: JSON.stringify({ signal }),
      }),

    getSavedSessions: async () => {
      const res = await request<{ sessions?: V1SessionState[] }>(`${prefix}sessions/saved`);
      return { ...res, data: res.data?.sessions };
    },

    restoreSession: (id) =>
      request<void>(`${prefix}sessions/${encodeURIComponent(id)}/restore`, {
        method: 'POST',
        body: JSON.stringify({}),
      }),

    saveSessions: () =>
      request<V1SaveSessionsResponse>(`${prefix}sessions/save`, { method: 'POST' }),

    buildWsUrl: (sessionId) => {
      const loc = window.location;
      const protocol = loc.protocol === 'https:' ? 'wss:' : 'ws:';
      return `${protocol}//${loc.host}${prefix}sessions/${encodeURIComponent(sessionId)}/ws`;
    },

    listTemplates: async () => {
      const res = await request<{ templates?: V1SessionTemplate[] }>(`${prefix}templates`);
      return { ...res, data: normalizeList(res.data?.templates) };
    },

    createTemplate: async (tpl) => {
      const res = await request<V1SessionTemplate>(`${prefix}templates`, {
        method: 'POST',
        body: JSON.stringify({ template: tpl }),
      });
      return normalizeSingle(res);
    },

    getTemplate: async (id) => {
      const res = await request<V1SessionTemplate>(`${prefix}templates/${encodeURIComponent(id)}`);
      return normalizeSingle(res);
    },

    updateTemplate: async (id, tpl) => {
      const res = await request<V1SessionTemplate>(`${prefix}templates/${encodeURIComponent(id)}`, {
        method: 'PUT',
        body: JSON.stringify({ template: tpl }),
      });
      return normalizeSingle(res);
    },

    deleteTemplate: (id) =>
      request<void>(`${prefix}templates/${encodeURIComponent(id)}`, { method: 'DELETE' }),

    getTemplateConfig: (id) =>
      request<V1TemplateConfigView>(`${prefix}templates/${encodeURIComponent(id)}/config`),

    updateTemplateConfig: (id, req) =>
      request<V1TemplateConfigView>(`${prefix}templates/${encodeURIComponent(id)}/config`, {
        method: 'PUT',
        body: JSON.stringify(req),
      }),

    getSessionConfig: (id) =>
      request<V1SessionConfigView>(`${prefix}sessions/${encodeURIComponent(id)}/config`),

    updateSessionConfig: (id, req) =>
      request<V1SessionConfigView>(`${prefix}sessions/${encodeURIComponent(id)}/config`, {
        method: 'PUT',
        body: JSON.stringify(req),
      }),

    getLocalConfig: () => request<V1LocalAgentConfig>(`${prefix}config`),

    updateLocalConfig: (cfg) =>
      request<V1LocalAgentConfig>(`${prefix}config`, {
        method: 'PUT',
        body: JSON.stringify(cfg),
      }),
  };
}
