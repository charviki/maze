package transport

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestAttachBearerToken(t *testing.T) {
	ctx := attachBearerToken(context.Background(), "secret-token")

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}

	got := md.Get("authorization")
	if len(got) != 1 || got[0] != "Bearer secret-token" {
		t.Fatalf("authorization metadata = %v, want [Bearer secret-token]", got)
	}
}

func TestAttachBearerToken_EmptyToken(t *testing.T) {
	base := context.Background()
	ctx := attachBearerToken(base, "")

	if ctx != base {
		t.Fatal("expected empty token to keep original context")
	}
}
