package configutil

import "strings"

// ServerConfig 提供跨服务复用的基础服务端配置字段与便捷判断。
// 具体模块可以内嵌它，再补充各自专属字段，避免重复实现 dev/CORS 判断。
type ServerConfig struct {
	ListenAddr     string   `yaml:"listen_addr"`
	JWTSecret      string   `yaml:"jwt_secret"`
	AllowedOrigins []string `yaml:"allowed_origins,omitempty"`
}

// IsDevMode 已废弃：jwt.secret 现为必填项，不再支持空 secret 开发模式。
// 保留函数签名以兼容调用方，始终返回 false。
func (c ServerConfig) IsDevMode() bool {
	return false
}

// Origins 返回已去空白的来源列表。
func (c ServerConfig) Origins() []string {
	return SplitCSV(strings.Join(c.AllowedOrigins, ","))
}
