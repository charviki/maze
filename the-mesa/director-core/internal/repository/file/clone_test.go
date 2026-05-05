package file

import (
	"testing"

	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/maze/the-mesa/director-core/internal/service"
)

func TestCloneHostSpec_Nil(t *testing.T) {
	if cloneHostSpec(nil) != nil {
		t.Error("cloneHostSpec(nil) should return nil")
	}
}

func TestCloneHostSpec_DeepCopy(t *testing.T) {
	spec := &protocol.HostSpec{Name: "test", Tools: []string{"claude", "go"}}
	cloned := cloneHostSpec(spec)

	spec.Tools[0] = "tampered"
	if cloned.Tools[0] != "claude" {
		t.Errorf("clone should have independent Tools: %v", cloned.Tools)
	}
}

func TestCloneNode_Nil(t *testing.T) {
	if cloneNode(nil) != nil {
		t.Error("cloneNode(nil) should return nil")
	}
}

func TestCloneNode_DeepCopy(t *testing.T) {
	node := &service.Node{
		Name: "agent-1",
		Capabilities: protocol.AgentCapabilities{
			SupportedTemplates: []string{"claude"},
			Tools:              []string{"tmux"},
		},
		AgentStatus: protocol.AgentStatus{
			SessionDetails: []protocol.SessionDetail{{ID: "s1"}},
		},
	}
	cloned := cloneNode(node)

	node.Capabilities.SupportedTemplates[0] = "tampered"
	node.AgentStatus.SessionDetails[0].ID = "tampered"

	if cloned.Capabilities.SupportedTemplates[0] != "claude" {
		t.Errorf("clone should have independent Capabilities: %v", cloned.Capabilities)
	}
	if cloned.AgentStatus.SessionDetails[0].ID != "s1" {
		t.Errorf("clone should have independent SessionDetails: %+v", cloned.AgentStatus.SessionDetails)
	}
}

func TestCloneAgentCapabilities_NilSlice(t *testing.T) {
	caps := protocol.AgentCapabilities{}
	cloned := cloneAgentCapabilities(caps)
	if cloned.SupportedTemplates != nil {
		t.Errorf("nil SupportedTemplates should remain nil, got %v", cloned.SupportedTemplates)
	}
	if cloned.Tools != nil {
		t.Errorf("nil Tools should remain nil, got %v", cloned.Tools)
	}
}

func TestCloneAgentStatus_LocalConfig(t *testing.T) {
	status := protocol.AgentStatus{
		LocalConfig: &protocol.LocalAgentConfig{
			WorkingDir: "/workspace",
			Env:        map[string]string{"KEY": "val"},
		},
	}
	cloned := cloneAgentStatus(status)
	status.LocalConfig.Env["KEY"] = "tampered"

	if cloned.LocalConfig.Env["KEY"] != "val" {
		t.Errorf("clone should have independent Env: %v", cloned.LocalConfig.Env)
	}
	if cloned.WorkspaceRoot != "" {
		t.Errorf("WorkspaceRoot = %q, want empty", cloned.WorkspaceRoot)
	}
}

func TestCloneStringMap_Nil(t *testing.T) {
	if cloneStringMap(nil) != nil {
		t.Error("cloneStringMap(nil) should return nil")
	}
}

func TestCloneStringMap_Empty(t *testing.T) {
	cloned := cloneStringMap(map[string]string{})
	if len(cloned) != 0 {
		t.Errorf("empty map clone length = %d, want 0", len(cloned))
	}
}
