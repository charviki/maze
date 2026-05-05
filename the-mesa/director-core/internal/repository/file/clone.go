package file

import (
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/maze/the-mesa/director-core/internal/service"
)

func cloneHostSpec(spec *protocol.HostSpec) *protocol.HostSpec {
	if spec == nil {
		return nil
	}
	cloned := *spec
	cloned.Tools = append([]string(nil), spec.Tools...)
	return &cloned
}

func cloneNode(node *service.Node) *service.Node {
	if node == nil {
		return nil
	}
	cloned := *node
	cloned.Capabilities = cloneAgentCapabilities(node.Capabilities)
	cloned.AgentStatus = cloneAgentStatus(node.AgentStatus)
	return &cloned
}

func cloneAgentCapabilities(capabilities protocol.AgentCapabilities) protocol.AgentCapabilities {
	return protocol.AgentCapabilities{
		SupportedTemplates: append([]string(nil), capabilities.SupportedTemplates...),
		MaxSessions:        capabilities.MaxSessions,
		Tools:              append([]string(nil), capabilities.Tools...),
	}
}

func cloneAgentStatus(status protocol.AgentStatus) protocol.AgentStatus {
	cloned := protocol.AgentStatus{
		ActiveSessions: status.ActiveSessions,
		CPUUsage:       status.CPUUsage,
		MemoryUsageMB:  status.MemoryUsageMB,
		WorkspaceRoot:  status.WorkspaceRoot,
		SessionDetails: append([]protocol.SessionDetail(nil), status.SessionDetails...),
	}
	if status.LocalConfig != nil {
		cloned.LocalConfig = &protocol.LocalAgentConfig{
			WorkingDir: status.LocalConfig.WorkingDir,
			Env:        cloneStringMap(status.LocalConfig.Env),
		}
	}
	return cloned
}

func cloneStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}
