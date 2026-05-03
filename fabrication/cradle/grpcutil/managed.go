package grpcutil

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/charviki/maze-cradle/logutil"
	"google.golang.org/grpc"
)

// ManagedGRPCServer 将 grpc.Server 适配为 lifecycle.Server，
// 让业务模块可以把 gRPC 与 HTTP 一起交给统一的 lifecycle.Manager 管理。
type ManagedGRPCServer struct {
	addr   string
	server *grpc.Server
	logger logutil.Logger

	mu       sync.Mutex
	listener net.Listener
}

// NewManagedGRPCServer 创建一个可被统一生命周期管理的 gRPC 服务器包装器。
func NewManagedGRPCServer(addr string, server *grpc.Server, logger logutil.Logger) *ManagedGRPCServer {
	return &ManagedGRPCServer{
		addr:   addr,
		server: server,
		logger: logger,
	}
}

// ListenAndServe 在目标地址上监听并阻塞服务。
func (s *ManagedGRPCServer) ListenAndServe() error {
	if s.server == nil {
		return errors.New("grpc server is nil")
	}

	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("grpc listen %s: %w", s.addr, err)
	}

	s.mu.Lock()
	s.listener = lis
	s.mu.Unlock()

	if s.logger != nil {
		s.logger.Infof("[grpc] server started on %s", s.addr)
	}

	if err := s.server.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return err
	}

	return nil
}

// Shutdown 优先尝试 GracefulStop，超时后回退到 Stop，避免进程无限卡住。
func (s *ManagedGRPCServer) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.server.Stop()
		<-done
		return nil
	}
}
