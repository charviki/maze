package protocol

import (
	"encoding/json"
	"testing"
	"time"
)

func newFixedTime() time.Time {
	return time.Date(2025, 4, 1, 12, 0, 0, 0, time.UTC)
}

func TestRegisterRequest_JSONRoundTrip(t *testing.T) {
	original := RegisterRequest{
		Name:         "agent-dolores-01",
		Address:      "192.168.1.10:8080",
		ExternalAddr: "dolores.maze.local:8080",
		Capabilities: AgentCapabilities{
			SupportedTemplates: []string{"claude", "bash"},
			MaxSessions:        5,
			Tools:              []string{"tmux", "filesystem"},
		},
		Status: AgentStatus{
			ActiveSessions: 2,
			CPUUsage:       45.5,
			MemoryUsageMB:  512.0,
			WorkspaceRoot:  "/home/agent/workspace",
			SessionDetails: []SessionDetail{
				{ID: "sess-1", Template: "claude", WorkingDir: "/home/agent/workspace/proj-a", UptimeSeconds: 3600},
				{ID: "sess-2", Template: "bash", WorkingDir: "/home/agent/workspace/proj-b", UptimeSeconds: 1800},
			},
			LocalConfig: &LocalAgentConfig{
				WorkingDir: "/home/agent/workspace",
				Env:        map[string]string{"PATH": "/usr/bin", "HOME": "/home/agent"},
			},
		},
		Metadata: AgentMetadata{
			Version:   "1.0.0",
			Hostname:  "dolores-host",
			StartedAt: newFixedTime(),
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal 失败: %v", err)
	}

	var decoded RegisterRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal 失败: %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name: 期望 %q, 实际 %q", original.Name, decoded.Name)
	}
	if decoded.Address != original.Address {
		t.Errorf("Address: 期望 %q, 实际 %q", original.Address, decoded.Address)
	}
	if decoded.ExternalAddr != original.ExternalAddr {
		t.Errorf("ExternalAddr: 期望 %q, 实际 %q", original.ExternalAddr, decoded.ExternalAddr)
	}
	if len(decoded.Capabilities.SupportedTemplates) != len(original.Capabilities.SupportedTemplates) {
		t.Errorf("SupportedTemplates 长度: 期望 %d, 实际 %d", len(original.Capabilities.SupportedTemplates), len(decoded.Capabilities.SupportedTemplates))
	}
	if decoded.Capabilities.MaxSessions != original.Capabilities.MaxSessions {
		t.Errorf("MaxSessions: 期望 %d, 实际 %d", original.Capabilities.MaxSessions, decoded.Capabilities.MaxSessions)
	}
	if decoded.Status.ActiveSessions != original.Status.ActiveSessions {
		t.Errorf("ActiveSessions: 期望 %d, 实际 %d", original.Status.ActiveSessions, decoded.Status.ActiveSessions)
	}
	if decoded.Status.CPUUsage != original.Status.CPUUsage {
		t.Errorf("CPUUsage: 期望 %f, 实际 %f", original.Status.CPUUsage, decoded.Status.CPUUsage)
	}
	if decoded.Status.WorkspaceRoot != original.Status.WorkspaceRoot {
		t.Errorf("WorkspaceRoot: 期望 %q, 实际 %q", original.Status.WorkspaceRoot, decoded.Status.WorkspaceRoot)
	}
	if len(decoded.Status.SessionDetails) != len(original.Status.SessionDetails) {
		t.Fatalf("SessionDetails 长度: 期望 %d, 实际 %d", len(original.Status.SessionDetails), len(decoded.Status.SessionDetails))
	}
	for i, sd := range decoded.Status.SessionDetails {
		if sd.ID != original.Status.SessionDetails[i].ID {
			t.Errorf("SessionDetails[%d].ID: 期望 %q, 实际 %q", i, original.Status.SessionDetails[i].ID, sd.ID)
		}
		if sd.UptimeSeconds != original.Status.SessionDetails[i].UptimeSeconds {
			t.Errorf("SessionDetails[%d].UptimeSeconds: 期望 %d, 实际 %d", i, original.Status.SessionDetails[i].UptimeSeconds, sd.UptimeSeconds)
		}
	}
	if decoded.Status.LocalConfig == nil {
		t.Fatal("LocalConfig 不应为 nil")
	}
	if decoded.Status.LocalConfig.WorkingDir != original.Status.LocalConfig.WorkingDir {
		t.Errorf("LocalConfig.WorkingDir: 期望 %q, 实际 %q", original.Status.LocalConfig.WorkingDir, decoded.Status.LocalConfig.WorkingDir)
	}
	if decoded.Metadata.Version != original.Metadata.Version {
		t.Errorf("Metadata.Version: 期望 %q, 实际 %q", original.Metadata.Version, decoded.Metadata.Version)
	}
	if !decoded.Metadata.StartedAt.Equal(original.Metadata.StartedAt) {
		t.Errorf("Metadata.StartedAt: 期望 %v, 实际 %v", original.Metadata.StartedAt, decoded.Metadata.StartedAt)
	}
}

func TestHeartbeatRequest_JSONRoundTrip(t *testing.T) {
	original := HeartbeatRequest{
		Name: "agent-dolores-01",
		Status: AgentStatus{
			ActiveSessions: 1,
			CPUUsage:       30.2,
			MemoryUsageMB:  256.0,
			WorkspaceRoot:  "/home/agent/workspace",
			SessionDetails: []SessionDetail{
				{ID: "sess-1", Template: "claude", WorkingDir: "/home/agent/workspace/proj-a", UptimeSeconds: 7200},
			},
			LocalConfig: &LocalAgentConfig{
				WorkingDir: "/home/agent/workspace",
				Env:        map[string]string{"PATH": "/usr/bin"},
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal 失败: %v", err)
	}

	var decoded HeartbeatRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal 失败: %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name: 期望 %q, 实际 %q", original.Name, decoded.Name)
	}
	if decoded.Status.ActiveSessions != original.Status.ActiveSessions {
		t.Errorf("ActiveSessions: 期望 %d, 实际 %d", original.Status.ActiveSessions, decoded.Status.ActiveSessions)
	}
	if len(decoded.Status.SessionDetails) != 1 {
		t.Fatalf("SessionDetails 长度: 期望 1, 实际 %d", len(decoded.Status.SessionDetails))
	}
	if decoded.Status.SessionDetails[0].ID != "sess-1" {
		t.Errorf("SessionDetails[0].ID: 期望 %q, 实际 %q", "sess-1", decoded.Status.SessionDetails[0].ID)
	}
	if decoded.Status.LocalConfig == nil {
		t.Fatal("LocalConfig 不应为 nil")
	}
	if decoded.Status.LocalConfig.WorkingDir != "/home/agent/workspace" {
		t.Errorf("LocalConfig.WorkingDir: 期望 %q, 实际 %q", "/home/agent/workspace", decoded.Status.LocalConfig.WorkingDir)
	}
}

func TestAuditLogEntry_JSONRoundTrip(t *testing.T) {
	original := AuditLogEntry{
		ID:             "audit-001",
		Timestamp:      newFixedTime(),
		Operator:       "frontend",
		TargetNode:     "agent-dolores-01",
		Action:         "create_session",
		PayloadSummary: `{"template":"claude","working_dir":"/home/agent"}`,
		Result:         "success",
		StatusCode:     200,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal 失败: %v", err)
	}

	var decoded AuditLogEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal 失败: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID: 期望 %q, 实际 %q", original.ID, decoded.ID)
	}
	if !decoded.Timestamp.Equal(original.Timestamp) {
		t.Errorf("Timestamp: 期望 %v, 实际 %v", original.Timestamp, decoded.Timestamp)
	}
	if decoded.Operator != original.Operator {
		t.Errorf("Operator: 期望 %q, 实际 %q", original.Operator, decoded.Operator)
	}
	if decoded.TargetNode != original.TargetNode {
		t.Errorf("TargetNode: 期望 %q, 实际 %q", original.TargetNode, decoded.TargetNode)
	}
	if decoded.Action != original.Action {
		t.Errorf("Action: 期望 %q, 实际 %q", original.Action, decoded.Action)
	}
	if decoded.Result != original.Result {
		t.Errorf("Result: 期望 %q, 实际 %q", original.Result, decoded.Result)
	}
	if decoded.StatusCode != original.StatusCode {
		t.Errorf("StatusCode: 期望 %d, 实际 %d", original.StatusCode, decoded.StatusCode)
	}
}

func TestEmptyCapabilities_JSONRoundTrip(t *testing.T) {
	original := AgentCapabilities{
		SupportedTemplates: nil,
		MaxSessions:        0,
		Tools:              nil,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal 失败: %v", err)
	}

	var decoded AgentCapabilities
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal 失败: %v", err)
	}
	if decoded.MaxSessions != 0 {
		t.Errorf("MaxSessions: 期望 0, 实际 %d", decoded.MaxSessions)
	}
	if len(decoded.SupportedTemplates) != 0 {
		t.Errorf("SupportedTemplates 长度: 期望 0, 实际 %d", len(decoded.SupportedTemplates))
	}
}

func TestEmptySessionDetails_OmitEmpty(t *testing.T) {
	original := AgentStatus{
		ActiveSessions: 0,
		CPUUsage:       0,
		MemoryUsageMB:  0,
		WorkspaceRoot:  "",
		SessionDetails: nil,
		LocalConfig:    nil,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal 失败: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal 到 raw map 失败: %v", err)
	}
	if _, ok := raw["session_details"]; ok {
		t.Error("session_details 应被 omitempty 省略，但出现在 JSON 中")
	}
	if _, ok := raw["local_config"]; ok {
		t.Error("local_config 应被 omitempty 省略，但出现在 JSON 中")
	}

	var decoded AgentStatus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal 失败: %v", err)
	}
	if decoded.LocalConfig != nil {
		t.Error("LocalConfig 应为 nil")
	}
}

func TestJSONTags_FieldMapping(t *testing.T) {
	cases := []struct {
		name     string
		input    interface{}
		expected map[string]bool
	}{
		{
			name:  "AgentCapabilities",
			input: AgentCapabilities{SupportedTemplates: []string{"a"}, MaxSessions: 1, Tools: []string{"b"}},
			expected: map[string]bool{
				"supported_templates": true,
				"max_sessions":        true,
				"tools":               true,
			},
		},
		{
			name:  "SessionDetail",
			input: SessionDetail{ID: "s1", Template: "bash", WorkingDir: "/tmp", UptimeSeconds: 100},
			expected: map[string]bool{
				"id":             true,
				"template":       true,
				"working_dir":    true,
				"uptime_seconds": true,
			},
		},
		{
			name:  "AgentMetadata",
			input: AgentMetadata{Version: "1.0", Hostname: "h", StartedAt: newFixedTime()},
			expected: map[string]bool{
				"version":    true,
				"hostname":   true,
				"started_at": true,
			},
		},
		{
			name:  "AuditLogEntry",
			input: AuditLogEntry{ID: "a1", Timestamp: newFixedTime(), Operator: "op", TargetNode: "n1", Action: "act", PayloadSummary: "p", Result: "r", StatusCode: 200},
			expected: map[string]bool{
				"id":              true,
				"timestamp":       true,
				"operator":        true,
				"target_node":     true,
				"action":          true,
				"payload_summary": true,
				"result":          true,
				"status_code":     true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("Marshal 失败: %v", err)
			}
			var raw map[string]json.RawMessage
			if err := json.Unmarshal(data, &raw); err != nil {
				t.Fatalf("Unmarshal 到 raw map 失败: %v", err)
			}
			for field := range tc.expected {
				if _, ok := raw[field]; !ok {
					t.Errorf("缺少 json 字段 %q, 实际 JSON: %s", field, string(data))
				}
			}
		})
	}
}
