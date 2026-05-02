package grpc

import (
	"errors"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/sweetwater-black-ridge/biz/model"
	"github.com/charviki/sweetwater-black-ridge/biz/service"
)

// Server Agent 端 gRPC 服务器
type Server struct {
	pb.UnimplementedSessionServiceServer
	pb.UnimplementedTemplateServiceServer
	pb.UnimplementedConfigServiceServer

	tmuxService   service.TmuxService
	localConfig   *service.LocalConfigStore
	templateStore *model.TemplateStore
	logger        logutil.Logger

	grpcServer *grpc.Server
}

// NewServer 创建 Agent gRPC Server
func NewServer(
	tmuxService service.TmuxService,
	localConfig *service.LocalConfigStore,
	templateStore *model.TemplateStore,
	logger logutil.Logger,
) *Server {
	return &Server{
		tmuxService:   tmuxService,
		localConfig:   localConfig,
		templateStore: templateStore,
		logger:        logger,
	}
}

// Start 启动 gRPC server（非阻塞）
func (s *Server) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("grpc listen %s: %w", addr, err)
	}

	s.grpcServer = grpc.NewServer()
	pb.RegisterSessionServiceServer(s.grpcServer, s)
	pb.RegisterTemplateServiceServer(s.grpcServer, s)
	pb.RegisterConfigServiceServer(s.grpcServer, s)

	go func() {
		s.logger.Infof("[grpc] server started on %s", addr)
		if err := s.grpcServer.Serve(lis); err != nil {
			s.logger.Errorf("[grpc] server error: %v", err)
		}
	}()
	return nil
}

// Stop 优雅关闭 gRPC server
func (s *Server) Stop() {
	if s.grpcServer != nil {
		stopped := make(chan struct{})
		go func() {
			s.grpcServer.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
		case <-time.After(5 * time.Second):
			s.grpcServer.Stop()
		}
	}
}

// errToStatus 将业务错误映射为 gRPC status code
func errToStatus(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, service.ErrSessionNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}

// 确保 Server 实现了所有 gRPC 接口
var (
	_ pb.SessionServiceServer  = (*Server)(nil)
	_ pb.TemplateServiceServer = (*Server)(nil)
	_ pb.ConfigServiceServer   = (*Server)(nil)
)
