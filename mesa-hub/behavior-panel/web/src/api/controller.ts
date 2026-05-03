import type { Node, Tool, CreateHostRequest, Host } from '@maze/fabrication';
import { createRequest } from '@maze/fabrication';

const request = createRequest('/api/v1');

export const controllerApi = {
  listNodes: async () => {
    const res = await request<{ nodes: Node[] }>('/nodes');
    return { ...res, data: res.data?.nodes };
  },

  getNode: (name: string) => request<Node>(`/nodes/${name}`),

  deleteNode: (name: string) => request<void>(`/nodes/${name}`, { method: 'DELETE' }),

  listTools: async () => {
    const res = await request<{ tools: Tool[] }>('/host/tools');
    return { ...res, data: res.data?.tools };
  },

  createHost: (data: CreateHostRequest) =>
    request<Host>('/hosts', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  listHosts: async () => {
    const res = await request<{ hosts: Host[] }>('/hosts');
    return { ...res, data: res.data?.hosts };
  },

  getHost: (name: string) => request<Host>(`/hosts/${name}`),

  getHostBuildLog: async (name: string) => {
    const res = await request<{ log: string }>(`/hosts/${name}/logs/build`);
    return { ...res, data: res.data?.log };
  },

  getHostRuntimeLog: async (name: string) => {
    const res = await request<{ log: string }>(`/hosts/${name}/logs/runtime`);
    return { ...res, data: res.data?.log };
  },

  deleteHost: (name: string) => request<void>(`/hosts/${name}`, { method: 'DELETE' }),
};
