# Portal AGENTS.md

## 职责

The Maze 的统一入口门户，模拟 Delos 的 Mesa Hub 控制中心。通过沉浸式的两阶段体验（Landing（含登录）→ 主界面），将用户引入 Westworld 风格的管理操作台。Landing 页以交互式迷宫（Canvas 粒子 + SVG 结构）和经典台词营造氛围，内嵌登录表单完成身份验证，主界面以卡片网格呈现 6 个子系统的导航入口，其中仅 Behavior Panel 可用，其余为未来扩展的锁定占位。

## 核心原则

- **沉浸式体验** — Landing 页内嵌登录表单，不跳转页面，迷宫缩小到左上角作为装饰，背景保持连续。不复用 Manager 的 BootSequence 避免同质化
- **OIDC 预留** — 认证模块（auth.ts）接口化，当前使用硬编码用户 + localStorage，未来可无缝切换到 OIDC/SSO 流程
- **声明式导航** — 模块卡片配置化（MODULES 数组），状态、路由、图标、描述集中管理
- **Westworld 主题** — 融入经典台词、交互式迷宫、THE MAZE IS NOT MEANT FOR YOU 等元素

## 依赖关系

- 依赖: [@maze/fabrication](../../fabrication/skin/AGENTS.md) (UI 组件库: DecryptText, TerrainBackground, HexWaterfall, GlitchEffect, ReverieEffect, HostVitalSign, RadarView, Button, Input, AnimationSettings, ErrorBoundary, Toast)
- 被依赖: 无

## 关键文件

| 路径 | 职责 |
|------|------|
| web/src/App.tsx | 主组件：两阶段渲染（Landing → PortalLayout），LandingPage 内部处理登录 |
| web/src/auth/auth.ts | 认证服务：login/logout/isAuthenticated/getCurrentUser（OIDC 预留接口） |
| web/src/data/mock-data.ts | 集中 Mock 数据源：节点、主机、会话、台词、诊断、事件等 |
| web/src/components/LandingPage.tsx | Landing 页：迷宫（Canvas+SVG）+ 台词轮播 + ENTER THE PARK + 内嵌登录表单 |
| web/src/components/MazeCanvas.tsx | Canvas 粒子系统：200+ 轨道粒子、鼠标斥力场、中心涟漪爆发、拖尾效果 |
| web/src/components/MazeSvg.tsx | 增强版 SVG 迷宫：多层发光、路径光点动画、中心光晕、环标签 |
| web/src/components/ModuleCard.tsx | 模块入口卡片：在线/锁定两种状态，hover 展示诊断信息 |
| web/src/components/SystemMetric.tsx | 侧栏指标卡片：Panel 容器 + 数值展示 |
| web/src/components/StatusBar.tsx | 底部状态栏：实时时钟 + 状态指标 + Westworld 台词轮播 |
| web/src/components/EventFeed.tsx | 侧栏事件流：随机系统事件 15-30s 间隔，DecryptText 解密最新条目 |
| web/src/components/ConsciousnessBar.tsx | 意识层级进度条：0-100 刻度 + 呼吸光效 + DORMANT/STIRRING/AWAKENING/CONSCIOUS 标签 |
| web/package.json | 包配置：@maze/portal-web，依赖 @maze/fabrication |
| web/vite.config.ts | Vite 配置：port 3002，build to ../web-dist |
| web/tailwind.config.js | Tailwind 配置：复用 fabrication preset |

### 保留但不再参与主流程

| 路径 | 说明 |
|------|------|
| web/src/auth/AuthGate.tsx | 认证网关：登录已内嵌 LandingPage，此组件保留供未来独立使用 |
| web/src/auth/LoginPage.tsx | 独立登录页：保留供未来 OIDC 回调等场景使用 |

## 设计要点

### 界面流程

```
Landing Page（含登录）→ Portal 主界面
```

Landing 页内嵌登录表单，点击 "ENTER THE PARK" 后迷宫缩小到左上角，表单从中央滑入。登录成功后 fade-out 进入 Portal。

### 交互式迷宫

Canvas 粒子 + SVG 结构的混合方案：

- **Canvas 粒子层**（MazeCanvas）：200+ 粒子在 6 条环形轨道上流动，带拖尾效果。鼠标靠近时粒子被推开（斥力场），进入中心触发能量涟漪扩散
- **SVG 结构层**（MazeSvg）：多层 feGaussianBlur 发光、4 个光点沿轨道滑动、中心 3 层光晕。鼠标高亮圆环并显示层级标签（MEMORY → REVERIE → IMPROVISATION → SELF → CONSCIOUSNESS → ...）
- 迷宫可复用，登录阶段缩小到左上角作为装饰，点击可返回首页

### 双层瀑布流

Landing 页左右各两层 HexWaterfall：外层宽 8rem / opacity 0.08，内层宽 4rem / opacity 0.2，营造深度感。

### 意识层级

侧栏 `ConsciousnessBar` 显示当前意识水平（Mock 值 82%），10 刻度进度条 + 呼吸光效 + 阶段标签。`EventFeed` 每 15-30 秒产生随机系统事件，最新条目使用 `DecryptText` 解密效果。

### 经典台词

底栏和 Landing 页共用 Westworld 经典台词库（10 条），通过 `DecryptText` 解密效果轮播展示。

### 模块卡片

6 个子系统：Behavior Panel（在线）、The Forge、Saloon、Loop Monitor、Reveries、Abernathy Ranch（锁定）。锁定卡片底部显示 "THE MAZE IS NOT MEANT FOR YOU"。

## 构建集成

### Dockerfile

| 文件 | 说明 |
|------|------|
| `mesa-hub/portal/Dockerfile.web` | Portal 独立构建（仅 Portal + nginx） |
| `mesa-hub/behavior-panel/Dockerfile.web` | 组合构建（Portal + Behavior Panel + nginx），docker-compose 使用 |

两个 Dockerfile 都需要复制 `pnpm-workspace.yaml` 中声明的**所有**子项目的 `package.json`（pnpm 要求完整 workspace 声明），但只复制当前构建所需的源码。Vite 不设 `base`，由 nginx `alias` 指令处理 `/portal/` 子路径映射。

### nginx 路由（组合构建）

| 路径 | 说明 |
|------|------|
| `/` | 302 重定向到 `/portal/` |
| `/portal/` | Portal 首页（alias → nginx/html/portal/） |
| `/behavior-panel/` | Behavior Panel SPA |
| `/api/` | 反向代理到 agent-manager:8080 |

### 开发环境

```bash
pnpm --filter @maze/portal-web dev    # port 3002
```

### 生产构建

```bash
# 仅 Portal
docker build -f mesa-hub/portal/Dockerfile.web -t maze-portal .

# 组合构建（Portal + Behavior Panel）
docker build -f mesa-hub/behavior-panel/Dockerfile.web -t maze-web .
```

## 详细文档

暂无额外文档。部署相关细节参见上方「构建集成」和 [Docker 构建规范](../../fabrication/docs/docker-build-guide.md)。
