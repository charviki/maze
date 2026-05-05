# 前端测试规范

这份文档定义仓库内前端测试的长期约束，重点回答以下问题：

- 前端测试应采用什么分层策略
- 各层测试应使用什么工具和模式
- 测试基础设施文件如何组织
- 新增测试时应遵循什么规范
- 禁止使用什么模式

## 测试分层

采用 Testing Trophy 模型（Kent C. Dodds），测试重心放在集成测试层：

| 层级 | 占比 | 测试什么 | 工具 |
|------|------|---------|------|
| 静态分析 | 基座 | 类型安全 + 代码规范 | tsc + eslint |
| 单元测试 | ~30% | 纯函数、工具函数、自定义 hooks | vitest |
| 集成测试 | ~50%（核心） | 组件交互 + API 调用 + 状态管理 | vitest + @testing-library/react + MSW |
| E2E 测试 | ~10% | 关键用户流程 | Playwright（远期规划） |

## 工具栈

| 工具 | 用途 | 必需 |
|------|------|------|
| vitest | 测试运行器 + 断言 + mock | 是 |
| @testing-library/react | 组件测试（用户行为视角） | 是 |
| @testing-library/jest-dom | DOM 断言扩展 | 是 |
| @testing-library/user-event | 真实用户事件模拟 | 是 |
| MSW | 网络层 API mock | 是 |
| @vitest/coverage-v8 | 覆盖率报告 | 是 |
| jsdom | DOM 环境模拟 | 是 |
| @vitejs/plugin-react | vitest 中编译 JSX | 是 |

## 测试文件组织

```
src/
  components/
    Panel.tsx
    Panel.test.tsx           # 测试文件与源文件同目录
  api/
    client.ts
    client.test.ts           # 测试文件与源文件同目录
  utils/
    mask.ts
    mask.test.ts             # 测试文件与源文件同目录
  test/
    setup.ts                 # 全局 setup（jest-dom + MSW lifecycle）
    mocks/
      server.ts              # MSW node server
      handlers.ts            # 默认 API mock handlers
```

**规范**：

- 测试文件与源文件**同目录**，命名为 `*.test.ts` 或 `*.test.tsx`
- 测试基础设施统一放在 `src/test/` 目录
- 不使用 `__tests__/` 子目录模式
- 不使用 `.spec.ts` 命名

## 全局 Setup 规范

每个前端模块必须有 `src/test/setup.ts`，内容统一为：

```typescript
import '@testing-library/jest-dom/vitest';
import { server } from './mocks/server';

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());
```

每个模块的 `vitest.config.ts` 必须配置：

```typescript
test: {
  environment: 'jsdom',
  globals: true,
  setupFiles: ['./src/test/setup.ts'],
}
```

**禁止**在测试文件中手动 `import '@testing-library/jest-dom'`。

## API Mock 规范

### 使用 MSW

所有涉及 HTTP 请求的测试**必须**使用 MSW（Mock Service Worker）进行 mock。

**MSW 文件结构**：

```typescript
// src/test/mocks/server.ts
import { setupServer } from 'msw/node';
import { handlers } from './handlers';
export const server = setupServer(...handlers);
```

```typescript
// src/test/mocks/handlers.ts
import { http, HttpResponse } from 'msw';
export const handlers = [
  http.get('/api/v1/sessions', () => {
    return HttpResponse.json({ sessions: [] });
  }),
];
```

### 在测试中使用 MSW

- 默认 handler 在 `handlers.ts` 中定义，提供通用的成功响应
- 测试用例级别使用 `server.use()` 覆盖默认行为：

```typescript
import { http, HttpResponse } from 'msw';
import { server } from '@/test/mocks/server';

it('should handle 500 error', async () => {
  server.use(
    http.get('/api/v1/sessions', () => {
      return new HttpResponse(null, { status: 500 });
    }),
  );
});
```

### 禁止事项

- **禁止**使用 `vi.stubGlobal('fetch', vi.fn())` 手动 mock fetch
- **禁止**手动构建 `{ ok: true, json: () => ... }` 这样的假 Response 对象
- **禁止**在多个测试文件中重复定义 `mockOk` / `mockError` 辅助函数

## 组件测试规范

### 使用 Testing Library

组件测试**必须**使用 `@testing-library/react`，遵循以下原则：

1. **测试行为，不测实现** — 使用 `screen.getByRole`、`screen.getByText` 等用户可见的查询方式
2. **优先使用 accessible 查询** — `getByRole` > `getByLabelText` > `getByText` > `getByTestId`
3. **避免查询内部实现** — 不使用 `container.querySelector` 直接查 CSS class 或 DOM 结构

### 使用 user-event

所有用户交互测试**必须**使用 `@testing-library/user-event`：

```typescript
import userEvent from '@testing-library/user-event';

it('should submit form', async () => {
  const user = userEvent.setup();
  render(<MyForm onSubmit={onSubmit} />);

  await user.type(screen.getByLabelText('Name'), 'test');
  await user.click(screen.getByRole('button', { name: /submit/i }));

  expect(onSubmit).toHaveBeenCalledWith({ name: 'test' });
});
```

**禁止**在新测试中使用 `fireEvent`。`fireEvent` 只触发单个 DOM 事件，不模拟真实用户操作链。

### 组件 mock 规范

- 使用 `vi.mock()` mock 不支持 jsdom 的组件（Canvas、WebGL、xterm 等）
- mock 时提供 `data-testid` 以便验证组件是否被渲染
- 使用 `vi.importActual()` 保留其余非 Canvas 组件的真实行为

```typescript
vi.mock('./MazeCanvas', () => ({
  MazeCanvas: ({ className }: { className?: string }) => (
    <div data-testid="maze-canvas" className={className} />
  ),
}));
```

### API client mock 规范

对于消费 `IAgentApiClient` 接口的组件测试，使用接口 mock 对象：

```typescript
function createMockApi(overrides: Partial<IAgentApiClient> = {}): IAgentApiClient {
  const ok = <T,>(data: T) => Promise.resolve({ status: 'ok' as const, data });
  return {
    listSessions: vi.fn(() => ok([])),
    createSession: vi.fn(() => ok(undefined)),
    // ... 其余方法
    ...overrides,
  };
}
```

## 单元测试规范

适用于纯函数、工具函数、自定义 hooks：

- **纯函数** — 直接调用并断言返回值
- **自定义 hooks** — 使用 `renderHook` from `@testing-library/react`
- **定时器相关** — 使用 `vi.useFakeTimers()` + `vi.advanceTimersByTime()`

## ESLint 测试规则

所有模块的 `eslint.config.js` 必须包含以下测试文件宽松规则：

```javascript
{
  files: ['**/*.test.{ts,tsx}'],
  rules: {
    '@typescript-eslint/no-unsafe-assignment': 'off',
    '@typescript-eslint/no-unsafe-member-access': 'off',
    '@typescript-eslint/no-unsafe-return': 'off',
    '@typescript-eslint/no-unsafe-call': 'off',
  },
}
```

同时将 `vitest.config.ts` 和 `src/test/setup.ts` 纳入配置文件宽松规则。

## 覆盖率

- 运行 `make coverage-frontend` 生成 4 个模块的覆盖率报告
- 覆盖率报告用于**量化参考**，不设硬性门禁阈值
- 重点关注意外下降的覆盖率，而非追求 100%

## 不测试的组件

以下组件依赖 Canvas API / WebGL / xterm，jsdom 环境不支持，暂不测试：

- `XtermTerminal` — 依赖 @xterm/xterm
- `TerrainBackground` — 依赖 Canvas
- `HexWaterfall` — 依赖 Canvas
- `BootSequence` — 依赖 Canvas
- `RadarView` — 依赖 Canvas
- `ReverieEffect` — 依赖 Canvas
- `GlitchEffect` — 依赖 Canvas

如需测试这些组件的交互逻辑，应将其交互逻辑抽离为不依赖 Canvas 的 hooks 或工具函数，然后对抽离的逻辑编写测试。
