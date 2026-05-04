package casbin

import (
	"fmt"
	"net/http"

	"github.com/charviki/maze-cradle/auth"
	"github.com/charviki/maze-cradle/httputil"
)

// HTTPResourceResolver 将 HTTP 请求映射成资源动作。
type HTTPResourceResolver func(r *http.Request) (ResourceAction, bool, error)

// RequireHTTPPermission 返回 HTTP 侧的权限检查中间件。
// 该中间件用于 grpc-gateway 直连 server handler 的场景，补足 REST 路径不会经过 gRPC interceptor 的缺口。
func RequireHTTPPermission(
	enforcer *Enforcer,
	extractUser auth.UserInfoExtractor,
	resolve HTTPResourceResolver,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ra, protected, err := resolve(r)
			if err != nil {
				httputil.Error(w, r, http.StatusInternalServerError, err.Error())
				return
			}
			if !protected {
				next.ServeHTTP(w, r)
				return
			}

			user, err := extractUser(r.Context())
			if err != nil {
				httputil.Error(w, r, http.StatusUnauthorized, "unauthorized")
				return
			}
			if user == nil || user.SubjectKey == "" {
				httputil.Error(w, r, http.StatusUnauthorized, "unauthorized")
				return
			}

			allowed, err := enforcer.Enforce(user.SubjectKey, ra.Resource, ra.Action)
			if err != nil {
				httputil.Error(w, r, http.StatusInternalServerError, fmt.Sprintf("permission check failed: %v", err))
				return
			}
			if !allowed {
				httputil.Error(w, r, http.StatusForbidden, "forbidden")
				return
			}

			next.ServeHTTP(w, r.WithContext(auth.WithUserInfo(r.Context(), user)))
		})
	}
}
