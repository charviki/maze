import type { Node, Tool, CreateHostRequest, Host } from '@maze/fabrication';
import { createRequest } from '@maze/fabrication';

const request = createRequest('/api/v1');

export const controllerApi = {
  listNodes: () => request<Node[]>('/nodes'),

  getNode: (name: string) => request<Node>(`/nodes/${name}`),

  deleteNode: (name: string) =>
    request<void>(`/nodes/${name}`, { method: 'DELETE' }),

  listTools: () => request<Tool[]>('/host/tools'),

  createHost: (data: CreateHostRequest) =>
    request<Host>('/hosts', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  listHosts: () => request<Host[]>('/hosts'),

  getHost: (name: string) => request<Host>(`/hosts/${name}`),

  getHostBuildLog: (name: string) =>
    request<string>(`/hosts/${name}/logs/build`),

  getHostRuntimeLog: (name: string) =>
    request<string>(`/hosts/${name}/logs/runtime`),

  deleteHost: (name: string) =>
    request<void>(`/hosts/${name}`, { method: 'DELETE' }),
};
