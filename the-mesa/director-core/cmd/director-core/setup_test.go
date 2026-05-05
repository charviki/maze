package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gorillaws "github.com/gorilla/websocket"

	"github.com/charviki/maze-cradle/httputil"
)

func TestAccessLogMiddlewarePreservesWebSocketUpgrade(t *testing.T) {
	handler := chainHTTP(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := httputil.NewUpgrader(nil).Upgrade(w, r, nil)
			if err != nil {
				t.Fatalf("upgrade failed: %v", err)
			}
			_ = conn.Close()
		}),
		accessLogMiddleware(nil),
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
