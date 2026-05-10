package transport

import (
	"context"
	"math"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
	"github.com/charviki/maze/fabrication/cradle/auth"
	"github.com/charviki/maze/the-mesa/the-forge/internal/service"
)

// KnowledgeGRPCTransport 实现 KnowledgeService gRPC 服务。
type KnowledgeGRPCTransport struct {
	pb.UnimplementedKnowledgeServiceServer

	knowledgeSvc *service.KnowledgeService
	taskSvc      *service.TaskService
}

// NewKnowledgeGRPCTransport 创建 KnowledgeService gRPC transport。
func NewKnowledgeGRPCTransport(knowledgeSvc *service.KnowledgeService, taskSvc *service.TaskService) *KnowledgeGRPCTransport {
	return &KnowledgeGRPCTransport{knowledgeSvc: knowledgeSvc, taskSvc: taskSvc}
}

// RegisterGRPC 注册 KnowledgeService 和 Health 服务到 gRPC server。
func (t *KnowledgeGRPCTransport) RegisterGRPC(server *grpc.Server) {
	pb.RegisterKnowledgeServiceServer(server, t)
}

// --- Archive ---

// CreateArchive implements KnowledgeServiceServer.
func (t *KnowledgeGRPCTransport) CreateArchive(ctx context.Context, req *pb.CreateArchiveRequest) (*pb.Archive, error) {
	author := extractAuthor(ctx)
	archive, err := t.knowledgeSvc.CreateArchive(ctx, req.GetName(), req.GetDescription(), req.GetIcon(), author)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create archive: %v", err)
	}
	return archiveToProto(archive), nil
}

// GetArchive implements KnowledgeServiceServer.
func (t *KnowledgeGRPCTransport) GetArchive(ctx context.Context, req *pb.GetArchiveRequest) (*pb.Archive, error) {
	archive, err := t.knowledgeSvc.GetArchive(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "get archive: %v", err)
	}
	return archiveToProto(archive), nil
}

// ListArchives implements KnowledgeServiceServer.
func (t *KnowledgeGRPCTransport) ListArchives(ctx context.Context, req *pb.ListArchivesRequest) (*pb.ListArchivesResponse, error) {
	archives, err := t.knowledgeSvc.ListArchives(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list archives: %v", err)
	}
	pbArchives := make([]*pb.Archive, len(archives))
	for i := range archives {
		pbArchives[i] = archiveToProto(&archives[i])
	}
	return &pb.ListArchivesResponse{Archives: pbArchives}, nil
}

// UpdateArchive implements KnowledgeServiceServer.
func (t *KnowledgeGRPCTransport) UpdateArchive(ctx context.Context, req *pb.UpdateArchiveRequest) (*pb.Archive, error) {
	archive, err := t.knowledgeSvc.UpdateArchive(ctx, req.GetId(), req.GetName(), req.GetDescription(), req.GetIcon())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update archive: %v", err)
	}
	return archiveToProto(archive), nil
}

// DeleteArchive implements KnowledgeServiceServer.
func (t *KnowledgeGRPCTransport) DeleteArchive(ctx context.Context, req *pb.DeleteArchiveRequest) (*emptypb.Empty, error) {
	if err := t.knowledgeSvc.DeleteArchive(ctx, req.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "delete archive: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// --- Memory ---

// CreateMemory implements KnowledgeServiceServer.
func (t *KnowledgeGRPCTransport) CreateMemory(ctx context.Context, req *pb.CreateMemoryRequest) (*pb.Memory, error) {
	var parentID *string
	if req.GetParentId() != "" {
		pid := req.GetParentId()
		parentID = &pid
	}
	attachments := protoToAttachments(req.GetAttachments())
	author := extractAuthor(ctx)

	memory, err := t.knowledgeSvc.CreateMemory(ctx,
		req.GetArchiveId(), parentID, req.GetKind(), req.GetTitle(),
		req.GetContent(), req.GetType(), req.GetTags(), author,
		req.GetVisibility(), req.GetSharedWith(), attachments,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create memory: %v", err)
	}
	return memoryToProto(memory), nil
}

// GetMemory implements KnowledgeServiceServer.
func (t *KnowledgeGRPCTransport) GetMemory(ctx context.Context, req *pb.GetMemoryRequest) (*pb.ParsedMemory, error) {
	parsed, err := t.knowledgeSvc.GetMemory(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "get memory: %v", err)
	}
	return parsedMemoryToProto(parsed), nil
}

// ListMemories implements KnowledgeServiceServer.
func (t *KnowledgeGRPCTransport) ListMemories(ctx context.Context, req *pb.ListMemoriesRequest) (*pb.ListMemoriesResponse, error) {
	var archiveID *string
	if req.GetArchiveId() != "" {
		aid := req.GetArchiveId()
		archiveID = &aid
	}
	var parentID *string
	if req.GetParentId() != "" {
		pid := req.GetParentId()
		parentID = &pid
	}

	limit := req.GetLimit()
	if limit <= 0 {
		limit = 50
	}

	items, total, err := t.knowledgeSvc.ListMemories(ctx, archiveID, parentID, req.GetKind(), req.GetType(), req.GetVisibility(), req.GetAuthor(), limit, req.GetOffset())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list memories: %v", err)
	}
	pbItems := make([]*pb.Memory, len(items))
	for i := range items {
		pbItems[i] = memoryToProto(&items[i])
	}
	return &pb.ListMemoriesResponse{Items: pbItems, Total: safeInt32(total)}, nil
}

// UpdateMemory implements KnowledgeServiceServer.
func (t *KnowledgeGRPCTransport) UpdateMemory(ctx context.Context, req *pb.UpdateMemoryRequest) (*pb.Memory, error) {
	params := service.UpdateMemoryParams{
		Tags:        req.GetTags(),
		SharedWith:  req.GetSharedWith(),
		Attachments: protoToAttachments(req.GetAttachments()),
	}
	if req.GetTitle() != "" {
		params.Title = strPtr(req.GetTitle())
	}
	if req.GetContent() != "" {
		params.Content = strPtr(req.GetContent())
	}
	if req.GetType() != "" {
		params.Type = strPtr(req.GetType())
	}
	if req.GetSummary() != "" {
		params.Summary = strPtr(req.GetSummary())
	}
	if req.GetVisibility() != "" {
		params.Visibility = strPtr(req.GetVisibility())
	}
	if req.GetParentId() != "" {
		params.ParentID = strPtr(req.GetParentId())
	}
	if req.GetKind() != "" {
		params.Kind = strPtr(req.GetKind())
	}

	memory, err := t.knowledgeSvc.UpdateMemory(ctx, req.GetId(), params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update memory: %v", err)
	}
	return memoryToProto(memory), nil
}

// DeleteMemory implements KnowledgeServiceServer.
func (t *KnowledgeGRPCTransport) DeleteMemory(ctx context.Context, req *pb.DeleteMemoryRequest) (*emptypb.Empty, error) {
	if err := t.knowledgeSvc.DeleteMemory(ctx, req.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "delete memory: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// SearchMemories implements KnowledgeServiceServer.
func (t *KnowledgeGRPCTransport) SearchMemories(ctx context.Context, req *pb.SearchMemoriesRequest) (*pb.SearchMemoriesResponse, error) {
	var archiveID *string
	if req.GetArchiveId() != "" {
		aid := req.GetArchiveId()
		archiveID = &aid
	}
	items, err := t.knowledgeSvc.SearchMemories(ctx, req.GetQ(), archiveID, req.GetVisibility(), req.GetAuthor())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "search memories: %v", err)
	}
	pbItems := make([]*pb.Memory, len(items))
	for i := range items {
		pbItems[i] = memoryToProto(&items[i])
	}
	return &pb.SearchMemoriesResponse{Items: pbItems}, nil
}

// --- Tree & Ancestors ---

// GetMemoryTree implements KnowledgeServiceServer.
func (t *KnowledgeGRPCTransport) GetMemoryTree(ctx context.Context, req *pb.GetMemoryTreeRequest) (*pb.GetMemoryTreeResponse, error) {
	var parentID *string
	if req.GetParentId() != "" {
		pid := req.GetParentId()
		parentID = &pid
	}
	memories, err := t.knowledgeSvc.GetMemoryTree(ctx, req.GetArchiveId(), parentID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get memory tree: %v", err)
	}
	nodes := make([]*pb.MemoryTreeNode, len(memories))
	for i := range memories {
		nodes[i] = &pb.MemoryTreeNode{Memory: memoryToProto(&memories[i])}
	}
	return &pb.GetMemoryTreeResponse{Nodes: nodes}, nil
}

// GetMemoryAncestors implements KnowledgeServiceServer.
func (t *KnowledgeGRPCTransport) GetMemoryAncestors(ctx context.Context, req *pb.GetMemoryAncestorsRequest) (*pb.GetMemoryAncestorsResponse, error) {
	ancestors, err := t.knowledgeSvc.GetMemoryAncestors(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get ancestors: %v", err)
	}
	pbAncestors := make([]*pb.Memory, len(ancestors))
	for i := range ancestors {
		pbAncestors[i] = memoryToProto(&ancestors[i])
	}
	return &pb.GetMemoryAncestorsResponse{Ancestors: pbAncestors}, nil
}

// GetMemoryDirectives implements KnowledgeServiceServer.
func (t *KnowledgeGRPCTransport) GetMemoryDirectives(ctx context.Context, req *pb.GetMemoryDirectivesRequest) (*pb.GetMemoryDirectivesResponse, error) {
	directives, err := t.knowledgeSvc.GetMemoryDirectives(ctx, req.GetId(), t.taskSvc.Repo())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get memory directives: %v", err)
	}
	pbDirectives := make([]*pb.DirectiveRef, len(directives))
	for i := range directives {
		pbDirectives[i] = &pb.DirectiveRef{
			Id:       directives[i].ID,
			Title:    directives[i].Title,
			Status:   directives[i].Status,
			Priority: directives[i].Priority,
		}
	}
	return &pb.GetMemoryDirectivesResponse{Directives: pbDirectives}, nil
}

// --- Neural Links ---

// GetLinks implements KnowledgeServiceServer.
func (t *KnowledgeGRPCTransport) GetLinks(ctx context.Context, req *pb.GetLinksRequest) (*pb.GetLinksResponse, error) {
	links, err := t.knowledgeSvc.GetLinks(ctx, req.GetId(), req.GetDirection(), req.GetRelationType())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get links: %v", err)
	}
	pbLinks := make([]*pb.NeuralLink, len(links))
	for i := range links {
		pbLinks[i] = neuralLinkToProto(&links[i])
	}
	return &pb.GetLinksResponse{Links: pbLinks}, nil
}

// CreateLink implements KnowledgeServiceServer.
func (t *KnowledgeGRPCTransport) CreateLink(ctx context.Context, req *pb.CreateLinkRequest) (*pb.NeuralLink, error) {
	link, err := t.knowledgeSvc.CreateLink(ctx, req.GetId(), req.GetTargetId(), req.GetRelationType())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create link: %v", err)
	}
	return neuralLinkToProto(link), nil
}

// DeleteLink implements KnowledgeServiceServer.
func (t *KnowledgeGRPCTransport) DeleteLink(ctx context.Context, req *pb.DeleteLinkRequest) (*emptypb.Empty, error) {
	if err := t.knowledgeSvc.DeleteLink(ctx, req.GetLinkId()); err != nil {
		return nil, status.Errorf(codes.Internal, "delete link: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// --- Stats ---

// GetStats implements KnowledgeServiceServer.
func (t *KnowledgeGRPCTransport) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	stats, err := t.knowledgeSvc.GetStats(ctx, t.taskSvc.Repo())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get stats: %v", err)
	}

	byStatus := make(map[string]int32, len(stats.DirectivesByStatus))
	for k, v := range stats.DirectivesByStatus {
		byStatus[k] = safeInt32FromInt(v)
	}

	pbRecent := make([]*pb.Memory, len(stats.RecentMemories))
	for i := range stats.RecentMemories {
		pbRecent[i] = memoryToProto(&stats.RecentMemories[i])
	}

	return &pb.GetStatsResponse{
		Stats: &pb.Stats{
			TotalMemories:      safeInt32FromInt(stats.TotalMemories),
			TotalDirectives:    safeInt32FromInt(stats.TotalDirectives),
			DirectivesByStatus: byStatus,
			RecentMemories:     pbRecent,
		},
	}, nil
}

// --- DirectiveGRPCTransport ---

// DirectiveGRPCTransport 实现 DirectiveService gRPC 服务。
type DirectiveGRPCTransport struct {
	pb.UnimplementedDirectiveServiceServer

	taskSvc *service.TaskService
}

// NewDirectiveGRPCTransport 创建 DirectiveService gRPC transport。
func NewDirectiveGRPCTransport(taskSvc *service.TaskService) *DirectiveGRPCTransport {
	return &DirectiveGRPCTransport{taskSvc: taskSvc}
}

// RegisterGRPC 注册 DirectiveService 到 gRPC server。
func (t *DirectiveGRPCTransport) RegisterGRPC(server *grpc.Server) {
	pb.RegisterDirectiveServiceServer(server, t)
}

// CreateDirective implements DirectiveServiceServer.
func (t *DirectiveGRPCTransport) CreateDirective(ctx context.Context, req *pb.CreateDirectiveRequest) (*pb.Directive, error) {
	author := extractAuthor(ctx)
	var archiveID *string
	if req.GetArchiveId() != "" {
		aid := req.GetArchiveId()
		archiveID = &aid
	}
	directive, err := t.taskSvc.CreateDirective(ctx,
		req.GetTitle(), req.GetDescription(), req.GetStatus(), req.GetPriority(),
		req.GetAssignee(), author, req.GetRequireDocIds(), req.GetNarrativeId(),
		archiveID, req.GetVisibility(),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create directive: %v", err)
	}
	return directiveToProto(directive), nil
}

// GetDirective implements DirectiveServiceServer.
func (t *DirectiveGRPCTransport) GetDirective(ctx context.Context, req *pb.GetDirectiveRequest) (*pb.Directive, error) {
	directive, err := t.taskSvc.GetDirective(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "get directive: %v", err)
	}
	return directiveToProto(directive), nil
}

// ListDirectives implements DirectiveServiceServer.
func (t *DirectiveGRPCTransport) ListDirectives(ctx context.Context, req *pb.ListDirectivesRequest) (*pb.ListDirectivesResponse, error) {
	limit := req.GetLimit()
	if limit <= 0 {
		limit = 50
	}
	items, total, err := t.taskSvc.ListDirectives(ctx, req.GetStatus(), req.GetAssignee(), req.GetPriority(), req.GetVisibility(), limit, req.GetOffset())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list directives: %v", err)
	}
	pbItems := make([]*pb.Directive, len(items))
	for i := range items {
		pbItems[i] = directiveToProto(&items[i])
	}
	return &pb.ListDirectivesResponse{Items: pbItems, Total: safeInt32(total)}, nil
}

// UpdateDirective implements DirectiveServiceServer.
func (t *DirectiveGRPCTransport) UpdateDirective(ctx context.Context, req *pb.UpdateDirectiveRequest) (*pb.Directive, error) {
	params := service.UpdateDirectiveParams{
		RequireDocIDs: req.GetRequireDocIds(),
	}
	if req.GetTitle() != "" {
		params.Title = strPtr(req.GetTitle())
	}
	if req.GetDescription() != "" {
		params.Description = strPtr(req.GetDescription())
	}
	if req.GetStatus() != "" {
		params.Status = strPtr(req.GetStatus())
	}
	if req.GetPriority() != "" {
		params.Priority = strPtr(req.GetPriority())
	}
	if req.GetAssignee() != "" {
		params.Assignee = strPtr(req.GetAssignee())
	}
	if req.GetAuthor() != "" {
		params.Author = strPtr(req.GetAuthor())
	}
	if req.GetNarrativeId() != "" {
		params.NarrativeID = strPtr(req.GetNarrativeId())
	}
	if req.GetArchiveId() != "" {
		params.ArchiveID = strPtr(req.GetArchiveId())
	}
	if req.GetVisibility() != "" {
		params.Visibility = strPtr(req.GetVisibility())
	}

	directive, err := t.taskSvc.UpdateDirective(ctx, req.GetId(), params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update directive: %v", err)
	}
	return directiveToProto(directive), nil
}

// DeleteDirective implements DirectiveServiceServer.
func (t *DirectiveGRPCTransport) DeleteDirective(ctx context.Context, req *pb.DeleteDirectiveRequest) (*emptypb.Empty, error) {
	if err := t.taskSvc.DeleteDirective(ctx, req.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "delete directive: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// --- Conversion helpers ---

func archiveToProto(a *service.Archive) *pb.Archive {
	return &pb.Archive{
		Id:          a.ID,
		Name:        a.Name,
		Description: a.Description,
		Icon:        a.Icon,
		Author:      a.Author,
		CreatedAt:   a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   a.UpdatedAt.Format(time.RFC3339),
	}
}

func memoryToProto(m *service.Memory) *pb.Memory {
	return &pb.Memory{
		Id:          m.ID,
		ArchiveId:   m.ArchiveID,
		ParentId:    strValue(m.ParentID),
		Kind:        m.Kind,
		Title:       m.Title,
		Content:     m.Content,
		Type:        m.Type,
		Summary:     m.Summary,
		Tags:        m.Tags,
		Author:      m.Author,
		Visibility:  m.Visibility,
		SharedWith:  m.SharedWith,
		Attachments: attachmentsToProto(m.Attachments),
		CreatedAt:   m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   m.UpdatedAt.Format(time.RFC3339),
	}
}

func parsedMemoryToProto(m *service.ParsedMemory) *pb.ParsedMemory {
	return &pb.ParsedMemory{
		Meta: &pb.MemoryMeta{
			Id:          m.Meta.ID,
			ArchiveId:   m.Meta.ArchiveID,
			ParentId:    strValue(m.Meta.ParentID),
			Kind:        m.Meta.Kind,
			Title:       m.Meta.Title,
			Type:        m.Meta.Type,
			Summary:     m.Meta.Summary,
			Tags:        m.Meta.Tags,
			Author:      m.Meta.Author,
			Visibility:  m.Meta.Visibility,
			SharedWith:  m.Meta.SharedWith,
			Attachments: attachmentsToProto(m.Meta.Attachments),
			CreatedAt:   m.Meta.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   m.Meta.UpdatedAt.Format(time.RFC3339),
		},
		Summary: m.Summary,
		Content: m.Content,
	}
}

func neuralLinkToProto(l *service.NeuralLink) *pb.NeuralLink {
	return &pb.NeuralLink{
		Id:           l.ID,
		SourceId:     l.SourceID,
		TargetId:     l.TargetID,
		RelationType: l.RelationType,
		SourceTitle:  l.SourceTitle,
		TargetTitle:  l.TargetTitle,
		CreatedAt:    l.CreatedAt.Format(time.RFC3339),
	}
}

func directiveToProto(d *service.Directive) *pb.Directive {
	return &pb.Directive{
		Id:            d.ID,
		Title:         d.Title,
		Description:   d.Description,
		Status:        d.Status,
		Priority:      d.Priority,
		Assignee:      d.Assignee,
		Author:        d.Author,
		RequireDocIds: d.RequireDocIDs,
		NarrativeId:   d.NarrativeID,
		ArchiveId:     strValue(d.ArchiveID),
		Visibility:    d.Visibility,
		CreatedAt:     d.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     d.UpdatedAt.Format(time.RFC3339),
	}
}

func attachmentsToProto(attachments []service.Attachment) []*pb.Attachment {
	if attachments == nil {
		return nil
	}
	result := make([]*pb.Attachment, len(attachments))
	for i, a := range attachments {
		result[i] = &pb.Attachment{
			Id:          a.ID,
			Key:         a.Key,
			Name:        a.Name,
			ContentType: a.ContentType,
			Size:        a.Size,
		}
	}
	return result
}

func protoToAttachments(attachments []*pb.Attachment) []service.Attachment {
	if attachments == nil {
		return nil
	}
	result := make([]service.Attachment, len(attachments))
	for i, a := range attachments {
		result[i] = service.Attachment{
			ID:          a.GetId(),
			Key:         a.GetKey(),
			Name:        a.GetName(),
			ContentType: a.GetContentType(),
			Size:        a.GetSize(),
		}
	}
	return result
}

// extractAuthor 从 context 中提取作者信息（优先从 JWT claims，否则返回 "anonymous"）。
func extractAuthor(ctx context.Context) string {
	if user := auth.GetUserInfo(ctx); user != nil && user.SubjectKey != "" {
		return user.SubjectKey
	}
	return "anonymous"
}

func strPtr(s string) *string { return &s }

func strValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func safeInt32(n int64) int32 {
	if n > math.MaxInt32 {
		return math.MaxInt32
	}
	if n < math.MinInt32 {
		return math.MinInt32
	}
	return int32(n)
}

func safeInt32FromInt(n int) int32 {
	if n > math.MaxInt32 {
		return math.MaxInt32
	}
	if n < math.MinInt32 {
		return math.MinInt32
	}
	return int32(n)
}
