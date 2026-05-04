package hostbuilder

import (
	"slices"
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

	ids := make([]string, len(tools))
	for i, tool := range tools {
		ids[i] = tool.ID
	}
	if !slices.IsSorted(ids) {
		t.Fatalf("期望工具列表按 ID 稳定排序，实际=%v", ids)
	}
}

func TestGenerateHostDockerfile_SingleTool(t *testing.T) {
	result := GenerateHostDockerfile([]string{"claude"}, "test-base:latest")

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
	result := GenerateHostDockerfile([]string{"claude", "go"}, "test-base:latest")

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
	result := GenerateHostDockerfile([]string{"go"}, "test-base:latest")

	if !strings.Contains(result, "ENV GOROOT=/opt/go") {
		t.Error("Go 工具应设置 GOROOT")
	}
	if !strings.Contains(result, "GOPROXY=https://goproxy.cn,direct") {
		t.Error("Go 工具应设置 GOPROXY")
	}
}

func TestGenerateHostDockerfile_EmptyTools(t *testing.T) {
	result := GenerateHostDockerfile([]string{}, "test-base:latest")

	if !strings.Contains(result, "FROM test-base:latest") {
		t.Error("即使无工具也应包含 FROM 指令")
	}
	if strings.Contains(result, "COPY --from=") {
		t.Error("无工具时不应包含 COPY 指令")
	}
}

func TestGenerateHostDockerfile_UnknownToolSkipped(t *testing.T) {
	result := GenerateHostDockerfile([]string{"claude", "nonexistent"}, "test-base:latest")

	claudeCopyCount := strings.Count(result, "COPY --from=maze-deps-claude:latest")
	if claudeCopyCount != 1 {
		t.Errorf("期望 1 条 claude COPY 指令, 实际=%d", claudeCopyCount)
	}
	if strings.Contains(result, "nonexistent") {
		t.Error("未知工具不应出现在 Dockerfile 中")
	}
}

func TestGenerateHostDockerfile_ContainsHashLabel(t *testing.T) {
	result := GenerateHostDockerfile([]string{"claude"}, "test-base:latest")

	if !strings.Contains(result, "LABEL maze.dockerfile-hash=") {
		t.Error("Dockerfile 应包含 maze.dockerfile-hash label")
	}
}

func TestGenerateHostDockerfile_HashChangesWithTools(t *testing.T) {
	result1 := GenerateHostDockerfile([]string{"claude"}, "test-base:latest")
	result2 := GenerateHostDockerfile([]string{"go"}, "test-base:latest")

	hash1 := extractHashFromDockerfile(result1)
	hash2 := extractHashFromDockerfile(result2)

	if hash1 == "" || hash2 == "" {
		t.Fatal("hash 不应为空")
	}
	if hash1 == hash2 {
		t.Errorf("不同工具组合应产生不同的 hash: %s == %s", hash1, hash2)
	}
}

func TestGenerateHostDockerfile_HashChangesWithBaseImage(t *testing.T) {
	result1 := GenerateHostDockerfile([]string{"claude"}, "base-v1:latest")
	result2 := GenerateHostDockerfile([]string{"claude"}, "base-v2:latest")

	hash1 := extractHashFromDockerfile(result1)
	hash2 := extractHashFromDockerfile(result2)

	if hash1 == hash2 {
		t.Errorf("不同基础镜像应产生不同的 hash: %s == %s", hash1, hash2)
	}
}

func TestGenerateHostDockerfile_SameInputSameHash(t *testing.T) {
	tools := []string{"claude", "go"}
	result1 := GenerateHostDockerfile(tools, "test-base:latest")
	result2 := GenerateHostDockerfile(tools, "test-base:latest")

	hash1 := extractHashFromDockerfile(result1)
	hash2 := extractHashFromDockerfile(result2)

	if hash1 != hash2 {
		t.Errorf("相同输入应产生相同的 hash: %s != %s", hash1, hash2)
	}
}

func TestDockerfileHash_Deterministic(t *testing.T) {
	content := "FROM test-base\nRUN echo hello"
	hash1 := DockerfileHash(content)
	hash2 := DockerfileHash(content)

	if hash1 != hash2 {
		t.Errorf("DockerfileHash 应是确定性的: %s != %s", hash1, hash2)
	}
}

func TestDockerfileHash_DifferentContent(t *testing.T) {
	hash1 := DockerfileHash("FROM base-v1\nRUN echo 1")
	hash2 := DockerfileHash("FROM base-v2\nRUN echo 2")

	if hash1 == hash2 {
		t.Errorf("不同内容应产生不同的 hash")
	}
}

func TestDockerfileHash_Length(t *testing.T) {
	hash := DockerfileHash("some content")
	if len(hash) != 16 {
		t.Errorf("hash 长度应为 16, 实际=%d", len(hash))
	}
}

// extractHashFromDockerfile 从生成的 Dockerfile 中提取 hash label 值
func extractHashFromDockerfile(dockerfile string) string {
	for _, line := range strings.Split(dockerfile, "\n") {
		if strings.HasPrefix(line, "LABEL maze.dockerfile-hash=") {
			return strings.TrimPrefix(line, "LABEL maze.dockerfile-hash=")
		}
	}
	return ""
}
