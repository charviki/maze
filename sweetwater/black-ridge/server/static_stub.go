//go:build !embed

package main

import "embed"

// staticFiles 空占位，在没有 web-dist 的环境（CI、本地 Go 检查）下使用
// Docker 构建时通过 -tags embed 切换到真实 embed 版本
var staticFiles embed.FS
