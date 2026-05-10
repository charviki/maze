package transport

import (
	"net/http"

	"github.com/charviki/maze/fabrication/cradle/httputil"

	"github.com/charviki/maze/the-mesa/the-forge/internal/service"
)

// NewHealthHandler 创建健康检查 HTTP handler。
func NewHealthHandler(healthService *service.HealthService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httputil.Success(w, r, map[string]string{"status": healthService.Status()})
	})
}
