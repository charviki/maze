package service

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
// 按地址复用连接，避免每次代理请求都建立/断开新连接的开销。
// 空闲超过 idleTTL 的连接由后台 goroutine 自动清理。
type ConnectionManager struct {
	mu        sync.RWMutex
	conns     map[string]*connEntry
	logger    logutil.Logger
	authToken string
	idleTTL   time.Duration
	stopCh    chan struct{}
}

// NewConnectionManager 创建连接管理器并启动空闲连接清理协程。
// authToken 用于代理请求时注入到 gRPC metadata，使 Agent 端认证通过。
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

// GetConn 获取指定地址的 gRPC 连接。同一地址复用已有连接，首次访问时建立新连接。
func (m *ConnectionManager) GetConn(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	m.mu.Lock()
	entry, ok := m.conns[addr]
	if ok {
		entry.lastUsed = time.Now()
		m.mu.Unlock()
		return entry.conn, nil
	}
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	entry, ok = m.conns[addr]
	if ok {
		entry.lastUsed = time.Now()
		return entry.conn, nil
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Manager 代理请求到 Agent 时注入全局认证 token，
	// 否则 Agent 端的 UnaryAuthInterceptor 会拒绝请求。
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

// Remove 关闭并移除指定地址的连接（节点下线时调用）
func (m *ConnectionManager) Remove(addr string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if entry, ok := m.conns[addr]; ok {
		delete(m.conns, addr)
		_ = entry.conn.Close()
	}
}

// CloseAll 关闭所有连接并停止清理协程，用于优雅关闭
func (m *ConnectionManager) CloseAll() {
	close(m.stopCh)
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
