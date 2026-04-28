import type {
  Session,
  CreateSessionRequest,
  TerminalOutput,
  SavedSession,
  ApiResponse,
  SessionTemplate,
  LocalAgentConfig,
  TemplateConfigView,
  SessionConfigView,
  SaveConfigRequest,
} from './types';

/**
 * 统一的 Agent API Client 接口。
 * 规范所有与 Agent 交互的方法，无论底层是直连还是经过 Manager 代理。
 */
export interface IAgentApiClient {
  /**
   * 获取所有运行中的会话列表
   */
  listSessions: () => Promise<ApiResponse<Session[]>>;

  /**
   * 创建新的会话
   */
  createSession: (data: CreateSessionRequest) => Promise<ApiResponse<Session>>;

  /**
   * 获取指定会话的详情
   */
  getSession: (id: string) => Promise<ApiResponse<Session>>;

  /**
   * 终止/删除指定的会话
   */
  deleteSession: (id: string) => Promise<ApiResponse<void>>;

  /**
   * 获取指定会话的终端输出
   */
  getOutput: (id: string, lines?: number) => Promise<ApiResponse<TerminalOutput>>;

  /**
   * 向指定会话发送命令输入
   */
  sendInput: (id: string, command: string) => Promise<ApiResponse<void>>;

  /**
   * 向指定会话发送控制信号（如 sigint）
   */
  sendSignal: (id: string, signal: string) => Promise<ApiResponse<void>>;

  /**
   * 获取所有已保存（持久化）的会话元数据列表
   */
  getSavedSessions: () => Promise<ApiResponse<SavedSession[]>>;

  /**
   * 恢复指定的已保存会话
   */
  restoreSession: (id: string) => Promise<ApiResponse<void>>;

  /**
   * 触发会话状态的手动全量保存
   */
  saveSessions: () => Promise<ApiResponse<{ saved_at: string }>>;

  /**
   * 动态生成 WebSocket 终端连接的 URL。
   * @param sessionId 会话的唯一标识
   */
  buildWsUrl: (sessionId: string) => string;

  /** 获取节点模板列表 */
  listTemplates: () => Promise<ApiResponse<SessionTemplate[]>>;

  /** 创建模板 */
  createTemplate: (tpl: SessionTemplate) => Promise<ApiResponse<SessionTemplate>>;

  /** 获取单个模板 */
  getTemplate: (id: string) => Promise<ApiResponse<SessionTemplate>>;

  /** 获取模板的真实全局配置 */
  getTemplateConfig: (id: string) => Promise<ApiResponse<TemplateConfigView>>;

  /** 更新模板 */
  updateTemplate: (id: string, tpl: SessionTemplate) => Promise<ApiResponse<SessionTemplate>>;

  /** 保存模板的真实全局配置 */
  updateTemplateConfig: (id: string, req: SaveConfigRequest) => Promise<ApiResponse<TemplateConfigView>>;

  /** 删除模板 */
  deleteTemplate: (id: string) => Promise<ApiResponse<void>>;

  /** 获取 session 的真实项目级配置 */
  getSessionConfig: (id: string) => Promise<ApiResponse<SessionConfigView>>;

  /** 保存 session 的真实项目级配置 */
  updateSessionConfig: (id: string, req: SaveConfigRequest) => Promise<ApiResponse<SessionConfigView>>;

  /** 获取节点本地配置 */
  getLocalConfig: () => Promise<ApiResponse<LocalAgentConfig>>;

  /** 更新节点本地配置 */
  updateLocalConfig: (cfg: Partial<LocalAgentConfig>) => Promise<ApiResponse<LocalAgentConfig>>;
}
