package grpc

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/sweetwater-black-ridge/biz/model"
	"github.com/charviki/sweetwater-black-ridge/biz/service"
)

// --- SessionService ---

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

func (s *Server) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.Session, error) {
	sessionName := req.Name
	if sessionName == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	confs := make([]model.ConfigItem, len(req.SessionConfs))
	for i, c := range req.SessionConfs {
		confs[i] = model.ConfigItem{Type: c.Type, Key: c.Key, Value: c.Value}
	}

	session, err := s.tmuxService.CreateSession(
		sessionName, req.Command, req.WorkingDir,
		confs, req.RestoreStrategy, req.TemplateId, "",
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate session") {
			return nil, status.Error(codes.AlreadyExists, "session already exists")
		}
		return nil, errToStatus(err)
	}
	return modelSessionToProto(session), nil
}

func (s *Server) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.Session, error) {
	session, err := s.tmuxService.GetSession(req.Id)
	if err != nil {
		return nil, errToStatus(err)
	}
	return modelSessionToProto(session), nil
}

func (s *Server) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*emptypb.Empty, error) {
	if err := s.tmuxService.SaveAllPipelineStates(); err != nil {
		s.logger.Warnf("[grpc] save pipeline states before delete %s: %v", req.Id, err)
	}
	if err := s.tmuxService.KillSession(req.Id); err != nil {
		return nil, errToStatus(err)
	}
	if err := s.tmuxService.DeleteSessionState(req.Id); err != nil {
		s.logger.Warnf("[grpc] delete session state %s: %v", req.Id, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) RestoreSession(ctx context.Context, req *pb.RestoreSessionRequest) (*emptypb.Empty, error) {
	if err := s.tmuxService.RestoreSession(req.Id); err != nil {
		return nil, errToStatus(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) SaveSessions(ctx context.Context, req *pb.SaveSessionsRequest) (*pb.SaveSessionsResponse, error) {
	if err := s.tmuxService.SaveAllPipelineStates(); err != nil {
		return nil, errToStatus(err)
	}
	return &pb.SaveSessionsResponse{SavedAt: time.Now().Format(time.RFC3339)}, nil
}

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

func (s *Server) GetSessionConfig(ctx context.Context, req *pb.GetSessionConfigRequest) (*pb.SessionConfigView, error) {
	state, err := s.tmuxService.GetSessionState(req.Id)
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
		SessionId:  req.Id,
		TemplateId: tpl.ID,
		WorkingDir: state.WorkingDir,
		Scope:      string(model.ConfigScopeProject),
		Files:      configSnapshotsToProto(files),
	}, nil
}

func (s *Server) UpdateSessionConfig(ctx context.Context, req *pb.UpdateSessionConfigRequest) (*pb.SessionConfigView, error) {
	state, err := s.tmuxService.GetSessionState(req.Id)
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
		protoConfigUpdatesToModel(req.Files),
	)
	if err != nil {
		return nil, errToStatus(err)
	}
	return &pb.SessionConfigView{
		SessionId:  req.Id,
		TemplateId: tpl.ID,
		WorkingDir: state.WorkingDir,
		Scope:      string(model.ConfigScopeProject),
		Files:      configSnapshotsToProto(files),
	}, nil
}

// --- Terminal ---

func (s *Server) GetOutput(ctx context.Context, req *pb.GetOutputRequest) (*pb.TerminalOutput, error) {
	lines := int(req.Lines)
	if lines <= 0 {
		lines = 100
	}
	output, err := s.tmuxService.CapturePane(req.Id, lines)
	if err != nil {
		return nil, errToStatus(err)
	}
	return &pb.TerminalOutput{
		SessionId: req.Id,
		Lines:     int32(lines),
		Output:    output,
	}, nil
}

func (s *Server) SendInput(ctx context.Context, req *pb.SendInputRequest) (*emptypb.Empty, error) {
	if req.Command == "" {
		return nil, status.Error(codes.InvalidArgument, "command is required")
	}
	if err := s.tmuxService.SendKeys(req.Id, req.Command); err != nil {
		return nil, errToStatus(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) SendSignal(ctx context.Context, req *pb.SendSignalRequest) (*emptypb.Empty, error) {
	if req.Signal == "" {
		return nil, status.Error(codes.InvalidArgument, "signal is required")
	}
	if err := s.tmuxService.SendSignal(req.Id, req.Signal); err != nil {
		return nil, errToStatus(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetEnv(ctx context.Context, req *pb.GetEnvRequest) (*pb.GetEnvResponse, error) {
	env, err := s.tmuxService.GetSessionEnv(req.Id)
	if err != nil {
		return nil, errToStatus(err)
	}
	return &pb.GetEnvResponse{Env: env}, nil
}

// --- ConfigService ---

func (s *Server) GetConfig(ctx context.Context, req *pb.GetConfigRequest) (*pb.LocalAgentConfig, error) {
	cfg := s.localConfig.Get()
	return &pb.LocalAgentConfig{
		WorkingDir: cfg.WorkingDir,
		Env:        cfg.Env,
	}, nil
}

func (s *Server) UpdateConfig(ctx context.Context, req *pb.UpdateConfigRequest) (*pb.LocalAgentConfig, error) {
	current := s.localConfig.Get()
	if req.WorkingDir != "" && req.WorkingDir != current.WorkingDir {
		return nil, status.Error(codes.InvalidArgument, "working_dir is read-only")
	}
	if req.Env != nil {
		if err := s.localConfig.UpdateEnv(req.Env); err != nil {
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
		Id:          sess.ID,
		Name:        sess.Name,
		Status:      sess.Status,
		CreatedAt:   sess.CreatedAt,
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
			Path:     file.Path,
			Content:  file.Content,
			BaseHash: file.BaseHash,
		}
	}
	return updates
}
