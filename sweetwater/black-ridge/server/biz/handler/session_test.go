package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	hertzconfig "github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/route"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/sweetwater-black-ridge/biz/config"
	"github.com/charviki/sweetwater-black-ridge/biz/model"
	"github.com/charviki/sweetwater-black-ridge/biz/service"
)

type mockTmuxService struct {
	sessions      map[string]*model.Session
	savedStates   []model.SessionState
	sessionStates map[string]*model.SessionState
	pipelineState map[string]model.Pipeline
	saveAllErr    error
}

func newHandlerMock() *mockTmuxService {
	return &mockTmuxService{
		sessions:      make(map[string]*model.Session),
		sessionStates: make(map[string]*model.SessionState),
		pipelineState: make(map[string]model.Pipeline),
	}
}

func (m *mockTmuxService) ListSessions() ([]model.Session, error) {
	var result []model.Session
	for _, s := range m.sessions {
		result = append(result, *s)
	}
	return result, nil
}

func (m *mockTmuxService) CreateSession(name string, command string, workingDir string, configs []model.ConfigItem, restoreStrategy string, templateID string, restoreCommand string) (*model.Session, error) {
	if _, exists := m.sessions[name]; exists {
		return nil, fmt.Errorf("duplicate session")
	}
	s := &model.Session{ID: name, Name: name, Status: "running"}
	m.sessions[name] = s
	m.pipelineState[name] = model.Pipeline{
		{ID: "sys-cd", Type: model.StepCD, Phase: model.PhaseSystem, Order: 0, Key: workingDir},
	}
	m.sessionStates[name] = &model.SessionState{
		SessionName: name,
		WorkingDir:  workingDir,
		TemplateID:  templateID,
	}
	return s, nil
}

func (m *mockTmuxService) KillSession(name string) error {
	if _, ok := m.sessions[name]; !ok {
		return fmt.Errorf("%w: %s", service.ErrSessionNotFound, name)
	}
	delete(m.sessions, name)
	return nil
}

func (m *mockTmuxService) GetSession(name string) (*model.Session, error) {
	if s, ok := m.sessions[name]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("%w: %s", service.ErrSessionNotFound, name)
}

func (m *mockTmuxService) CapturePane(name string, lines int) (string, error) {
	return "mock output", nil
}

func (m *mockTmuxService) SendKeys(name string, command string) error { return nil }

func (m *mockTmuxService) SendSignal(name string, signal string) error { return nil }

func (m *mockTmuxService) AttachSession(name string, rows, cols uint16) (*os.File, error) {
	return nil, nil
}

func (m *mockTmuxService) ResizeSession(name string, rows, cols uint16) error { return nil }

func (m *mockTmuxService) GetSessionEnv(name string) (map[string]string, error) {
	return map[string]string{"PATH": "/usr/bin"}, nil
}

func (m *mockTmuxService) ExecutePipeline(sessionName string, pipeline model.Pipeline) error {
	return nil
}

func (m *mockTmuxService) BuildPipeline(workingDir string, command string, configs []model.ConfigItem) model.Pipeline {
	return model.Pipeline{
		{ID: "sys-cd", Type: model.StepCD, Phase: model.PhaseSystem, Order: 0, Key: workingDir},
	}
}

func (m *mockTmuxService) SavePipelineState(sessionName string, pipeline model.Pipeline, restoreStrategy string, templateID string, cliSessionID string, restoreCommand string) error {
	m.pipelineState[sessionName] = pipeline
	return nil
}

func (m *mockTmuxService) SaveAllPipelineStates() error { return m.saveAllErr }

func (m *mockTmuxService) GetSavedSessions() ([]model.SessionState, error) {
	return m.savedStates, nil
}

func (m *mockTmuxService) GetSessionState(sessionName string) (*model.SessionState, error) {
	if state, ok := m.sessionStates[sessionName]; ok {
		return state, nil
	}
	return nil, os.ErrNotExist
}

func (m *mockTmuxService) RestoreSession(sessionName string) error {
	m.sessions[sessionName] = &model.Session{ID: sessionName, Name: sessionName, Status: "running"}
	return nil
}

func (m *mockTmuxService) DeleteSessionWorkspace(sessionName string, workspaceRoot string) error {
	delete(m.sessionStates, sessionName)
	return nil
}

func (m *mockTmuxService) DeleteSessionState(sessionName string) error {
	delete(m.sessionStates, sessionName)
	return nil
}

type mockTemplateStore struct {
	templates map[string]*model.SessionTemplate
}

func (m *mockTemplateStore) Get(id string) *model.SessionTemplate {
	return m.templates[id]
}

func setupRouter(mock *mockTmuxService) *route.Engine {
	return setupRouterWithTemplates(mock, make(map[string]*model.SessionTemplate))
}

func setupRouterWithTemplates(mock *mockTmuxService, templates map[string]*model.SessionTemplate) *route.Engine {
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{RootDir: "/home/agent"},
	}
	h := NewSessionHandler(mock, &mockTemplateStore{templates: templates}, cfg, logutil.NewNop())

	r := route.NewEngine(hertzconfig.NewOptions(nil))
	r.POST("/api/v1/sessions", h.CreateSession)
	r.GET("/api/v1/sessions", h.ListSessions)
	r.GET("/api/v1/sessions/:id", h.GetSession)
	r.DELETE("/api/v1/sessions/:id", h.DeleteSession)
	r.GET("/api/v1/sessions/:id/config", h.GetSessionConfig)
	r.PUT("/api/v1/sessions/:id/config", h.UpdateSessionConfig)
	r.GET("/api/v1/sessions/saved", h.GetSavedSessions)
	r.POST("/api/v1/sessions/:id/restore", h.RestoreSession)
	r.POST("/api/v1/sessions/save", h.SaveSessions)
	return r
}

func performPost(r *route.Engine, path string, body string) *ut.ResponseRecorder {
	return ut.PerformRequest(r, http.MethodPost, path, &ut.Body{Body: strings.NewReader(body), Len: len(body)},
		ut.Header{Key: "Content-Type", Value: "application/json"})
}

func parseResponse(body []byte) map[string]interface{} {
	var resp map[string]interface{}
	json.Unmarshal(body, &resp)
	return resp
}

func TestCreateSession_Success(t *testing.T) {
	mock := newHandlerMock()
	r := setupRouter(mock)

	w := performPost(r, "/api/v1/sessions", `{"name":"test-session","command":"claude","working_dir":"/home/agent/test-session"}`)

	assert.DeepEqual(t, http.StatusOK, w.Code)
	resp := parseResponse(w.Body.Bytes())
	assert.DeepEqual(t, "ok", resp["status"])
}

func TestCreateSession_EmptyName(t *testing.T) {
	mock := newHandlerMock()
	r := setupRouter(mock)

	w := performPost(r, "/api/v1/sessions", `{"name":"","command":"claude"}`)

	assert.DeepEqual(t, http.StatusBadRequest, w.Code)
}

func TestCreateSession_Duplicate(t *testing.T) {
	mock := newHandlerMock()
	mock.sessions["existing"] = &model.Session{ID: "existing", Name: "existing"}
	r := setupRouter(mock)

	w := performPost(r, "/api/v1/sessions", `{"name":"existing","command":"bash","working_dir":"/home/agent/existing"}`)

	assert.DeepEqual(t, http.StatusConflict, w.Code)
}

func TestGetSession_Found(t *testing.T) {
	mock := newHandlerMock()
	mock.sessions["my-session"] = &model.Session{ID: "my-session", Name: "my-session", Status: "running"}
	r := setupRouter(mock)

	w := ut.PerformRequest(r, http.MethodGet, "/api/v1/sessions/my-session", nil)

	assert.DeepEqual(t, http.StatusOK, w.Code)
}

func TestGetSession_NotFound(t *testing.T) {
	mock := newHandlerMock()
	r := setupRouter(mock)

	w := ut.PerformRequest(r, http.MethodGet, "/api/v1/sessions/nonexistent", nil)

	assert.DeepEqual(t, http.StatusNotFound, w.Code)
}

func TestListSessions(t *testing.T) {
	mock := newHandlerMock()
	mock.sessions["s1"] = &model.Session{ID: "s1", Name: "s1", Status: "running"}
	mock.sessions["s2"] = &model.Session{ID: "s2", Name: "s2", Status: "running"}
	r := setupRouter(mock)

	w := ut.PerformRequest(r, http.MethodGet, "/api/v1/sessions", nil)

	assert.DeepEqual(t, http.StatusOK, w.Code)
	resp := parseResponse(w.Body.Bytes())
	assert.DeepEqual(t, "ok", resp["status"])
}

func TestDeleteSession_Success(t *testing.T) {
	mock := newHandlerMock()
	mock.sessions["to-delete"] = &model.Session{ID: "to-delete", Name: "to-delete"}
	r := setupRouter(mock)

	w := ut.PerformRequest(r, http.MethodDelete, "/api/v1/sessions/to-delete", nil)

	assert.DeepEqual(t, http.StatusOK, w.Code)

	if _, exists := mock.sessions["to-delete"]; exists {
		t.Error("期望 session 已被删除")
	}
}

func TestDeleteSession_NotFound(t *testing.T) {
	mock := newHandlerMock()
	r := setupRouter(mock)

	// tmux session 不存在时（如 saved session），仍应成功清理状态文件
	w := ut.PerformRequest(r, http.MethodDelete, "/api/v1/sessions/nonexistent", nil)

	assert.DeepEqual(t, http.StatusOK, w.Code)
}

func TestGetSavedSessions(t *testing.T) {
	mock := newHandlerMock()
	mock.savedStates = []model.SessionState{
		{SessionName: "saved-1", RestoreStrategy: "manual", SavedAt: "2025-01-01T00:00:00Z"},
	}
	r := setupRouter(mock)

	w := ut.PerformRequest(r, http.MethodGet, "/api/v1/sessions/saved", nil)

	assert.DeepEqual(t, http.StatusOK, w.Code)
	resp := parseResponse(w.Body.Bytes())
	assert.DeepEqual(t, "ok", resp["status"])
}

func TestRestoreSession_Success(t *testing.T) {
	mock := newHandlerMock()
	r := setupRouter(mock)

	w := ut.PerformRequest(r, http.MethodPost, "/api/v1/sessions/saved-1/restore", nil)

	assert.DeepEqual(t, http.StatusOK, w.Code)

	if _, ok := mock.sessions["saved-1"]; !ok {
		t.Error("期望 session 已被恢复")
	}
}

func TestSaveSessions(t *testing.T) {
	mock := newHandlerMock()
	r := setupRouter(mock)

	w := ut.PerformRequest(r, http.MethodPost, "/api/v1/sessions/save", nil)

	assert.DeepEqual(t, http.StatusOK, w.Code)
}

// 验证 SaveSessions 返回 saved_at 时间戳
func TestSaveSessions_ReturnsTimestamp(t *testing.T) {
	mock := newHandlerMock()
	r := setupRouter(mock)

	w := ut.PerformRequest(r, http.MethodPost, "/api/v1/sessions/save", nil)

	assert.DeepEqual(t, http.StatusOK, w.Code)
	resp := parseResponse(w.Body.Bytes())
	assert.DeepEqual(t, "ok", resp["status"])

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("响应中缺少 data 字段或 data 不是对象")
	}
	savedAt, ok := data["saved_at"].(string)
	if !ok || savedAt == "" {
		t.Errorf("data.saved_at 应为非空字符串, 实际: %v", data["saved_at"])
	}
}

// 验证 SaveSessions 保存失败时返回 500
func TestSaveSessions_Error(t *testing.T) {
	mock := newHandlerMock()
	mock.saveAllErr = fmt.Errorf("tmux not available")
	r := setupRouter(mock)

	w := ut.PerformRequest(r, http.MethodPost, "/api/v1/sessions/save", nil)

	assert.DeepEqual(t, http.StatusInternalServerError, w.Code)
}

func TestCreateSession_InvalidBody(t *testing.T) {
	mock := newHandlerMock()
	r := setupRouter(mock)

	w := performPost(r, "/api/v1/sessions", "not json")

	assert.DeepEqual(t, http.StatusBadRequest, w.Code)
}

func TestCreateSession_RejectsSessionConfigWithoutTemplate(t *testing.T) {
	mock := newHandlerMock()
	r := setupRouter(mock)

	w := performPost(r, "/api/v1/sessions", `{"name":"test-session","command":"claude","working_dir":"/tmp/work","session_confs":[{"type":"file","key":".claude/settings.json","value":"{}"}]}`)

	assert.DeepEqual(t, http.StatusBadRequest, w.Code)
}

func TestGetSessionConfig_ReadsWorkingDirFile(t *testing.T) {
	mock := newHandlerMock()
	workingDir := t.TempDir()
	configDir := filepath.Join(workingDir, ".claude")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("创建配置目录失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "settings.json"), []byte("{\"theme\":\"dark\"}"), 0644); err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	mock.sessionStates["sess-1"] = &model.SessionState{
		SessionName: "sess-1",
		WorkingDir:  workingDir,
		TemplateID:  "claude",
	}
	r := setupRouterWithTemplates(mock, map[string]*model.SessionTemplate{
		"claude": {
			ID: "claude",
			SessionSchema: model.SessionSchema{
				FileDefs: []model.FileDef{{Path: ".claude/settings.json"}},
			},
		},
	})

	w := ut.PerformRequest(r, http.MethodGet, "/api/v1/sessions/sess-1/config", nil)

	assert.DeepEqual(t, http.StatusOK, w.Code)
	resp := parseResponse(w.Body.Bytes())
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("期望 data 为对象")
	}
	files, ok := data["files"].([]interface{})
	if !ok || len(files) != 1 {
		t.Fatalf("期望 files 长度为 1, 实际 %#v", data["files"])
	}
	fileObj, ok := files[0].(map[string]interface{})
	if !ok {
		t.Fatal("期望 files[0] 为对象")
	}
	assert.DeepEqual(t, ".claude/settings.json", fileObj["path"])
	assert.DeepEqual(t, "{\"theme\":\"dark\"}", fileObj["content"])
}

func TestUpdateSessionConfig_ReturnsConflictPayload(t *testing.T) {
	mock := newHandlerMock()
	workingDir := t.TempDir()
	configPath := filepath.Join(workingDir, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("创建配置目录失败: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("{\"version\":1}"), 0644); err != nil {
		t.Fatalf("写入初始配置失败: %v", err)
	}

	mock.sessionStates["sess-1"] = &model.SessionState{
		SessionName: "sess-1",
		WorkingDir:  workingDir,
		TemplateID:  "claude",
	}
	r := setupRouterWithTemplates(mock, map[string]*model.SessionTemplate{
		"claude": {
			ID: "claude",
			SessionSchema: model.SessionSchema{
				FileDefs: []model.FileDef{{Path: ".claude/settings.json"}},
			},
		},
	})

	getResp := ut.PerformRequest(r, http.MethodGet, "/api/v1/sessions/sess-1/config", nil)
	if getResp.Code != http.StatusOK {
		t.Fatalf("读取 session config 失败: %d", getResp.Code)
	}
	payload := parseResponse(getResp.Body.Bytes())
	data := payload["data"].(map[string]interface{})
	files := data["files"].([]interface{})
	fileObj := files[0].(map[string]interface{})
	baseHash := fileObj["hash"].(string)

	if err := os.WriteFile(configPath, []byte("{\"version\":2}"), 0644); err != nil {
		t.Fatalf("模拟外部改写失败: %v", err)
	}

	body := fmt.Sprintf(`{"files":[{"path":".claude/settings.json","content":"{\"version\":3}","base_hash":"%s"}]}`, baseHash)
	w := performPostMethod(r, http.MethodPut, "/api/v1/sessions/sess-1/config", body)

	assert.DeepEqual(t, http.StatusConflict, w.Code)
	resp := parseResponse(w.Body.Bytes())
	assert.DeepEqual(t, "config_conflict", resp["code"])
	assert.DeepEqual(t, "配置已变更，请重新加载后再修改", resp["message"])
}

func performPostMethod(r *route.Engine, method string, path string, body string) *ut.ResponseRecorder {
	return ut.PerformRequest(r, method, path, &ut.Body{Body: strings.NewReader(body), Len: len(body)},
		ut.Header{Key: "Content-Type", Value: "application/json"})
}
