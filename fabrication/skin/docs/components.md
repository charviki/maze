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
- createRequest — HTTP 请求工厂
- usePollingWithBackoff — 指数退避轮询 Hook

## 接口
- IAgentApiClient — Agent API 客户端接口（会话 CRUD、终端 I/O、模板管理、配置管理、WebSocket URL 生成）
