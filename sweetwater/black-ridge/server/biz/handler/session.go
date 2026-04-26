package handler

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/charviki/maze-cradle/httputil"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/sweetwater-black-ridge/biz/config"
	"github.com/charviki/sweetwater-black-ridge/biz/model"
	"github.com/charviki/sweetwater-black-ridge/biz/service"
)

// SessionHandler Session 管理 handler，封装 TmuxService 调用
type SessionHandler struct {
	tmuxService   service.TmuxService
	templateStore TemplateStore
	configFiles   *service.ConfigFileService
	cfg           *config.Config
	logger        logutil.Logger
}

// TemplateStore 模板查询接口，handler 只需读取能力
type TemplateStore interface {
	Get(id string) *model.SessionTemplate
}

// NewSessionHandler 创建 SessionHandler 实例
func NewSessionHandler(tmuxService service.TmuxService, templateStore TemplateStore, cfg *config.Config, logger logutil.Logger) *SessionHandler {
	return &SessionHandler{
		tmuxService:   tmuxService,
		templateStore: templateStore,
		configFiles:   service.NewConfigFileService(),
		cfg:           cfg,
		logger:        logger,
	}
}

// ListSessions 列出所有 tmux 会话
func (h *SessionHandler) ListSessions(ctx context.Context, c *app.RequestContext) {
	sessions, err := h.tmuxService.ListSessions()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.Success(c, sessions)
}

// CreateSession 创建新的 tmux 会话，基于管线执行模式
func (h *SessionHandler) CreateSession(ctx context.Context, c *app.RequestContext) {
	var req model.CreateSessionRequest
	if err := c.Bind(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		httputil.Error(c, http.StatusBadRequest, "name is required")
		return
	}

	resolvedWorkingDir, err := resolveWorkingDir(req.WorkingDir, h.cfg.Workspace.RootDir)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	req.WorkingDir = resolvedWorkingDir

	// 从模板中获取 RestoreCommand，用于恢复时使用专用的恢复命令（如 --dangerously-skip-permissions）
	var restoreCommand string
	var tpl *model.SessionTemplate
	if req.TemplateID != "" {
		tpl = h.templateStore.Get(req.TemplateID)
		if tpl == nil {
			httputil.Error(c, http.StatusBadRequest, "template not found")
			return
		}
		restoreCommand = tpl.RestoreCommand
	}
	if err := validateSessionConfs(req.SessionConfs, tpl); err != nil {
		httputil.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	session, err := h.tmuxService.CreateSession(req.Name, req.Command, req.WorkingDir, req.SessionConfs, req.RestoreStrategy, req.TemplateID, restoreCommand)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "duplicate session") {
			httputil.Error(c, http.StatusConflict, "session already exists")
			return
		}
		httputil.Error(c, http.StatusInternalServerError, errMsg)
		return
	}
	httputil.Success(c, session)
}

// GetSessionConfig 返回 session 工作目录下固定项目级配置文件的真实快照。
func (h *SessionHandler) GetSessionConfig(ctx context.Context, c *app.RequestContext) {
	sessionID := c.Param("id")
	if sessionID == "" {
		httputil.Error(c, http.StatusBadRequest, "session id is required")
		return
	}

	state, tpl, statusCode, err := h.loadSessionConfigContext(sessionID)
	if err != nil {
		httputil.Error(c, statusCode, err.Error())
		return
	}

	files, err := h.configFiles.ReadProjectFiles(state.WorkingDir, tpl.SessionSchema.FileDefs)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.Success(c, model.SessionConfigView{
		SessionID:  sessionID,
		TemplateID: tpl.ID,
		WorkingDir: state.WorkingDir,
		Scope:      model.ConfigScopeProject,
		Files:      files,
	})
}

// UpdateSessionConfig 保存 session 工作目录下的固定项目级配置文件。
func (h *SessionHandler) UpdateSessionConfig(ctx context.Context, c *app.RequestContext) {
	sessionID := c.Param("id")
	if sessionID == "" {
		httputil.Error(c, http.StatusBadRequest, "session id is required")
		return
	}

	var req model.SaveConfigRequest
	if err := c.Bind(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	state, tpl, statusCode, err := h.loadSessionConfigContext(sessionID)
	if err != nil {
		httputil.Error(c, statusCode, err.Error())
		return
	}

	files, err := h.configFiles.SaveProjectFiles(state.WorkingDir, tpl.SessionSchema.FileDefs, req.Files)
	if err != nil {
		h.handleConfigSaveError(c, err)
		return
	}

	httputil.Success(c, model.SessionConfigView{
		SessionID:  sessionID,
		TemplateID: tpl.ID,
		WorkingDir: state.WorkingDir,
		Scope:      model.ConfigScopeProject,
		Files:      files,
	})
}

// GetSession 获取指定会话详情
func (h *SessionHandler) GetSession(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, http.StatusBadRequest, "id is required")
		return
	}

	session, err := h.tmuxService.GetSession(id)
	if err != nil {
		if errors.Is(err, service.ErrSessionNotFound) {
			httputil.Error(c, http.StatusNotFound, "session not found")
		} else {
			httputil.Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	httputil.Success(c, session)
}

// DeleteSession 终止并删除指定会话，同时清理管线状态文件
func (h *SessionHandler) DeleteSession(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, http.StatusBadRequest, "id is required")
		return
	}

	if err := h.tmuxService.SaveAllPipelineStates(); err != nil {
		h.logger.WithSession(id).WithAction("delete").Errorf("save pipeline states before delete failed: %v", err)
	}

	if err := h.tmuxService.KillSession(id); err != nil {
		// tmux session 不存在时（如 saved session），不中断流程，继续清理状态文件
		if errors.Is(err, service.ErrSessionNotFound) {
			h.logger.WithSession(id).WithAction("delete").Warnf("tmux session not found, proceeding to clean state file")
		} else {
			httputil.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// 删除的是该 session 实际落地的工作目录；若它恰好等于基础根目录，则仅保护并跳过。
	if err := h.tmuxService.DeleteSessionWorkspace(id, h.cfg.Workspace.RootDir); err != nil {
		if errors.Is(err, service.ErrWorkspaceRootProtected) {
			h.logger.WithSession(id).WithAction("delete").Warnf("skip deleting protected workspace root: %v", err)
		} else {
			httputil.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// 清理管线状态文件，失败时记录日志但不影响删除结果
	if err := h.tmuxService.DeleteSessionState(id); err != nil {
		h.logger.WithSession(id).WithAction("delete").Errorf("delete session state failed: %v", err)
	}

	httputil.Success(c, nil)
}

// GetSavedSessions 返回已保存的 session 列表
func (h *SessionHandler) GetSavedSessions(ctx context.Context, c *app.RequestContext) {
	states, err := h.tmuxService.GetSavedSessions()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if states == nil {
		states = []model.SessionState{}
	}
	httputil.Success(c, states)
}

// RestoreSession 触发单个 session 的管线重放恢复
func (h *SessionHandler) RestoreSession(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, http.StatusBadRequest, "session id is required")
		return
	}

	if err := h.tmuxService.RestoreSession(id); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.Success(c, nil)
}

// SaveSessions 触发所有活跃 session 的管线状态保存，返回保存完成时间
func (h *SessionHandler) SaveSessions(ctx context.Context, c *app.RequestContext) {
	if err := h.tmuxService.SaveAllPipelineStates(); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.Success(c, map[string]string{"saved_at": time.Now().Format(time.RFC3339)})
}

func (h *SessionHandler) loadSessionConfigContext(sessionID string) (*model.SessionState, *model.SessionTemplate, int, error) {
	state, err := h.tmuxService.GetSessionState(sessionID)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, http.StatusNotFound, errors.New("session state not found")
		}
		return nil, nil, http.StatusInternalServerError, err
	}
	if state.TemplateID == "" {
		return nil, nil, http.StatusBadRequest, errors.New("session template is required")
	}

	tpl := h.templateStore.Get(state.TemplateID)
	if tpl == nil {
		return nil, nil, http.StatusNotFound, errors.New("template not found")
	}
	return state, tpl, http.StatusOK, nil
}

func (h *SessionHandler) handleConfigSaveError(c *app.RequestContext, err error) {
	var conflictErr *service.ConfigConflictError
	if errors.As(err, &conflictErr) {
		c.JSON(http.StatusConflict, map[string]interface{}{
			"status":    "error",
			"code":      "config_conflict",
			"message":   conflictErr.Error(),
			"conflicts": conflictErr.Conflicts,
		})
		return
	}
	httputil.Error(c, http.StatusInternalServerError, err.Error())
}

func validateSessionConfs(configs []model.ConfigItem, tpl *model.SessionTemplate) error {
	if len(configs) == 0 {
		return nil
	}
	if tpl == nil {
		return errors.New("template_id is required when session_confs are provided")
	}

	allowedEnvKeys := make(map[string]struct{}, len(tpl.SessionSchema.EnvDefs))
	for _, def := range tpl.SessionSchema.EnvDefs {
		allowedEnvKeys[def.Key] = struct{}{}
	}

	allowedFilePaths := make(map[string]struct{}, len(tpl.SessionSchema.FileDefs))
	for _, def := range tpl.SessionSchema.FileDefs {
		allowedFilePaths[filepath.Clean(def.Path)] = struct{}{}
	}

	for i := range configs {
		cfg := &configs[i]
		switch cfg.Type {
		case "env":
			if _, ok := allowedEnvKeys[cfg.Key]; !ok {
				return errors.New("session env key is not allowed by template")
			}
		case "file":
			cleaned := filepath.Clean(strings.TrimSpace(cfg.Key))
			// 这里强制路径规范化并限制为模板声明的相对路径，
			// 避免通过 ./、../ 或绝对路径把 session 保存偷偷扩散到全局文件。
			if filepath.IsAbs(cleaned) || cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
				return errors.New("session file path must stay under working directory")
			}
			if _, ok := allowedFilePaths[cleaned]; !ok {
				return errors.New("session file path is not allowed by template")
			}
			cfg.Key = cleaned
		default:
			return errors.New("unsupported session config type")
		}
	}

	return nil
}

func resolveWorkingDir(rawWorkingDir string, workspaceRoot string) (string, error) {
	trimmed := strings.TrimSpace(rawWorkingDir)
	if trimmed == "" {
		return "", errors.New("working_dir is required")
	}

	root := filepath.Clean(workspaceRoot)
	var resolved string
	if filepath.IsAbs(trimmed) {
		resolved = filepath.Clean(trimmed)
	} else {
		cleanedRelative := filepath.Clean(trimmed)
		// 相对目录必须留在基础根目录内，避免通过 ../ 跳出工作区。
		if cleanedRelative == "." || cleanedRelative == ".." || strings.HasPrefix(cleanedRelative, ".."+string(filepath.Separator)) {
			return "", errors.New("working_dir must stay under workspace root")
		}
		resolved = filepath.Join(root, cleanedRelative)
	}

	// session 的工作目录不能直接落在根目录本身，否则删除 session 时会面临整根目录清理风险。
	if resolved == root {
		return "", errors.New("working_dir cannot be workspace root")
	}
	return resolved, nil
}
