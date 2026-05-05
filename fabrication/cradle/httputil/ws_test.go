package httputil

import (
	"net"
	"net/http"
	"testing"

	gorillaws "github.com/gorilla/websocket"
)

func TestNewUpgrader_NilOrigins(t *testing.T) {
	upgrader := NewUpgrader(nil)
	if upgrader == nil {
		t.Fatal("NewUpgrader returned nil")
	}
	if upgrader.CheckOrigin == nil {
		t.Error("CheckOrigin should not be nil")
	}
	if !upgrader.CheckOrigin(&http.Request{}) {
		t.Error("nil origins should allow all")
	}
}

func TestNewUpgrader_EmptyOrigins(t *testing.T) {
	upgrader := NewUpgrader([]string{})
	if !upgrader.CheckOrigin(&http.Request{}) {
		t.Error("empty origins should allow all")
	}
}

func TestNewUpgrader_RestrictedOrigins(t *testing.T) {
	upgrader := NewUpgrader([]string{"http://myapp.com"})

	ok := upgrader.CheckOrigin(&http.Request{
		Header: http.Header{"Origin": []string{"http://myapp.com"}},
	})
	if !ok {
		t.Error("allowed origin should pass")
	}

	notOk := upgrader.CheckOrigin(&http.Request{
		Header: http.Header{"Origin": []string{"http://evil.com"}},
	})
	if notOk {
		t.Error("disallowed origin should fail")
	}
}

func TestNewUpgrader_NoOriginHeader(t *testing.T) {
	upgrader := NewUpgrader([]string{"http://myapp.com"})
	if !upgrader.CheckOrigin(&http.Request{}) {
		t.Error("no origin header should be allowed (browser direct access)")
	}
}

func TestCheckOrigin_CaseInsensitive(t *testing.T) {
	fn := CheckOrigin([]string{"HTTP://myapp.com"})

	if !fn(&http.Request{Header: http.Header{"Origin": []string{"http://myapp.com"}}}) {
		t.Error("origin check should be case-insensitive")
	}
	if !fn(&http.Request{Header: http.Header{"Origin": []string{"http://MYAPP.COM"}}}) {
		t.Error("origin check should be case-insensitive")
	}
}

func TestCheckOrigin_AllowsWhenEmptyList(t *testing.T) {
	fn := CheckOrigin(nil)
	if !fn(&http.Request{Header: http.Header{"Origin": []string{"http://anything.com"}}}) {
		t.Error("empty list should allow all")
	}
}

func TestCheckOrigin_AllowsWhenNoOrigin(t *testing.T) {
	fn := CheckOrigin([]string{"http://myapp.com"})
	if !fn(&http.Request{}) {
		t.Error("no Origin header should be allowed")
	}
}

func TestCheckOrigin_RejectsUnknown(t *testing.T) {
	fn := CheckOrigin([]string{"http://myapp.com"})
	if fn(&http.Request{Header: http.Header{"Origin": []string{"http://evil.com"}}}) {
		t.Error("unknown origin should be rejected")
	}
}

func TestRelayWebSocket_BidirectionalClose(t *testing.T) {
	clientConn, serverConn := gorillawsPair(t)

	errCh := make(chan error, 1)
	go func() {
		errCh <- RelayWebSocket(clientConn, serverConn)
	}()

	_ = clientConn.Close()

	err := <-errCh
	if err == nil {
		t.Error("expected relay error from close")
	}
}

func gorillawsPair(t *testing.T) (*gorillaws.Conn, *gorillaws.Conn) {
	t.Helper()
	srv := &http.Server{Addr: "127.0.0.1:0"}
	var upgrader = NewUpgrader(nil)

	connCh := make(chan *gorillaws.Conn, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade failed: %v", err)
			return
		}
		connCh <- conn
	})
	srv.Handler = mux

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := l.Addr().String()
	go func() { _ = srv.Serve(l) }()
	t.Cleanup(func() { srv.Close() })

	client, _, err := gorillaws.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { client.Close() })

	server := <-connCh
	t.Cleanup(func() { server.Close() })

	return client, server
}
