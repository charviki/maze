import type {
  IAgentApiClient,
  ApiResponse,
  V1SessionTemplate,
  NormalizedTemplate,
} from '@maze/fabrication';
import {
  SessionServiceApi,
  TemplateServiceApi,
  ConfigServiceApi,
  createSdkConfiguration,
  unwrapSdkResponse,
  unwrapVoidResponse,
  normalizeTemplate,
} from '@maze/fabrication';

export function createAgentApi(managerBase: string, nodeName: string): IAgentApiClient {
  const config = createSdkConfiguration(managerBase);
  const sessionApi = new SessionServiceApi(config);
  const templateApi = new TemplateServiceApi(config);
  const configApi = new ConfigServiceApi(config);

  function normalizeList(result: V1SessionTemplate[] | undefined): NormalizedTemplate[] {
    return (result || []).map(normalizeTemplate);
  }

  function normalizeSingle(res: ApiResponse<V1SessionTemplate>): ApiResponse<NormalizedTemplate> {
    if (res.status === 'ok' && res.data) {
      return { ...res, data: normalizeTemplate(res.data) };
    }
    return res as ApiResponse<NormalizedTemplate>;
  }

  return {
    listSessions: async () => {
      const res = await unwrapSdkResponse(sessionApi.sessionServiceListSessions({ nodeName }));
      return { ...res, data: res.data?.sessions };
    },

    createSession: (data) =>
      unwrapSdkResponse(sessionApi.sessionServiceCreateSession({ nodeName, body: data })),

    getSession: (id) => unwrapSdkResponse(sessionApi.sessionServiceGetSession({ nodeName, id })),

    deleteSession: (id) =>
      unwrapVoidResponse(sessionApi.sessionServiceDeleteSession({ nodeName, id })),

    getOutput: (id, lines = 50) =>
      unwrapSdkResponse(sessionApi.sessionServiceGetOutput({ nodeName, id, lines })),

    sendInput: (id, command) =>
      unwrapVoidResponse(sessionApi.sessionServiceSendInput({ nodeName, id, body: { command } })),

    sendSignal: (id, signal) =>
      unwrapVoidResponse(sessionApi.sessionServiceSendSignal({ nodeName, id, body: { signal } })),

    getSavedSessions: async () => {
      const res = await unwrapSdkResponse(sessionApi.sessionServiceGetSavedSessions({ nodeName }));
      return { ...res, data: res.data?.sessions };
    },

    restoreSession: (id) =>
      unwrapVoidResponse(sessionApi.sessionServiceRestoreSession({ nodeName, id, body: {} })),

    saveSessions: () => unwrapSdkResponse(sessionApi.sessionServiceSaveSessions({ nodeName })),

    buildWsUrl: (sessionId) => {
      const loc = window.location;
      const protocol = loc.protocol === 'https:' ? 'wss:' : 'ws:';
      return `${protocol}//${loc.host}/api/v1/nodes/${encodeURIComponent(nodeName)}/sessions/${sessionId}/ws`;
    },

    listTemplates: async () => {
      const res = await unwrapSdkResponse(templateApi.templateServiceListTemplates({ nodeName }));
      return { ...res, data: normalizeList(res.data?.templates) };
    },

    createTemplate: async (tpl) => {
      const res = await unwrapSdkResponse(
        templateApi.templateServiceCreateTemplate({ nodeName, body: { template: tpl } }),
      );
      return normalizeSingle(res);
    },

    getTemplate: async (id) => {
      const res = await unwrapSdkResponse(templateApi.templateServiceGetTemplate({ nodeName, id }));
      return normalizeSingle(res);
    },

    updateTemplate: async (id, tpl) => {
      const res = await unwrapSdkResponse(
        templateApi.templateServiceUpdateTemplate({ nodeName, id, body: { template: tpl } }),
      );
      return normalizeSingle(res);
    },

    deleteTemplate: (id) =>
      unwrapVoidResponse(templateApi.templateServiceDeleteTemplate({ nodeName, id })),

    getTemplateConfig: (id) =>
      unwrapSdkResponse(templateApi.templateServiceGetTemplateConfig({ nodeName, id })),

    updateTemplateConfig: (id, req) =>
      unwrapSdkResponse(
        templateApi.templateServiceUpdateTemplateConfig({ nodeName, id, body: req }),
      ),

    getSessionConfig: (id) =>
      unwrapSdkResponse(sessionApi.sessionServiceGetSessionConfig({ nodeName, id })),

    updateSessionConfig: (id, req) =>
      unwrapSdkResponse(sessionApi.sessionServiceUpdateSessionConfig({ nodeName, id, body: req })),

    getLocalConfig: () => unwrapSdkResponse(configApi.configServiceGetConfig({ nodeName })),

    updateLocalConfig: (cfg) =>
      unwrapSdkResponse(configApi.configServiceUpdateConfig({ nodeName, body: cfg })),
  };
}
