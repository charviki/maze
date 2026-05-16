package transport

import (
	"context"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
	"github.com/charviki/maze/fabrication/cradle/gatewayutil"
)

// GatewayRegistrationParams 包含 gateway 注册所需的参数。
type GatewayRegistrationParams struct {
	GWMux       *gwruntime.ServeMux
	GRPCAddr    string
	GRPCServer  *grpc.Server
	PermHandler *PermissionServiceServer
	AuthHandler *AuthHandler
}

// RegisterGatewayHandlers 将所有 gRPC 服务注册到 grpc-gateway mux。
// cmd 不再需要知道具体注册了哪些服务。
func RegisterGatewayHandlers(ctx context.Context, params GatewayRegistrationParams) error {
	endpoint := gatewayutil.LocalGRPCEndpoint(params.GRPCAddr)
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	registrations := []struct {
		name string
		fn   func(ctx context.Context, mux *gwruntime.ServeMux, endpoint string, opts []grpc.DialOption) error
	}{
		{"HostService", pb.RegisterHostServiceHandlerFromEndpoint},
		{"NodeService", pb.RegisterNodeServiceHandlerFromEndpoint},
		{"AuditService", pb.RegisterAuditServiceHandlerFromEndpoint},
		{"SessionService", pb.RegisterSessionServiceHandlerFromEndpoint},
		{"TemplateService", pb.RegisterTemplateServiceHandlerFromEndpoint},
		{"ConfigService", pb.RegisterConfigServiceHandlerFromEndpoint},
		{"AgentService", pb.RegisterAgentServiceHandlerFromEndpoint},
		{"SkillService", pb.RegisterSkillServiceHandlerFromEndpoint},
		{"MCPService", pb.RegisterMCPServiceHandlerFromEndpoint},
		{"RuleService", pb.RegisterRuleServiceHandlerFromEndpoint},
		{"GitKeyService", pb.RegisterGitKeyServiceHandlerFromEndpoint},
	}

	for _, reg := range registrations {
		if err := reg.fn(ctx, params.GWMux, endpoint, dialOpts); err != nil {
			return err
		}
	}

	if params.PermHandler != nil {
		if err := pb.RegisterPermissionServiceHandlerFromEndpoint(ctx, params.GWMux, endpoint, dialOpts); err != nil {
			return err
		}
	}

	if params.AuthHandler != nil {
		if err := pb.RegisterAuthServiceHandlerFromEndpoint(ctx, params.GWMux, endpoint, dialOpts); err != nil {
			return err
		}
	}

	return nil
}
