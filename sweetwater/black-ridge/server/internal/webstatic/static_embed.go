//go:build embed

package webstatic

import "embed"

// Files 嵌入前端构建产物，实现单二进制部署。
// 选择独立包而非 cmd 目录，是为了让 `cmd/black-ridge` 入口与 embed 资源可以解耦。
//
//go:embed all:web-dist
var Files embed.FS
