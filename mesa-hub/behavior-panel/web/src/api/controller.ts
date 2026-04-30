import type { Node, Tool, CreateHostRequest, CreateHostResponse } from '@maze/fabrication';
import { createRequest } from '@maze/fabrication';

const request = createRequest('/api/v1');

export const controllerApi = {
  listNodes: () => request<Node[]>('/nodes'),

  getNode: (name: string) => request<Node>(`/nodes/${name}`),

  deleteNode: (name: string) =>
    request<void>(`/nodes/${name}`, { method: 'DELETE' }),

  listTools: () => request<Tool[]>('/host/tools'),

  createHost: (data: CreateHostRequest) => {
    // docker build 耗时较长，使用自定义 5 分钟超时
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5 * 60 * 1000);
    const promise = request<CreateHostResponse>('/hosts', {
      method: 'POST',
      body: JSON.stringify(data),
      signal: controller.signal,
    });
    promise.finally(() => clearTimeout(timeoutId));
    return promise;
  },

  deleteHost: (name: string) =>
    request<void>(`/hosts/${name}`, { method: 'DELETE' }),
};
