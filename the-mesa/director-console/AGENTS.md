# Director Console

## 职责

The Mesa 控制台前端，负责节点视图、会话操作、模板管理和控制核心 API 的浏览器侧交互。

## 项目结构

纯前端项目（React + Vite），入口 `src/App.tsx`，API 适配在 `src/api/`，核心界面组件在 `src/components/`，共享 Fabrication 组件在 `src/components/shared/`，工具函数在 `src/utils/`。

### Fabrication 资源管理

`src/components/shared/` 提供 Fabrication 资源（Rules、Skills、GitKeys、MCP）的共享基础设施：
- `types.ts` — 通用列表/详情类型定义
- `useFabricationList.ts` — 统一 CRUD + 分页 hook
- `DetailPanelShell.tsx` — 创建/编辑面板外壳
- `FormLabel.tsx` — 表单标签组件
- `FabricationListColumn.tsx` — 列表列定义工具

每种资源（rules、skills、git-keys、mcp）由独立的 `XxxList.tsx` + `XxxDetailPanel.tsx` 组成，消费上述共享组件。`FabricationPanel.tsx` 作为顶层容器组装四种资源的 Tab 视图。

## 核心原则

- **控制台优先** — 所有页面围绕节点、会话、模板等控制面操作展开，不承载后端业务逻辑
- **接口隔离** — 浏览器侧只依赖 HTTP API / Web 端约定，不直接感知 Director Core 内部实现
- **正式入口** — 控制台前端以 `/director-console/` 作为正式 Web 路由

## 命令

- `pnpm --filter @maze/director-console dev` — 开发环境 (port 3000)
- `pnpm --filter @maze/director-console test` — 运行前端测试
- 组合构建见 The Mesa 的 `make build-web`

## 依赖

- 依赖: [@maze/fabrication](../../fabrication/skin/AGENTS.md)
- 被依赖: [The Mesa](../AGENTS.md)
