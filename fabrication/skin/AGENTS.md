# Skin AGENTS.md

## 职责

如同剧中的制造部负责 Host 的外观定制，Skin 是西部世界主题的 React 组件库。提供视觉特效组件、基础 UI 组件、Agent 业务组件及工具函数，所有面向使用者的界面都由这里"制造"出来。

## 核心原则

- **主题一致性** — 所有组件必须遵循西部世界科幻视觉风格（切角面板、解密动画、CRT 扫描线等）
- **特效可控** — 动画特效通过 AnimationSettings Context 全局管控，尊重用户 `prefers-reduced-motion` 系统设置
- **API 抽象** — Agent 业务组件仅依赖 `IAgentApiClient` 接口，不绑定具体传输层实现
- **敏感数据保护** — 涉及环境变量和文件内容展示时，必须使用 `maskEnvValue` / `maskFileContent` 脱敏

## 依赖关系

- 依赖: React, @xterm/xterm (+addon-fit/addon-attach/addon-webgl), @radix-ui/react-dialog, @radix-ui/react-select, @radix-ui/react-slot, class-variance-authority, clsx, tailwind-merge, lucide-react, tailwindcss
- 被依赖: [behavior-panel/web](../../mesa-hub/behavior-panel/AGENTS.md), [black-ridge/web](../../sweetwater/black-ridge/AGENTS.md)（消费组件和 IAgentApiClient）

## 关键文件

| 路径                               | 职责                                                                                              | 文档同步                            |
| ---------------------------------- | ------------------------------------------------------------------------------------------------- | ----------------------------------- |
| src/index.ts                       | 模块入口，统一导出所有组件、工具函数和类型                                                        | 本文件                              |
| src/types.ts                       | 全局类型定义（Session, Pipeline, Template, Config, Tool, CreateHostRequest, HostStatus, Host 等） | 本文件                              |
| src/api.ts                         | IAgentApiClient 接口定义，规范所有 Agent 交互方法                                                 | 本文件                              |
| src/utils.ts                       | cn 类名合并工具                                                                                   | 本文件                              |
| src/utils/mask.ts                  | 环境变量和文件内容脱敏工具                                                                        | 本文件                              |
| src/utils/request.ts               | HTTP 请求工厂函数 createRequest                                                                   | 本文件                              |
| src/hooks/usePollingWithBackoff.ts | 带指数退避的轮询 Hook                                                                             | 本文件                              |
| src/components/ui/                 | 视觉特效和基础 UI 组件（20 个文件）                                                               | [components.md](docs/components.md) |
| src/components/agent/              | Agent 业务组件（6 个文件）                                                                        | [components.md](docs/components.md) |

## 详细文档

| 文档                                     | 内容                                      |
| ---------------------------------------- | ----------------------------------------- |
| [docs/components.md](docs/components.md) | 组件清单 + 工具函数 + Hook + 接口使用方式 |
