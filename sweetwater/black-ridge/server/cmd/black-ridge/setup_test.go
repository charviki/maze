package main

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/charviki/maze-cradle/logutil"
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
