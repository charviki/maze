package lifecycle

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/charviki/maze-cradle/logutil"
	"golang.org/x/sync/errgroup"
)

const defaultShutdownTimeout = 5 * time.Second

// Server 统一抽象 HTTP / gRPC 等需要被统一管理生命周期的服务器。
// ListenAndServe 负责阻塞运行，Shutdown 负责在给定 deadline 内优雅关闭。
type Server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// Manager 负责统一管理多个服务器的并发启动、信号监听与优雅关闭。
// 任一服务器退出时都会触发整组服务收敛，避免单边存活造成半瘫痪状态。
type Manager struct {
	servers         []Server
	ShutdownTimeout time.Duration
	Logger          logutil.Logger

	// signalNotify 抽成可替换函数是为了在测试里精确模拟 SIGTERM，
	// 否则直接向测试进程发真实信号会把整个 `go test` 一起打断。
	signalNotify func(chan<- os.Signal, ...os.Signal) func()
}

// Add 注册一个需要由 Manager 管理的服务器。
func (m *Manager) Add(s Server) {
	if s == nil {
		return
	}
	m.servers = append(m.servers, s)
}

// Run 并发启动所有服务器，并在收到退出信号、上游 context 取消或任一服务器退出后，
// 统一触发整组服务优雅关闭。
func (m *Manager) Run(ctx context.Context) error {
	if len(m.servers) == 0 {
		return nil
	}

	timeout := m.ShutdownTimeout
	if timeout <= 0 {
		timeout = defaultShutdownTimeout
	}

	group := &errgroup.Group{}
	runCtx, cancelRun := context.WithCancel(ctx)
	defer cancelRun()

	var (
		shutdownOnce sync.Once
		shutdownErr  error
	)

	shutdownAll := func(reason string) error {
		shutdownOnce.Do(func() {
			cancelRun()
			m.logInfo("lifecycle shutdown triggered: %s", reason)

			shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			for _, srv := range m.servers {
				if err := srv.Shutdown(shutdownCtx); err != nil && shutdownErr == nil {
					// 仅记录第一个关闭错误，避免后续错误覆盖真正的根因。
					shutdownErr = err
				}
			}
		})

		return shutdownErr
	}

	for _, srv := range m.servers {
		group.Go(func() error {
			err := normalizeServeError(srv.ListenAndServe())
			if shutdownErr := shutdownAll("server exited"); shutdownErr != nil && err == nil {
				return shutdownErr
			}
			return err
		})
	}

	group.Go(func() error {
		sigCh := make(chan os.Signal, 1)
		stopNotify := m.newSignalNotify()(sigCh, syscall.SIGINT, syscall.SIGTERM)
		defer stopNotify()

		select {
		case sig := <-sigCh:
			return shutdownAll("signal: " + sig.String())
		case <-ctx.Done():
			return shutdownAll("context canceled")
		case <-runCtx.Done():
			return nil
		}
	})

	return group.Wait()
}

func (m *Manager) newSignalNotify() func(chan<- os.Signal, ...os.Signal) func() {
	if m.signalNotify != nil {
		return m.signalNotify
	}

	return func(ch chan<- os.Signal, sigs ...os.Signal) func() {
		signal.Notify(ch, sigs...)
		return func() {
			signal.Stop(ch)
		}
	}
}

func (m *Manager) logInfo(format string, args ...any) {
	if m.Logger == nil {
		return
	}
	m.Logger.Infof(format, args...)
}

func normalizeServeError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
