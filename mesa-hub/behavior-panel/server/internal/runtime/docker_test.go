package runtime

import (
	"testing"

	builder "github.com/charviki/mesa-hub-behavior-panel/internal/imagebuilder"
)

func TestExtractDockerfileHash_Found(t *testing.T) {
	content := `FROM test-base:latest
COPY --from=maze-deps-claude:latest /opt/claude /opt/claude
LABEL maze.dockerfile-hash=abc123def4567890`

	hash := builder.ExtractDockerfileHash(content)
	if hash != "abc123def4567890" {
		t.Errorf("hash 不匹配: got %q, want %q", hash, "abc123def4567890")
	}
}

func TestExtractDockerfileHash_NotFound(t *testing.T) {
	content := `FROM test-base:latest
RUN echo hello`

	hash := builder.ExtractDockerfileHash(content)
	if hash != "" {
		t.Errorf("无 hash label 时应返回空字符串, got %q", hash)
	}
}

func TestExtractDockerfileHash_EmptyContent(t *testing.T) {
	hash := builder.ExtractDockerfileHash("")
	if hash != "" {
		t.Errorf("空内容应返回空字符串, got %q", hash)
	}
}

func TestExtractDockerfileHash_LastLine(t *testing.T) {
	content := "FROM base\nLABEL maze.dockerfile-hash=deadbeef12345678"
	hash := builder.ExtractDockerfileHash(content)
	if hash != "deadbeef12345678" {
		t.Errorf("hash 不匹配: got %q, want %q", hash, "deadbeef12345678")
	}
}
