package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charviki/maze/fabrication/cradle/logutil"
	"github.com/charviki/maze/fabrication/cradle/protocol"
)

func TestOverwriteManagedSection_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	if err := overwriteManagedSection(path, "hello world"); err != nil {
		t.Fatalf("overwriteManagedSection: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, managedMarker) {
		t.Error("missing managed marker")
	}
	if !strings.Contains(content, "hello world") {
		t.Error("missing content")
	}
}

func TestOverwriteManagedSection_ReplaceExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	// Write initial content with marker
	if err := overwriteManagedSection(path, "first content"); err != nil {
		t.Fatalf("first write: %v", err)
	}

	// Write again — should replace marker section
	if err := overwriteManagedSection(path, "second content"); err != nil {
		t.Fatalf("second write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "second content") {
		t.Error("should contain second content")
	}
	if strings.Contains(content, "first content") {
		t.Error("should NOT contain first content after replacement")
	}
	// Marker should appear exactly once
	if count := strings.Count(content, managedMarker); count != 1 {
		t.Errorf("marker count = %d, want 1", count)
	}
}

func TestOverwriteManagedSection_AppendToExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	// Write user content first
	if err := os.WriteFile(path, []byte("# My Notes\nSome user content\n"), 0600); err != nil {
		t.Fatalf("write user content: %v", err)
	}

	// Write managed section
	if err := overwriteManagedSection(path, "managed rule"); err != nil {
		t.Fatalf("overwriteManagedSection: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "# My Notes") {
		t.Error("user content should be preserved")
	}
	if !strings.Contains(content, "managed rule") {
		t.Error("managed content should be appended")
	}
}

func TestBuildSkillContent(t *testing.T) {
	got := buildSkillContent(protocol.SkillConfig{
		Name:        "my-skill",
		Description: "does stuff",
		Config:      map[string]string{"key1": "val1"},
	})
	if !strings.Contains(got, "# my-skill") {
		t.Error("should contain skill heading")
	}
	if !strings.Contains(got, "does stuff") {
		t.Error("should contain description")
	}
	if !strings.Contains(got, "key1: val1") {
		t.Error("should contain config")
	}
}

func TestWriteSkills_WritesToBothClaudeAndCodex(t *testing.T) {
	dir := t.TempDir()
	svc := NewHostConfigService(nil, "test-host", "token", logutil.NewNop())

	errs := svc.writeSkills(dir, []protocol.SkillConfig{
		{Name: "skill-1", Description: "test"},
	})
	if errs != 0 {
		t.Fatalf("writeSkills returned %d errors", errs)
	}

	claudePath := filepath.Join(dir, ".claude", "skills", "skill-1.md")
	codexPath := filepath.Join(dir, ".codex", "skills", "skill-1.md")

	for _, p := range []string{claudePath, codexPath} {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("file %s not found: %v", p, err)
			continue
		}
		if !strings.Contains(string(data), "skill-1") {
			t.Errorf("file %s missing skill name", p)
		}
	}
}

func TestWriteRules_WritesToBothClaudeAndAgentsMD(t *testing.T) {
	dir := t.TempDir()
	svc := NewHostConfigService(nil, "test-host", "token", logutil.NewNop())

	errs := svc.writeRules(dir, []protocol.RuleConfig{
		{Name: "rule-1", Content: "do the thing"},
	})
	if errs != 0 {
		t.Fatalf("writeRules returned %d errors", errs)
	}

	claudeMD := filepath.Join(dir, ".claude", "CLAUDE.md")
	agentsMD := filepath.Join(dir, ".codex", "AGENTS.md")

	for _, p := range []string{claudeMD, agentsMD} {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("file %s not found: %v", p, err)
			continue
		}
		content := string(data)
		if !strings.Contains(content, "rule-1") {
			t.Errorf("file %s missing rule name", p)
		}
		if !strings.Contains(content, "do the thing") {
			t.Errorf("file %s missing rule content", p)
		}
	}
}

func TestWriteGitKeys_SSHKey(t *testing.T) {
	dir := t.TempDir()
	svc := NewHostConfigService(nil, "test-host", "token", logutil.NewNop())

	errs := svc.writeGitKeys(dir, []protocol.GitKeyItem{
		{Name: "my-ssh", TokenType: protocol.GitKeyTypeSSHKey, Host: "github.com", DecryptedToken: "ssh-ed25519 AAAAkey"},
	})
	if errs != 0 {
		t.Fatalf("writeGitKeys returned %d errors", errs)
	}

	// Check key file
	keyData, err := os.ReadFile(filepath.Join(dir, ".ssh", "id_ed25519"))
	if err != nil {
		t.Fatalf("ssh key file: %v", err)
	}
	if string(keyData) != "ssh-ed25519 AAAAkey" {
		t.Errorf("key content = %q, want %q", string(keyData), "ssh-ed25519 AAAAkey")
	}

	// Check ssh config
	configData, err := os.ReadFile(filepath.Join(dir, ".ssh", "config"))
	if err != nil {
		t.Fatalf("ssh config: %v", err)
	}
	config := string(configData)
	if !strings.Contains(config, "Host github.com") {
		t.Error("ssh config missing Host entry")
	}
	if !strings.Contains(config, "id_ed25519") {
		t.Error("ssh config missing IdentityFile")
	}
}

func TestWriteGitKeys_PAT(t *testing.T) {
	dir := t.TempDir()
	svc := NewHostConfigService(nil, "test-host", "token", logutil.NewNop())

	errs := svc.writeGitKeys(dir, []protocol.GitKeyItem{
		{Name: "my-pat", TokenType: protocol.GitKeyTypePAT, Host: "github.com", DecryptedToken: "ghp_abc123"},
	})
	if errs != 0 {
		t.Fatalf("writeGitKeys returned %d errors", errs)
	}

	// Check git-credentials
	credData, err := os.ReadFile(filepath.Join(dir, ".git-credentials"))
	if err != nil {
		t.Fatalf("git-credentials: %v", err)
	}
	cred := string(credData)
	if !strings.Contains(cred, "https://x-access-token:ghp_abc123@github.com") {
		t.Errorf("credentials = %q, want contains PAT line", cred)
	}

	// Check .gitconfig
	gitconfigData, err := os.ReadFile(filepath.Join(dir, ".gitconfig"))
	if err != nil {
		t.Fatalf(".gitconfig: %v", err)
	}
	if !strings.Contains(string(gitconfigData), "helper = store") {
		t.Errorf("gitconfig = %q, want contains helper = store", string(gitconfigData))
	}
}

func TestFetchAndApply_ReportsWriteErrors(t *testing.T) {
	// This test verifies error counting — we use a read-only dir to force write failures
	dir := t.TempDir()
	readOnlyDir := filepath.Join(dir, "readonly")
	if err := os.MkdirAll(readOnlyDir, 0555); err != nil {
		t.Fatal(err)
	}

	svc := NewHostConfigService(nil, "test-host", "token", logutil.NewNop())
	// writeSkills to read-only home should fail
	errs := svc.writeSkills(readOnlyDir, []protocol.SkillConfig{
		{Name: "fail-skill", Description: "should fail"},
	})
	if errs == 0 {
		t.Error("expected write errors when writing to read-only directory")
	}
}
