package gatewayutil

import "strings"

// LocalGRPCEndpoint 将 gRPC 监听地址转换为本地 dial 地址。
func LocalGRPCEndpoint(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "127.0.0.1" + addr
	}
	return addr
}
