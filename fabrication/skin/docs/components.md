# Skin 组件目录概览

## 视觉特效组件 (components/ui/)
- HexWaterfall / DecryptText / TerrainBackground / GlitchEffect / BootSequence / ReverieEffect — 西部世界主题装饰与动画
- RadarView / HostVitalSign — Agent / Host 状态可视化
- Panel / Button / Input / Select / Dialog / ConfirmDialog / Skeleton / Toast — 基础 UI 控件
- XtermTerminal / SessionPipeline — 终端与管线编辑器
- AnimationSettings — 特效全局开关
- ErrorBoundary — 错误边界

## Agent 业务组件 (components/agent/)
- AgentPanel — Agent 主面板容器
- SessionList / TerminalPane / SessionDialogs — 会话管理与终端
- CreateSessionWithTemplateDialog / TemplateManager — 模板与创建对话框
- NodeConfigPanel — 节点配置面板

## 工具函数
- cn — Tailwind 类名合并
- maskEnvValue / maskFileContent — 敏感数据脱敏
- createRequest — HTTP 请求工厂（已修复 error 分支丢失 code/conflicts 的 BUG）
- createSdkConfiguration — SDK Configuration 工厂函数，注入自定义 fetch
- normalizeTemplate — 模板规范化共享函数（确保 defaults/sessionSchema 不为 null）
- unwrapSdkResponse / unwrapVoidResponse — SDK 响应解包辅助函数
- usePollingWithBackoff — 指数退避轮询 Hook

## 接口
- ISessionApi — 会话 API 子接口（会话 CRUD、终端 I/O、WebSocket URL 生成）
- ITemplateApi — 模板 API 子接口（模板 CRUD、配置管理）
- IConfigApi — 远程配置 API 子接口（通过 Manager 代理的节点配置管理）
- ILocalConfigApi — 本地配置 API 子接口（直连 Agent 的本地配置管理）
- IAgentApiClient — 以上四个子接口的联合类型别名
