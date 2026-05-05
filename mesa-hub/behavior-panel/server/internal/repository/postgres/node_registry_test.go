package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"

	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/internal/service"
)

func TestNodeRegistryHostTokenLifecycle(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("new pgxmock pool: %v", err)
	}
	defer mock.Close()

	registry := NewNodeRegistry(mock)
	mock.ExpectExec("INSERT INTO host_tokens").
		WithArgs("host-1", "token-1").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectQuery("SELECT token FROM host_tokens WHERE name = \\$1").
		WithArgs("host-1").
		WillReturnRows(pgxmock.NewRows([]string{"token"}).AddRow("token-1"))
	mock.ExpectExec("DELETE FROM host_tokens WHERE name = \\$1").
		WithArgs("host-1").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	if err := registry.StoreHostToken(context.Background(), "host-1", "token-1"); err != nil {
		t.Fatalf("StoreHostToken 返回错误: %v", err)
	}
	exists, matched, err := registry.ValidateHostToken(context.Background(), "host-1", "token-1")
	if err != nil {
		t.Fatalf("ValidateHostToken 返回错误: %v", err)
	}
	if !exists || !matched {
		t.Fatalf("ValidateHostToken = exists:%v matched:%v, want true true", exists, matched)
	}
	if err := registry.RemoveHostToken(context.Background(), "host-1"); err != nil {
		t.Fatalf("RemoveHostToken 返回错误: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("存在未满足的数据库预期: %v", err)
	}
}

func TestNodeRegistryGetRefreshesOfflineStatus(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("new pgxmock pool: %v", err)
	}
	defer mock.Close()

	lastHeartbeat := time.Now().Add(-2 * nodeOfflineThreshold)
	registeredAt := lastHeartbeat.Add(-time.Minute)
	mock.ExpectQuery("SELECT .* FROM nodes WHERE name = \\$1").
		WithArgs("node-1").
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "name", "address", "external_addr", "grpc_address", "status", "capabilities", "agent_status", "metadata", "registered_at", "last_heartbeat",
		}).AddRow(
			int64(1),
			"node-1",
			"http://node-1:8080",
			"",
			"node-1:9090",
			service.NodeStatusOnline,
			[]byte(`{"supported_templates":["claude"],"max_sessions":8,"tools":["tmux"]}`),
			[]byte(`{"active_sessions":2}`),
			[]byte(`{"version":"1.0.0"}`),
			pgtype.Timestamptz{Time: registeredAt, Valid: true},
			pgtype.Timestamptz{Time: lastHeartbeat, Valid: true},
		))
	mock.ExpectExec("UPDATE nodes SET status = 'offline'").
		WithArgs("node-1").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	registry := NewNodeRegistry(mock)
	node, err := registry.Get(context.Background(), "node-1")
	if err != nil {
		t.Fatalf("Get 返回错误: %v", err)
	}
	if node == nil {
		t.Fatal("Get 应返回节点")
	}
	if node.Status != service.NodeStatusOffline {
		t.Fatalf("status = %q, want offline", node.Status)
	}
	if node.AgentStatus.ActiveSessions != 2 {
		t.Fatalf("active_sessions = %d, want 2", node.AgentStatus.ActiveSessions)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("存在未满足的数据库预期: %v", err)
	}
}

func TestNodeRegistryDeleteUsesTransactionExecutor(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("new pgxmock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM nodes WHERE name = \\$1").
		WithArgs("node-1").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectCommit()

	registry := NewNodeRegistry(mock)
	txm := NewHostTxManager(mock)
	err = txm.WithinTx(context.Background(), func(ctx context.Context) error {
		deleted, deleteErr := registry.Delete(ctx, "node-1")
		if deleteErr != nil {
			return deleteErr
		}
		if !deleted {
			t.Fatal("Delete 应返回 true")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WithinTx 返回错误: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("存在未满足的数据库预期: %v", err)
	}
}

func TestNodeRegistryRegisterAndHeartbeat(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("new pgxmock pool: %v", err)
	}
	defer mock.Close()

	now := time.Now().UTC().Round(time.Microsecond)
	row := []any{
		int64(1),
		"node-1",
		"http://node-1:8080",
		"http://public-node-1:8080",
		"node-1:9090",
		service.NodeStatusOnline,
		[]byte(`{"supported_templates":["claude"],"max_sessions":8,"tools":["tmux"]}`),
		[]byte(`{"active_sessions":1}`),
		[]byte(`{"version":"1.0.0"}`),
		pgtype.Timestamptz{Time: now, Valid: true},
		pgtype.Timestamptz{Time: now, Valid: true},
	}

	mock.ExpectQuery("INSERT INTO nodes").
		WithArgs("node-1", "http://node-1:8080", "http://public-node-1:8080", "node-1:9090", pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "name", "address", "external_addr", "grpc_address", "status", "capabilities", "agent_status", "metadata", "registered_at", "last_heartbeat",
		}).AddRow(row...))
	mock.ExpectQuery("UPDATE nodes").
		WithArgs("node-1", pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "name", "address", "external_addr", "grpc_address", "status", "capabilities", "agent_status", "metadata", "registered_at", "last_heartbeat",
		}).AddRow(row...))

	registry := NewNodeRegistry(mock)
	registered, err := registry.Register(context.Background(), protocol.RegisterRequest{
		Name:         "node-1",
		Address:      "http://node-1:8080",
		ExternalAddr: "http://public-node-1:8080",
		GrpcAddress:  "node-1:9090",
		Capabilities: protocol.AgentCapabilities{SupportedTemplates: []string{"claude"}, MaxSessions: 8, Tools: []string{"tmux"}},
		Status:       protocol.AgentStatus{ActiveSessions: 1},
		Metadata:     protocol.AgentMetadata{Version: "1.0.0"},
	})
	if err != nil {
		t.Fatalf("Register 返回错误: %v", err)
	}
	if registered == nil || registered.GrpcAddress != "node-1:9090" {
		t.Fatalf("registered = %+v, want grpc node-1:9090", registered)
	}

	heartbeat, err := registry.Heartbeat(context.Background(), protocol.HeartbeatRequest{
		Name:   "node-1",
		Status: protocol.AgentStatus{ActiveSessions: 1},
	})
	if err != nil {
		t.Fatalf("Heartbeat 返回错误: %v", err)
	}
	if heartbeat == nil || heartbeat.Status != service.NodeStatusOnline {
		t.Fatalf("heartbeat = %+v, want online node", heartbeat)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("存在未满足的数据库预期: %v", err)
	}
}
