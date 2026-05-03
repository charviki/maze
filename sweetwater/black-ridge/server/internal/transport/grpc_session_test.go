package transport

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/sweetwater-black-ridge/internal/model"
	"github.com/charviki/sweetwater-black-ridge/internal/service"
)

// --- resolveWorkingDir 测试 ---

func TestResolveWorkingDir_AbsolutePath(t *testing.T) {
	resolved, err := resolveWorkingDir("/home/agent/project", "/home/agent")
	if err != nil {
		t.Fatalf("绝对路径应成功: %v", err)
	}
	expected := filepath.Clean("/home/agent/project")
	if resolved != expected {
		t.Errorf("resolved = %q, 期望 %q", resolved, expected)
	}
}

func TestResolveWorkingDir_RelativePath(t *testing.T) {
	resolved, err := resolveWorkingDir("my-project", "/home/agent")
	if err != nil {
		t.Fatalf("相对路径应成功: %v", err)
	}
	expected := filepath.Join("/home/agent", "my-project")
	if resolved != expected {
		t.Errorf("resolved = %q, 期望 %q", resolved, expected)
	}
}

func TestResolveWorkingDir_Empty(t *testing.T) {
	_, err := resolveWorkingDir("", "/home/agent")
	if err == nil {
		t.Fatal("空路径应返回错误")
	}
}

func TestResolveWorkingDir_WhitespaceOnly(t *testing.T) {
	_, err := resolveWorkingDir("   ", "/home/agent")
	if err == nil {
		t.Fatal("纯空格路径应返回错误")
	}
}

func TestResolveWorkingDir_ParentTraversal(t *testing.T) {
	_, err := resolveWorkingDir("../../etc", "/home/agent")
	if err == nil {
		t.Fatal("路径遍历应被拒绝")
	}
}

func TestResolveWorkingDir_DotDot(t *testing.T) {
	_, err := resolveWorkingDir("..", "/home/agent")
	if err == nil {
		t.Fatal(".. 应被拒绝")
	}
}

func TestResolveWorkingDir_Dot(t *testing.T) {
	_, err := resolveWorkingDir(".", "/home/agent")
	if err == nil {
		t.Fatal(". 应被拒绝（等于 workspace root）")
	}
}

func TestResolveWorkingDir_EqualsWorkspaceRoot(t *testing.T) {
	_, err := resolveWorkingDir("/home/agent", "/home/agent")
	if err == nil {
		t.Fatal("workspace root 本身应被拒绝")
	}
}

func TestResolveWorkingDir_SubdirectoryAllowed(t *testing.T) {
	resolved, err := resolveWorkingDir("sub/dir", "/home/agent")
	if err != nil {
		t.Fatalf("子目录应成功: %v", err)
	}
	expected := filepath.Join("/home/agent", "sub/dir")
	if resolved != expected {
		t.Errorf("resolved = %q, 期望 %q", resolved, expected)
	}
}

// --- validateSessionConfs 测试 ---

func TestValidateSessionConfs_Empty(t *testing.T) {
	err := validateSessionConfs(nil, nil)
	if err != nil {
		t.Fatalf("空配置应通过: %v", err)
	}
}

func TestValidateSessionConfs_NoTemplate(t *testing.T) {
	configs := []model.ConfigItem{{Type: "env", Key: "FOO", Value: "bar"}}
	err := validateSessionConfs(configs, nil)
	if err == nil {
		t.Fatal("有配置但无模板应返回错误")
	}
}

func TestValidateSessionConfs_ValidEnv(t *testing.T) {
	tpl := &model.SessionTemplate{
		SessionSchema: model.SessionSchema{
			EnvDefs: []model.EnvDef{{Key: "API_KEY"}},
		},
	}
	configs := []model.ConfigItem{{Type: "env", Key: "API_KEY", Value: "secret"}}
	err := validateSessionConfs(configs, tpl)
	if err != nil {
		t.Fatalf("合法 env key 应通过: %v", err)
	}
}

func TestValidateSessionConfs_InvalidEnv(t *testing.T) {
	tpl := &model.SessionTemplate{
		SessionSchema: model.SessionSchema{
			EnvDefs: []model.EnvDef{{Key: "API_KEY"}},
		},
	}
	configs := []model.ConfigItem{{Type: "env", Key: "UNAUTHORIZED", Value: "val"}}
	err := validateSessionConfs(configs, tpl)
	if err == nil {
		t.Fatal("未声明 env key 应被拒绝")
	}
}

func TestValidateSessionConfs_ValidFile(t *testing.T) {
	tpl := &model.SessionTemplate{
		SessionSchema: model.SessionSchema{
			FileDefs: []model.FileDef{{Path: "CLAUDE.md"}},
		},
	}
	configs := []model.ConfigItem{{Type: "file", Key: "CLAUDE.md", Value: "# Instructions"}}
	err := validateSessionConfs(configs, tpl)
	if err != nil {
		t.Fatalf("合法 file path 应通过: %v", err)
	}
}

func TestValidateSessionConfs_InvalidFile(t *testing.T) {
	tpl := &model.SessionTemplate{
		SessionSchema: model.SessionSchema{
			FileDefs: []model.FileDef{{Path: "CLAUDE.md"}},
		},
	}
	configs := []model.ConfigItem{{Type: "file", Key: "/etc/passwd", Value: "hacked"}}
	err := validateSessionConfs(configs, tpl)
	if err == nil {
		t.Fatal("绝对路径文件应被拒绝")
	}
}

func TestValidateSessionConfs_UnsupportedType(t *testing.T) {
	tpl := &model.SessionTemplate{}
	configs := []model.ConfigItem{{Type: "unknown", Key: "K", Value: "V"}}
	err := validateSessionConfs(configs, tpl)
	if err == nil {
		t.Fatal("未知类型应被拒绝")
	}
}

func TestValidateSessionConfs_FilePathTraversal(t *testing.T) {
	tpl := &model.SessionTemplate{
		SessionSchema: model.SessionSchema{
			FileDefs: []model.FileDef{{Path: "CLAUDE.md"}},
		},
	}
	configs := []model.ConfigItem{{Type: "file", Key: "../secret", Value: "data"}}
	err := validateSessionConfs(configs, tpl)
	if err == nil {
		t.Fatal("路径遍历文件应被拒绝")
	}
}

// --- gRPC 方法测试（使用 mock TmuxService）---

type mockTmuxService struct {
	sessions     []model.Session
	sessionState *model.SessionState
	created      *model.Session
	killed       string
	workspace    string
	stateDeleted string
	saveAllErr   error
	createErr    error
}

func (m *mockTmuxService) ListSessions() ([]model.Session, error) {
	return m.sessions, nil
}

func (m *mockTmuxService) CreateSession(name, command, workingDir string, configs []model.ConfigItem, restoreStrategy, templateID, restoreCommand string) (*model.Session, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	m.created = &model.Session{ID: name, Name: name, Status: "running", CreatedAt: "2026-01-01T00:00:00Z"}
	return m.created, nil
}

func (m *mockTmuxService) KillSession(name string) error {
	m.killed = name
	return nil
}

func (m *mockTmuxService) GetSession(name string) (*model.Session, error) {
	for _, s := range m.sessions {
		if s.ID == name {
			return &s, nil
		}
	}
	return nil, service.ErrSessionNotFound
}

func (m *mockTmuxService) CapturePane(name string, lines int) (string, error) {
	return "output", nil
}

func (m *mockTmuxService) SendKeys(name, command string) error {
	return nil
}

func (m *mockTmuxService) SendSignal(name, signal string) error {
	return nil
}

func (m *mockTmuxService) AttachSession(name string, rows, cols uint16) (*os.File, error) {
	return nil, nil
}

func (m *mockTmuxService) ResizeSession(name string, rows, cols uint16) error {
	return nil
}

func (m *mockTmuxService) GetSessionEnv(name string) (map[string]string, error) {
	return map[string]string{"PATH": "/usr/bin"}, nil
}

func (m *mockTmuxService) ExecutePipeline(sessionName string, pipeline model.Pipeline) error {
	return nil
}

func (m *mockTmuxService) BuildPipeline(workingDir, command string, configs []model.ConfigItem) model.Pipeline {
	return nil
}

func (m *mockTmuxService) SavePipelineState(sessionName string, pipeline model.Pipeline, restoreStrategy, templateID, cliSessionID, restoreCommand string) error {
	return nil
}

func (m *mockTmuxService) SaveAllPipelineStates() error {
	return m.saveAllErr
}

func (m *mockTmuxService) GetSavedSessions() ([]model.SessionState, error) {
	return nil, nil
}

func (m *mockTmuxService) GetSessionState(sessionName string) (*model.SessionState, error) {
	if m.sessionState != nil {
		return m.sessionState, nil
	}
	return nil, os.ErrNotExist
}

func (m *mockTmuxService) RestoreSession(sessionName string) error {
	return nil
}

func (m *mockTmuxService) DeleteSessionWorkspace(sessionName, workspaceRoot string) error {
	m.workspace = sessionName
	return nil
}

func (m *mockTmuxService) DeleteSessionState(sessionName string) error {
	m.stateDeleted = sessionName
	return nil
}

func newTestServer(t *testing.T, mock *mockTmuxService) *Server {
	t.Helper()
	templateStore := model.NewTemplateStore(
		filepath.Join(t.TempDir(), "templates.json"),
		logutil.NewNop(),
	)
	return NewServer(
		mock,
		service.NewLocalConfigStore(t.TempDir(), logutil.NewNop()),
		templateStore,
		"/home/agent",
		logutil.NewNop(),
	)
}

func TestListSessions(t *testing.T) {
	mock := &mockTmuxService{
		sessions: []model.Session{
			{ID: "s1", Name: "session-1", Status: "running"},
		},
	}
	srv := newTestServer(t, mock)

	resp, err := srv.ListSessions(context.Background(), &pb.ListSessionsRequest{})
	if err != nil {
		t.Fatalf("ListSessions 失败: %v", err)
	}
	if len(resp.GetSessions()) != 1 {
		t.Fatalf("sessions 数量 = %d, 期望 1", len(resp.GetSessions()))
	}
	if resp.GetSessions()[0].GetId() != "s1" {
		t.Errorf("session id = %q, 期望 %q", resp.GetSessions()[0].GetId(), "s1")
	}
}

func TestCreateSession_Success(t *testing.T) {
	mock := &mockTmuxService{}
	srv := newTestServer(t, mock)

	resp, err := srv.CreateSession(context.Background(), &pb.CreateSessionRequest{
		Name:       "test-session",
		WorkingDir: "project-a",
	})
	if err != nil {
		t.Fatalf("CreateSession 失败: %v", err)
	}
	if resp.GetId() != "test-session" {
		t.Errorf("id = %q, 期望 %q", resp.GetId(), "test-session")
	}
}

func TestCreateSession_MissingName(t *testing.T) {
	mock := &mockTmuxService{}
	srv := newTestServer(t, mock)

	_, err := srv.CreateSession(context.Background(), &pb.CreateSessionRequest{})
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("应返回 gRPC status")
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, 期望 InvalidArgument", st.Code())
	}
}

func TestCreateSession_MissingWorkingDir(t *testing.T) {
	mock := &mockTmuxService{}
	srv := newTestServer(t, mock)

	_, err := srv.CreateSession(context.Background(), &pb.CreateSessionRequest{
		Name: "test",
	})
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("应返回 gRPC status")
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, 期望 InvalidArgument", st.Code())
	}
}

func TestCreateSession_Duplicate(t *testing.T) {
	mock := &mockTmuxService{
		createErr: errors.New("duplicate session: test"),
	}
	srv := newTestServer(t, mock)

	_, err := srv.CreateSession(context.Background(), &pb.CreateSessionRequest{
		Name:       "test",
		WorkingDir: "project",
	})
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("应返回 gRPC status")
	}
	if st.Code() != codes.AlreadyExists {
		t.Errorf("code = %v, 期望 AlreadyExists", st.Code())
	}
}

func TestCreateSession_TemplateNotFound(t *testing.T) {
	mock := &mockTmuxService{}
	srv := newTestServer(t, mock)

	_, err := srv.CreateSession(context.Background(), &pb.CreateSessionRequest{
		Name:       "test",
		WorkingDir: "project",
		TemplateId: "nonexistent",
	})
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("应返回 gRPC status")
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, 期望 InvalidArgument", st.Code())
	}
}

func TestGetSession_Found(t *testing.T) {
	mock := &mockTmuxService{
		sessions: []model.Session{
			{ID: "s1", Name: "session-1", Status: "running"},
		},
	}
	srv := newTestServer(t, mock)

	resp, err := srv.GetSession(context.Background(), &pb.GetSessionRequest{Id: "s1"})
	if err != nil {
		t.Fatalf("GetSession 失败: %v", err)
	}
	if resp.GetId() != "s1" {
		t.Errorf("id = %q, 期望 %q", resp.GetId(), "s1")
	}
}

func TestGetSession_NotFound(t *testing.T) {
	mock := &mockTmuxService{sessions: []model.Session{}}
	srv := newTestServer(t, mock)

	_, err := srv.GetSession(context.Background(), &pb.GetSessionRequest{Id: "missing"})
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("应返回 gRPC status")
	}
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, 期望 NotFound", st.Code())
	}
}

func TestDeleteSession_Success(t *testing.T) {
	mock := &mockTmuxService{}
	srv := newTestServer(t, mock)

	_, err := srv.DeleteSession(context.Background(), &pb.DeleteSessionRequest{Id: "test-session"})
	if err != nil {
		t.Fatalf("DeleteSession 失败: %v", err)
	}
	if mock.killed != "test-session" {
		t.Errorf("killed = %q, 期望 %q", mock.killed, "test-session")
	}
	if mock.workspace != "test-session" {
		t.Errorf("workspace = %q, 期望 %q", mock.workspace, "test-session")
	}
	if mock.stateDeleted != "test-session" {
		t.Errorf("stateDeleted = %q, 期望 %q", mock.stateDeleted, "test-session")
	}
}

func TestDeleteSession_SaveAllFails(t *testing.T) {
	mock := &mockTmuxService{
		saveAllErr: errors.New("disk full"),
	}
	srv := newTestServer(t, mock)

	// 保存失败不应阻断删除流程
	_, err := srv.DeleteSession(context.Background(), &pb.DeleteSessionRequest{Id: "test"})
	if err != nil {
		t.Fatalf("保存失败不应阻断删除: %v", err)
	}
}

func TestGetConfig(t *testing.T) {
	mock := &mockTmuxService{}
	srv := newTestServer(t, mock)

	resp, err := srv.GetConfig(context.Background(), &pb.GetConfigRequest{})
	if err != nil {
		t.Fatalf("GetConfig 失败: %v", err)
	}
	if resp.GetWorkingDir() == "" {
		t.Error("working_dir 不应为空")
	}
}

func TestUpdateConfig_ReadOnlyWorkingDir(t *testing.T) {
	mock := &mockTmuxService{}
	srv := newTestServer(t, mock)

	_, err := srv.UpdateConfig(context.Background(), &pb.UpdateConfigRequest{
		WorkingDir: "/hacked",
	})
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("应返回 gRPC status")
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, 期望 InvalidArgument", st.Code())
	}
}

func TestSendInput_EmptyCommand(t *testing.T) {
	mock := &mockTmuxService{}
	srv := newTestServer(t, mock)

	_, err := srv.SendInput(context.Background(), &pb.SendInputRequest{
		Id:      "s1",
		Command: "",
	})
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("应返回 gRPC status")
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, 期望 InvalidArgument", st.Code())
	}
}

func TestSendSignal_EmptySignal(t *testing.T) {
	mock := &mockTmuxService{}
	srv := newTestServer(t, mock)

	_, err := srv.SendSignal(context.Background(), &pb.SendSignalRequest{
		Id:     "s1",
		Signal: "",
	})
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("应返回 gRPC status")
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, 期望 InvalidArgument", st.Code())
	}
}
