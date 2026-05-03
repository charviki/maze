package transport

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/sweetwater-black-ridge/internal/model"
	"github.com/charviki/sweetwater-black-ridge/internal/service"
)

// --- SessionService ---

// ListSessions 返回所有 tmux Session 列表
func (s *Server) ListSessions(ctx context.Context, req *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
	sessions, err := s.tmuxService.ListSessions()
	if err != nil {
		return nil, errToStatus(err)
	}
	pbSessions := make([]*pb.Session, len(sessions))
	for i, sess := range sessions {
		pbSessions[i] = modelSessionToProto(&sess)
	}
	return &pb.ListSessionsResponse{Sessions: pbSessions}, nil
}

// CreateSession 创建新的 tmux Session
func (s *Server) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.Session, error) {
	sessionName := req.GetName()
	if sessionName == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	// working_dir 安全解析：相对路径基于 workspace root 解析为绝对路径
	resolvedWorkingDir, err := resolveWorkingDir(req.GetWorkingDir(), s.workspaceRootDir)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 从模板获取 restoreCommand，用于恢复时使用专用恢复命令
	var restoreCommand string
	var tpl *model.SessionTemplate
	if req.GetTemplateId() != "" {
		tpl = s.templateStore.Get(req.GetTemplateId())
		if tpl == nil {
			return nil, status.Error(codes.InvalidArgument, "template not found")
		}
		restoreCommand = tpl.RestoreCommand
	}

	confs := make([]model.ConfigItem, len(req.GetSessionConfs()))
	for i, c := range req.GetSessionConfs() {
		confs[i] = model.ConfigItem{Type: c.GetType(), Key: c.GetKey(), Value: c.GetValue()}
	}

	// session_confs 校验：确保 env key 和 file path 在模板声明的范围内
	if err := validateSessionConfs(confs, tpl); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	session, err := s.tmuxService.CreateSession(
		sessionName, req.GetCommand(), resolvedWorkingDir,
		confs, req.GetRestoreStrategy(), req.GetTemplateId(), restoreCommand,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate session") {
			return nil, status.Error(codes.AlreadyExists, "session already exists")
		}
		return nil, errToStatus(err)
	}
	return modelSessionToProto(session), nil
}

// GetSession 获取指定 Session 详情
func (s *Server) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.Session, error) {
	session, err := s.tmuxService.GetSession(req.GetId())
	if err != nil {
		return nil, errToStatus(err)
	}
	return modelSessionToProto(session), nil
}

// DeleteSession 删除指定 Session，同时清理工作目录
func (s *Server) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*emptypb.Empty, error) {
	if err := s.tmuxService.SaveAllPipelineStates(); err != nil {
		s.logger.Warnf("[grpc] save pipeline states before delete %s: %v", req.GetId(), err)
	}
	if err := s.tmuxService.KillSession(req.GetId()); err != nil {
		// tmux session 不存在时（如 saved session），不中断流程
		if !errors.Is(err, service.ErrSessionNotFound) {
			return nil, errToStatus(err)
		}
		s.logger.Warnf("[grpc] tmux session %s not found, proceeding to clean state", req.GetId())
	}
	// 清理 session 工作目录，保护 workspace root 不被误删
	if err := s.tmuxService.DeleteSessionWorkspace(req.GetId(), s.workspaceRootDir); err != nil {
		if errors.Is(err, service.ErrWorkspaceRootProtected) {
			s.logger.Warnf("[grpc] skip deleting protected workspace root: %v", err)
		} else {
			s.logger.Warnf("[grpc] delete session workspace %s: %v", req.GetId(), err)
		}
	}
	if err := s.tmuxService.DeleteSessionState(req.GetId()); err != nil {
		s.logger.Warnf("[grpc] delete session state %s: %v", req.GetId(), err)
	}
	return &emptypb.Empty{}, nil
}

// RestoreSession 恢复已终止的 Session
func (s *Server) RestoreSession(ctx context.Context, req *pb.RestoreSessionRequest) (*emptypb.Empty, error) {
	if err := s.tmuxService.RestoreSession(req.GetId()); err != nil {
		return nil, errToStatus(err)
	}
	return &emptypb.Empty{}, nil
}

// SaveSessions 保存所有 Session 管线状态
func (s *Server) SaveSessions(ctx context.Context, req *pb.SaveSessionsRequest) (*pb.SaveSessionsResponse, error) {
	if err := s.tmuxService.SaveAllPipelineStates(); err != nil {
		return nil, errToStatus(err)
	}
	return &pb.SaveSessionsResponse{SavedAt: time.Now().Format(time.RFC3339)}, nil
}

// GetSavedSessions 获取已保存的 Session 列表
func (s *Server) GetSavedSessions(ctx context.Context, req *pb.GetSavedSessionsRequest) (*pb.GetSavedSessionsResponse, error) {
	states, err := s.tmuxService.GetSavedSessions()
	if err != nil {
		return nil, errToStatus(err)
	}
	pbStates := make([]*pb.SessionState, len(states))
	for i, st := range states {
		pipeline, _ := st.ToJSON()
		pbStates[i] = &pb.SessionState{
			SessionName:      st.SessionName,
			Pipeline:         pipeline,
			RestoreStrategy:  st.RestoreStrategy,
			RestoreCommand:   st.RestoreCommand,
			WorkingDir:       st.WorkingDir,
			TemplateId:       st.TemplateID,
			CliSessionId:     st.CLISessionID,
			EnvSnapshot:      st.EnvSnapshot,
			TerminalSnapshot: st.TerminalSnapshot,
			SavedAt:          st.SavedAt,
		}
	}
	return &pb.GetSavedSessionsResponse{Sessions: pbStates}, nil
}

// GetSessionConfig 获取 Session 项目配置
func (s *Server) GetSessionConfig(ctx context.Context, req *pb.GetSessionConfigRequest) (*pb.SessionConfigView, error) {
	state, err := s.tmuxService.GetSessionState(req.GetId())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, status.Error(codes.NotFound, "session state not found")
		}
		return nil, errToStatus(err)
	}
	if state.TemplateID == "" {
		return nil, status.Error(codes.InvalidArgument, "session template is required")
	}
	tpl := s.templateStore.Get(state.TemplateID)
	if tpl == nil {
		return nil, status.Error(codes.NotFound, "template not found")
	}

	files, err := service.NewConfigFileService().ReadProjectFiles(state.WorkingDir, tpl.SessionSchema.FileDefs)
	if err != nil {
		return nil, errToStatus(err)
	}
	return &pb.SessionConfigView{
		SessionId:  req.GetId(),
		TemplateId: tpl.ID,
		WorkingDir: state.WorkingDir,
		Scope:      string(model.ConfigScopeProject),
		Files:      configSnapshotsToProto(files),
	}, nil
}

// UpdateSessionConfig 更新 Session 项目配置
func (s *Server) UpdateSessionConfig(ctx context.Context, req *pb.UpdateSessionConfigRequest) (*pb.SessionConfigView, error) {
	state, err := s.tmuxService.GetSessionState(req.GetId())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, status.Error(codes.NotFound, "session state not found")
		}
		return nil, errToStatus(err)
	}
	if state.TemplateID == "" {
		return nil, status.Error(codes.InvalidArgument, "session template is required")
	}
	tpl := s.templateStore.Get(state.TemplateID)
	if tpl == nil {
		return nil, status.Error(codes.NotFound, "template not found")
	}

	files, err := service.NewConfigFileService().SaveProjectFiles(
		state.WorkingDir,
		tpl.SessionSchema.FileDefs,
		protoConfigUpdatesToModel(req.GetFiles()),
	)
	if err != nil {
		return nil, configConflictToStatus(err)
	}
	return &pb.SessionConfigView{
		SessionId:  req.GetId(),
		TemplateId: tpl.ID,
		WorkingDir: state.WorkingDir,
		Scope:      string(model.ConfigScopeProject),
		Files:      configSnapshotsToProto(files),
	}, nil
}

// --- Terminal ---

// GetOutput 获取终端输出
func (s *Server) GetOutput(ctx context.Context, req *pb.GetOutputRequest) (*pb.TerminalOutput, error) {
	lines := int(req.GetLines())
	if lines <= 0 {
		lines = 100
	}
	output, err := s.tmuxService.CapturePane(req.GetId(), lines)
	if err != nil {
		return nil, errToStatus(err)
	}
	return &pb.TerminalOutput{
		SessionId: req.GetId(),
		Lines:     int32(lines),
		Output:    output,
	}, nil
}

// SendInput 发送终端输入
func (s *Server) SendInput(ctx context.Context, req *pb.SendInputRequest) (*emptypb.Empty, error) {
	if req.GetCommand() == "" {
		return nil, status.Error(codes.InvalidArgument, "command is required")
	}
	if err := s.tmuxService.SendKeys(req.GetId(), req.GetCommand()); err != nil {
		return nil, errToStatus(err)
	}
	return &emptypb.Empty{}, nil
}

// SendSignal 发送终端信号
func (s *Server) SendSignal(ctx context.Context, req *pb.SendSignalRequest) (*emptypb.Empty, error) {
	if req.GetSignal() == "" {
		return nil, status.Error(codes.InvalidArgument, "signal is required")
	}
	if err := s.tmuxService.SendSignal(req.GetId(), req.GetSignal()); err != nil {
		return nil, errToStatus(err)
	}
	return &emptypb.Empty{}, nil
}

// GetEnv 获取 Session 环境变量
func (s *Server) GetEnv(ctx context.Context, req *pb.GetEnvRequest) (*pb.GetEnvResponse, error) {
	env, err := s.tmuxService.GetSessionEnv(req.GetId())
	if err != nil {
		return nil, errToStatus(err)
	}
	return &pb.GetEnvResponse{Env: env}, nil
}

// --- ConfigService ---

// GetConfig 获取 Agent 本地配置
func (s *Server) GetConfig(ctx context.Context, req *pb.GetConfigRequest) (*pb.LocalAgentConfig, error) {
	cfg := s.localConfig.Get()
	return &pb.LocalAgentConfig{
		WorkingDir: cfg.WorkingDir,
		Env:        cfg.Env,
	}, nil
}

// UpdateConfig 更新 Agent 本地配置
func (s *Server) UpdateConfig(ctx context.Context, req *pb.UpdateConfigRequest) (*pb.LocalAgentConfig, error) {
	current := s.localConfig.Get()
	if req.GetWorkingDir() != "" && req.GetWorkingDir() != current.WorkingDir {
		return nil, status.Error(codes.InvalidArgument, "working_dir is read-only")
	}
	if req.Env != nil {
		if err := s.localConfig.UpdateEnv(req.GetEnv()); err != nil {
			return nil, errToStatus(err)
		}
	}
	cfg := s.localConfig.Get()
	return &pb.LocalAgentConfig{
		WorkingDir: cfg.WorkingDir,
		Env:        cfg.Env,
	}, nil
}

// modelSessionToProto 将 model.Session 转换为 protobuf Session
func modelSessionToProto(sess *model.Session) *pb.Session {
	if sess == nil {
		return nil
	}
	return &pb.Session{
		Id:        sess.ID,
		Name:      sess.Name,
		Status:    sess.Status,
		CreatedAt: sess.CreatedAt,
		//nolint:gosec
		WindowCount: int32(sess.WindowCount),
	}
}

func configSnapshotsToProto(files []model.ConfigFileSnapshot) []*pb.ConfigFileSnapshot {
	pbFiles := make([]*pb.ConfigFileSnapshot, len(files))
	for i, file := range files {
		pbFiles[i] = &pb.ConfigFileSnapshot{
			Path:    file.Path,
			Content: file.Content,
			Exists:  file.Exists,
			Hash:    file.Hash,
		}
	}
	return pbFiles
}

func protoConfigUpdatesToModel(files []*pb.ConfigFileUpdate) []model.ConfigFileUpdate {
	updates := make([]model.ConfigFileUpdate, len(files))
	for i, file := range files {
		updates[i] = model.ConfigFileUpdate{
			Path:     file.GetPath(),
			Content:  file.GetContent(),
			BaseHash: file.GetBaseHash(),
		}
	}
	return updates
}

// resolveWorkingDir 解析 working_dir：相对路径基于 workspace root 转为绝对路径，
// 防止路径遍历（../）和 workspace root 本身被占用。
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
		// 相对目录必须留在基础根目录内，避免通过 ../ 跳出工作区
		if cleanedRelative == "." || cleanedRelative == ".." || strings.HasPrefix(cleanedRelative, ".."+string(filepath.Separator)) {
			return "", errors.New("working_dir must stay under workspace root")
		}
		resolved = filepath.Join(root, cleanedRelative)
	}

	// session 工作目录不能是 workspace root 本身，否则删除时有整根清理风险
	if resolved == root {
		return "", errors.New("working_dir cannot be workspace root")
	}
	return resolved, nil
}

// validateSessionConfs 校验 session 配置项：env key 和 file path 必须在模板声明范围内
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
			// 强制路径规范化并限制为模板声明的相对路径
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

// configConflictToStatus 将 ConfigConflictError 转为 gRPC status，携带冲突详情 JSON
func configConflictToStatus(err error) error {
	var confErr *service.ConfigConflictError
	if errors.As(err, &confErr) {
		detail, _ := json.Marshal(confErr.Conflicts)
		return status.Error(codes.FailedPrecondition, string(detail))
	}
	return errToStatus(err)
}
