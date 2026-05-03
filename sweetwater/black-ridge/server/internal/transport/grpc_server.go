package transport

import (
	"encoding/json"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/sweetwater-black-ridge/internal/model"
	"github.com/charviki/sweetwater-black-ridge/internal/service"
)

// Server Agent 端 gRPC 服务器
type Server struct {
	pb.UnimplementedSessionServiceServer
	pb.UnimplementedTemplateServiceServer
	pb.UnimplementedConfigServiceServer

	tmuxService      service.TmuxService
	localConfig      *service.LocalConfigStore
	templateStore    *model.TemplateStore
	workspaceRootDir string
	logger           logutil.Logger

	grpcServer *grpc.Server
}

// NewServer 创建 Agent gRPC Server
func NewServer(
	tmuxService service.TmuxService,
	localConfig *service.LocalConfigStore,
	templateStore *model.TemplateStore,
	workspaceRootDir string,
	logger logutil.Logger,
) *Server {
	return &Server{
		tmuxService:      tmuxService,
		localConfig:      localConfig,
		templateStore:    templateStore,
		workspaceRootDir: workspaceRootDir,
		logger:           logger,
	}
}

// RegisterGRPC 将服务实现注册到统一管理的 gRPC server。
func (s *Server) RegisterGRPC(grpcServer *grpc.Server) {
	s.grpcServer = grpcServer
	pb.RegisterSessionServiceServer(grpcServer, s)
	pb.RegisterTemplateServiceServer(grpcServer, s)
	pb.RegisterConfigServiceServer(grpcServer, s)
}

// errToStatus 将业务错误映射为 gRPC status code
func errToStatus(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, service.ErrSessionNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}
	// ConfigConflictError 携带冲突详情，映射为 FailedPrecondition
	var confErr *service.ConfigConflictError
	if errors.As(err, &confErr) {
		detail, _ := json.Marshal(confErr.Conflicts)
		return status.Error(codes.FailedPrecondition, string(detail))
	}
	return status.Error(codes.Internal, err.Error())
}

// 确保 Server 实现了所有 gRPC 接口
var (
	_ pb.SessionServiceServer  = (*Server)(nil)
	_ pb.TemplateServiceServer = (*Server)(nil)
	_ pb.ConfigServiceServer   = (*Server)(nil)
)
