import {
  NodeServiceApi,
  HostServiceApi,
  createSdkConfiguration,
  unwrapSdkResponse,
  unwrapVoidResponse,
} from '@maze/fabrication';
import type { HostSpec } from '@maze/fabrication';

const config = createSdkConfiguration('');
const nodeApi = new NodeServiceApi(config);
const hostApi = new HostServiceApi(config);

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
};
