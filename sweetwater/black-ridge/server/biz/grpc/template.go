package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/sweetwater-black-ridge/biz/model"
	"github.com/charviki/sweetwater-black-ridge/biz/service"
)

// --- TemplateService ---

func (s *Server) ListTemplates(ctx context.Context, req *pb.ListTemplatesRequest) (*pb.ListTemplatesResponse, error) {
	templates := s.templateStore.List()
	pbTemplates := make([]*pb.SessionTemplate, len(templates))
	for i, t := range templates {
		pbTemplates[i] = modelTemplateToProto(t)
	}
	return &pb.ListTemplatesResponse{Templates: pbTemplates}, nil
}

func (s *Server) CreateTemplate(ctx context.Context, req *pb.CreateTemplateRequest) (*pb.SessionTemplate, error) {
	if req.Template == nil {
		return nil, status.Error(codes.InvalidArgument, "template is required")
	}
	tpl := protoToModelTemplate(req.Template)
	if tpl.ID == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if tpl.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if s.templateStore.Get(tpl.ID) != nil {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("template %s already exists", tpl.ID))
	}
	tpl.Builtin = false
	if err := s.templateStore.Set(&tpl); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	saved := s.templateStore.Get(tpl.ID)
	return modelTemplateToProto(saved), nil
}

func (s *Server) GetTemplate(ctx context.Context, req *pb.GetTemplateRequest) (*pb.SessionTemplate, error) {
	tpl := s.templateStore.Get(req.Id)
	if tpl == nil {
		return nil, status.Error(codes.NotFound, "template not found")
	}
	return modelTemplateToProto(tpl), nil
}

func (s *Server) UpdateTemplate(ctx context.Context, req *pb.UpdateTemplateRequest) (*pb.SessionTemplate, error) {
	existing := s.templateStore.Get(req.Id)
	if existing == nil {
		return nil, status.Error(codes.NotFound, "template not found")
	}
	updated := protoToModelTemplate(req.Template)
	updated.ID = req.Id
	updated.Builtin = existing.Builtin
	if err := s.templateStore.Set(&updated); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	saved := s.templateStore.Get(req.Id)
	return modelTemplateToProto(saved), nil
}

func (s *Server) DeleteTemplate(ctx context.Context, req *pb.DeleteTemplateRequest) (*emptypb.Empty, error) {
	existing := s.templateStore.Get(req.Id)
	if existing == nil {
		return nil, status.Error(codes.NotFound, "template not found")
	}
	if existing.Builtin {
		return nil, status.Error(codes.PermissionDenied, "cannot delete built-in template")
	}
	if err := s.templateStore.Delete(req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetTemplateConfig(ctx context.Context, req *pb.GetTemplateConfigRequest) (*pb.TemplateConfigView, error) {
	tpl := s.templateStore.Get(req.Id)
	if tpl == nil {
		return nil, status.Error(codes.NotFound, "template not found")
	}

	files, err := service.NewConfigFileService().ReadGlobalFiles(tpl.Defaults.Files)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.TemplateConfigView{
		TemplateId: tpl.ID,
		Scope:      string(model.ConfigScopeGlobal),
		Files:      configSnapshotsToProto(files),
	}, nil
}

func (s *Server) UpdateTemplateConfig(ctx context.Context, req *pb.UpdateTemplateConfigRequest) (*pb.TemplateConfigView, error) {
	existing := s.templateStore.Get(req.Id)
	if existing == nil {
		return nil, status.Error(codes.NotFound, "template not found")
	}

	updates := protoConfigUpdatesToModel(req.Files)
	files, err := service.NewConfigFileService().SaveGlobalFiles(existing.Defaults.Files, updates)
	if err != nil {
		return nil, errToStatus(err)
	}
	updated := cloneTemplateWithGlobalContents(existing, updates)
	if err := s.templateStore.Set(updated); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.TemplateConfigView{
		TemplateId: updated.ID,
		Scope:      string(model.ConfigScopeGlobal),
		Files:      configSnapshotsToProto(files),
	}, nil
}

func modelTemplateToProto(t *model.SessionTemplate) *pb.SessionTemplate {
	if t == nil {
		return nil
	}
	return &pb.SessionTemplate{
		Id:                 t.ID,
		Name:               t.Name,
		Command:            t.Command,
		RestoreCommand:     t.RestoreCommand,
		SessionFilePattern: t.SessionFilePattern,
		Description:        t.Description,
		Icon:               t.Icon,
		Builtin:            t.Builtin,
		Defaults:           configLayerToProto(t.Defaults),
		SessionSchema:      sessionSchemaToProto(t.SessionSchema),
	}
}

func protoToModelTemplate(t *pb.SessionTemplate) model.SessionTemplate {
	if t == nil {
		return model.SessionTemplate{}
	}
	return model.SessionTemplate{
		ID:                 t.Id,
		Name:               t.Name,
		Command:            t.Command,
		RestoreCommand:     t.RestoreCommand,
		SessionFilePattern: t.SessionFilePattern,
		Description:        t.Description,
		Icon:               t.Icon,
		Builtin:            t.Builtin,
		Defaults:           protoToConfigLayer(t.Defaults),
		SessionSchema:      protoToSessionSchema(t.SessionSchema),
	}
}

func configLayerToProto(layer model.ConfigLayer) *pb.ConfigLayer {
	files := make([]*pb.ConfigFile, len(layer.Files))
	for i, file := range layer.Files {
		files[i] = &pb.ConfigFile{Path: file.Path, Content: file.Content}
	}
	return &pb.ConfigLayer{Env: layer.Env, Files: files}
}

func protoToConfigLayer(layer *pb.ConfigLayer) model.ConfigLayer {
	if layer == nil {
		return model.ConfigLayer{}
	}
	files := make([]model.ConfigFile, len(layer.Files))
	for i, file := range layer.Files {
		files[i] = model.ConfigFile{Path: file.Path, Content: file.Content}
	}
	return model.ConfigLayer{Env: layer.Env, Files: files}
}

func sessionSchemaToProto(schema model.SessionSchema) *pb.SessionSchema {
	envDefs := make([]*pb.EnvDef, len(schema.EnvDefs))
	for i, envDef := range schema.EnvDefs {
		envDefs[i] = &pb.EnvDef{
			Key:         envDef.Key,
			Label:       envDef.Label,
			Required:    envDef.Required,
			Placeholder: envDef.Placeholder,
			Sensitive:   envDef.Sensitive,
		}
	}
	fileDefs := make([]*pb.FileDef, len(schema.FileDefs))
	for i, fileDef := range schema.FileDefs {
		fileDefs[i] = &pb.FileDef{
			Path:           fileDef.Path,
			Label:          fileDef.Label,
			Required:       fileDef.Required,
			DefaultContent: fileDef.DefaultContent,
		}
	}
	return &pb.SessionSchema{EnvDefs: envDefs, FileDefs: fileDefs}
}

func protoToSessionSchema(schema *pb.SessionSchema) model.SessionSchema {
	if schema == nil {
		return model.SessionSchema{}
	}
	envDefs := make([]model.EnvDef, len(schema.EnvDefs))
	for i, envDef := range schema.EnvDefs {
		envDefs[i] = model.EnvDef{
			Key:         envDef.Key,
			Label:       envDef.Label,
			Required:    envDef.Required,
			Placeholder: envDef.Placeholder,
			Sensitive:   envDef.Sensitive,
		}
	}
	fileDefs := make([]model.FileDef, len(schema.FileDefs))
	for i, fileDef := range schema.FileDefs {
		fileDefs[i] = model.FileDef{
			Path:           fileDef.Path,
			Label:          fileDef.Label,
			Required:       fileDef.Required,
			DefaultContent: fileDef.DefaultContent,
		}
	}
	return model.SessionSchema{EnvDefs: envDefs, FileDefs: fileDefs}
}

func cloneTemplateWithGlobalContents(existing *model.SessionTemplate, updates []model.ConfigFileUpdate) *model.SessionTemplate {
	cloned := *existing
	cloned.Defaults.Env = cloneStringMap(existing.Defaults.Env)
	cloned.Defaults.Files = cloneConfigFiles(existing.Defaults.Files)
	cloned.SessionSchema.EnvDefs = append([]model.EnvDef(nil), existing.SessionSchema.EnvDefs...)
	cloned.SessionSchema.FileDefs = append([]model.FileDef(nil), existing.SessionSchema.FileDefs...)

	contentByPath := make(map[string]string, len(updates))
	for _, update := range updates {
		contentByPath[update.Path] = update.Content
	}

	// gRPC 侧与旧 HTTP 一致，只同步固定路径文件内容，不允许顺手改路径集合。
	for i := range cloned.Defaults.Files {
		if content, ok := contentByPath[cloned.Defaults.Files[i].Path]; ok {
			cloned.Defaults.Files[i].Content = content
		}
	}
	return &cloned
}

func cloneStringMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func cloneConfigFiles(src []model.ConfigFile) []model.ConfigFile {
	if src == nil {
		return nil
	}
	dst := make([]model.ConfigFile, len(src))
	copy(dst, src)
	return dst
}
