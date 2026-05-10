package transport

import (
	"context"
	"strings"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
)

// localGRPCEndpoint 将 gRPC 监听地址转换为本地 dial 地址。
func localGRPCEndpoint(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "127.0.0.1" + addr
	}
	return addr
}

// RegisterGatewayHandlers 将 KnowledgeService 和 DirectiveService 注册到 grpc-gateway mux。
func RegisterGatewayHandlers(ctx context.Context, gwMux *gwruntime.ServeMux, grpcAddr string) error {
	endpoint := localGRPCEndpoint(grpcAddr)
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	registrations := []struct {
		name string
		fn   func(ctx context.Context, mux *gwruntime.ServeMux, endpoint string, opts []grpc.DialOption) error
	}{
		{"KnowledgeService", pb.RegisterKnowledgeServiceHandlerFromEndpoint},
		{"DirectiveService", pb.RegisterDirectiveServiceHandlerFromEndpoint},
	}

	for _, reg := range registrations {
		if err := reg.fn(ctx, gwMux, endpoint, dialOpts); err != nil {
			return err
		}
	}

	return nil
}
