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
	GWMux    *gwruntime.ServeMux
	GRPCAddr string
}

// RegisterGatewayHandlers 将 KnowledgeService 注册到 grpc-gateway mux。
func RegisterGatewayHandlers(ctx context.Context, params GatewayRegistrationParams) error {
	endpoint := gatewayutil.LocalGRPCEndpoint(params.GRPCAddr)
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	registrations := []struct {
		name string
		fn   func(ctx context.Context, mux *gwruntime.ServeMux, endpoint string, opts []grpc.DialOption) error
	}{
		{"KnowledgeService", pb.RegisterKnowledgeServiceHandlerFromEndpoint},
	}

	for _, reg := range registrations {
		if err := reg.fn(ctx, params.GWMux, endpoint, dialOpts); err != nil {
			return err
		}
	}

	return nil
}
