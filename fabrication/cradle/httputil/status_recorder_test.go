package httputil

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatusRecorderDefaultsToOK(t *testing.T) {
	rec := NewStatusRecorder(httptest.NewRecorder())

	if rec.Status() != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Status(), http.StatusOK)
	}
}

func TestStatusRecorderWriteHeaderTracksStatus(t *testing.T) {
	base := httptest.NewRecorder()
	rec := NewStatusRecorder(base)

	rec.WriteHeader(http.StatusCreated)

	if rec.Status() != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Status(), http.StatusCreated)
	}
	if base.Code != http.StatusCreated {
		t.Fatalf("base code = %d, want %d", base.Code, http.StatusCreated)
	}
}

func TestStatusRecorderHijackPassThrough(t *testing.T) {
	base := &hijackableRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		t:                t,
	}
	rec := NewStatusRecorder(base)

	conn, rw, err := rec.Hijack()
	if err != nil {
		t.Fatalf("Hijack() error = %v", err)
	}
	if conn == nil || rw == nil {
		t.Fatal("Hijack() returned nil conn or readwriter")
	}
}

func TestStatusRecorderHijackUnsupported(t *testing.T) {
	rec := NewStatusRecorder(httptest.NewRecorder())

	if _, _, err := rec.Hijack(); err == nil {
		t.Fatal("Hijack() error = nil, want unsupported error")
	}
}

type hijackableRecorder struct {
	*httptest.ResponseRecorder
	t *testing.T
}

func (r *hijackableRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	serverConn, clientConn := net.Pipe()
	r.t.Cleanup(func() { _ = clientConn.Close() })
	reader := bufio.NewReader(serverConn)
	writer := bufio.NewWriter(serverConn)
	return serverConn, bufio.NewReadWriter(reader, writer), nil
}
