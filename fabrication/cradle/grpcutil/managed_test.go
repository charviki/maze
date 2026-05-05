package grpcutil

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/charviki/maze-cradle/logutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func waitForListener(t *testing.T, server *ManagedGRPCServer) string {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		server.mu.Lock()
		listener := server.listener
		server.mu.Unlock()
		if listener != nil {
			return listener.Addr().String()
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("listener 未在预期时间内就绪")
	return ""
}

func TestNewManagedGRPCServer(t *testing.T) {
	srv := grpc.NewServer()
	defer srv.Stop()

	m := NewManagedGRPCServer(":0", srv, logutil.NewNop())
	if m == nil {
		t.Fatal("NewManagedGRPCServer returned nil")
	}
	if m.server != srv {
		t.Error("server field not set correctly")
	}
	if m.addr != ":0" {
		t.Errorf("addr = %q, want %q", m.addr, ":0")
	}
}

func TestManagedGRPCServer_ListenAndServe_NilServer(t *testing.T) {
	m := NewManagedGRPCServer(":0", nil, logutil.NewNop())
	err := m.ListenAndServe()
	if err == nil {
		t.Error("expected error for nil server")
	}
}

func TestManagedGRPCServer_ListenAndServe_InvalidAddr(t *testing.T) {
	srv := grpc.NewServer()
	defer srv.Stop()

	m := NewManagedGRPCServer("invalid-addr", srv, logutil.NewNop())
	err := m.ListenAndServe()
	if err == nil {
		t.Error("expected error for invalid address")
	}
}

func TestManagedGRPCServer_ListenAndServe_Success(t *testing.T) {
	srv := grpc.NewServer()
	defer srv.GracefulStop()

	m := NewManagedGRPCServer("127.0.0.1:0", srv, logutil.NewNop())
	errCh := make(chan error, 1)
	go func() {
		errCh <- m.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		t.Fatalf("unexpected early return: %v", err)
	case <-time.After(100 * time.Millisecond):
	}

	addr := waitForListener(t, m)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	conn.Close()
}

func TestManagedGRPCServer_Shutdown_NilServer(t *testing.T) {
	m := NewManagedGRPCServer(":0", nil, logutil.NewNop())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := m.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown with nil server should not error: %v", err)
	}
}

func TestManagedGRPCServer_Shutdown_GracefulStop(t *testing.T) {
	srv := grpc.NewServer()
	m := NewManagedGRPCServer("127.0.0.1:0", srv, logutil.NewNop())

	go func() {
		_ = m.ListenAndServe()
	}()

	waitForListener(t, m)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := m.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestManagedGRPCServer_Shutdown_Timeout(t *testing.T) {
	srv := grpc.NewServer()
	m := NewManagedGRPCServer("127.0.0.1:0", srv, logutil.NewNop())

	go func() {
		_ = m.ListenAndServe()
	}()

	waitForListener(t, m)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := m.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown with timeout should not error: %v", err)
	}
}

func TestManagedGRPCServer_ListenerSet(t *testing.T) {
	srv := grpc.NewServer()
	defer srv.GracefulStop()

	m := NewManagedGRPCServer("127.0.0.1:0", srv, logutil.NewNop())
	go func() { _ = m.ListenAndServe() }()

	addr := waitForListener(t, m)
	listener, err := net.Listen("tcp", addr)
	if err == nil {
		_ = listener.Close()
		t.Fatalf("监听地址 %s 仍可被重复占用，说明 gRPC listener 未启动", addr)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.listener.(*net.TCPListener)
	if !ok {
		t.Errorf("expected *net.TCPListener, got %T", m.listener)
	}
}
