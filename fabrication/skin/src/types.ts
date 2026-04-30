export interface Node {
  name: string;
  address: string;
  external_addr: string;
  session_count: number;
  status: string;
  registered_at: string;
  last_heartbeat: string;
}

export interface Session {
  id: string;
  name: string;
  status: string;
  created_at: string;
  window_count: number;
}

export interface ConfigItem {
  type: string;
  key: string;
  value: string;
}

export interface CreateSessionRequest {
  name: string;
  command?: string;
  working_dir?: string;
  session_confs?: ConfigItem[];
  restore_strategy?: string;
  template_id?: string;
}

export interface TerminalOutput {
  session_id: string;
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

// --- 统一配置层 ---

export interface ConfigFile {
  path: string;
  content: string;
}

export interface ConfigLayer {
  env: Record<string, string>;
  files: ConfigFile[];
}

// --- Pipeline ---

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

// --- SessionState ---

export interface SessionState {
  session_name: string;
  pipeline: PipelineStep[];
  restore_strategy: string;
  working_dir: string;
  env_snapshot: Record<string, string>;
  terminal_snapshot: string;
  saved_at: string;
}

// --- SavedSession (前端展示用) ---

export interface SavedSession {
  session_name: string;
  pipeline: PipelineStep[];
  restore_strategy: string;
  working_dir: string;
  terminal_snapshot: string;
  saved_at: string;
}

// --- Template ---

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
  default_content: string;
}

export interface SessionSchema {
  env_defs: EnvDef[];
  file_defs: FileDef[];
}

export interface SessionTemplate {
  id: string;
  name: string;
  command: string;
  description: string;
  icon: string;
  builtin: boolean;
  defaults: ConfigLayer;
  session_schema: SessionSchema;
}

export type ConfigScope = 'global' | 'project';

export interface ConfigFileSnapshot {
  path: string;
  content: string;
  exists: boolean;
  hash: string;
}

export interface TemplateConfigView {
  template_id: string;
  scope: ConfigScope;
  files: ConfigFileSnapshot[];
}

export interface SessionConfigView {
  session_id: string;
  template_id: string;
  working_dir: string;
  scope: ConfigScope;
  files: ConfigFileSnapshot[];
}

export interface ConfigFileUpdate {
  path: string;
  content: string;
  base_hash: string;
}

export interface SaveConfigRequest {
  files: ConfigFileUpdate[];
}

export interface ConfigConflict {
  path: string;
  current_hash: string;
}

// --- LocalAgentConfig ---

export interface LocalAgentConfig {
  working_dir: string;
  env: Record<string, string>;
}

// --- NodeConfig ---

export interface NodeConfig {
  node_name: string;
  working_dir: string;
  env: Record<string, string>;
}

// --- SessionMeta ---

export interface SessionMeta {
  session_id: string;
  node_name: string;
  template_id: string;
  template_name: string;
  command: string;
  working_dir: string;
  configs: ConfigLayer;
  pipeline: PipelineStep[];
  restore_strategy: string;
  created_at: string;
}

export interface SaveSessionMetaRequest {
  session_id: string;
  template_id: string;
  template_name: string;
  command: string;
  working_dir: string;
  configs: ConfigLayer;
  pipeline: PipelineStep[];
  restore_strategy: string;
}

// --- Host ---

export interface Tool {
  id: string;
  image: string;
  source_path: string;
  dest_path: string;
  bin_paths: string[];
  env_vars?: Record<string, string>;
  description: string;
  category: string;
}

export interface ResourceLimits {
  cpu_limit?: string;
  memory_limit?: string;
}

export interface CreateHostRequest {
  name: string;
  tools: string[];
  display_name?: string;
  resources?: ResourceLimits;
}

export interface CreateHostResponse {
  name: string;
  tools: string[];
  image_tag: string;
  container_id: string;
  status: string;
  build_log?: string;
}
