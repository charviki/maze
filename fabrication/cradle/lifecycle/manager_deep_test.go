package lifecycle

import (
	"context"
	"testing"
	"time"

	"github.com/charviki/maze-cradle/logutil"
)

func TestManager_AddNil(t *testing.T) {
	mgr := &Manager{ShutdownTimeout: time.Second}
	mgr.Add(nil)
	if len(mgr.servers) != 0 {
		t.Errorf("servers length = %d, want 0", len(mgr.servers))
	}
}

func TestManager_Run_NoServers(t *testing.T) {
	mgr := &Manager{}
	err := mgr.Run(context.Background())
	if err != nil {
		t.Errorf("Run with no servers should return nil, got %v", err)
	}
}

func TestManager_Run_DefaultTimeoutUsed(t *testing.T) {
	srv := newFakeServer()
	srv.serveCh <- nil
	srv.shutdownFn = func(ctx context.Context) error {
		select {
		case srv.serveCh <- nil:
		default:
		}
		return nil
	}

	mgr := &Manager{Logger: logutil.NewNop()}
	mgr.Add(srv)

	err := mgr.Run(context.Background())
	if err != nil {
		t.Errorf("Run error: %v", err)
	}
}

func TestManager_AddMultiple(t *testing.T) {
	mgr := &Manager{ShutdownTimeout: time.Second}
	for i := 0; i < 10; i++ {
		mgr.Add(newFakeServer())
	}
	if len(mgr.servers) != 10 {
		t.Errorf("servers length = %d, want 10", len(mgr.servers))
	}
}

func TestManager_SignalNotifyNilLogger(t *testing.T) {
	mgr := &Manager{
		ShutdownTimeout: time.Second,
	}
	mgr.Add(newFakeServer())
	fn := mgr.newSignalNotify()
	if fn == nil {
		t.Error("newSignalNotify should return a valid function")
	}
}

func TestNormalizeServeError_Nil(t *testing.T) {
	if err := normalizeServeError(nil); err != nil {
		t.Errorf("nil error should stay nil, got %v", err)
	}
}

func TestNormalizeServeError_ServerClosed(t *testing.T) {
	err := normalizeServeError(httpErrServerClosed())
	if err != nil {
		t.Errorf("http.ErrServerClosed should be normalized to nil, got %v", err)
	}
}
