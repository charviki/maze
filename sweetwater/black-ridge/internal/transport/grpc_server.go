package transport

import (
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
	"github.com/charviki/maze/fabrication/cradle/errutil"
	"github.com/charviki/maze/fabrication/cradle/logutil"

	"github.com/charviki/sweetwater-black-ridge/internal/service"
)

// Server Agent 端 gRPC 服务器
type Server struct {
	pb.UnimplementedSessionServiceServer
	pb.UnimplementedTemplateServiceServer
	pb.UnimplementedConfigServiceServer

	tmuxService      service.TmuxService
	localConfig      *service.LocalConfigStore
	templateStore    *service.TemplateStore
	configFiles      service.ConfigFileProvider
	workspaceRootDir string
	logger           logutil.Logger

	grpcServer *grpc.Server
}

// NewServer 创建 Agent gRPC Server
func NewServer(
	tmuxService service.TmuxService,
	localConfig *service.LocalConfigStore,
	templateStore *service.TemplateStore,
	configFiles service.ConfigFileProvider,
	workspaceRootDir string,
	logger logutil.Logger,
) *Server {
	return &Server{
		tmuxService:      tmuxService,
		localConfig:      localConfig,
		templateStore:    templateStore,
		configFiles:      configFiles,
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
		return errutil.NewError(codes.NotFound, pb.ErrorReason_ERROR_REASON_SESSION_NOT_FOUND, err.Error())
	}
	var confErr *service.ConfigConflictError
	if errors.As(err, &confErr) {
		violations := make([]errutil.PreconditionViolation, len(confErr.Conflicts))
		for i, c := range confErr.Conflicts {
			violations[i] = errutil.PreconditionViolation{
				Type:        "CONFIG_CONFLICT",
				Subject:     c.Path,
				Description: c.CurrentHash,
			}
		}
		return errutil.NewPreconditionError(
			codes.FailedPrecondition,
			pb.ErrorReason_ERROR_REASON_CONFIG_CONFLICT,
			"config conflict",
			violations,
		)
	}
	return errutil.NewError(codes.Internal, pb.ErrorReason_ERROR_REASON_UNSPECIFIED, err.Error())
}

// 确保 Server 实现了所有 gRPC 接口
var (
	_ pb.SessionServiceServer  = (*Server)(nil)
	_ pb.TemplateServiceServer = (*Server)(nil)
	_ pb.ConfigServiceServer   = (*Server)(nil)
)
