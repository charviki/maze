import {
  createSdkConfiguration,
  unwrapSdkResponse,
  unwrapVoidResponse,
  nodeServiceListNodes,
  nodeServiceGetNode,
  nodeServiceDeleteNode,
  hostServiceListTools,
  hostServiceCreateHost,
  hostServiceListHosts,
  hostServiceGetHost,
  hostServiceGetBuildLog,
  hostServiceGetRuntimeLog,
  hostServiceDeleteHost,
  skillServiceListSkills,
  skillServiceCreateSkill,
  skillServiceGetSkill,
  skillServiceUpdateSkill,
  skillServiceDeleteSkill,
  mcpServiceListMcpServers,
  mcpServiceCreateMcpServer,
  mcpServiceGetMcpServer,
  mcpServiceUpdateMcpServer,
  mcpServiceDeleteMcpServer,
  ruleServiceListRules,
  ruleServiceCreateRule,
  ruleServiceGetRule,
  ruleServiceUpdateRule,
  ruleServiceDeleteRule,
  gitKeyServiceListGitKeys,
  gitKeyServiceCreateGitKey,
  gitKeyServiceGetGitKey,
  gitKeyServiceDeleteGitKey,
} from '@maze/fabrication';
import type { HostSpec, V1Skill, V1McpServer, V1Rule, V1GitKey } from '@maze/fabrication';
import type { CreateGitKeyData } from '../components/git-keys/constants';

// 配置全局 hey-api client（auth 预刷新 + 401 重试 + 30s 超时经自定义 fetch 注入）。
createSdkConfiguration('');

// hey-api 默认 throwOnError=false（失败返回 {error} 而非抛错）。对需要直接拿到数据（非 ApiResponse）
// 的调用点，unwrapData 统一解包并在失败时抛错，保持与旧 typescript-fetch 一致的失败语义。
async function unwrapData<T>(
  result: Promise<{ data?: T; error?: unknown; request?: Request; response?: Response }>,
): Promise<T> {
  const res = await unwrapSdkResponse<T>(result);
  if (res.status === 'error') throw new Error(res.message);
  return res.data as T;
}

export const controllerApi = {
  listNodes: async () => {
    const res = await unwrapSdkResponse(nodeServiceListNodes());
    return { ...res, data: res.data?.nodes };
  },
  getNode: (name: string) => unwrapSdkResponse(nodeServiceGetNode({ path: { name } })),
  deleteNode: (name: string) => unwrapVoidResponse(nodeServiceDeleteNode({ path: { name } })),

  listTools: async () => {
    const res = await unwrapSdkResponse(hostServiceListTools());
    return { ...res, data: res.data?.tools };
  },
  createHost: (data: HostSpec) => unwrapSdkResponse(hostServiceCreateHost({ body: data })),
  listHosts: async () => {
    const res = await unwrapSdkResponse(hostServiceListHosts());
    return { ...res, data: res.data?.hosts };
  },
  getHost: (name: string) => unwrapSdkResponse(hostServiceGetHost({ path: { name } })),
  getHostBuildLog: async (name: string) => {
    const res = await unwrapSdkResponse(hostServiceGetBuildLog({ path: { name } }));
    return { ...res, data: res.data?.log };
  },
  getHostRuntimeLog: async (name: string) => {
    const res = await unwrapSdkResponse(hostServiceGetRuntimeLog({ path: { name } }));
    return { ...res, data: res.data?.log };
  },
  deleteHost: (name: string) => unwrapVoidResponse(hostServiceDeleteHost({ path: { name } })),

  // Skills
  listSkills: async (): Promise<V1Skill[]> => {
    const data = await unwrapData(skillServiceListSkills());
    return data?.skills ?? [];
  },
  createSkill: async (data: {
    name: string;
    description?: string;
    config?: Record<string, string>;
  }): Promise<V1Skill> => unwrapData(skillServiceCreateSkill({ body: data })),
  getSkill: (name: string) => unwrapData(skillServiceGetSkill({ path: { name } })),
  updateSkill: async (
    name: string,
    data: { description?: string; config?: Record<string, string> },
  ): Promise<V1Skill> => unwrapData(skillServiceUpdateSkill({ path: { name }, body: data })),
  deleteSkill: (name: string) => unwrapVoidResponse(skillServiceDeleteSkill({ path: { name } })),

  // MCP Servers
  listMCPServers: async (): Promise<V1McpServer[]> => {
    const data = await unwrapData(mcpServiceListMcpServers());
    return data?.servers ?? [];
  },
  createMCPServer: async (data: {
    name: string;
    type: string;
    command?: string;
    url?: string;
    args?: string[];
    env?: Record<string, string>;
  }): Promise<V1McpServer> => unwrapData(mcpServiceCreateMcpServer({ body: data })),
  getMCPServer: (name: string) => unwrapData(mcpServiceGetMcpServer({ path: { name } })),
  updateMCPServer: async (
    name: string,
    data: {
      type: string;
      command?: string;
      url?: string;
      args?: string[];
      env?: Record<string, string>;
    },
  ): Promise<V1McpServer> => unwrapData(mcpServiceUpdateMcpServer({ path: { name }, body: data })),
  deleteMCPServer: (name: string) =>
    unwrapVoidResponse(mcpServiceDeleteMcpServer({ path: { name } })),

  // Rules
  listRules: async (): Promise<V1Rule[]> => {
    const data = await unwrapData(ruleServiceListRules());
    return data?.rules ?? [];
  },
  createRule: async (data: { name: string; content?: string }): Promise<V1Rule> =>
    unwrapData(ruleServiceCreateRule({ body: data })),
  getRule: (name: string) => unwrapData(ruleServiceGetRule({ path: { name } })),
  updateRule: async (name: string, data: { content?: string }): Promise<V1Rule> =>
    unwrapData(ruleServiceUpdateRule({ path: { name }, body: data })),
  deleteRule: (name: string) => unwrapVoidResponse(ruleServiceDeleteRule({ path: { name } })),

  // Git Keys
  listGitKeys: async (): Promise<V1GitKey[]> => {
    const data = await unwrapData(gitKeyServiceListGitKeys());
    return data?.gitKeys ?? [];
  },
  createGitKey: async (data: CreateGitKeyData): Promise<V1GitKey> =>
    unwrapData(gitKeyServiceCreateGitKey({ body: data })),
  getGitKey: (name: string) => unwrapData(gitKeyServiceGetGitKey({ path: { name } })),
  deleteGitKey: (name: string) => unwrapVoidResponse(gitKeyServiceDeleteGitKey({ path: { name } })),
};
