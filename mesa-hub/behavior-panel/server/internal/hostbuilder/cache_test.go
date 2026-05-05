package hostbuilder

import "testing"

func TestToolsetImageTag(t *testing.T) {
	tests := []struct {
		name  string
		tools []string
		want  string
	}{
		{name: "empty", tools: nil, want: "maze-agent:"},
		{name: "single", tools: []string{"claude"}, want: "maze-agent:claude"},
		{name: "sorted", tools: []string{"go", "claude"}, want: "maze-agent:claude-go"},
		{name: "presorted", tools: []string{"claude", "go"}, want: "maze-agent:claude-go"},
		{name: "three tools", tools: []string{"python", "go", "claude"}, want: "maze-agent:claude-go-python"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToolsetImageTag(tt.tools)
			if got != tt.want {
				t.Errorf("ToolsetImageTag(%v) = %q, want %q", tt.tools, got, tt.want)
			}
		})
	}
}

func TestToolsetImageTag_DoesNotMutateInput(t *testing.T) {
	input := []string{"b", "a"}
	ToolsetImageTag(input)
	if input[0] != "b" {
		t.Error("ToolsetImageTag should not mutate input slice")
	}
}

func TestExtractDockerfileHash(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{name: "has label", content: "FROM test\nLABEL maze.dockerfile-hash=abc123\nRUN echo", want: "abc123"},
		{name: "no label", content: "FROM test\nRUN echo", want: ""},
		{name: "empty", content: "", want: ""},
		{name: "partial match", content: "LABEL maze.dockerfile=abc", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractDockerfileHash(tt.content)
			if got != tt.want {
				t.Errorf("ExtractDockerfileHash = %q, want %q", got, tt.want)
			}
		})
	}
}
