# AGENTS.md 编写规范

## 什么是 AGENTS.md

AGENTS.md 是由 OpenAI 发起、Google/GitHub/Cursor/Windsurf 等 20+ AI 编码工具共同支持的**跨工具开放标准**。定位为 "README for agents"——一个存放 README 不适合但 coding agent 需要的技术上下文的 Markdown 文件。

**来源**：[agents.md](https://agents.md/) · [agentsmd.io](https://agentsmd.io/) · [agent.md 开放规范 (GitHub)](https://github.com/agentmd/agent.md)

## 与传统 README 的区别

| <br /> | README.md           | AGENTS.md         |
| ------ | ------------------- | ----------------- |
| 读者     | 人类开发者               | AI Coding Agent   |
| 内容     | 项目介绍、快速开始、贡献指南      | 构建命令、代码约定、架构决策    |
| 风格     | 面向阅读体验，可以有 emoji/截图 | 面向解析执行，偏好精确指令     |
| 变更频率   | 低频（项目不变就不改）         | 迭代演进（AI 反复犯错就加进去） |

## 文件位置与分层发现

```
仓库根目录/AGENTS.md       # 全局项目指令
子目录/AGENTS.md            # 子系统/子包指令（覆盖父级）
~/.config/AGENTS.md         # 用户全局偏好（可选）
```

Agent 读取规则：**最近文件优先**（编辑子目录文件时，子目录 AGENTS.md 覆盖根级）。

## 推荐内容结构

无强制章节，以下为社区共识的推荐结构：

```markdown
# Project Overview        — 一句话项目简介
# Setup Commands          — 安装/构建/运行命令
# Architecture Overview   — 简要架构描述（2-3 句话）
# Code Style              — 代码约定
# Testing                 — 测试命令和规范
# Git Workflows           — 分支/PR/提交规范（可选）
```

**本项目的模板**（所有模块 AGENTS.md 统一遵循）：

```markdown
# [模块名]

## 职责
（1 句话）

## 项目结构
（2-3 句话描述目录组织，不枚举文件）

## 核心原则
（仅模块特有约束，2-4 条）

## 命令
（agent 可能猜错的构建/测试/代码生成命令）

## 依赖
- 依赖: [模块](路径)
- 被依赖: [模块](路径)

## 详细文档
（链接表，指向 docs/ 下文件）
```

## 维护原则

1. **只放 Agent 需要的** — 如果 Agent 用 LS/Glob 就能看出来的（如标准语言约定、目录结构细目），不放
2. **迭代演进** — 观察到 AI 反复犯同一个错误时，把那件事写进去；不要试图一次写完美
3. **禁止搬运代码** — 不在 AGENTS.md 中复制函数签名、类型定义、API 端点清单。Proto/Swagger 已是 API 契约源
4. **不枚举文件** — 不创建逐文件的表格或清单。文件查找是 Agent 工具链（Glob/LS/Grep）的职责
5. **架构深度知识放 docs/** — 架构图、数据流、设计决策放在 docs/ 下，AGENTS.md 只放链接

