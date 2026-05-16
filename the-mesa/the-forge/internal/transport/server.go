package transport

import (
	"context"

	"google.golang.org/grpc"

	"github.com/charviki/maze/fabrication/cradle/gatewayutil"
	"github.com/charviki/maze/fabrication/cradle/grpcutil"
	"github.com/charviki/maze/fabrication/cradle/lifecycle"
	"github.com/charviki/maze/fabrication/cradle/logutil"

	intconfig "github.com/charviki/maze/the-mesa/the-forge/internal/config"
	"github.com/charviki/maze/the-mesa/the-forge/internal/service"
)

// ServerParams 包含创建 the-forge 全套服务所需的参数。
type ServerParams struct {
	Config            *intconfig.Config
	Logger            logutil.Logger
	DocSvc            *service.DocService
	ExtraInterceptors []grpc.UnaryServerInterceptor
}

// NewGRPCGatewayServer 创建 gRPC + gateway + HTTP 全套服务。
// 返回 (httpServer, grpcServer)，两者都实现了 lifecycle.Server。
func NewGRPCGatewayServer(params ServerParams) (lifecycle.Server, lifecycle.Server, error) {
	validationInterceptor, err := gatewayutil.NewValidationInterceptor()
	if err != nil {
		return nil, nil, err
	}

	interceptors := make([]grpc.UnaryServerInterceptor, 0, 2+len(params.ExtraInterceptors))
	interceptors = append(interceptors,
		validationInterceptor,
		gatewayutil.UnaryAuthInterceptor(params.Config.Server.JWTSecret),
	)
	interceptors = append(interceptors, params.ExtraInterceptors...)

	grpcCore := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
	docTransport := NewServer(params.DocSvc)
	docTransport.RegisterGRPC(grpcCore)
	managedGRPC := grpcutil.NewManagedGRPCServer(params.Config.Server.GRPCAddr, grpcCore, params.Logger)

	gwMux := gatewayutil.NewServeMux()
	if err := RegisterGatewayHandlers(context.Background(), GatewayRegistrationParams{
		GWMux:    gwMux,
		GRPCAddr: params.Config.Server.GRPCAddr,
	}); err != nil {
		return nil, nil, err
	}

	httpServer := NewHTTPServer(HTTPHandlerParams{
		Config: params.Config,
		GWMux:  gwMux,
		Logger: params.Logger,
	})

	return httpServer, managedGRPC, nil
}
