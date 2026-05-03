import type {
  V1Session,
  V1NodeInfo,
  V1HostInfo,
  V1HostSpec,
  V1SessionTemplate,
  V1CreateSessionRequest,
  V1TerminalOutput,
  V1ConfigItem,
  V1ConfigFile,
  V1ConfigLayer,
  V1EnvDef,
  V1FileDef,
  V1SessionSchema,
  V1ConfigFileSnapshot,
  V1TemplateConfigView,
  V1SessionConfigView,
  V1ConfigFileUpdate,
  V1LocalAgentConfig,
  V1ToolConfig,
  V1ResourceLimits,
  V1CreateHostRequest,
} from './api/gen/models';

// ===== SDK 类型 re-export（向后兼容别名） =====

export type {
  V1SessionTemplate,
  V1ConfigLayer,
  V1SessionSchema,
  V1Session,
  V1NodeInfo,
  V1HostInfo,
  V1HostSpec,
  V1CreateHostRequest,
  V1CreateSessionRequest,
  V1TerminalOutput,
  V1ConfigItem,
  V1ConfigFile,
  V1EnvDef,
  V1FileDef,
  V1ConfigFileSnapshot,
  V1TemplateConfigView,
  V1SessionConfigView,
  V1ConfigFileUpdate,
  V1LocalAgentConfig,
  V1ToolConfig,
  V1ResourceLimits,
  V1SessionState,
  V1SaveSessionsResponse,
} from './api/gen/models';

export type Session = V1Session;
export type Node = V1NodeInfo;
export type Host = V1HostInfo;
export type HostSpec = V1HostSpec;
export type SessionTemplate = V1SessionTemplate;
export type CreateSessionRequest = V1CreateSessionRequest;
export type TerminalOutput = V1TerminalOutput;
export type ConfigItem = V1ConfigItem;
export type ConfigFile = V1ConfigFile;
export type ConfigLayer = V1ConfigLayer;
export type EnvDef = V1EnvDef;
export type FileDef = V1FileDef;
export type SessionSchema = V1SessionSchema;
export type ConfigFileSnapshot = Omit<V1ConfigFileSnapshot, '_exists'> & { exists: boolean };
export type TemplateConfigView = V1TemplateConfigView;
export type SessionConfigView = V1SessionConfigView;
export type ConfigFileUpdate = V1ConfigFileUpdate;
export type LocalAgentConfig = V1LocalAgentConfig;
export type Tool = V1ToolConfig;
export type ResourceLimits = V1ResourceLimits;
export type CreateHostRequest = V1CreateHostRequest;
export type { NormalizedTemplate } from './api/normalize';

// ===== 前端特有类型 =====

export interface ApiResponse<T> {
  status: string;
  data?: T;
  message?: string;
  code?: string;
  conflicts?: ConfigConflict[];
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

export type ConfigScope = 'global' | 'project';

export interface SaveConfigRequest {
  files: ConfigFileUpdate[];
}

export interface ConfigConflict {
  path: string;
  currentHash: string;
}

export type HostStatus = 'pending' | 'deploying' | 'online' | 'offline' | 'failed';
