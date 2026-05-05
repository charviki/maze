# Skin

## 职责

西部世界主题 React 组件库，提供视觉特效组件、基础 UI 组件、Agent 业务组件及工具函数。

## 项目结构

组件在 src/components/（分 ui/ 视觉特效 + agent/ Agent 业务），API 接口在 src/api.ts（ISessionApi/ITemplateApi/IConfigApi/ILocalConfigApi 子接口），工具函数在 src/utils/ + src/hooks/，自动生成的 TypeScript SDK 在 src/api/gen/。

## 核心原则

- **主题一致性** — 所有组件遵循西部世界科幻视觉风格（切角面板、解密动画、CRT 扫描线）
- **特效可控** — 动画特效通过 AnimationSettings Context 全局管控，尊重 prefers-reduced-motion
- **API 抽象** — Agent 业务组件仅依赖子接口，不绑定具体传输层
- **敏感数据保护** — 环境变量和文件内容展示时必须使用脱敏工具

## 命令

- `pnpm --filter @maze/fabrication build` — 构建
- `pnpm --filter @maze/fabrication typecheck` — 类型检查

## 依赖

- 依赖: React, @xterm/xterm, @radix-ui/*, tailwindcss
- 被依赖: [portal](../../mesa-hub/portal/AGENTS.md), [behavior-panel](../../mesa-hub/behavior-panel/AGENTS.md), [black-ridge](../../sweetwater/black-ridge/AGENTS.md)

## 详细文档

| 文档 | 内容 |
|------|------|
| [docs/components.md](docs/components.md) | 组件分类总览 + 使用方式 |
