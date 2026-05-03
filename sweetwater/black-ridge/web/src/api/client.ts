import type { IAgentApiClient } from '@maze/fabrication';
import { createAgentApiClient } from '@maze/fabrication';

// 直连模式：直接访问本地 Agent 节点的 API
export function createAgentApi(): IAgentApiClient {
  return createAgentApiClient('/api/v1/');
}

export const api = createAgentApi();
