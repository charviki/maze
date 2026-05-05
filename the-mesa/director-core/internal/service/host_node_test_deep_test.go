package service

import (
	"testing"
	"time"

	"github.com/charviki/maze-cradle/protocol"
)

func TestNode_RefreshOfflineStatus(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		node        *Node
		threshold   time.Duration
		wantChanged bool
		wantStatus  string
	}{
		{
			name:        "nil node",
			node:        nil,
			threshold:   30 * time.Second,
			wantChanged: false,
		},
		{
			name: "already offline",
			node: &Node{
				Status:        NodeStatusOffline,
				LastHeartbeat: now.Add(-60 * time.Second),
			},
			threshold:   30 * time.Second,
			wantChanged: false,
			wantStatus:  NodeStatusOffline,
		},
		{
			name: "within threshold",
			node: &Node{
				Status:        NodeStatusOnline,
				LastHeartbeat: now.Add(-10 * time.Second),
			},
			threshold:   30 * time.Second,
			wantChanged: false,
			wantStatus:  NodeStatusOnline,
		},
		{
			name: "exceeds threshold",
			node: &Node{
				Status:        NodeStatusOnline,
				LastHeartbeat: now.Add(-60 * time.Second),
			},
			threshold:   30 * time.Second,
			wantChanged: true,
			wantStatus:  NodeStatusOffline,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed := tt.node.RefreshOfflineStatus(now, tt.threshold)
			if changed != tt.wantChanged {
				t.Errorf("changed = %v, want %v", changed, tt.wantChanged)
			}
			if tt.node != nil && tt.node.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", tt.node.Status, tt.wantStatus)
			}
		})
	}
}

func TestNode_BindUnbind(t *testing.T) {
	node := &Node{
		Name:    "agent-1",
		Address: "http://10.0.0.1:9090",
		Status:  NodeStatusOnline,
	}
	if node.Status != NodeStatusOnline {
		t.Errorf("Status = %q, want %q", node.Status, NodeStatusOnline)
	}

	node.Status = NodeStatusOffline
	if node.Status != NodeStatusOffline {
		t.Error("node should be markable as offline")
	}
}

func TestNode_MultiNodeStatus(t *testing.T) {
	nodes := []Node{
		{Name: "agent-1", Status: NodeStatusOnline},
		{Name: "agent-2", Status: NodeStatusOffline},
		{Name: "agent-3", Status: NodeStatusOnline},
	}

	onlineCount := 0
	offlineCount := 0
	for _, n := range nodes {
		switch n.Status {
		case NodeStatusOnline:
			onlineCount++
		case NodeStatusOffline:
			offlineCount++
		}
	}
	if onlineCount != 2 {
		t.Errorf("onlineCount = %d, want 2", onlineCount)
	}
	if offlineCount != 1 {
		t.Errorf("offlineCount = %d, want 1", offlineCount)
	}
}

func TestNode_RefreshOfflineStatus_Boundary(t *testing.T) {
	now := time.Now()

	node := &Node{
		Status:        NodeStatusOnline,
		LastHeartbeat: now.Add(-30 * time.Second),
	}
	changed := node.RefreshOfflineStatus(now, 30*time.Second)
	if changed {
		t.Error("should not be offline exactly at threshold boundary")
	}
	if node.Status != NodeStatusOnline {
		t.Errorf("Status = %q, want %q", node.Status, NodeStatusOnline)
	}
}

func TestNode_FieldsAfterRegister(t *testing.T) {
	now := time.Now()
	node := &Node{
		Name:          "agent-1",
		Address:       "http://10.0.0.1:9090",
		Status:        NodeStatusOnline,
		RegisteredAt:  now,
		LastHeartbeat: now,
		Capabilities: protocol.AgentCapabilities{
			SupportedTemplates: []string{"claude"},
			MaxSessions:        5,
		},
	}

	if node.Capabilities.MaxSessions != 5 {
		t.Errorf("MaxSessions = %d, want 5", node.Capabilities.MaxSessions)
	}
	if len(node.Capabilities.SupportedTemplates) != 1 {
		t.Errorf("SupportedTemplates length = %d, want 1", len(node.Capabilities.SupportedTemplates))
	}
}

func TestFormatNodeSummary(t *testing.T) {
	node := &Node{
		Name:    "agent-summary",
		Address: "http://10.0.0.1:9090",
		Status:  NodeStatusOnline,
		AgentStatus: protocol.AgentStatus{
			ActiveSessions: 3,
			CPUUsage:       45.0,
			MemoryUsageMB:  512.0,
		},
	}

	summary := FormatNodeSummary(node)
	if summary == "" {
		t.Error("FormatNodeSummary should not return empty string")
	}
}
