package transport

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/configutil"
	
	"github.com/charviki/sweetwater-black-ridge/internal/service"
)

// --- TemplateService ---

// ListTemplates 返回所有模板列表
func (s *Server) ListTemplates(ctx context.Context, req *pb.ListTemplatesRequest) (*pb.ListTemplatesResponse, error) {
	templates := s.templateStore.List()
	pbTemplates := make([]*pb.SessionTemplate, len(templates))
	for i, t := range templates {
		pbTemplates[i] = modelTemplateToProto(t)
	}
	return &pb.ListTemplatesResponse{Templates: pbTemplates}, nil
}

// CreateTemplate 创建新模板
func (s *Server) CreateTemplate(ctx context.Context, req *pb.CreateTemplateRequest) (*pb.SessionTemplate, error) {
	if req.GetTemplate() == nil {
		return nil, status.Error(codes.InvalidArgument, "template is required")
	}
	tpl := protoToModelTemplate(req.GetTemplate())
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

// GetTemplate 获取指定模板
func (s *Server) GetTemplate(ctx context.Context, req *pb.GetTemplateRequest) (*pb.SessionTemplate, error) {
	tpl := s.templateStore.Get(req.GetId())
	if tpl == nil {
		return nil, status.Error(codes.NotFound, "template not found")
	}
	return modelTemplateToProto(tpl), nil
}

// UpdateTemplate 更新模板
func (s *Server) UpdateTemplate(ctx context.Context, req *pb.UpdateTemplateRequest) (*pb.SessionTemplate, error) {
	existing := s.templateStore.Get(req.GetId())
	if existing == nil {
		return nil, status.Error(codes.NotFound, "template not found")
	}
	updated := protoToModelTemplate(req.GetTemplate())
	updated.ID = req.GetId()
	updated.Builtin = existing.Builtin
	if err := s.templateStore.Set(&updated); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	saved := s.templateStore.Get(req.GetId())
	return modelTemplateToProto(saved), nil
}

// DeleteTemplate 删除模板（禁止删除内置模板）
func (s *Server) DeleteTemplate(ctx context.Context, req *pb.DeleteTemplateRequest) (*emptypb.Empty, error) {
	existing := s.templateStore.Get(req.GetId())
	if existing == nil {
		return nil, status.Error(codes.NotFound, "template not found")
	}
	if existing.Builtin {
		return nil, status.Error(codes.PermissionDenied, "cannot delete built-in template")
	}
	if err := s.templateStore.Delete(req.GetId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

// GetTemplateConfig 获取模板全局配置
func (s *Server) GetTemplateConfig(ctx context.Context, req *pb.GetTemplateConfigRequest) (*pb.TemplateConfigView, error) {
	tpl := s.templateStore.Get(req.GetId())
	if tpl == nil {
		return nil, status.Error(codes.NotFound, "template not found")
	}

	files, err := s.configFiles.ReadGlobalFiles(tpl.Defaults.Files)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.TemplateConfigView{
		TemplateId: tpl.ID,
		Scope:      string(service.ConfigScopeGlobal),
		Files:      configSnapshotsToProto(files),
	}, nil
}

// UpdateTemplateConfig 更新模板全局配置
func (s *Server) UpdateTemplateConfig(ctx context.Context, req *pb.UpdateTemplateConfigRequest) (*pb.TemplateConfigView, error) {
	existing := s.templateStore.Get(req.GetId())
	if existing == nil {
		return nil, status.Error(codes.NotFound, "template not found")
	}

	updates := protoConfigUpdatesToModel(req.GetFiles())
	files, err := s.configFiles.SaveGlobalFiles(existing.Defaults.Files, updates)
	if err != nil {
		return nil, errToStatus(err)
	}
	updated := cloneTemplateWithGlobalContents(existing, updates)
	if err := s.templateStore.Set(updated); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.TemplateConfigView{
		TemplateId: updated.ID,
		Scope:      string(service.ConfigScopeGlobal),
		Files:      configSnapshotsToProto(files),
	}, nil
}

func modelTemplateToProto(t *service.SessionTemplate) *pb.SessionTemplate {
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

func protoToModelTemplate(t *pb.SessionTemplate) service.SessionTemplate {
	if t == nil {
		return service.SessionTemplate{}
	}
	return service.SessionTemplate{
		ID:                 t.GetId(),
		Name:               t.GetName(),
		Command:            t.GetCommand(),
		RestoreCommand:     t.GetRestoreCommand(),
		SessionFilePattern: t.GetSessionFilePattern(),
		Description:        t.GetDescription(),
		Icon:               t.GetIcon(),
		Builtin:            t.GetBuiltin(),
		Defaults:           protoToConfigLayer(t.GetDefaults()),
		SessionSchema:      protoToSessionSchema(t.GetSessionSchema()),
	}
}

func configLayerToProto(layer configutil.ConfigLayer) *pb.ConfigLayer {
	files := make([]*pb.ConfigFile, len(layer.Files))
	for i, file := range layer.Files {
		files[i] = &pb.ConfigFile{Path: file.Path, Content: file.Content}
	}
	return &pb.ConfigLayer{Env: layer.Env, Files: files}
}

func protoToConfigLayer(layer *pb.ConfigLayer) configutil.ConfigLayer {
	if layer == nil {
		return configutil.ConfigLayer{}
	}
	files := make([]configutil.ConfigFile, len(layer.GetFiles()))
	for i, file := range layer.GetFiles() {
		files[i] = configutil.ConfigFile{Path: file.GetPath(), Content: file.GetContent()}
	}
	return configutil.ConfigLayer{Env: layer.GetEnv(), Files: files}
}

func sessionSchemaToProto(schema configutil.SessionSchema) *pb.SessionSchema {
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

func protoToSessionSchema(schema *pb.SessionSchema) configutil.SessionSchema {
	if schema == nil {
		return configutil.SessionSchema{}
	}
	envDefs := make([]configutil.EnvDef, len(schema.GetEnvDefs()))
	for i, envDef := range schema.GetEnvDefs() {
		envDefs[i] = configutil.EnvDef{
			Key:         envDef.GetKey(),
			Label:       envDef.GetLabel(),
			Required:    envDef.GetRequired(),
			Placeholder: envDef.GetPlaceholder(),
			Sensitive:   envDef.GetSensitive(),
		}
	}
	fileDefs := make([]configutil.FileDef, len(schema.GetFileDefs()))
	for i, fileDef := range schema.GetFileDefs() {
		fileDefs[i] = configutil.FileDef{
			Path:           fileDef.GetPath(),
			Label:          fileDef.GetLabel(),
			Required:       fileDef.GetRequired(),
			DefaultContent: fileDef.GetDefaultContent(),
		}
	}
	return configutil.SessionSchema{EnvDefs: envDefs, FileDefs: fileDefs}
}

func cloneTemplateWithGlobalContents(existing *service.SessionTemplate, updates []service.ConfigFileUpdate) *service.SessionTemplate {
	cloned := *existing
	cloned.Defaults.Env = cloneStringMap(existing.Defaults.Env)
	cloned.Defaults.Files = cloneConfigFiles(existing.Defaults.Files)
	cloned.SessionSchema.EnvDefs = append([]configutil.EnvDef(nil), existing.SessionSchema.EnvDefs...)
	cloned.SessionSchema.FileDefs = append([]configutil.FileDef(nil), existing.SessionSchema.FileDefs...)

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

func cloneConfigFiles(src []configutil.ConfigFile) []configutil.ConfigFile {
	if src == nil {
		return nil
	}
	dst := make([]configutil.ConfigFile, len(src))
	copy(dst, src)
	return dst
}
