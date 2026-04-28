package middleware

import (
	"context"
	"net/http"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
)

func TestCORS_OptionsRequest(t *testing.T) {
	handler := CORS()

	rc := app.NewContext(0)
	rc.Request.SetMethod("OPTIONS")
	rc.Request.SetRequestURI("/test")
	rc.SetHandlers(app.HandlersChain{handler})

	handler(context.Background(), rc)

	if rc.Response.StatusCode() != http.StatusNoContent {
		t.Errorf("状态码 = %d, 期望 %d", rc.Response.StatusCode(), http.StatusNoContent)
	}

	origin := string(rc.Response.Header.Get("Access-Control-Allow-Origin"))
	if origin != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, 期望 %q", origin, "*")
	}
	methods := string(rc.Response.Header.Get("Access-Control-Allow-Methods"))
	if methods == "" {
		t.Error("Access-Control-Allow-Methods 为空, 期望非空")
	}
	headers := string(rc.Response.Header.Get("Access-Control-Allow-Headers"))
	if headers == "" {
		t.Error("Access-Control-Allow-Headers 为空, 期望非空")
	}
}

func TestCORS_NormalRequest(t *testing.T) {
	handler := CORS()
	called := false

	next := func(ctx context.Context, rc *app.RequestContext) {
		called = true
	}

	rc := app.NewContext(0)
	rc.Request.SetMethod("GET")
	rc.Request.SetRequestURI("/test")
	rc.SetHandlers(app.HandlersChain{handler, next})

	handler(context.Background(), rc)

	if !called {
		t.Error("期望 next 被调用（普通请求应放行），但未调用")
	}

	origin := string(rc.Response.Header.Get("Access-Control-Allow-Origin"))
	if origin != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, 期望 %q", origin, "*")
	}
}
