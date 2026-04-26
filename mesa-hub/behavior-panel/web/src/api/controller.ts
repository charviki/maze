import type { Node } from '@maze/fabrication';
import { createRequest } from '@maze/fabrication';

const request = createRequest('/api/v1');

export const controllerApi = {
  listNodes: () => request<Node[]>('/nodes'),

  getNode: (name: string) => request<Node>(`/nodes/${name}`),

  deleteNode: (name: string) =>
    request<void>(`/nodes/${name}`, { method: 'DELETE' }),
};
