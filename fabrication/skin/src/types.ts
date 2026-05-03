export interface Node {
  name: string;
  address: string;
  externalAddr: string;
  sessionCount: number;
  status: string;
  registeredAt: string;
  lastHeartbeat: string;
}

export interface Session {
  id: string;
  name: string;
  status: string;
  createdAt: string;
  windowCount: number;
}

export interface ConfigItem {
  type: string;
  key: string;
  value: string;
}

export interface CreateSessionRequest {
  name: string;
  command?: string;
  workingDir?: string;
  sessionConfs?: ConfigItem[];
  restoreStrategy?: string;
  templateId?: string;
}

export interface TerminalOutput {
  sessionId: string;
  lines: number;
  output: string;
}

export interface ApiResponse<T> {
  status: string;
  data?: T;
  message?: string;
  code?: string;
  conflicts?: ConfigConflict[];
}

export interface ConfigFile {
  path: string;
  content: string;
}

export interface ConfigLayer {
  env: Record<string, string>;
  files: ConfigFile[];
}

export type PipelinePhase = 'system' | 'template' | 'user';
export type PipelineStepType = 'cd' | 'env' | 'file' | 'command';

export interface PipelineStep {
  id: string;
  type: PipelineStepType;
  phase: PipelinePhase;
  order: number;
  key?: string;
  value: string;
}

export interface SessionState {
  sessionName: string;
  pipeline: PipelineStep[];
  restoreStrategy: string;
  workingDir: string;
  envSnapshot: Record<string, string>;
  terminalSnapshot: string;
  savedAt: string;
}

export interface SavedSession {
  sessionName: string;
  pipeline: PipelineStep[];
  restoreStrategy: string;
  workingDir: string;
  terminalSnapshot: string;
  savedAt: string;
}

export interface EnvDef {
  key: string;
  label: string;
  required: boolean;
  placeholder: string;
  sensitive: boolean;
}

export interface FileDef {
  path: string;
  label: string;
  required: boolean;
  defaultContent: string;
}

export interface SessionSchema {
  envDefs: EnvDef[];
  fileDefs: FileDef[];
}

export interface SessionTemplate {
  id: string;
  name: string;
  command: string;
  description: string;
  icon: string;
  builtin: boolean;
  defaults: ConfigLayer;
  sessionSchema: SessionSchema;
}

export type ConfigScope = 'global' | 'project';

export interface ConfigFileSnapshot {
  path: string;
  content: string;
  exists: boolean;
  hash: string;
}

export interface TemplateConfigView {
  templateId: string;
  scope: ConfigScope;
  files: ConfigFileSnapshot[];
}

export interface SessionConfigView {
  sessionId: string;
  templateId: string;
  workingDir: string;
  scope: ConfigScope;
  files: ConfigFileSnapshot[];
}

export interface ConfigFileUpdate {
  path: string;
  content: string;
  baseHash: string;
}

export interface SaveConfigRequest {
  files: ConfigFileUpdate[];
}

export interface ConfigConflict {
  path: string;
  currentHash: string;
}

export interface LocalAgentConfig {
  workingDir: string;
  env: Record<string, string>;
}

export interface NodeConfig {
  nodeName: string;
  workingDir: string;
  env: Record<string, string>;
}

export interface SessionMeta {
  sessionId: string;
  nodeName: string;
  templateId: string;
  templateName: string;
  command: string;
  workingDir: string;
  configs: ConfigLayer;
  pipeline: PipelineStep[];
  restoreStrategy: string;
  createdAt: string;
}

export interface SaveSessionMetaRequest {
  sessionId: string;
  templateId: string;
  templateName: string;
  command: string;
  workingDir: string;
  configs: ConfigLayer;
  pipeline: PipelineStep[];
  restoreStrategy: string;
}

export interface Tool {
  id: string;
  image: string;
  sourcePath: string;
  destPath: string;
  binPaths: string[];
  envVars?: Record<string, string>;
  description: string;
  category: string;
}

export interface ResourceLimits {
  cpuLimit?: string;
  memoryLimit?: string;
}

export interface CreateHostRequest {
  name: string;
  tools: string[];
  displayName?: string;
  resources?: ResourceLimits;
}

export interface CreateHostResponse {
  name: string;
  tools: string[];
  imageTag: string;
  containerId: string;
  status: string;
  buildLog?: string;
}

export type HostStatus = 'pending' | 'deploying' | 'online' | 'offline' | 'failed';

export interface Host {
  name: string;
  displayName?: string;
  tools: string[];
  resources?: ResourceLimits;
  authToken: string;
  createdAt: string;
  updatedAt: string;
  status: HostStatus;
  errorMsg?: string;
  retryCount: number;
  address?: string;
  sessionCount: number;
  lastHeartbeat?: string;
}
