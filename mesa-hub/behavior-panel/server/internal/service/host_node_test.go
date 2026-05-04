package service

import (
	"testing"
	"time"

	"github.com/charviki/maze-cradle/protocol"
)

func sampleHostSpec(name string) *protocol.HostSpec {
	return &protocol.HostSpec{
		Name:      name,
		Tools:     []string{"claude"},
		Status:    protocol.HostStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestBuildHostInfo_DeployingWithOnlineNode(t *testing.T) {
	spec := sampleHostSpec("host-merged")
	spec.Status = protocol.HostStatusDeploying
	heartbeat := time.Now().Add(-2 * time.Second).UTC().Truncate(time.Second)

	info := BuildHostInfo(spec, &Node{
		Name:          spec.Name,
		Address:       "http://host-merged:8080",
		Status:        NodeStatusOnline,
		LastHeartbeat: heartbeat,
		AgentStatus: protocol.AgentStatus{
			ActiveSessions: 4,
		},
	})
	if info == nil {
		t.Fatal("BuildHostInfo 应返回 HostInfo")
	}
	if info.Status != protocol.HostStatusOnline {
		t.Fatalf("Status = %q, want %q", info.Status, protocol.HostStatusOnline)
	}
	if info.Address != "http://host-merged:8080" {
		t.Fatalf("Address = %q, want %q", info.Address, "http://host-merged:8080")
	}
	if info.SessionCount != 4 {
		t.Fatalf("SessionCount = %d, want 4", info.SessionCount)
	}
	if info.LastHeartbeat != heartbeat.Format(time.RFC3339) {
		t.Fatalf("LastHeartbeat = %q, want %q", info.LastHeartbeat, heartbeat.Format(time.RFC3339))
	}
}

func TestBuildHostInfo_DeployingWithoutNodeKeepsSpecState(t *testing.T) {
	spec := sampleHostSpec("host-pending-runtime")
	spec.Status = protocol.HostStatusDeploying

	info := BuildHostInfo(spec, nil)
	if info == nil {
		t.Fatal("BuildHostInfo 应返回 HostInfo")
	}
	if info.Status != protocol.HostStatusDeploying {
		t.Fatalf("Status = %q, want %q", info.Status, protocol.HostStatusDeploying)
	}
}

func TestBuildHostInfo_NonDeployingDoesNotProjectNodeState(t *testing.T) {
	spec := sampleHostSpec("host-failed")
	spec.Status = protocol.HostStatusFailed

	info := BuildHostInfo(spec, &Node{
		Name:    spec.Name,
		Address: "http://host-failed:8080",
		Status:  NodeStatusOnline,
	})
	if info == nil {
		t.Fatal("BuildHostInfo 应返回 HostInfo")
	}
	if info.Status != protocol.HostStatusFailed {
		t.Fatalf("Status = %q, want %q", info.Status, protocol.HostStatusFailed)
	}
	if info.Address != "" {
		t.Fatalf("Address = %q, want empty", info.Address)
	}
}
