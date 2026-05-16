package transport

import (
	"context"
	"encoding/json"
	"errors"
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

const defaultPageSize int32 = 50

// Server 实现 KnowledgeService gRPC 服务。
type Server struct {
	pb.UnimplementedKnowledgeServiceServer

	docSvc *service.DocService
}

// NewServer 创建 KnowledgeService gRPC transport。
func NewServer(docSvc *service.DocService) *Server {
	return &Server{docSvc: docSvc}
}

// RegisterGRPC 注册 KnowledgeService 到 gRPC server。
func (t *Server) RegisterGRPC(server *grpc.Server) {
	pb.RegisterKnowledgeServiceServer(server, t)
}

// --- Archive ---

// CreateArchive implements KnowledgeServiceServer.
func (t *Server) CreateArchive(ctx context.Context, req *pb.CreateArchiveRequest) (*pb.Archive, error) {
	author := extractAuthor(ctx)
	archive, err := t.docSvc.CreateArchive(ctx, req.GetName(), req.GetDescription(), req.GetIcon(), author)
	if err != nil {
		return nil, toStatusError(err)
	}
	return archiveToProto(archive), nil
}

// GetArchive implements KnowledgeServiceServer.
func (t *Server) GetArchive(ctx context.Context, req *pb.GetArchiveRequest) (*pb.Archive, error) {
	archive, err := t.docSvc.GetArchive(ctx, req.GetId())
	if err != nil {
		return nil, toStatusError(err)
	}
	return archiveToProto(archive), nil
}

// ListArchives implements KnowledgeServiceServer.
func (t *Server) ListArchives(ctx context.Context, req *pb.ListArchivesRequest) (*pb.ListArchivesResponse, error) {
	archives, err := t.docSvc.ListArchives(ctx)
	if err != nil {
		return nil, toStatusError(err)
	}
	pbArchives := make([]*pb.Archive, len(archives))
	for i := range archives {
		pbArchives[i] = archiveToProto(&archives[i])
	}
	return &pb.ListArchivesResponse{Archives: pbArchives}, nil
}

// UpdateArchive implements KnowledgeServiceServer.
func (t *Server) UpdateArchive(ctx context.Context, req *pb.UpdateArchiveRequest) (*pb.Archive, error) {
	archive, err := t.docSvc.UpdateArchive(ctx, req.GetId(), req.GetName(), req.GetDescription(), req.GetIcon())
	if err != nil {
		return nil, toStatusError(err)
	}
	return archiveToProto(archive), nil
}

// DeleteArchive implements KnowledgeServiceServer.
func (t *Server) DeleteArchive(ctx context.Context, req *pb.DeleteArchiveRequest) (*emptypb.Empty, error) {
	if err := t.docSvc.DeleteArchive(ctx, req.GetId()); err != nil {
		return nil, toStatusError(err)
	}
	return &emptypb.Empty{}, nil
}

// --- Doc ---

// CreateDoc implements KnowledgeServiceServer.
func (t *Server) CreateDoc(ctx context.Context, req *pb.CreateDocRequest) (*pb.Doc, error) {
	var parentID *string
	if req.GetParentId() != "" {
		pid := req.GetParentId()
		parentID = &pid
	}
	author := extractAuthor(ctx)

	doc, err := t.docSvc.CreateDoc(ctx, service.CreateDocParams{
		ArchiveID:   req.GetArchiveId(),
		ParentID:    parentID,
		Title:       req.GetTitle(),
		Content:     req.GetContent(),
		Summary:     req.GetSummary(),
		Status:      strPtrOrNil(req.GetStatus()),
		Priority:    strPtrOrNil(req.GetPriority()),
		Assignee:    req.GetAssignee(),
		Tags:        req.GetTags(),
		Author:      author,
		Visibility:  req.GetVisibility(),
		SharedWith:  req.GetSharedWith(),
		Attachments: protoToAttachments(req.GetAttachments()),
	})
	if err != nil {
		return nil, toStatusError(err)
	}
	return docToProto(doc), nil
}

// GetDoc implements KnowledgeServiceServer.
func (t *Server) GetDoc(ctx context.Context, req *pb.GetDocRequest) (*pb.Doc, error) {
	doc, err := t.docSvc.GetDoc(ctx, req.GetId())
	if err != nil {
		return nil, toStatusError(err)
	}
	return docToProto(doc), nil
}

// ListDocs implements KnowledgeServiceServer.
func (t *Server) ListDocs(ctx context.Context, req *pb.ListDocsRequest) (*pb.ListDocsResponse, error) {
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
		limit = defaultPageSize
	}

	items, total, err := t.docSvc.ListDocs(ctx, archiveID, parentID, req.GetStatus(), req.GetStatus() != "", req.GetVisibility(), req.GetAuthor(), limit, req.GetOffset())
	if err != nil {
		return nil, toStatusError(err)
	}
	pbItems := make([]*pb.Doc, len(items))
	for i := range items {
		pbItems[i] = docToProto(&items[i])
	}
	return &pb.ListDocsResponse{Items: pbItems, Total: safeInt32(total)}, nil
}

// UpdateDoc implements KnowledgeServiceServer.
func (t *Server) UpdateDoc(ctx context.Context, req *pb.UpdateDocRequest) (*pb.Doc, error) {
	// Only serialize slice fields when explicitly provided (non-empty),
	// so that COALESCE preserves existing values when not set.
	params := service.UpdateDocParams{
		ClearStatus:   false,
		ClearPriority: false,
	}
	if tags := req.GetTags(); len(tags) > 0 {
		tagsJSON, _ := json.Marshal(tags)
		params.Tags = tagsJSON
	}
	if shared := req.GetSharedWith(); len(shared) > 0 {
		sharedJSON, _ := json.Marshal(shared)
		params.SharedWith = sharedJSON
	}
	if attachments := protoToAttachments(req.GetAttachments()); len(attachments) > 0 {
		attachJSON, _ := json.Marshal(attachments)
		params.Attachments = attachJSON
	}
	if req.GetTitle() != "" {
		params.Title = strPtr(req.GetTitle())
	}
	if req.GetContent() != "" {
		params.Content = strPtr(req.GetContent())
	}
	if req.GetSummary() != "" {
		params.Summary = strPtr(req.GetSummary())
	}
	// Status: proto3 can't distinguish empty from unset.
	// We use ClearStatus when client explicitly wants to clear.
	// For now: non-empty sets, empty does nothing (backward compatible).
	if req.GetStatus() != "" {
		params.Status = strPtr(req.GetStatus())
	}
	if req.GetPriority() != "" {
		params.Priority = strPtr(req.GetPriority())
	}
	if req.GetAssignee() != "" {
		params.Assignee = strPtr(req.GetAssignee())
	}
	if req.GetVisibility() != "" {
		params.Visibility = strPtr(req.GetVisibility())
	}
	if req.GetParentId() != "" {
		params.ParentID = strPtr(req.GetParentId())
	}

	doc, err := t.docSvc.UpdateDoc(ctx, req.GetId(), params)
	if err != nil {
		return nil, toStatusError(err)
	}
	return docToProto(doc), nil
}

// DeleteDoc implements KnowledgeServiceServer.
func (t *Server) DeleteDoc(ctx context.Context, req *pb.DeleteDocRequest) (*emptypb.Empty, error) {
	if err := t.docSvc.DeleteDoc(ctx, req.GetId()); err != nil {
		return nil, toStatusError(err)
	}
	return &emptypb.Empty{}, nil
}

// SearchDocs implements KnowledgeServiceServer.
func (t *Server) SearchDocs(ctx context.Context, req *pb.SearchDocsRequest) (*pb.SearchDocsResponse, error) {
	var archiveID *string
	if req.GetArchiveId() != "" {
		aid := req.GetArchiveId()
		archiveID = &aid
	}
	items, err := t.docSvc.SearchDocs(ctx, req.GetQ(), archiveID, req.GetVisibility(), req.GetAuthor())
	if err != nil {
		return nil, toStatusError(err)
	}
	pbItems := make([]*pb.Doc, len(items))
	for i := range items {
		pbItems[i] = docToProto(&items[i])
	}
	return &pb.SearchDocsResponse{Items: pbItems}, nil
}

// --- Tree & Ancestors ---

// GetDocTree implements KnowledgeServiceServer.
func (t *Server) GetDocTree(ctx context.Context, req *pb.GetDocTreeRequest) (*pb.GetDocTreeResponse, error) {
	var parentID *string
	if req.GetParentId() != "" {
		pid := req.GetParentId()
		parentID = &pid
	}
	docs, err := t.docSvc.GetDocTree(ctx, req.GetArchiveId(), parentID)
	if err != nil {
		return nil, toStatusError(err)
	}

	// 有 parentId 时返回 flat 列表（懒加载子节点）
	if parentID != nil {
		nodes := make([]*pb.DocTreeNode, len(docs))
		for i := range docs {
			nodes[i] = &pb.DocTreeNode{Doc: docToProto(&docs[i])}
		}
		return &pb.GetDocTreeResponse{Nodes: nodes}, nil
	}

	// 无 parentId 时构建完整嵌套树
	return &pb.GetDocTreeResponse{Nodes: buildTree(docs)}, nil
}

// buildTree 将 flat doc 列表构建为嵌套 DocTreeNode 树。
func buildTree(flatDocs []service.Doc) []*pb.DocTreeNode {
	childrenMap := make(map[string][]*service.Doc)
	var roots []*service.Doc

	for i := range flatDocs {
		d := &flatDocs[i]
		if d.ParentID != nil {
			childrenMap[*d.ParentID] = append(childrenMap[*d.ParentID], d)
		} else {
			roots = append(roots, d)
		}
	}

	const maxTreeDepth = 100
	var buildNodes func(docs []*service.Doc, depth int) []*pb.DocTreeNode
	buildNodes = func(docs []*service.Doc, depth int) []*pb.DocTreeNode {
		if depth > maxTreeDepth {
			return nil
		}
		nodes := make([]*pb.DocTreeNode, 0, len(docs))
		for _, d := range docs {
			node := &pb.DocTreeNode{Doc: docToProto(d)}
			if children, ok := childrenMap[d.ID]; ok {
				node.Children = buildNodes(children, depth+1)
			}
			nodes = append(nodes, node)
		}
		return nodes
	}

	return buildNodes(roots, 0)
}

// GetDocAncestors implements KnowledgeServiceServer.
func (t *Server) GetDocAncestors(ctx context.Context, req *pb.GetDocAncestorsRequest) (*pb.GetDocAncestorsResponse, error) {
	ancestors, err := t.docSvc.GetDocAncestors(ctx, req.GetId())
	if err != nil {
		return nil, toStatusError(err)
	}
	pbAncestors := make([]*pb.Doc, len(ancestors))
	for i := range ancestors {
		pbAncestors[i] = docToProto(&ancestors[i])
	}
	return &pb.GetDocAncestorsResponse{Ancestors: pbAncestors}, nil
}

// --- Doc Links ---

// GetLinks implements KnowledgeServiceServer.
func (t *Server) GetLinks(ctx context.Context, req *pb.GetLinksRequest) (*pb.GetLinksResponse, error) {
	links, err := t.docSvc.GetLinks(ctx, req.GetId(), req.GetDirection(), req.GetRelationType())
	if err != nil {
		return nil, toStatusError(err)
	}
	pbLinks := make([]*pb.DocLink, len(links))
	for i := range links {
		pbLinks[i] = docLinkToProto(&links[i])
	}
	return &pb.GetLinksResponse{Links: pbLinks}, nil
}

// CreateLink implements KnowledgeServiceServer.
func (t *Server) CreateLink(ctx context.Context, req *pb.CreateLinkRequest) (*pb.DocLink, error) {
	link, err := t.docSvc.CreateLink(ctx, req.GetId(), req.GetTargetId(), req.GetRelationType())
	if err != nil {
		return nil, toStatusError(err)
	}
	return docLinkToProto(link), nil
}

// DeleteLink implements KnowledgeServiceServer.
func (t *Server) DeleteLink(ctx context.Context, req *pb.DeleteLinkRequest) (*emptypb.Empty, error) {
	if err := t.docSvc.DeleteLink(ctx, req.GetLinkId()); err != nil {
		return nil, toStatusError(err)
	}
	return &emptypb.Empty{}, nil
}

// --- Error mapping ---

func toStatusError(err error) error {
	var ve *service.ValidationError
	switch {
	case err == nil:
		return nil
	case status.Code(err) != codes.Unknown:
		return err
	case errors.As(err, &ve):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrArchiveNotFound),
		errors.Is(err, service.ErrDocNotFound),
		errors.Is(err, service.ErrLinkNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
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

func docToProto(d *service.Doc) *pb.Doc {
	return &pb.Doc{
		Id:          d.ID,
		ArchiveId:   d.ArchiveID,
		ParentId:    strValue(d.ParentID),
		Title:       d.Title,
		Content:     d.Content,
		Summary:     d.Summary,
		Status:      strValue(d.Status),
		Priority:    strValue(d.Priority),
		Assignee:    d.Assignee,
		Tags:        d.Tags,
		Author:      d.Author,
		Visibility:  d.Visibility,
		SharedWith:  d.SharedWith,
		Attachments: attachmentsToProto(d.Attachments),
		CreatedAt:   d.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   d.UpdatedAt.Format(time.RFC3339),
	}
}

func docLinkToProto(l *service.DocLink) *pb.DocLink {
	return &pb.DocLink{
		Id:           l.ID,
		SourceId:     l.SourceID,
		TargetId:     l.TargetID,
		RelationType: l.RelationType,
		SourceTitle:  l.SourceTitle,
		TargetTitle:  l.TargetTitle,
		CreatedAt:    l.CreatedAt.Format(time.RFC3339),
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

func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

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

var _ pb.KnowledgeServiceServer = (*Server)(nil)
