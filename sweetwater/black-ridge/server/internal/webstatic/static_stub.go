//go:build !embed

package webstatic

import "embed"

// Files 空占位，在没有前端产物的开发/CI 环境下允许 Go 代码先编译。
var Files embed.FS
