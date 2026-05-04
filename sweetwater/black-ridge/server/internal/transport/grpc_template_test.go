package transport

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
)

func TestListTemplates_ReturnsBuiltins(t *testing.T) {
	srv := newTestServer(t, &mockTmuxService{})

	resp, err := srv.ListTemplates(context.Background(), &pb.ListTemplatesRequest{})
	if err != nil {
		t.Fatalf("ListTemplates 失败: %v", err)
	}

	ids := make(map[string]bool)
	for _, tpl := range resp.GetTemplates() {
		ids[tpl.GetId()] = true
	}
	for _, id := range []string{"claude", "codex", "bash"} {
		if !ids[id] {
			t.Errorf("缺少内置模板 %q", id)
		}
	}
	if len(resp.GetTemplates()) < 3 {
		t.Errorf("模板数量 = %d, 期望至少 3", len(resp.GetTemplates()))
	}
}

func TestCreateTemplate_Success(t *testing.T) {
	srv := newTestServer(t, &mockTmuxService{})

	resp, err := srv.CreateTemplate(context.Background(), &pb.CreateTemplateRequest{
		Template: &pb.SessionTemplate{
			Id:      "my-custom",
			Name:    "My Custom Template",
			Command: "echo hello",
		},
	})
	if err != nil {
		t.Fatalf("CreateTemplate 失败: %v", err)
	}
	if resp.GetId() != "my-custom" {
		t.Errorf("id = %q, 期望 %q", resp.GetId(), "my-custom")
	}
	if resp.GetBuiltin() {
		t.Error("自定义模板不应标记为 builtin")
	}
}

func TestCreateTemplate_DuplicateID(t *testing.T) {
	srv := newTestServer(t, &mockTmuxService{})

	_, err := srv.CreateTemplate(context.Background(), &pb.CreateTemplateRequest{
		Template: &pb.SessionTemplate{
			Id:      "claude",
			Name:    "Duplicate Claude",
			Command: "claude",
		},
	})
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("应返回 gRPC status")
	}
	if st.Code() != codes.AlreadyExists {
		t.Errorf("code = %v, 期望 AlreadyExists", st.Code())
	}
}

func TestCreateTemplate_MissingID(t *testing.T) {
	srv := newTestServer(t, &mockTmuxService{})

	_, err := srv.CreateTemplate(context.Background(), &pb.CreateTemplateRequest{
		Template: &pb.SessionTemplate{
			Name:    "No ID",
			Command: "echo",
		},
	})
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("应返回 gRPC status")
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, 期望 InvalidArgument", st.Code())
	}
}

func TestGetTemplate_Found(t *testing.T) {
	srv := newTestServer(t, &mockTmuxService{})

	resp, err := srv.GetTemplate(context.Background(), &pb.GetTemplateRequest{Id: "claude"})
	if err != nil {
		t.Fatalf("GetTemplate 失败: %v", err)
	}
	if resp.GetId() != "claude" {
		t.Errorf("id = %q, 期望 %q", resp.GetId(), "claude")
	}
	if !resp.GetBuiltin() {
		t.Error("claude 应为内置模板")
	}
}

func TestGetTemplate_NotFound(t *testing.T) {
	srv := newTestServer(t, &mockTmuxService{})

	_, err := srv.GetTemplate(context.Background(), &pb.GetTemplateRequest{Id: "nonexistent"})
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("应返回 gRPC status")
	}
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, 期望 NotFound", st.Code())
	}
}

func TestUpdateTemplate_Success(t *testing.T) {
	srv := newTestServer(t, &mockTmuxService{})

	_, err := srv.CreateTemplate(context.Background(), &pb.CreateTemplateRequest{
		Template: &pb.SessionTemplate{
			Id:      "custom-tpl",
			Name:    "Original",
			Command: "echo v1",
		},
	})
	if err != nil {
		t.Fatalf("CreateTemplate 失败: %v", err)
	}

	resp, err := srv.UpdateTemplate(context.Background(), &pb.UpdateTemplateRequest{
		Id: "custom-tpl",
		Template: &pb.SessionTemplate{
			Name:    "Updated",
			Command: "echo v2",
		},
	})
	if err != nil {
		t.Fatalf("UpdateTemplate 失败: %v", err)
	}
	if resp.GetName() != "Updated" {
		t.Errorf("name = %q, 期望 %q", resp.GetName(), "Updated")
	}
	if resp.GetCommand() != "echo v2" {
		t.Errorf("command = %q, 期望 %q", resp.GetCommand(), "echo v2")
	}
}

func TestDeleteTemplate_Success(t *testing.T) {
	srv := newTestServer(t, &mockTmuxService{})

	_, err := srv.CreateTemplate(context.Background(), &pb.CreateTemplateRequest{
		Template: &pb.SessionTemplate{
			Id:      "to-delete",
			Name:    "To Delete",
			Command: "echo bye",
		},
	})
	if err != nil {
		t.Fatalf("CreateTemplate 失败: %v", err)
	}

	_, err = srv.DeleteTemplate(context.Background(), &pb.DeleteTemplateRequest{Id: "to-delete"})
	if err != nil {
		t.Fatalf("DeleteTemplate 失败: %v", err)
	}

	_, err = srv.GetTemplate(context.Background(), &pb.GetTemplateRequest{Id: "to-delete"})
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("应返回 gRPC status")
	}
	if st.Code() != codes.NotFound {
		t.Errorf("删除后 GetTemplate code = %v, 期望 NotFound", st.Code())
	}
}

func TestDeleteTemplate_BuiltinProtected(t *testing.T) {
	srv := newTestServer(t, &mockTmuxService{})

	_, err := srv.DeleteTemplate(context.Background(), &pb.DeleteTemplateRequest{Id: "claude"})
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("应返回 gRPC status")
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("code = %v, 期望 PermissionDenied", st.Code())
	}
}

func TestGetTemplateConfig_ClaudeHasFiles(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	srv := newTestServer(t, &mockTmuxService{})

	resp, err := srv.GetTemplateConfig(context.Background(), &pb.GetTemplateConfigRequest{Id: "claude"})
	if err != nil {
		t.Fatalf("GetTemplateConfig 失败: %v", err)
	}
	if resp.GetTemplateId() != "claude" {
		t.Errorf("template_id = %q, 期望 %q", resp.GetTemplateId(), "claude")
	}
	if len(resp.GetFiles()) == 0 {
		t.Error("claude 模板应返回文件列表")
	}
}

func TestUpdateTemplateConfig_ClaudeUpdatesGlobalFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	srv := newTestServer(t, &mockTmuxService{})

	// 先读取 claude 配置拿到初始文件快照
	configResp, err := srv.GetTemplateConfig(context.Background(), &pb.GetTemplateConfigRequest{Id: "claude"})
	if err != nil {
		t.Fatalf("GetTemplateConfig 失败: %v", err)
	}

	// 找到 ~/.claude/CLAUDE.md 的初始 hash，用于 base_hash
	var initialHash string
	for _, f := range configResp.GetFiles() {
		if f.GetPath() == "~/.claude/CLAUDE.md" {
			initialHash = f.GetHash()
			break
		}
	}
	if initialHash == "" {
		t.Fatal("未找到 ~/.claude/CLAUDE.md 文件快照")
	}

	// 更新该文件内容
	newContent := "# Updated Global Instructions\n"
	resp, err := srv.UpdateTemplateConfig(context.Background(), &pb.UpdateTemplateConfigRequest{
		Id: "claude",
		Files: []*pb.ConfigFileUpdate{
			{Path: "~/.claude/CLAUDE.md", Content: newContent, BaseHash: initialHash},
		},
	})
	if err != nil {
		t.Fatalf("UpdateTemplateConfig 失败: %v", err)
	}

	found := false
	for _, f := range resp.GetFiles() {
		if f.GetPath() == "~/.claude/CLAUDE.md" {
			found = true
			if f.GetContent() != newContent {
				t.Errorf("content = %q, 期望 %q", f.GetContent(), newContent)
			}
			break
		}
	}
	if !found {
		t.Error("响应中未找到 ~/.claude/CLAUDE.md")
	}
}
