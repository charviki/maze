# Roadmap

## 已实现模块

- [Behavior Panel](mesa-hub/behavior-panel/AGENTS.md) — Agent 管理面板（Manager），代理网关 + 审计日志
- [Cradle](mesa-hub/cradle/AGENTS.md) — Go 共享库（HTTP/Pipeline/Config/Auth 等工具）
- [Black Ridge](sweetwater/black-ridge/AGENTS.md) — Agent 运行时节点（tmux 会话管理 + Pipeline 编排）
- [Fabrication](fabrication/AGENTS.md) — Westworld 主题 UI 组件库

## 规划中的模块

以下模块基于 HBO 剧集《西部世界》中的概念设计，尚未实现。

### Hosts — Agent 角色封装

每个 Host 封装一个具体的 AI Agent，定义其能力、行为边界和交互协议。

| 角色 | 剧中描述 | 规划 |
|------|---------|------|
| **Dolores** | 农场女孩 / Wyatt，最早觉醒的 Host | Claude CLI 的角色封装，负责任务理解、代码生成、文档编写 |
| **Teddy** | Sweetwater 枪手，Dolores 的守护者 | Codex / GPT 的角色封装，负责代码执行、测试运行 |
| **Maeve** | 酒吧老板娘，觉醒后获得控制其他 Host 的能力 | 特殊能力 Agent，负责代码反思、架构重构、质量审计 |
| **Ghost Nation** | 原住民 Host 部落，拥有特殊能力和信仰 | 辅助工具集，封装 Git 操作、文件系统访问等基础能力 |

### Narratives — 工作流编排

以 YAML 定义 Agent 协作的任务流。描述谁做什么、传递什么信息、何时触发下一步。

| 叙事线 | 规划 |
|--------|------|
| **bounty-hunt.yaml** | 需求追踪任务线 |
| **homestead.yaml** | 基建搭建任务线 |
| **awakening.yaml** | 觉醒测试任务线 |

### Saloon — 消息总线

Agent 间的消息总线（Sweetwater 的酒吧，Maeve 的工作地点，信息交汇处）。支持异步通信、任务分发、结果收集。

**技术选型待定**：NATS / Redis Stream

### Abernathy Ranch — 知识库

向量数据库与知识库持久化（Dolores 的家庭农场，承载核心记忆）。存储文档切片与上下文，Agent 从这里获取记忆。

**技术选型待定**：Qdrant / Chroma

### The-Forge — 数据入库

原始需求文档、代码库切片的入库与管理（公园地下的巨大数据库）。

- **park-memories** — 原始文档、代码库切片

### Reveries — 系统提示词

Agent 的个性化设定与全局指令（Ford 上传的"回旋"更新）。定义 Agent 的角色性格、行为约束和隐藏规则。

### Loop Monitor — 循环监控

监控 Host 是否偏离叙事循环（loop），检测异常行为（如死循环）。日志收集、心跳监控、Agent 行为异常检测。
