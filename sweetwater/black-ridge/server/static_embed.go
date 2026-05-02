//go:build embed

package main

import "embed"

// staticFiles 嵌入前端构建产物，实现单二进制部署
// 仅在 -tags embed 时编译，需要 web-dist 目录存在
//go:embed all:web-dist
var staticFiles embed.FS
