import {
  NodeServiceApi,
  HostServiceApi,
  SkillServiceApi,
  MCPServiceApi,
  RuleServiceApi,
  GitKeyServiceApi,
  createSdkConfiguration,
  unwrapSdkResponse,
  unwrapVoidResponse,
} from '@maze/fabrication';
import type { HostSpec, V1Skill, V1MCPServer, V1Rule, V1GitKey } from '@maze/fabrication';
import type { CreateGitKeyData } from '../components/git-keys/constants';

const config = createSdkConfiguration('');
const nodeApi = new NodeServiceApi(config);
const hostApi = new HostServiceApi(config);
const skillApi = new SkillServiceApi(config);
const mcpApi = new MCPServiceApi(config);
const ruleApi = new RuleServiceApi(config);
const gitKeyApi = new GitKeyServiceApi(config);

export const controllerApi = {
  listNodes: async () => {
    const res = await unwrapSdkResponse(nodeApi.nodeServiceListNodes());
    return { ...res, data: res.data?.nodes };
  },

  getNode: (name: string) => unwrapSdkResponse(nodeApi.nodeServiceGetNode({ name })),

  deleteNode: (name: string) => unwrapVoidResponse(nodeApi.nodeServiceDeleteNode({ name })),

  listTools: async () => {
    const res = await unwrapSdkResponse(hostApi.hostServiceListTools());
    return { ...res, data: res.data?.tools };
  },

  createHost: (data: HostSpec) => unwrapSdkResponse(hostApi.hostServiceCreateHost({ body: data })),

  listHosts: async () => {
    const res = await unwrapSdkResponse(hostApi.hostServiceListHosts());
    return { ...res, data: res.data?.hosts };
  },

  getHost: (name: string) => unwrapSdkResponse(hostApi.hostServiceGetHost({ name })),

  getHostBuildLog: async (name: string) => {
    const res = await unwrapSdkResponse(hostApi.hostServiceGetBuildLog({ name }));
    return { ...res, data: res.data?.log };
  },

  getHostRuntimeLog: async (name: string) => {
    const res = await unwrapSdkResponse(hostApi.hostServiceGetRuntimeLog({ name }));
    return { ...res, data: res.data?.log };
  },

  deleteHost: (name: string) => unwrapVoidResponse(hostApi.hostServiceDeleteHost({ name })),

  // Skills
  listSkills: async (): Promise<V1Skill[]> => {
    const res = await skillApi.skillServiceListSkills();
    return res.skills ?? [];
  },
  createSkill: async (data: {
    name: string;
    description?: string;
    config?: Record<string, string>;
  }): Promise<V1Skill> => {
    return skillApi.skillServiceCreateSkill({ body: data });
  },
  getSkill: (name: string) => skillApi.skillServiceGetSkill({ name }),
  updateSkill: async (
    name: string,
    data: { description?: string; config?: Record<string, string> },
  ): Promise<V1Skill> => {
    return skillApi.skillServiceUpdateSkill({ name, body: data });
  },
  deleteSkill: (name: string) => unwrapVoidResponse(skillApi.skillServiceDeleteSkill({ name })),

  // MCP Servers
  listMCPServers: async (): Promise<V1MCPServer[]> => {
    const res = await mcpApi.mCPServiceListMCPServers();
    return res.servers ?? [];
  },
  createMCPServer: async (data: {
    name: string;
    type: string;
    command?: string;
    url?: string;
    args?: string[];
    env?: Record<string, string>;
  }): Promise<V1MCPServer> => {
    return mcpApi.mCPServiceCreateMCPServer({ body: data });
  },
  getMCPServer: (name: string) => mcpApi.mCPServiceGetMCPServer({ name }),
  updateMCPServer: async (
    name: string,
    data: {
      type: string;
      command?: string;
      url?: string;
      args?: string[];
      env?: Record<string, string>;
    },
  ): Promise<V1MCPServer> => {
    return mcpApi.mCPServiceUpdateMCPServer({ name, body: data });
  },
  deleteMCPServer: (name: string) => unwrapVoidResponse(mcpApi.mCPServiceDeleteMCPServer({ name })),

  // Rules
  listRules: async (): Promise<V1Rule[]> => {
    const res = await ruleApi.ruleServiceListRules();
    return res.rules ?? [];
  },
  createRule: async (data: { name: string; content?: string }): Promise<V1Rule> => {
    return ruleApi.ruleServiceCreateRule({ body: data });
  },
  getRule: (name: string) => ruleApi.ruleServiceGetRule({ name }),
  updateRule: async (name: string, data: { content?: string }): Promise<V1Rule> => {
    return ruleApi.ruleServiceUpdateRule({ name, body: data });
  },
  deleteRule: (name: string) => unwrapVoidResponse(ruleApi.ruleServiceDeleteRule({ name })),

  // Git Keys
  listGitKeys: async (): Promise<V1GitKey[]> => {
    const res = await gitKeyApi.gitKeyServiceListGitKeys();
    return res.gitKeys ?? [];
  },
  createGitKey: async (data: CreateGitKeyData): Promise<V1GitKey> => {
    return gitKeyApi.gitKeyServiceCreateGitKey({ body: data });
  },
  getGitKey: (name: string) => gitKeyApi.gitKeyServiceGetGitKey({ name }),
  deleteGitKey: (name: string) => unwrapVoidResponse(gitKeyApi.gitKeyServiceDeleteGitKey({ name })),
};
