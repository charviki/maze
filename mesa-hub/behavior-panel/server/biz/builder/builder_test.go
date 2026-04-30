package builder

import (
	"strings"
	"testing"
)

func TestDefaultToolRegistry_ContainsAllTools(t *testing.T) {
	expected := []string{"claude", "codex", "go", "python", "node"}
	for _, id := range expected {
		if _, ok := DefaultToolRegistry[id]; !ok {
			t.Errorf("DefaultToolRegistry 缺少工具 %q", id)
		}
	}
	if len(DefaultToolRegistry) != len(expected) {
		t.Errorf("DefaultToolRegistry 期望 %d 个工具, 实际=%d", len(expected), len(DefaultToolRegistry))
	}
}

func TestDefaultToolRegistry_ToolConfigComplete(t *testing.T) {
	for id, cfg := range DefaultToolRegistry {
		if cfg.ID != id {
			t.Errorf("工具 %q 的 ID 字段不匹配: %q", id, cfg.ID)
		}
		if cfg.Image == "" {
			t.Errorf("工具 %q 缺少 Image", id)
		}
		if cfg.SourcePath == "" {
			t.Errorf("工具 %q 缺少 SourcePath", id)
		}
		if cfg.DestPath == "" {
			t.Errorf("工具 %q 缺少 DestPath", id)
		}
		if len(cfg.BinPaths) == 0 {
			t.Errorf("工具 %q 缺少 BinPaths", id)
		}
		if cfg.Description == "" {
			t.Errorf("工具 %q 缺少 Description", id)
		}
		if cfg.Category == "" {
			t.Errorf("工具 %q 缺少 Category", id)
		}
	}
}

func TestValidateTools_AllValid(t *testing.T) {
	unknown := ValidateTools([]string{"claude", "go", "python"})
	if len(unknown) != 0 {
		t.Errorf("全部有效工具应返回空列表, 实际=%v", unknown)
	}
}

func TestValidateTools_UnknownTools(t *testing.T) {
	unknown := ValidateTools([]string{"claude", "rust", "unknown-tool"})
	if len(unknown) != 2 {
		t.Fatalf("期望 2 个未知工具, 实际=%d: %v", len(unknown), unknown)
	}
	if unknown[0] != "rust" {
		t.Errorf("第一个未知工具期望 rust, 实际=%s", unknown[0])
	}
	if unknown[1] != "unknown-tool" {
		t.Errorf("第二个未知工具期望 unknown-tool, 实际=%s", unknown[1])
	}
}

func TestValidateTools_EmptyList(t *testing.T) {
	unknown := ValidateTools([]string{})
	if len(unknown) != 0 {
		t.Errorf("空工具列表应返回空列表, 实际=%v", unknown)
	}
}

func TestListAvailableTools_ReturnsAll(t *testing.T) {
	tools := ListAvailableTools()
	if len(tools) != len(DefaultToolRegistry) {
		t.Errorf("ListAvailableTools 应返回 %d 个工具, 实际=%d", len(DefaultToolRegistry), len(tools))
	}
}

func TestGenerateHostDockerfile_SingleTool(t *testing.T) {
	result, err := GenerateHostDockerfile([]string{"claude"}, "test-base:latest")
	if err != nil {
		t.Fatalf("GenerateHostDockerfile 失败: %v", err)
	}

	if !strings.Contains(result, "FROM test-base:latest") {
		t.Error("Dockerfile 应包含 FROM test-base:latest")
	}
	if !strings.Contains(result, "COPY --from=maze-deps-claude:latest /opt/claude /opt/claude") {
		t.Error("Dockerfile 应包含 claude 的 COPY 指令")
	}
	if !strings.Contains(result, "/opt/claude/bin") {
		t.Error("Dockerfile 应包含 claude 的 bin 路径")
	}
}

func TestGenerateHostDockerfile_MultipleTools(t *testing.T) {
	result, err := GenerateHostDockerfile([]string{"claude", "go"}, "test-base:latest")
	if err != nil {
		t.Fatalf("GenerateHostDockerfile 失败: %v", err)
	}

	claudeCopyCount := strings.Count(result, "COPY --from=maze-deps-claude:latest")
	if claudeCopyCount != 1 {
		t.Errorf("期望 1 条 claude COPY 指令, 实际=%d", claudeCopyCount)
	}
	goCopyCount := strings.Count(result, "COPY --from=maze-deps-go:latest")
	if goCopyCount != 1 {
		t.Errorf("期望 1 条 go COPY 指令, 实际=%d", goCopyCount)
	}

	// PATH 应包含两个工具的 bin 目录
	if !strings.Contains(result, "/opt/claude/bin") {
		t.Error("PATH 应包含 /opt/claude/bin")
	}
	if !strings.Contains(result, "/opt/go/bin") {
		t.Error("PATH 应包含 /opt/go/bin")
	}
}

func TestGenerateHostDockerfile_GoEnvVars(t *testing.T) {
	result, err := GenerateHostDockerfile([]string{"go"}, "test-base:latest")
	if err != nil {
		t.Fatalf("GenerateHostDockerfile 失败: %v", err)
	}

	if !strings.Contains(result, "ENV GOROOT=/opt/go") {
		t.Error("Go 工具应设置 GOROOT")
	}
	if !strings.Contains(result, "GOPROXY=https://goproxy.cn,direct") {
		t.Error("Go 工具应设置 GOPROXY")
	}
}

func TestGenerateHostDockerfile_EmptyTools(t *testing.T) {
	result, err := GenerateHostDockerfile([]string{}, "test-base:latest")
	if err != nil {
		t.Fatalf("GenerateHostDockerfile 失败: %v", err)
	}

	if !strings.Contains(result, "FROM test-base:latest") {
		t.Error("即使无工具也应包含 FROM 指令")
	}
	if strings.Contains(result, "COPY --from=") {
		t.Error("无工具时不应包含 COPY 指令")
	}
}

func TestGenerateHostDockerfile_UnknownToolSkipped(t *testing.T) {
	result, err := GenerateHostDockerfile([]string{"claude", "nonexistent"}, "test-base:latest")
	if err != nil {
		t.Fatalf("GenerateHostDockerfile 失败: %v", err)
	}

	claudeCopyCount := strings.Count(result, "COPY --from=maze-deps-claude:latest")
	if claudeCopyCount != 1 {
		t.Errorf("期望 1 条 claude COPY 指令, 实际=%d", claudeCopyCount)
	}
	if strings.Contains(result, "nonexistent") {
		t.Error("未知工具不应出现在 Dockerfile 中")
	}
}
