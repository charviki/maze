import type { IAgentApiClient } from '@maze/fabrication';
import { createAgentApiClient } from '@maze/fabrication';

// 代理模式：通过 Manager 网关转发请求到指定 Agent 节点
export function createAgentApi(managerBase: string, nodeName: string): IAgentApiClient {
  return createAgentApiClient(`${managerBase}/api/v1/nodes/${encodeURIComponent(nodeName)}/`);
}
