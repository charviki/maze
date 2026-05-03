package configutil

import "strings"

// ServerConfig 提供跨服务复用的基础服务端配置字段与便捷判断。
// 具体模块可以内嵌它，再补充各自专属字段，避免重复实现 dev/CORS 判断。
type ServerConfig struct {
	ListenAddr     string   `yaml:"listen_addr"`
	AuthToken      string   `yaml:"auth_token"`
	AllowedOrigins []string `yaml:"allowed_origins,omitempty"`
}

// IsDevMode 当鉴权令牌为空时视为开发模式。
func (c ServerConfig) IsDevMode() bool {
	return c.AuthToken == ""
}

// Origins 返回已去空白的来源列表。
func (c ServerConfig) Origins() []string {
	return SplitCSV(strings.Join(c.AllowedOrigins, ","))
}
