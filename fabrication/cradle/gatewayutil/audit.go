package gatewayutil

import (
	"context"
	"strings"
	"unicode"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// AuditEntry 审计日志条目，与 handler/audit_logger.go 的 AuditLogEntry 字段对齐。
// 作为 gatewayutil 包的独立定义，避免 handler 对 cradle 的循环依赖。
type AuditEntry struct {
	Operator       string
	TargetNode     string
	Action         string
	PayloadSummary string
	Result         string
	StatusCode     int
}

// AuditLogger 审计日志接口，由调用方（Manager）注入具体实现。
// 方法签名与 handler.AuditLogger.Log 对齐：接收一个条目并持久化。
type AuditLogger interface {
	Log(entry AuditEntry)
}

// auditableServices 需要审计的 gRPC 服务前缀（Session/Template/Config 代理方法均含 node_name）
var auditableServices = []string{
	"/maze.v1.SessionService/",
	"/maze.v1.TemplateService/",
	"/maze.v1.ConfigService/",
}

// shouldAudit 判断 RPC 方法是否需要审计。
// 仅审计 Session/Template/Config 三类代理方法；HostService、NodeService、AuditService、AgentService 跳过。
func shouldAudit(method string) bool {
	for _, prefix := range auditableServices {
		if strings.HasPrefix(method, prefix) {
			return true
		}
	}
	return false
}

// extractNodeName 通过 protobuf 反射从请求中提取目标节点名称。
// 优先查找 node_name 字段（Session/Template/Config 请求），其次查找 name 字段。
func extractNodeName(req interface{}) string {
	msg, ok := req.(proto.Message)
	if !ok {
		return ""
	}

	reflectMsg := msg.ProtoReflect()
	if reflectMsg == nil {
		return ""
	}

	// 优先取 node_name 字段（Session/Template/Config 的代理请求都含此字段）
	if field := getField(reflectMsg, "node_name"); field.IsValid() {
		if s := field.String(); s != "" {
			return s
		}
	}

	// 兜底取 name 字段
	if field := getField(reflectMsg, "name"); field.IsValid() {
		return field.String()
	}

	return ""
}

// getField 从 protobuf 消息反射中安全获取指定字段描述的值。
// 字段不存在时返回零值 Value（IsValid() == false）。
func getField(msg protoreflect.Message, fieldName string) protoreflect.Value {
	desc := msg.Descriptor().Fields().ByName(protoreflect.Name(fieldName))
	if desc == nil {
		return protoreflect.Value{}
	}
	return msg.Get(desc)
}

// methodToAction 将 gRPC 完整方法名转换为审计 action。
// 示例："/maze.v1.SessionService/ListSessions" → "list_sessions"
func methodToAction(fullMethod string) string {
	// 提取最后一段方法名（如 "ListSessions"）
	idx := strings.LastIndex(fullMethod, "/")
	if idx < 0 {
		return strings.ToLower(fullMethod)
	}
	methodName := fullMethod[idx+1:]
	if methodName == "" {
		return ""
	}

	// CamelCase → snake_case：在每个大写字母前插入下划线，然后全转小写
	var b strings.Builder
	for i, r := range methodName {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// grpcCodeToHTTP 将 gRPC 错误码映射为 HTTP 状态码，用于审计日志的 StatusCode 字段。
func grpcCodeToHTTP(code codes.Code) int {
	switch code {
	case codes.OK:
		return 200
	case codes.Canceled:
		return 499
	case codes.Unknown:
		return 500
	case codes.InvalidArgument:
		return 400
	case codes.DeadlineExceeded:
		return 504
	case codes.NotFound:
		return 404
	case codes.AlreadyExists:
		return 409
	case codes.PermissionDenied:
		return 403
	case codes.ResourceExhausted:
		return 429
	case codes.FailedPrecondition:
		return 400
	case codes.Aborted:
		return 409
	case codes.OutOfRange:
		return 400
	case codes.Unimplemented:
		return 501
	case codes.Internal:
		return 500
	case codes.Unavailable:
		return 503
	case codes.DataLoss:
		return 500
	case codes.Unauthenticated:
		return 401
	default:
		return 500
	}
}

// extractOperator 从 gRPC metadata 中提取操作者信息。
// 前端请求经 grpc-gateway 转发时，HTTP Authorization header 会被映射到 gRPC metadata。
// 存在 Authorization 时标记为 "frontend"，否则标记为 "internal"。
func extractOperator(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "internal"
	}
	authHeaders := md.Get("authorization")
	if len(authHeaders) > 0 && authHeaders[0] != "" {
		return "frontend"
	}
	return "internal"
}

// UnaryAuditInterceptor 返回 gRPC 一元拦截器，自动审计代理方法调用。
// 拦截逻辑：
//  1. 跳过不需要审计的方法（Host/Node/Audit/Agent Service）
//  2. 提取操作者、目标节点、action 名称
//  3. 调用下游 handler 后记录成功/失败审计日志
func UnaryAuditInterceptor(logger AuditLogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 不需要审计的方法直接透传
		if !shouldAudit(info.FullMethod) {
			return handler(ctx, req)
		}

		operator := extractOperator(ctx)
		nodeName := extractNodeName(req)
		action := methodToAction(info.FullMethod)

		resp, err := handler(ctx, req)

		// 根据调用结果构建审计条目
		entry := AuditEntry{
			Operator:   operator,
			TargetNode: nodeName,
			Action:     action,
		}

		if err != nil {
			// 下游返回错误：记录错误信息和对应的 HTTP 状态码
			st, _ := status.FromError(err)
			entry.Result = "error: " + st.Message()
			entry.StatusCode = grpcCodeToHTTP(st.Code())
		} else {
			entry.Result = "success"
			entry.StatusCode = 200
		}

		logger.Log(entry)
		return resp, err
	}
}
