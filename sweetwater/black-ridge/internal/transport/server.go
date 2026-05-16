package transport

import (
	"context"
	"io/fs"
	"strings"

	"google.golang.org/grpc"

	pb "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
	"github.com/charviki/maze/fabrication/cradle/gatewayutil"
	"github.com/charviki/maze/fabrication/cradle/grpcutil"
	"github.com/charviki/maze/fabrication/cradle/lifecycle"
	"github.com/charviki/maze/fabrication/cradle/logutil"

	intconfig "github.com/charviki/sweetwater-black-ridge/internal/config"
	"github.com/charviki/sweetwater-black-ridge/internal/service"
)

const defaultGRPCListenAddr = ":9090"

// ServerParams 包含创建 black-ridge 全套服务所需的参数。
type ServerParams struct {
	Config            *intconfig.Config
	Logger            logutil.Logger
	TmuxService       service.TmuxService
	LocalConfig       *service.LocalConfigStore
	ConfigFileService service.ConfigFileProvider
	WorkspaceRootDir  string
	StaticFiles       fs.FS
	ExtraInterceptors []grpc.UnaryServerInterceptor
}

// GRPCListenAddrFor 从配置的 gRPC 地址中提取监听端口。
func GRPCListenAddrFor(grpcAddr string) string {
	if grpcAddr == "" {
		return defaultGRPCListenAddr
	}
	if idx := strings.LastIndex(grpcAddr, ":"); idx >= 0 {
		return ":" + grpcAddr[idx+1:]
	}
	return grpcAddr
}

// NewGRPCGatewayServer 创建 gRPC + gateway + HTTP 全套服务。
// 返回 (httpServer, grpcServer, templateStore, error)。
func NewGRPCGatewayServer(params ServerParams) (lifecycle.Server, lifecycle.Server, *service.TemplateStore, error) {
	gwMux := gatewayutil.NewServeMux()
	httpServer, templateStore := NewHTTPServer(HTTPHandlerParams{
		Config:         params.Config,
		TmuxService:    params.TmuxService,
		Logger:         params.Logger,
		GWMux:          gwMux,
		StaticFiles:    params.StaticFiles,
		JWTSecret:      params.Config.Server.JWTSecret,
		AllowedOrigins: params.Config.Server.AllowedOrigins,
	})

	grpcAddr := GRPCListenAddrFor(params.Config.Server.GRPCAddr)

	grpcServer := NewServer(params.TmuxService, params.LocalConfig, templateStore, params.ConfigFileService, params.WorkspaceRootDir, params.Logger)

	validationInterceptor, err := gatewayutil.NewValidationInterceptor()
	if err != nil {
		return nil, nil, nil, err
	}

	interceptors := make([]grpc.UnaryServerInterceptor, 0, 2+len(params.ExtraInterceptors))
	interceptors = append(interceptors,
		validationInterceptor,
		gatewayutil.UnaryAuthInterceptor(params.Config.Server.JWTSecret),
	)
	interceptors = append(interceptors, params.ExtraInterceptors...)

	grpcCore := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
	grpcServer.RegisterGRPC(grpcCore)
	managedGRPC := grpcutil.NewManagedGRPCServer(grpcAddr, grpcCore, params.Logger)

	ctx := context.Background()
	if err := pb.RegisterSessionServiceHandlerServer(ctx, gwMux, grpcServer); err != nil {
		return nil, nil, nil, err
	}
	if err := pb.RegisterTemplateServiceHandlerServer(ctx, gwMux, grpcServer); err != nil {
		return nil, nil, nil, err
	}
	if err := pb.RegisterConfigServiceHandlerServer(ctx, gwMux, grpcServer); err != nil {
		return nil, nil, nil, err
	}

	return httpServer, managedGRPC, templateStore, nil
}
