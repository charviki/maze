package lifecycle

import (
	"context"
	"errors"
	"net/http"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/charviki/maze-cradle/logutil"
)

type fakeServer struct {
	started chan struct{}
	serveCh chan error

	mu            sync.Mutex
	shutdownCalls int
	shutdownFn    func(context.Context) error
}

func newFakeServer() *fakeServer {
	return &fakeServer{
		started: make(chan struct{}),
		serveCh: make(chan error, 1),
	}
}

func (s *fakeServer) ListenAndServe() error {
	close(s.started)
	return <-s.serveCh
}

func (s *fakeServer) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	s.shutdownCalls++
	shutdownFn := s.shutdownFn
	s.mu.Unlock()

	if shutdownFn != nil {
		return shutdownFn(ctx)
	}

	select {
	case s.serveCh <- nil:
	default:
	}
	return nil
}

func (s *fakeServer) ShutdownCalls() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.shutdownCalls
}

func TestManagerRun_ShutsDownOnSignal(t *testing.T) {
	t.Parallel()

	srvA := newFakeServer()
	srvB := newFakeServer()
	sigCh := make(chan os.Signal, 1)

	mgr := &Manager{
		ShutdownTimeout: time.Second,
		Logger:          logutil.NewNop(),
		signalNotify:    singleSignalNotify(sigCh),
	}
	mgr.Add(srvA)
	mgr.Add(srvB)

	errCh := make(chan error, 1)
	go func() {
		errCh <- mgr.Run(context.Background())
	}()

	waitStarted(t, srvA.started)
	waitStarted(t, srvB.started)
	sigCh <- syscall.SIGTERM

	if err := waitRun(t, errCh); err != nil {
		t.Fatalf("Run() error = %v, 期望 nil", err)
	}
	if srvA.ShutdownCalls() != 1 {
		t.Fatalf("srvA Shutdown 调用次数 = %d, 期望 1", srvA.ShutdownCalls())
	}
	if srvB.ShutdownCalls() != 1 {
		t.Fatalf("srvB Shutdown 调用次数 = %d, 期望 1", srvB.ShutdownCalls())
	}
}

func TestManagerRun_PropagatesServerError(t *testing.T) {
	t.Parallel()

	srvA := newFakeServer()
	srvB := newFakeServer()
	expectedErr := errors.New("listen failed")

	srvA.serveCh <- expectedErr
	srvB.shutdownFn = func(ctx context.Context) error {
		select {
		case srvB.serveCh <- nil:
		default:
		}
		return nil
	}

	mgr := &Manager{
		ShutdownTimeout: time.Second,
		Logger:          logutil.NewNop(),
		signalNotify:    singleSignalNotify(make(chan os.Signal)),
	}
	mgr.Add(srvA)
	mgr.Add(srvB)

	err := mgr.Run(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("Run() error = %v, 期望包含 %v", err, expectedErr)
	}
	if srvB.ShutdownCalls() != 1 {
		t.Fatalf("srvB Shutdown 调用次数 = %d, 期望 1", srvB.ShutdownCalls())
	}
}

func TestManagerRun_ForcesShutdownOnTimeout(t *testing.T) {
	t.Parallel()

	srv := newFakeServer()
	sigCh := make(chan os.Signal, 1)
	forced := make(chan struct{}, 1)

	srv.shutdownFn = func(ctx context.Context) error {
		<-ctx.Done()
		forced <- struct{}{}
		select {
		case srv.serveCh <- nil:
		default:
		}
		return nil
	}

	mgr := &Manager{
		ShutdownTimeout: 20 * time.Millisecond,
		Logger:          logutil.NewNop(),
		signalNotify:    singleSignalNotify(sigCh),
	}
	mgr.Add(srv)

	errCh := make(chan error, 1)
	go func() {
		errCh <- mgr.Run(context.Background())
	}()

	waitStarted(t, srv.started)
	sigCh <- syscall.SIGTERM

	if err := waitRun(t, errCh); err != nil {
		t.Fatalf("Run() error = %v, 期望 nil", err)
	}
	select {
	case <-forced:
	case <-time.After(time.Second):
		t.Fatal("期望 Shutdown 收到超时 context，但未触发")
	}
}

func TestManagerRun_NormalizesHTTPServerClosed(t *testing.T) {
	t.Parallel()

	srvA := newFakeServer()
	srvB := newFakeServer()

	srvA.serveCh <- httpErrServerClosed()
	srvB.shutdownFn = func(ctx context.Context) error {
		select {
		case srvB.serveCh <- nil:
		default:
		}
		return nil
	}

	mgr := &Manager{
		ShutdownTimeout: time.Second,
		Logger:          logutil.NewNop(),
		signalNotify:    singleSignalNotify(make(chan os.Signal)),
	}
	mgr.Add(srvA)
	mgr.Add(srvB)

	if err := mgr.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v, 期望 nil", err)
	}
}

func singleSignalNotify(sigCh <-chan os.Signal) func(chan<- os.Signal, ...os.Signal) func() {
	return func(dst chan<- os.Signal, _ ...os.Signal) func() {
		stopCh := make(chan struct{})
		go func() {
			select {
			case sig, ok := <-sigCh:
				if !ok {
					return
				}
				dst <- sig
			case <-stopCh:
				return
			}
		}()
		return func() {
			close(stopCh)
		}
	}
}

func waitStarted(t *testing.T, started <-chan struct{}) {
	t.Helper()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("等待 server 启动超时")
	}
}

func waitRun(t *testing.T, errCh <-chan error) error {
	t.Helper()

	select {
	case err := <-errCh:
		return err
	case <-time.After(2 * time.Second):
		t.Fatal("等待 Run 返回超时")
		return nil
	}
}

func httpErrServerClosed() error {
	return http.ErrServerClosed
}
