# Arrival Gate

## 职责

The Mesa 统一入口门户，两阶段体验（Landing 含登录 → 主界面），以西部世界主题的 Canvas 迷宫 + SVG 结构呈现 6 个子系统导航入口。

## 项目结构

纯前端项目（React + Vite），入口 `src/App.tsx`，组件在 `src/components/`（迷宫粒子/SVG、Landing、模块卡片、状态栏等），认证模块在 `src/auth/`（OIDC 预留接口），Mock 数据在 `src/data/`。

## 核心原则

- **沉浸式体验** — Landing 内嵌登录，迷宫缩小到左上角作为装饰，背景保持连续
- **OIDC 预留** — 认证模块接口化，当前使用硬编码 + localStorage，未来可无缝切换
- **声明式导航** — 模块卡片配置化（MODULES 数组），状态/路由/图标集中管理

## 命令

- `pnpm --filter @maze/arrival-gate dev` — 开发环境 (port 3002)
- `docker build -f the-mesa/arrival-gate/Dockerfile.web -t maze-arrival-gate .` — 独立构建
- 组合构建见 The Mesa 的 `make build-web`

## 依赖

- 依赖: [@maze/fabrication](../../fabrication/skin/AGENTS.md)
- 被依赖: 无

## 详细文档

| 文档 | 内容 |
|------|------|
| [docs/design.md](docs/design.md) | 界面流程、迷宫交互设计、瀑布流、意识层级、登录流程 |
