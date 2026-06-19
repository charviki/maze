import { defineConfig } from '@hey-api/openapi-ts';

// Phase 0：A/B 对比配置。
// - @hey-api/typescript（生成 types.gen.ts message 接口）+ @hey-api/sdk（生成 sdk.gen.ts 函数）
//   是默认 plugin，无需显式声明；仅显式指定 client-fetch 作为 HTTP 客户端。
// - input 复用 A1 修正后的 swagger.json（openapiv2 插件直接产物，含完整枚举）。
export default defineConfig({
  input: '../cradle/api/gen/openapiv2/maze.swagger.json',
  output: 'src/api/gen',
  plugins: [{ name: '@hey-api/client-fetch' }],
});
