package handler

import (
	"fmt"
	"testing"
	"time"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
)

// ========== AuditLogger 测试 ==========

func TestAuditLogger_LogAndList(t *testing.T) {
	a := NewAuditLogger("", logutil.NewNop())

	for i := 1; i <= 3; i++ {
		a.Log(protocol.AuditLogEntry{
			TargetNode: fmt.Sprintf("agent-%d", i),
			Action:     "list_sessions",
			Result:     "success",
		})
	}

	logs := a.List()
	if len(logs) != 3 {
		t.Fatalf("期望 3 条日志, 实际=%d", len(logs))
	}
	// 最新的在前
	if logs[0].TargetNode != "agent-3" {
		t.Errorf("最新日志应为 agent-3, 实际=%s", logs[0].TargetNode)
	}
	if logs[2].TargetNode != "agent-1" {
		t.Errorf("最早日志应为 agent-1, 实际=%s", logs[2].TargetNode)
	}
}

func TestAuditLogger_AutoFillFields(t *testing.T) {
	a := NewAuditLogger("", logutil.NewNop())

	before := time.Now()
	a.Log(protocol.AuditLogEntry{
		TargetNode: "agent-1",
		Action:     "create_session",
		Result:     "success",
	})
	after := time.Now()

	logs := a.List()
	if len(logs) != 1 {
		t.Fatalf("期望 1 条日志, 实际=%d", len(logs))
	}
	entry := logs[0]
	if entry.ID == "" {
		t.Error("ID 应被自动填充")
	}
	if entry.Timestamp.Before(before) || entry.Timestamp.After(after) {
		t.Errorf("Timestamp 应在 [%v, %v] 范围内, 实际=%v", before, after, entry.Timestamp)
	}
}

func TestAuditLogger_QueryByNode(t *testing.T) {
	a := NewAuditLogger("", logutil.NewNop())
	a.Log(protocol.AuditLogEntry{TargetNode: "agent-1", Action: "list_sessions"})
	a.Log(protocol.AuditLogEntry{TargetNode: "agent-2", Action: "list_sessions"})
	a.Log(protocol.AuditLogEntry{TargetNode: "agent-1", Action: "create_session"})

	results := a.Query("agent-1", "")
	if len(results) != 2 {
		t.Fatalf("期望 2 条 agent-1 记录, 实际=%d", len(results))
	}
	for _, e := range results {
		if e.TargetNode != "agent-1" {
			t.Errorf("期望 TargetNode=agent-1, 实际=%s", e.TargetNode)
		}
	}
}

func TestAuditLogger_QueryByAction(t *testing.T) {
	a := NewAuditLogger("", logutil.NewNop())
	a.Log(protocol.AuditLogEntry{TargetNode: "agent-1", Action: "list_sessions"})
	a.Log(protocol.AuditLogEntry{TargetNode: "agent-2", Action: "create_session"})
	a.Log(protocol.AuditLogEntry{TargetNode: "agent-1", Action: "create_session"})

	results := a.Query("", "create")
	if len(results) != 2 {
		t.Fatalf("期望 2 条 create 操作记录, 实际=%d", len(results))
	}
	for _, e := range results {
		if e.Action != "create_session" {
			t.Errorf("期望 Action=create_session, 实际=%s", e.Action)
		}
	}
}

func TestAuditLogger_QueryByNodeAndAction(t *testing.T) {
	a := NewAuditLogger("", logutil.NewNop())
	a.Log(protocol.AuditLogEntry{TargetNode: "agent-1", Action: "list_sessions"})
	a.Log(protocol.AuditLogEntry{TargetNode: "agent-1", Action: "create_session"})
	a.Log(protocol.AuditLogEntry{TargetNode: "agent-2", Action: "create_session"})

	results := a.Query("agent-1", "create")
	if len(results) != 1 {
		t.Fatalf("期望 1 条 agent-1+create 记录, 实际=%d", len(results))
	}
	if results[0].TargetNode != "agent-1" || results[0].Action != "create_session" {
		t.Errorf("期望 agent-1+create_session, 实际 %s/%s", results[0].TargetNode, results[0].Action)
	}
}
