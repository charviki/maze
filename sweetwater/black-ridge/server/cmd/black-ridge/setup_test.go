package main

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/charviki/maze-cradle/configutil"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/sweetwater-black-ridge/internal/config"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func TestBackgroundRunner_ListenAndServe(t *testing.T) {
	started := make(chan struct{})
	stopped := make(chan struct{})

	runner := newBackgroundRunner("test-bg", logutil.NewNop(), func(stopCh <-chan struct{}) {
		close(started)
		<-stopCh
		close(stopped)
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := runner.ListenAndServe(); err != nil {
			t.Errorf("ListenAndServe returned error: %v", err)
		}
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("runner did not start")
	}

	if err := runner.Shutdown(t.Context()); err != nil {
		t.Errorf("Shutdown returned error: %v", err)
	}

	wg.Wait()

	select {
	case <-stopped:
	default:
		t.Error("runner did not stop")
	}
}

func TestBackgroundRunner_Shutdown_AlreadyStopped(t *testing.T) {
	started := make(chan struct{})
	runner := newBackgroundRunner("test", logutil.NewNop(), func(stopCh <-chan struct{}) {
		close(started)
		<-stopCh
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = runner.ListenAndServe()
	}()

	<-started

	ctx := t.Context()
	if err := runner.Shutdown(ctx); err != nil {
		t.Errorf("first Shutdown error: %v", err)
	}
	wg.Wait()

	if err := runner.Shutdown(ctx); err != nil {
		t.Errorf("second Shutdown should not error: %v", err)
	}
}

func TestBackgroundRunner_Shutdown_Timeout(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	runner := newBackgroundRunner("test", logutil.NewNop(), func(stopCh <-chan struct{}) {
		close(started)
		<-stopCh
		<-release
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = runner.ListenAndServe()
	}()

	<-started

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := runner.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown with timeout should not error: %v", err)
	}
	close(release)
	wg.Wait()
}

func TestCorsMiddleware_NilOrigins(t *testing.T) {
	mw := corsMiddleware(nil)
	if mw == nil {
		t.Error("corsMiddleware with nil should return a non-nil function")
	}
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	if handler == nil {
		t.Error("middleware returned nil handler")
	}
}

func TestCorsMiddleware_WithOrigins(t *testing.T) {
	mw := corsMiddleware([]string{"http://localhost:3000"})
	if mw == nil {
		t.Error("corsMiddleware with origins should return a non-nil function")
	}
}

func TestChainHTTP_NoMiddleware(t *testing.T) {
	called := false
	handler := chainHTTP(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}),
	)
	if handler == nil {
		t.Error("chainHTTP returned nil")
	}
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if !called {
		t.Error("handler should have been called")
	}
}

func TestChainHTTP_MultipleMiddleware(t *testing.T) {
	order := []string{}
	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw1-before")
			next.ServeHTTP(w, r)
			order = append(order, "mw1-after")
		})
	}
	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw2-before")
			next.ServeHTTP(w, r)
			order = append(order, "mw2-after")
		})
	}
	handler := chainHTTP(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "handler")
		}),
		mw1, mw2,
	)
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	expected := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	if len(order) != len(expected) {
		t.Fatalf("order length = %d, want %d: %v", len(order), len(expected), order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("order[%d] = %q, want %q", i, order[i], v)
		}
	}
}

func TestChainHTTP_ResponsePropagation(t *testing.T) {
	handler := chainHTTP(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, "hello")
		}),
	)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if rec.Body.String() != "hello" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "hello")
	}
}

func TestAccessLogMiddleware_LogsStatusAndPath(t *testing.T) {
	var buf bytes.Buffer
	logger := logutil.New("test")
	logger.SetOutput(&buf)

	handler := accessLogMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, "ok")
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/demo", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	output := buf.String()
	if !strings.Contains(output, "POST") || !strings.Contains(output, "/api/demo") || !strings.Contains(output, "status=201") {
		t.Fatalf("日志未包含关键字段: %s", output)
	}
}

func TestGetStaticFS_UsesInjectedFS(t *testing.T) {
	origStaticFiles := staticFiles
	t.Cleanup(func() { staticFiles = origStaticFiles })
	staticFiles = fstest.MapFS{
		"web-dist":            &fstest.MapFile{Mode: fs.ModeDir},
		"web-dist/index.html": &fstest.MapFile{Data: []byte("index")},
		"web-dist/app.js":     &fstest.MapFile{Data: []byte("console.log('ok')")},
	}

	subFS, err := getStaticFS()
	if err != nil {
		t.Fatalf("getStaticFS 返回错误: %v", err)
	}
	data, err := fs.ReadFile(subFS, "app.js")
	if err != nil {
		t.Fatalf("读取静态资源失败: %v", err)
	}
	if string(data) != "console.log('ok')" {
		t.Fatalf("data = %q, want injected asset", string(data))
	}
}

func TestGetStaticFS_ReturnsErrorWhenBundleMissing(t *testing.T) {
	origStaticFiles := staticFiles
	t.Cleanup(func() { staticFiles = origStaticFiles })
	staticFiles = fstest.MapFS{}

	if _, err := getStaticFS(); err == nil {
		t.Fatal("缺少 web-dist 时应返回错误")
	}
}

func TestServeSPA_FallbacksToIndexForUnknownRoute(t *testing.T) {
	origStaticFiles := staticFiles
	t.Cleanup(func() { staticFiles = origStaticFiles })
	staticFiles = fstest.MapFS{
		"web-dist":            &fstest.MapFile{Mode: fs.ModeDir},
		"web-dist/index.html": &fstest.MapFile{Data: []byte("<html>index</html>")},
		"web-dist/app.js":     &fstest.MapFile{Data: []byte("console.log('ok')")},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/hosts/123", nil)
	serveSPA(logutil.NewNop(), rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "index") {
		t.Fatalf("body = %q, want SPA index", rec.Body.String())
	}
}

func TestServeSPA_ServesStaticAsset(t *testing.T) {
	origStaticFiles := staticFiles
	t.Cleanup(func() { staticFiles = origStaticFiles })
	staticFiles = fstest.MapFS{
		"web-dist":            &fstest.MapFile{Mode: fs.ModeDir},
		"web-dist/index.html": &fstest.MapFile{Data: []byte("<html>index</html>")},
		"web-dist/app.js":     &fstest.MapFile{Data: []byte("console.log('ok')")},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/app.js", nil)
	serveSPA(logutil.NewNop(), rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "console.log") {
		t.Fatalf("body = %q, want JS asset", rec.Body.String())
	}
}

func TestNewHTTPServer_RoutesHealthAPIAndSPA(t *testing.T) {
	origStaticFiles := staticFiles
	t.Cleanup(func() { staticFiles = origStaticFiles })
	staticFiles = fstest.MapFS{
		"web-dist":            &fstest.MapFile{Mode: fs.ModeDir},
		"web-dist/index.html": &fstest.MapFile{Data: []byte("<html>index</html>")},
	}

	gwmux := gwruntime.NewServeMux()
	if err := gwmux.HandlePath("GET", "/api/echo", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		_, _ = io.WriteString(w, "gateway")
	}); err != nil {
		t.Fatalf("注册 gateway path 失败: %v", err)
	}

	cfg := &config.Config{
		Server: config.ServerConfig{
			ServerConfig: configutil.ServerConfig{
				ListenAddr: ":8080",
			},
		},
		Workspace: config.WorkspaceConfig{
			StateDir: t.TempDir(),
		},
	}
	server, templateStore := newHTTPServer(cfg, nil, logutil.NewNop(), gwmux)
	if templateStore == nil {
		t.Fatal("templateStore 不应为 nil")
	}

	health := httptest.NewRecorder()
	server.Handler.ServeHTTP(health, httptest.NewRequest("GET", "/health", nil))
	if health.Code != http.StatusOK || !strings.Contains(health.Body.String(), `"status":"ok"`) {
		t.Fatalf("health 响应异常: code=%d body=%q", health.Code, health.Body.String())
	}

	api := httptest.NewRecorder()
	server.Handler.ServeHTTP(api, httptest.NewRequest("GET", "/api/echo", nil))
	if api.Body.String() != "gateway" {
		t.Fatalf("api body = %q, want gateway", api.Body.String())
	}

	spa := httptest.NewRecorder()
	server.Handler.ServeHTTP(spa, httptest.NewRequest("GET", "/unknown", nil))
	if !strings.Contains(spa.Body.String(), "index") {
		t.Fatalf("spa body = %q, want index fallback", spa.Body.String())
	}
}

func TestGrpcListenAddrFor(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want string
	}{
		{name: "default", addr: "", want: ":9090"},
		{name: "host with port", addr: "127.0.0.1:19090", want: ":19090"},
		{name: "scheme-less hostname", addr: "agent.example:29090", want: ":29090"},
		{name: "already listen addr", addr: ":39090", want: ":39090"},
		{name: "raw token", addr: "grpc-socket", want: "grpc-socket"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := grpcListenAddrFor(tt.addr); got != tt.want {
				t.Fatalf("grpcListenAddrFor(%q) = %q, want %q", tt.addr, got, tt.want)
			}
		})
	}
}
