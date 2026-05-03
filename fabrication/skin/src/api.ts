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
