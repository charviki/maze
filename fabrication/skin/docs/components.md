# Fabrication 组件文档

> 本文档严格基于源码生成，列出 `@maze/fabrication` 导出的所有组件、工具函数、Hook 和接口。

---

## 1. 视觉特效组件

### BootSequence

西部世界风格的开机启动序列动画。分六个阶段（wake → brand → diag → cognitive → awaken → done），支持 `mesa-hub` 和 `sweetwater` 两种分区主题，展示 DELOS ASCII 品牌标识和诊断日志滚动。动画完成后调用 `onComplete` 回调。受 AnimationSettings 开关控制。

**Props**: `onComplete: () => void`, `division?: 'mesa-hub' | 'sweetwater'`

### DecryptText

文字解密动画效果。逐字符从随机字符集解码为目标文本，支持设置速度、最大迭代次数和自定义字符集。可配置为鼠标悬停时触发。

**Props**: `text: string`, `speed?: number`, `maxIterations?: number`, `characters?: string`, `className?: string`, `animateOnHover?: boolean`

### GlitchEffect

故障失真效果包装器。激活时通过双层偏移叠加实现 glitch 视觉。

**Props**: `children: ReactNode`, `isActive?: boolean`, `className?: string`

### HexWaterfall

十六进制字符瀑布流背景。基于 Canvas 绘制 `0-9A-F` 字符的 Matrix 风格下落动画，自动读取 CSS 变量 `--primary` 作为颜色。页面不可见时自动暂停。

**Props**: `className?: string`, `opacity?: number`, `color?: string`

### TerrainBackground

低多边形地形背景动画。包含三层视觉效果：低多边形地形网格、等高线脉冲波扩散、监测信标（带呼吸节奏和粒子效果）。使用确定性伪随机数保证每次挂载地形一致。

**Props**: `className?: string`

### RadarView

雷达扫描视图。显示同心圆、十字准线、旋转扫描线，以及各节点的状态指示点。当有多个 online 节点时，随机生成节点间连接脉冲线。

**Props**: `className?: string`, `nodes?: RadarNode[]`（`RadarNode: { name: string; status: 'online' | 'offline' }`）

### HostVitalSign

Host 生命体征指示器。根据 `running`（青色呼吸脉冲）、`saved`（黄色慢速呼吸）、`offline`（灰色不规则闪烁）三种状态展示不同节奏的光效。支持 sm/md 两种尺寸。

**Props**: `status: 'running' | 'saved' | 'offline'`, `size?: 'sm' | 'md'`, `className?: string`

### ReverieEffect

遗忆（Reverie）效果包装器。当 `isActive` 为 true 时，以 8-20 秒随机间隔触发 500ms 的短暂闪烁失真动画，模拟西部世界中 Host 的记忆回溯。

**Props**: `children: ReactNode`, `isActive?: boolean`, `className?: string`

### AnimationSettings / AnimationSettingsProvider / useAnimationSettings / AnimationSettingsPanel

视觉特效全局设置系统。通过 Context 提供五项开关：`canvasBackground`、`crtScanlines`、`decryptText`、`glitchEffect`、`bootSequence`。设置持久化到 `localStorage`（key: `maze:animation-settings`）。

---

## 2. 基础 UI 组件

### Panel

切角面板容器。使用 `clip-path` 实现八边形切角造型，支持 `default`/`destructive`/`warning`/`success` 四种颜色变体。带角标装饰线和状态标签文字。

**Props**: `children`, `className?`, `cornerSize?: number`, `showCrosshairs?: boolean`, `variant?`, `transparent?: boolean`

### Button

基于 CVA 的按钮组件。支持 6 种变体（default/destructive/outline/secondary/ghost/link）和 4 种尺寸（default/sm/lg/icon）。支持 `asChild` 模式。

### Dialog

基于 `@radix-ui/react-dialog` 的弹窗组件套件。包含完整的弹窗体系：Dialog, DialogTrigger, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter 等。

### Input

标准输入框组件，基于 `forwardRef` 实现。

### Select

基于 `@radix-ui/react-select` 的下拉选择组件套件。包含 Select, SelectTrigger, SelectContent, SelectItem, SelectGroup 等。

### Skeleton

骨架屏占位组件，带切角 clip-path 造型。

### ConfirmDialog

确认对话框。封装 Dialog + Button + AlertTriangle 图标，支持 `destructive`/`warning`/`default` 三种变体。

**Props**: `open`, `onOpenChange`, `title`, `description`, `confirmLabel?`, `cancelLabel?`, `variant?`, `onConfirm`

### CreateHostDialog

Host 创建对话框。西部世界科幻风格，支持工具多选（全选/取消全选）、CPU/内存资源配置、实时错误展示。创建过程中显示 "BUILDING IMAGE..." 状态，超时 5 分钟。

**Props**: `open`, `onOpenChange`, `tools: Tool[]`, `onSubmit: (req: CreateHostRequest) => Promise<void>`

### ErrorBoundary

错误边界组件。捕获子组件渲染异常后展示全屏 SYSTEM FAULT 错误面板，包含 RETRY 按钮。

### Toast / ToastProvider / useToast

Toast 通知系统。支持 `success`/`error`/`warning` 三种类型，固定右下角，4 秒自动消失。

### XtermTerminal

基于 `@xterm/xterm` 的终端组件。通过 WebSocket 连接后端 PTY，支持 FitAddon 自适应、AttachAddon 数据绑定。自动从 CSS 变量读取终端主题，支持背景透明。

**Props**: `wsUrl: string`, `backgroundComponent?: ReactNode`, `theme?`, `allowTransparency?: boolean`

### SessionPipeline

会话管线步骤编辑器。按 system（只读）/ template / user（可编辑）三层展示 PipelineStep 列表。支持添加用户自定义 shell 命令。

**Props**: `steps: PipelineStep[]`, `onChange: (steps: PipelineStep[]) => void`, `readOnly?: boolean`

---

## 3. Agent 业务组件

### AgentPanel

Agent 主面板组件，顶层容器。内部整合 SessionList + TerminalPane + SessionDialogs，通过 `useReducer` 管理完整状态。使用 `usePollingWithBackoff` 轮询刷新会话。支持 `renderCreateDialog` 插槽自定义创建对话框。

**Props**: `apiClient: IAgentApiClient`, `nodeName?: string`, `renderCreateDialog?`, `headerActions?`, `terminalBackground?`

### SessionList

会话列表组件。展示合并后的运行中和已保存会话，支持搜索。每个会话项展示 HostVitalSign 状态指示、DecryptText 名称、操作按钮。使用 ReverieEffect 包裹运行中的会话。

### TerminalPane

终端面板组件。展示选中会话的 XtermTerminal，顶部操作栏含 SYNC STATE 和 VIEW CONFIG 按钮。

### SessionDialogs

会话相关弹窗集合。聚合终止确认、恢复确认、管线查看、SessionConfigEditorDialog（项目级配置编辑器，支持加载/保存/冲突检测）。

### CreateSessionWithTemplateDialog

基于模板创建会话的弹窗。流程：选择模板 → 配置参数 → 预览管线 → 创建。自动从节点本地配置填充环境变量，敏感值脱敏展示。

### TemplateManager

模板管理弹窗。支持新建/编辑/删除/克隆模板。内置模板元数据锁定只读。保存时检测配置冲突。

### NodeConfigPanel

节点配置面板弹窗。管理 Host 默认环境变量（新增/编辑/删除），工作目录由服务端决定、只读。

---

## 4. 工具函数

### cn(...inputs: ClassValue[])

合并 Tailwind CSS 类名。内部使用 `clsx` + `twMerge`。

### maskEnvValue(key, value, sensitive?): string

环境变量值脱敏。按 key 名是否含敏感关键词（password/secret/token 等）自动判断。保留前 4 后 4，不足 8 位全部掩码。

### maskFileContent(path, content): string

文件内容脱敏。JSON 文件递归扫描键值对脱敏，文本文件用正则匹配 `key=value` / `key: value` 模式。

### createRequest(baseUrl?): (url, options?) => Promise<ApiResponse<T>>

创建带 baseUrl 的 HTTP 请求函数。统一处理 30s 超时、JSON 解析、HTTP 错误码。

---

## 5. Hook

### usePollingWithBackout\<T\>(options)

带指数退避的轮询 Hook。

**参数**: `fetchFn`, `baseInterval?`（5000ms）, `maxInterval?`（30000ms）, `enabled?`（true）

**返回**: `{ data, error, isLoading, refresh }`

连续失败按 `baseInterval * 2^failures` 退避。页面不可见时暂停，恢复时立即执行。

---

## 6. 接口

### IAgentApiClient

统一的 Agent API Client 接口。方法：

| 方法                                                                           | 说明               |
| ------------------------------------------------------------------------------ | ------------------ |
| listSessions / createSession / getSession / deleteSession                      | 会话 CRUD          |
| getOutput / sendInput / sendSignal                                             | 终端 I/O           |
| getSavedSessions / restoreSession / saveSessions                               | 持久化与恢复       |
| buildWsUrl                                                                     | WebSocket URL 生成 |
| listTemplates / createTemplate / getTemplate / updateTemplate / deleteTemplate | 模板 CRUD          |
| getTemplateConfig / updateTemplateConfig                                       | 模板全局配置       |
| getSessionConfig / updateSessionConfig                                         | 会话项目级配置     |
| getLocalConfig / updateLocalConfig                                             | 节点本地配置       |
