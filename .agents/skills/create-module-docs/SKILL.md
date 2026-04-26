---
name: create-module-docs
description: 当新增代码模块时，按标准模板创建 AGENTS.md + docs/ 目录结构，并更新根 AGENTS.md 的模块索引
---

# Create Module Docs

## When to Use

当用户新建一个代码模块（新目录），或你检测到某个模块目录缺少 AGENTS.md 时，使用此 Skill 引导创建标准化文档。

触发场景：
- 用户说"新建一个模块"、"创建新组件"、"新增 xxx 目录"
- 你发现某个代码目录没有 AGENTS.md

## Instructions

### Step 1: 确定模块类型

根据模块内容判断类型：

| 模块类型 | docs/ 结构 |
|----------|-----------|
| Go 后端服务 | docs/architecture.md + docs/api.md |
| Go 共享库 | docs/packages.md |
| 前端组件库 | docs/components.md |
| 其他 | 根据实际情况决定 |

### Step 2: 创建 AGENTS.md

在新模块根目录创建 `AGENTS.md`，使用以下模板：

```markdown
# [模块名] AGENTS.md

## 职责
（1-2 句话：这个模块做什么、不做什么。严格基于代码，不虚构）

## 核心原则
（2-4 条模块特有的约束）

## 依赖关系
- 依赖: （依赖了哪些其他模块）
- 被依赖: （被谁依赖）

## 关键文件
| 路径 | 职责 | 文档同步 |
|------|------|----------|
| ... | ... | → docs/xxx.md 或 本文件 |

## 详细文档
| 文档 | 内容 |
|------|------|
| docs/xxx.md | ... |
```

**"文档同步"列**：标注每个代码文件对应的文档位置，修改代码时一眼就知道该同步哪个文档。

### Step 3: 创建 docs/ 目录

根据 Step 1 确定的类型创建对应文件：

**Go 后端服务**：
- `docs/architecture.md` — 架构设计、数据流、部署
- `docs/api.md` — 全部 HTTP API 端点

**Go 共享库**：
- `docs/packages.md` — 各子包说明 + 导出 API

**前端组件库**：
- `docs/components.md` — 组件清单 + 使用方式

### Step 4: 更新根 AGENTS.md

在项目根目录的 `AGENTS.md` 的"模块索引"表格中添加新模块条目：

```markdown
| 模块名 | 目录路径 | 职责简述 | AGENTS.md + docs/ |
```

### Step 5: 验证

- 确认 AGENTS.md 中所有路径指向实际存在的文件
- 确认 docs/ 中的内容基于代码，无虚构
- 确认根 AGENTS.md 的模块索引已更新

## 重要约束

- 所有文档内容必须严格基于代码，严禁虚构
- 文档用中文编写
- 代码注释用中文
