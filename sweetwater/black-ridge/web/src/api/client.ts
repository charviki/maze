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

export function createAgentApi(): IAgentApiClient {
  const config = createSdkConfiguration('');
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
      const res = await unwrapSdkResponse(sessionApi.sessionServiceListSessions2());
      return { ...res, data: res.data?.sessions };
    },

    createSession: (data) =>
      unwrapSdkResponse(sessionApi.sessionServiceCreateSession2({ body: data })),

    getSession: (id) => unwrapSdkResponse(sessionApi.sessionServiceGetSession2({ id })),

    deleteSession: (id) => unwrapVoidResponse(sessionApi.sessionServiceDeleteSession2({ id })),

    getOutput: (id, lines = 50) =>
      unwrapSdkResponse(sessionApi.sessionServiceGetOutput2({ id, lines })),

    sendInput: (id, command) =>
      unwrapVoidResponse(sessionApi.sessionServiceSendInput2({ id, body: { command } })),

    sendSignal: (id, signal) =>
      unwrapVoidResponse(sessionApi.sessionServiceSendSignal2({ id, body: { signal } })),

    getSavedSessions: async () => {
      const res = await unwrapSdkResponse(sessionApi.sessionServiceGetSavedSessions2());
      return { ...res, data: res.data?.sessions };
    },

    restoreSession: (id) =>
      unwrapVoidResponse(sessionApi.sessionServiceRestoreSession2({ id, body: {} })),

    saveSessions: () => unwrapSdkResponse(sessionApi.sessionServiceSaveSessions2()),

    buildWsUrl: (sessionId) => {
      const loc = window.location;
      const protocol = loc.protocol === 'https:' ? 'wss:' : 'ws:';
      return `${protocol}//${loc.host}/api/v1/sessions/${sessionId}/ws`;
    },

    listTemplates: async () => {
      const res = await unwrapSdkResponse(templateApi.templateServiceListTemplates2());
      return { ...res, data: normalizeList(res.data?.templates) };
    },

    createTemplate: async (tpl) => {
      const res = await unwrapSdkResponse(
        templateApi.templateServiceCreateTemplate2({ body: { nodeName: '', template: tpl } }),
      );
      return normalizeSingle(res);
    },

    getTemplate: async (id) => {
      const res = await unwrapSdkResponse(templateApi.templateServiceGetTemplate2({ id }));
      return normalizeSingle(res);
    },

    updateTemplate: async (id, tpl) => {
      const res = await unwrapSdkResponse(
        templateApi.templateServiceUpdateTemplate2({ id, body: { template: tpl } }),
      );
      return normalizeSingle(res);
    },

    deleteTemplate: (id) => unwrapVoidResponse(templateApi.templateServiceDeleteTemplate2({ id })),

    getTemplateConfig: (id) =>
      unwrapSdkResponse(templateApi.templateServiceGetTemplateConfig2({ id })),

    updateTemplateConfig: (id, req) =>
      unwrapSdkResponse(templateApi.templateServiceUpdateTemplateConfig2({ id, body: req })),

    getSessionConfig: (id) => unwrapSdkResponse(sessionApi.sessionServiceGetSessionConfig2({ id })),

    updateSessionConfig: (id, req) =>
      unwrapSdkResponse(sessionApi.sessionServiceUpdateSessionConfig2({ id, body: req })),

    getLocalConfig: () => unwrapSdkResponse(configApi.configServiceGetConfig2()),

    updateLocalConfig: (cfg) =>
      unwrapSdkResponse(configApi.configServiceUpdateConfig2({ body: cfg })),
  };
}

export const api = createAgentApi();
