package agentclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/charviki/maze-cradle/logutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type connEntry struct {
	conn     *grpc.ClientConn
	lastUsed time.Time
}

// ConnectionManager 管理到 Agent 节点的 gRPC 长连接池。
// 按地址复用连接是为了避免每次代理请求都重复建连，降低代理开销。
// 空闲超过 idleTTL 的连接由后台协程自动清理，避免进程长时间运行后积累废连接。
type ConnectionManager struct {
	mu        sync.RWMutex
	conns     map[string]*connEntry
	logger    logutil.Logger
	authToken string
	idleTTL   time.Duration
	closeOnce sync.Once
	stopCh    chan struct{}
}

// NewConnectionManager 创建连接管理器并启动空闲连接清理协程。
// authToken 会被注入到代理请求的 gRPC metadata 中，确保 Agent 端鉴权链通过。
func NewConnectionManager(logger logutil.Logger, authToken string, idleTTL time.Duration) *ConnectionManager {
	cm := &ConnectionManager{
		conns:     make(map[string]*connEntry),
		logger:    logger,
		authToken: authToken,
		idleTTL:   idleTTL,
		stopCh:    make(chan struct{}),
	}
	go cm.cleanupLoop()
	return cm
}

// GetConn 获取指定地址的 gRPC 连接。
// 同一地址优先复用已有连接，首次访问才会建立新连接。
func (m *ConnectionManager) GetConn(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	m.mu.Lock()
	entry, ok := m.conns[addr]
	if ok {
		entry.lastUsed = time.Now()
		m.mu.Unlock()
		return entry.conn, nil
	}
	defer m.mu.Unlock()

	// 首次 miss 与真正建连之间仍可能被并发请求抢先填充，这里二次检查避免重复建连。
	entry, ok = m.conns[addr]
	if ok {
		entry.lastUsed = time.Now()
		return entry.conn, nil
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Manager 代理调用 Agent RPC 时需要补齐全局认证头，否则 Agent interceptor 会直接拒绝。
	if m.authToken != "" {
		opts = append(opts, grpc.WithUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req interface{},
			reply interface{},
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+m.authToken)
			return invoker(ctx, method, req, reply, cc, opts...)
		}))
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}

	m.logger.Infof("[connection-manager] new gRPC connection to %s", addr)
	m.conns[addr] = &connEntry{conn: conn, lastUsed: time.Now()}
	return conn, nil
}

// Remove 关闭并移除指定地址的连接，供节点下线或地址切换时主动清理。
func (m *ConnectionManager) Remove(addr string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if entry, ok := m.conns[addr]; ok {
		delete(m.conns, addr)
		_ = entry.conn.Close()
	}
}

// CloseAll 关闭所有连接并停止后台清理协程，供优雅关闭时统一释放资源。
func (m *ConnectionManager) CloseAll() {
	m.closeOnce.Do(func() {
		close(m.stopCh)
	})
	m.mu.Lock()
	defer m.mu.Unlock()
	for addr, entry := range m.conns {
		_ = entry.conn.Close()
		delete(m.conns, addr)
	}
}

func (m *ConnectionManager) cleanupLoop() {
	ticker := time.NewTicker(m.idleTTL / 2)
	defer ticker.Stop()
	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.cleanupIdle()
		}
	}
}

func (m *ConnectionManager) cleanupIdle() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for addr, entry := range m.conns {
		if now.Sub(entry.lastUsed) > m.idleTTL {
			m.logger.Infof("[connection-manager] closing idle gRPC connection to %s", addr)
			_ = entry.conn.Close()
			delete(m.conns, addr)
		}
	}
}
