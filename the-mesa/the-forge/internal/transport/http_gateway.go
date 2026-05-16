package transport

import (
	"context"
	"strings"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
)

// GatewayRegistrationParams 包含 gateway 注册所需的参数。
type GatewayRegistrationParams struct {
	GWMux    *gwruntime.ServeMux
	GRPCAddr string
}

// localGRPCEndpoint 将 gRPC 监听地址转换为本地 dial 地址。
func localGRPCEndpoint(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "127.0.0.1" + addr
	}
	return addr
}

// RegisterGatewayHandlers 将 KnowledgeService 注册到 grpc-gateway mux。
func RegisterGatewayHandlers(ctx context.Context, params GatewayRegistrationParams) error {
	endpoint := localGRPCEndpoint(params.GRPCAddr)
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
