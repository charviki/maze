package gatewayutil

import (
	"context"
	"testing"

	mazev1 "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// mockAuditLogger 收集审计日志条目，用于断言
type mockAuditLogger struct {
	entries []AuditEntry
}

func (m *mockAuditLogger) Log(entry AuditEntry) {
	m.entries = append(m.entries, entry)
}

func TestShouldAudit(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   bool
	}{
		// 应审计的三类服务
		{"SessionService 方法", "/maze.v1.SessionService/ListSessions", true},
		{"SessionService CreateSession", "/maze.v1.SessionService/CreateSession", true},
		{"TemplateService 方法", "/maze.v1.TemplateService/ListTemplates", true},
		{"ConfigService 方法", "/maze.v1.ConfigService/GetConfig", true},

		// 不需要审计的服务
		{"AgentService 注册", "/maze.v1.AgentService/Register", false},
		{"AgentService 心跳", "/maze.v1.AgentService/Heartbeat", false},
		{"HostService 方法", "/maze.v1.HostService/CreateHost", false},
		{"NodeService 方法", "/maze.v1.NodeService/ListNode", false},
		{"AuditService 方法", "/maze.v1.AuditService/QueryLogs", false},
		{"空方法名", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldAudit(tt.method)
			if got != tt.want {
				t.Errorf("shouldAudit(%q) = %v, 期望 %v", tt.method, got, tt.want)
			}
		})
	}
}

func TestMethodToAction(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   string
	}{
		{"ListSessions", "/maze.v1.SessionService/ListSessions", "list_sessions"},
		{"CreateSession", "/maze.v1.SessionService/CreateSession", "create_session"},
		{"GetConfig", "/maze.v1.ConfigService/GetConfig", "get_config"},
		{"ListTemplates", "/maze.v1.TemplateService/ListTemplates", "list_templates"},
		// 无斜杠前缀时直接走 strings.ToLower 分支
		{"无斜杠", "SimpleMethod", "simplemethod"},
		// 空方法名
		{"空尾部", "/maze.v1.SessionService/", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := methodToAction(tt.method)
			if got != tt.want {
				t.Errorf("methodToAction(%q) = %q, 期望 %q", tt.method, got, tt.want)
			}
		})
	}
}

func TestExtractNodeName(t *testing.T) {
	tests := []struct {
		name string
		req  interface{}
		want string
	}{
		{
			name: "从 node_name 字段提取",
			// Session/Template/Config 请求都有 node_name 字段
			req:  &mazev1.ListSessionsRequest{NodeName: "node-1"},
			want: "node-1",
		},
		{
			name: "node_name 优先于 name",
			// CreateSessionRequest 同时有 node_name 和 name，应优先取 node_name
			req:  &mazev1.CreateSessionRequest{NodeName: "node-2", Name: "session-x"},
			want: "node-2",
		},
		{
			name: "node_name 为空时取 name",
			// 某些请求可能只有 name 字段，node_name 为零值
			req:  &mazev1.RegisterRequest{Name: "agent-1"},
			want: "agent-1",
		},
		{
			name: "非 proto.Message 返回空",
			req:  "not a proto message",
			want: "",
		},
		{
			name: "零值请求",
			req:  &mazev1.ListSessionsRequest{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractNodeName(tt.req)
			if got != tt.want {
				t.Errorf("extractNodeName() = %q, 期望 %q", got, tt.want)
			}
		})
	}
}

func TestUnaryAuditInterceptor_Success(t *testing.T) {
	logger := &mockAuditLogger{}
	interceptor := UnaryAuditInterceptor(logger)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer token"))
	req := &mazev1.ListSessionsRequest{NodeName: "node-1"}
	info := &grpc.UnaryServerInfo{FullMethod: "/maze.v1.SessionService/ListSessions"}

	_, err := interceptor(ctx, req, info, func(_ context.Context, _ interface{}) (interface{}, error) {
		return "ok", nil
	})

	if err != nil {
		t.Fatalf("成功调用不应返回错误: %v", err)
	}
	if len(logger.entries) != 1 {
		t.Fatalf("日志条目数 = %d, 期望 1", len(logger.entries))
	}

	entry := logger.entries[0]
	if entry.Operator != "frontend" {
		t.Errorf("Operator = %q, 期望 %q", entry.Operator, "frontend")
	}
	if entry.TargetNode != "node-1" {
		t.Errorf("TargetNode = %q, 期望 %q", entry.TargetNode, "node-1")
	}
	if entry.Action != "list_sessions" {
		t.Errorf("Action = %q, 期望 %q", entry.Action, "list_sessions")
	}
	if entry.Result != "success" {
		t.Errorf("Result = %q, 期望 %q", entry.Result, "success")
	}
	if entry.StatusCode != 200 {
		t.Errorf("StatusCode = %d, 期望 200", entry.StatusCode)
	}
}

func TestUnaryAuditInterceptor_Error(t *testing.T) {
	logger := &mockAuditLogger{}
	interceptor := UnaryAuditInterceptor(logger)

	ctx := context.Background()
	// 无 metadata → operator 为 "internal"
	req := &mazev1.CreateSessionRequest{NodeName: "node-2"}
	info := &grpc.UnaryServerInfo{FullMethod: "/maze.v1.SessionService/CreateSession"}

	_, err := interceptor(ctx, req, info, func(_ context.Context, _ interface{}) (interface{}, error) {
		return nil, status.Error(codes.NotFound, "session not found")
	})

	if err == nil {
		t.Fatal("期望返回错误，但得到 nil")
	}

	entry := logger.entries[0]
	if entry.Operator != "internal" {
		t.Errorf("Operator = %q, 期望 %q", entry.Operator, "internal")
	}
	if entry.Result != "error: session not found" {
		t.Errorf("Result = %q, 期望 %q", entry.Result, "error: session not found")
	}
	if entry.StatusCode != 404 {
		t.Errorf("StatusCode = %d, 期望 404", entry.StatusCode)
	}
}

func TestUnaryAuditInterceptor_SkippedMethods(t *testing.T) {
	logger := &mockAuditLogger{}
	interceptor := UnaryAuditInterceptor(logger)

	// 不在审计范围内的方法不应产生审计日志
	skippedMethods := []string{
		"/maze.v1.AgentService/Register",
		"/maze.v1.AgentService/Heartbeat",
		"/maze.v1.HostService/CreateHost",
		"/maze.v1.NodeService/ListNode",
		"/maze.v1.AuditService/QueryLogs",
	}

	for _, method := range skippedMethods {
		_, err := interceptor(
			context.Background(),
			nil,
			&grpc.UnaryServerInfo{FullMethod: method},
			successHandler,
		)
		if err != nil {
			t.Errorf("method %q 不应返回错误: %v", method, err)
		}
	}

	if len(logger.entries) != 0 {
		t.Errorf("审计日志条目数 = %d, 期望 0（所有方法都应跳过审计）", len(logger.entries))
	}
}

func TestExtractOperator(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "有 Authorization header → frontend",
			ctx:  metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer token")),
			want: "frontend",
		},
		{
			name: "无 metadata → internal",
			ctx:  context.Background(),
			want: "internal",
		},
		{
			name: "有 metadata 但无 Authorization → internal",
			ctx:  metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-request-id", "123")),
			want: "internal",
		},
		{
			name: "空 Authorization 值 → internal",
			ctx:  metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "")),
			want: "internal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractOperator(tt.ctx)
			if got != tt.want {
				t.Errorf("extractOperator() = %q, 期望 %q", got, tt.want)
			}
		})
	}
}
