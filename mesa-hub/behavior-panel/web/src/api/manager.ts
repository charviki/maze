import type { SessionTemplate } from '@maze/fabrication';
import { createRequest } from '@maze/fabrication';

const request = createRequest('/api/v1');

/**
 * 创建 Manager API 客户端（模板管理）。
 * NodeConfig 和 SessionMeta 相关 API 已移除：
 * - 节点配置由 Agent 本地记忆管理，Manager 通过心跳获取只读视图
 * - Session 元数据由 Agent 的 Pipeline 状态文件管理，Manager 通过代理获取
 */
export function createManagerApi() {
  return {
    listTemplates: () => request<SessionTemplate[]>('/templates'),

    getTemplate: (id: string) =>
      request<SessionTemplate>(`/templates/${encodeURIComponent(id)}`),

    createTemplate: (template: SessionTemplate) =>
      request<SessionTemplate>('/templates', {
        method: 'POST',
        body: JSON.stringify(template),
      }),

    updateTemplate: (id: string, template: SessionTemplate) =>
      request<SessionTemplate>(`/templates/${encodeURIComponent(id)}`, {
        method: 'PUT',
        body: JSON.stringify(template),
      }),

    deleteTemplate: (id: string) =>
      request<void>(`/templates/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  };
}

export type ManagerApiClient = ReturnType<typeof createManagerApi>;
