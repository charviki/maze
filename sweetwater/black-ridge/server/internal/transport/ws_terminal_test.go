package transport

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/pipeline"
	"github.com/charviki/sweetwater-black-ridge/internal/service"
)

type testTmuxService struct{}

func (s *testTmuxService) ListSessions() ([]service.Session, error)         { return nil, nil }
func (s *testTmuxService) GetSession(name string) (*service.Session, error) { return nil, nil }
func (s *testTmuxService) CreateSession(opts service.CreateSessionOptions) (*service.Session, error) {
	return nil, nil
}
func (s *testTmuxService) KillSession(name string) error    { return nil }
func (s *testTmuxService) RestoreSession(name string) error { return nil }
func (s *testTmuxService) DeleteSessionWorkspace(sessionName string, workspaceRoot string) error {
	return nil
}
func (s *testTmuxService) DeleteSessionState(sessionName string) error { return nil }
func (s *testTmuxService) BuildPipeline(workingDir string, command string, configs []service.ConfigItem) pipeline.Pipeline {
	return nil
}
func (s *testTmuxService) ExecutePipeline(sessionName string, pipeline pipeline.Pipeline) error {
	return nil
}
func (s *testTmuxService) CapturePane(name string, lines int) (string, error) { return "", nil }
func (s *testTmuxService) SendKeys(name string, command string) error         { return nil }
func (s *testTmuxService) SendSignal(name string, signal string) error        { return nil }
func (s *testTmuxService) AttachSession(name string, rows, cols uint16) (*os.File, error) {
	return nil, nil
}
func (s *testTmuxService) ResizeSession(name string, rows, cols uint16) error            { return nil }
func (s *testTmuxService) SavePipelineState(opts service.SavePipelineStateOptions) error { return nil }
func (s *testTmuxService) SaveAllPipelineStates() error                                  { return nil }
func (s *testTmuxService) GetSavedSessions() ([]service.SessionState, error)             { return nil, nil }
func (s *testTmuxService) GetSessionState(sessionName string) (*service.SessionState, error) {
	return nil, nil
}
func (s *testTmuxService) GetSessionEnv(name string) (map[string]string, error) { return nil, nil }

func TestNewTerminalHandler(t *testing.T) {
	h := NewTerminalHandler(&testTmuxService{}, 40, logutil.NewNop(), nil)
	if h == nil {
		t.Fatal("NewTerminalHandler returned nil")
	}
	if h.defaultLines != 40 {
		t.Errorf("defaultLines = %d, want 40", h.defaultLines)
	}
	if h.tmuxService == nil {
		t.Error("tmuxService should not be nil")
	}
}

func TestTerminalHandler_HandleWs_MissingID(t *testing.T) {
	h := NewTerminalHandler(&testTmuxService{}, 24, logutil.NewNop(), nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/sessions/ws", nil)

	h.HandleWs(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestTerminalHandler_DefaultValues(t *testing.T) {
	h := NewTerminalHandler(&testTmuxService{}, 0, logutil.NewNop(), nil)
	if h == nil {
		t.Fatal("zero defaultLines should be allowed")
	}
}

func TestTerminalHandler_AllowedOrigins(t *testing.T) {
	h := NewTerminalHandler(&testTmuxService{}, 24, logutil.NewNop(), []string{"http://myapp.com"})
	if len(h.allowedOrigins) != 1 {
		t.Errorf("allowedOrigins length = %d, want 1", len(h.allowedOrigins))
	}
}
