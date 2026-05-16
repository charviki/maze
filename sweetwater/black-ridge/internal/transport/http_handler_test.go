package transport

import (
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/charviki/maze/fabrication/cradle/configutil"
	"github.com/charviki/maze/fabrication/cradle/logutil"
	"github.com/charviki/sweetwater-black-ridge/internal/config"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	gorillaws "github.com/gorilla/websocket"

	"github.com/charviki/maze/fabrication/cradle/httputil"
)

func TestCORSMiddleware_NilOrigins(t *testing.T) {
	mw := httputil.CORSMiddleware(nil)
	if mw == nil {
		t.Error("CORSMiddleware with nil should return a non-nil function")
	}
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	if handler == nil {
		t.Error("middleware returned nil handler")
	}
}

func TestCORSMiddleware_WithOrigins(t *testing.T) {
	mw := httputil.CORSMiddleware([]string{"http://localhost:3000"})
	if mw == nil {
		t.Error("CORSMiddleware with origins should return a non-nil function")
	}
}

func TestChainHTTP_NoMiddleware(t *testing.T) {
	called := false
	handler := httputil.ChainHTTP(
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
	handler := httputil.ChainHTTP(
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
	handler := httputil.ChainHTTP(
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

	handler := httputil.AccessLogMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		t.Fatalf("log missing key fields: %s", output)
	}
}

func TestAccessLogMiddlewarePreservesWebSocketUpgrade(t *testing.T) {
	handler := httputil.ChainHTTP(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := httputil.NewUpgrader(nil).Upgrade(w, r, nil)
			if err != nil {
				t.Fatalf("upgrade failed: %v", err)
			}
			_ = conn.Close()
		}),
		httputil.AccessLogMiddleware(nil),
	)

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := gorillaws.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket failed: %v", err)
	}
	_ = conn.Close()
}

func TestGetStaticFS_UsesInjectedFS(t *testing.T) {
	staticFS := fstest.MapFS{
		"web-dist":            &fstest.MapFile{Mode: fs.ModeDir},
		"web-dist/index.html": &fstest.MapFile{Data: []byte("index")},
		"web-dist/app.js":     &fstest.MapFile{Data: []byte("console.log('ok')")},
	}

	subFS, err := getStaticFS(staticFS)
	if err != nil {
		t.Fatalf("getStaticFS returned error: %v", err)
	}
	data, err := fs.ReadFile(subFS, "app.js")
	if err != nil {
		t.Fatalf("read static asset failed: %v", err)
	}
	if string(data) != "console.log('ok')" {
		t.Fatalf("data = %q, want injected asset", string(data))
	}
}

func TestGetStaticFS_ReturnsErrorWhenBundleMissing(t *testing.T) {
	staticFS := fstest.MapFS{}

	if _, err := getStaticFS(staticFS); err == nil {
		t.Fatal("missing web-dist should return error")
	}
}

func TestServeSPA_FallbacksToIndexForUnknownRoute(t *testing.T) {
	staticFS := fstest.MapFS{
		"web-dist":            &fstest.MapFile{Mode: fs.ModeDir},
		"web-dist/index.html": &fstest.MapFile{Data: []byte("<html>index</html>")},
		"web-dist/app.js":     &fstest.MapFile{Data: []byte("console.log('ok')")},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/hosts/123", nil)
	serveSPA(logutil.NewNop(), staticFS, rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "index") {
		t.Fatalf("body = %q, want SPA index", rec.Body.String())
	}
}

func TestServeSPA_ServesStaticAsset(t *testing.T) {
	staticFS := fstest.MapFS{
		"web-dist":            &fstest.MapFile{Mode: fs.ModeDir},
		"web-dist/index.html": &fstest.MapFile{Data: []byte("<html>index</html>")},
		"web-dist/app.js":     &fstest.MapFile{Data: []byte("console.log('ok')")},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/app.js", nil)
	serveSPA(logutil.NewNop(), staticFS, rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "console.log") {
		t.Fatalf("body = %q, want JS asset", rec.Body.String())
	}
}

func TestNewHTTPServer_RoutesHealthAPIAndSPA(t *testing.T) {
	staticFS := fstest.MapFS{
		"web-dist":            &fstest.MapFile{Mode: fs.ModeDir},
		"web-dist/index.html": &fstest.MapFile{Data: []byte("<html>index</html>")},
	}

	gwmux := gwruntime.NewServeMux()
	if err := gwmux.HandlePath("GET", "/api/echo", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		_, _ = io.WriteString(w, "gateway")
	}); err != nil {
		t.Fatalf("register gateway path failed: %v", err)
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

	server, templateStore := NewHTTPServer(HTTPHandlerParams{
		Config:       cfg,
		Logger:       logutil.NewNop(),
		GWMux:        gwmux,
		StaticFiles:  staticFS,
	})
	if templateStore == nil {
		t.Fatal("templateStore should not be nil")
	}

	health := httptest.NewRecorder()
	server.Handler.ServeHTTP(health, httptest.NewRequest("GET", "/health", nil))
	if health.Code != http.StatusOK || !strings.Contains(health.Body.String(), `"status":"ok"`) {
		t.Fatalf("health response unexpected: code=%d body=%q", health.Code, health.Body.String())
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
