package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/charviki/maze/the-mesa/the-forge/internal/repository/postgres/sqlc"
	"github.com/charviki/maze/the-mesa/the-forge/internal/service"
)

// KnowledgeRepository 实现 service.KnowledgeRepository。
type KnowledgeRepository struct {
	pool *pgxpool.Pool
	q    *gen.Queries
}

// NewKnowledgeRepository 创建 KnowledgeRepository。
func NewKnowledgeRepository(pool *pgxpool.Pool) *KnowledgeRepository {
	return &KnowledgeRepository{pool: pool, q: gen.New(pool)}
}

// --- Archive ---

// CreateArchive inserts a new knowledge archive into the database.
func (r *KnowledgeRepository) CreateArchive(ctx context.Context, name, description, icon, author string) (*service.Archive, error) {
	row, err := r.q.CreateArchive(ctx, gen.CreateArchiveParams{Name: name, Description: description, Icon: icon, Author: author})
	if err != nil {
		return nil, err
	}
	return convertArchive(row), nil
}

// GetArchive retrieves a single archive by its ID.
func (r *KnowledgeRepository) GetArchive(ctx context.Context, id string) (*service.Archive, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetArchive(ctx, uid)
	if err != nil {
		return nil, err
	}
	return convertArchive(row), nil
}

// ListArchives returns all knowledge archives.
func (r *KnowledgeRepository) ListArchives(ctx context.Context) ([]service.Archive, error) {
	rows, err := r.q.ListArchives(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]service.Archive, len(rows))
	for i, row := range rows {
		result[i] = *convertArchive(row)
	}
	return result, nil
}

// UpdateArchive updates the name, description, and icon of an existing archive.
func (r *KnowledgeRepository) UpdateArchive(ctx context.Context, id, name, description, icon string) (*service.Archive, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	row, err := r.q.UpdateArchive(ctx, gen.UpdateArchiveParams{ID: uid, Name: name, Description: description, Icon: icon})
	if err != nil {
		return nil, err
	}
	return convertArchive(row), nil
}

// DeleteArchive removes an archive by its ID.
func (r *KnowledgeRepository) DeleteArchive(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return err
	}
	return r.q.DeleteArchive(ctx, uid)
}

// --- Memory ---

// CreateMemory inserts a new memory document into the database.
func (r *KnowledgeRepository) CreateMemory(ctx context.Context, archiveID string, parentID *string, kind, title, content, summary, memType string, tags []string, author, visibility string, sharedWith []string, attachments []service.Attachment) (*service.Memory, error) {
	aid, err := parseUUID(archiveID)
	if err != nil {
		return nil, err
	}
	var pid pgtype.UUID
	if parentID != nil {
		pid, err = parseUUID(*parentID)
		if err != nil {
			return nil, err
		}
	}
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return nil, fmt.Errorf("marshal tags: %w", err)
	}
	sharedJSON, err := json.Marshal(sharedWith)
	if err != nil {
		return nil, fmt.Errorf("marshal sharedWith: %w", err)
	}
	attachJSON, err := json.Marshal(attachments)
	if err != nil {
		return nil, fmt.Errorf("marshal attachments: %w", err)
	}

	row, err := r.q.CreateMemory(ctx, gen.CreateMemoryParams{
		ArchiveID: aid, ParentID: pid, Kind: kind, Title: title, Content: content,
		Summary: summary, Type: memType, Tags: tagsJSON, Author: author,
		Visibility: visibility, SharedWith: sharedJSON, Attachments: attachJSON,
	})
	if err != nil {
		return nil, err
	}
	return convertMemory(row), nil
}

// GetMemory retrieves a single memory by its ID.
func (r *KnowledgeRepository) GetMemory(ctx context.Context, id string) (*service.Memory, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetMemory(ctx, uid)
	if err != nil {
		return nil, err
	}
	return convertMemory(row), nil
}

// ListMemories returns a paginated list of memories matching the given filters.
func (r *KnowledgeRepository) ListMemories(ctx context.Context, archiveID *string, parentID *string, kind, memType, visibility, author string, limit, offset int32) ([]service.Memory, error) {
	var aid pgtype.UUID
	if archiveID != nil {
		aid, _ = parseUUID(*archiveID)
	}
	var pid pgtype.UUID
	if parentID != nil {
		pid, _ = parseUUID(*parentID)
	}
	rows, err := r.q.ListMemories(ctx, gen.ListMemoriesParams{
		ArchiveID: aid, ParentID: pid, Kind: kind, Type: memType,
		Visibility: visibility, Author: author, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	result := make([]service.Memory, len(rows))
	for i, row := range rows {
		result[i] = *convertMemory(row)
	}
	return result, nil
}

// CountMemories returns the total number of memories matching the given filters.
func (r *KnowledgeRepository) CountMemories(ctx context.Context, archiveID *string, parentID *string, kind, memType, visibility, author string) (int64, error) {
	var aid pgtype.UUID
	if archiveID != nil {
		aid, _ = parseUUID(*archiveID)
	}
	var pid pgtype.UUID
	if parentID != nil {
		pid, _ = parseUUID(*parentID)
	}
	return r.q.CountMemories(ctx, gen.CountMemoriesParams{
		ArchiveID: aid, ParentID: pid, Kind: kind, Type: memType,
		Visibility: visibility, Author: author,
	})
}

// UpdateMemory updates the fields of an existing memory.
func (r *KnowledgeRepository) UpdateMemory(ctx context.Context, id string, params service.UpdateMemoryParams) (*service.Memory, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	tagsJSON, err := json.Marshal(params.Tags)
	if err != nil {
		return nil, fmt.Errorf("marshal tags: %w", err)
	}
	sharedJSON, err := json.Marshal(params.SharedWith)
	if err != nil {
		return nil, fmt.Errorf("marshal sharedWith: %w", err)
	}
	attachJSON, err := json.Marshal(params.Attachments)
	if err != nil {
		return nil, fmt.Errorf("marshal attachments: %w", err)
	}

	var parentID pgtype.UUID
	if params.ParentID != nil {
		parentID, _ = parseUUID(*params.ParentID)
	}

	row, err := r.q.UpdateMemory(ctx, gen.UpdateMemoryParams{
		ID: uid,
		Title:       derefString(params.Title),
		Content:     derefString(params.Content),
		Summary:     derefString(params.Summary),
		Type:        derefString(params.Type),
		Tags:        tagsJSON,
		Visibility:  derefString(params.Visibility),
		SharedWith:  sharedJSON,
		Attachments: attachJSON,
		ParentID:    parentID,
		Kind:        derefString(params.Kind),
	})
	if err != nil {
		return nil, err
	}
	return convertMemory(row), nil
}

// DeleteMemory removes a memory by its ID.
func (r *KnowledgeRepository) DeleteMemory(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return err
	}
	return r.q.DeleteMemory(ctx, uid)
}

// SearchMemories performs a full-text search for memories matching the query.
func (r *KnowledgeRepository) SearchMemories(ctx context.Context, query string, archiveID *string, visibility, author string) ([]service.Memory, error) {
	var aid pgtype.UUID
	if archiveID != nil {
		aid, _ = parseUUID(*archiveID)
	}
	rows, err := r.q.SearchMemories(ctx, gen.SearchMemoriesParams{
		PlaintoTsquery: query, ArchiveID: aid, Visibility: visibility, Author: author,
	})
	if err != nil {
		return nil, err
	}
	result := make([]service.Memory, len(rows))
	for i, row := range rows {
		result[i] = *convertMemory(row)
	}
	return result, nil
}

// GetMemoryChildren returns all immediate child memories of the given parent.
func (r *KnowledgeRepository) GetMemoryChildren(ctx context.Context, parentID string) ([]service.Memory, error) {
	uid, err := parseUUID(parentID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetMemoryChildren(ctx, uid)
	if err != nil {
		return nil, err
	}
	result := make([]service.Memory, len(rows))
	for i, row := range rows {
		result[i] = *convertMemory(row)
	}
	return result, nil
}

// GetMemoryRootChildren returns root-level memories (parent_id IS NULL) for the given archive.
func (r *KnowledgeRepository) GetMemoryRootChildren(ctx context.Context, archiveID string) ([]service.Memory, error) {
	aid, err := parseUUID(archiveID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetMemoryRootChildren(ctx, aid)
	if err != nil {
		return nil, err
	}
	result := make([]service.Memory, len(rows))
	for i, row := range rows {
		result[i] = *convertMemory(row)
	}
	return result, nil
}

// GetMemoryAncestors returns the ancestor chain of a memory using a recursive CTE.
func (r *KnowledgeRepository) GetMemoryAncestors(ctx context.Context, id string) ([]service.Memory, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	// 递归 CTE 手写查询（sqlc 无法解析）
	rows, err := r.pool.Query(ctx, `
		WITH RECURSIVE ancestor_path(id, archive_id, parent_id, kind, title, content, summary, type, tags, author, visibility, shared_with, attachments, created_at, updated_at) AS (
			SELECT id, archive_id, parent_id, kind, title, content, summary, type, tags, author, visibility, shared_with, attachments, created_at, updated_at
			FROM memories WHERE id = $1
			UNION
			SELECT m.id, m.archive_id, m.parent_id, m.kind, m.title, m.content, m.summary, m.type, m.tags, m.author, m.visibility, m.shared_with, m.attachments, m.created_at, m.updated_at
			FROM memories m JOIN ancestor_path ap ON m.id = ap.parent_id
		)
		SELECT id, archive_id, parent_id, kind, title, content, summary, type, tags, author, visibility, shared_with, attachments, created_at, updated_at
		FROM ancestor_path WHERE id != $1 ORDER BY created_at ASC`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []service.Memory
	for rows.Next() {
		var m gen.Memory
		if err := rows.Scan(&m.ID, &m.ArchiveID, &m.ParentID, &m.Kind, &m.Title, &m.Content, &m.Summary, &m.Type, &m.Tags, &m.Author, &m.Visibility, &m.SharedWith, &m.Attachments, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, *convertMemory(m))
	}
	return result, nil
}

// --- Neural Link ---

// CreateLink creates a neural link between two memories.
func (r *KnowledgeRepository) CreateLink(ctx context.Context, sourceID, targetID, relationType string) (*service.NeuralLink, error) {
	sid, err := parseUUID(sourceID)
	if err != nil {
		return nil, err
	}
	tid, err := parseUUID(targetID)
	if err != nil {
		return nil, err
	}
	row, err := r.q.CreateLink(ctx, gen.CreateLinkParams{SourceID: sid, TargetID: tid, RelationType: relationType})
	if err != nil {
		return nil, err
	}
	return &service.NeuralLink{
		ID:           formatUUID(row.ID),
		SourceID:     formatUUID(row.SourceID),
		TargetID:     formatUUID(row.TargetID),
		RelationType: row.RelationType,
		CreatedAt:    row.CreatedAt.Time,
	}, nil
}

// GetOutLinks returns all outgoing neural links from the given source memory.
func (r *KnowledgeRepository) GetOutLinks(ctx context.Context, sourceID, relationType string) ([]service.NeuralLink, error) {
	sid, err := parseUUID(sourceID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetLinksBySource(ctx, gen.GetLinksBySourceParams{SourceID: sid, RelationType: relationType})
	if err != nil {
		return nil, err
	}
	result := make([]service.NeuralLink, len(rows))
	for i, row := range rows {
		result[i] = service.NeuralLink{
			ID:           formatUUID(row.ID),
			SourceID:     formatUUID(row.SourceID),
			TargetID:     formatUUID(row.TargetID),
			RelationType: row.RelationType,
			TargetTitle:  row.TargetTitle,
			CreatedAt:    row.CreatedAt.Time,
		}
	}
	return result, nil
}

// GetInLinks returns all incoming neural links to the given target memory.
func (r *KnowledgeRepository) GetInLinks(ctx context.Context, targetID, relationType string) ([]service.NeuralLink, error) {
	tid, err := parseUUID(targetID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetLinksByTarget(ctx, gen.GetLinksByTargetParams{TargetID: tid, RelationType: relationType})
	if err != nil {
		return nil, err
	}
	result := make([]service.NeuralLink, len(rows))
	for i, row := range rows {
		result[i] = service.NeuralLink{
			ID:           formatUUID(row.ID),
			SourceID:     formatUUID(row.SourceID),
			TargetID:     formatUUID(row.TargetID),
			RelationType: row.RelationType,
			SourceTitle:  row.SourceTitle,
			CreatedAt:    row.CreatedAt.Time,
		}
	}
	return result, nil
}

// DeleteLink removes a neural link by its ID.
func (r *KnowledgeRepository) DeleteLink(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return err
	}
	return r.q.DeleteLink(ctx, uid)
}

// --- Stats ---

// GetTotalMemories returns the total count of memories in the database.
func (r *KnowledgeRepository) GetTotalMemories(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, "SELECT count(*) FROM memories").Scan(&count)
	return count, err
}

// GetRecentMemories returns the most recently updated memories up to the given limit.
func (r *KnowledgeRepository) GetRecentMemories(ctx context.Context, limit int32) ([]service.Memory, error) {
	rows, err := r.pool.Query(ctx, "SELECT id, archive_id, parent_id, kind, title, content, summary, type, tags, author, visibility, shared_with, attachments, created_at, updated_at FROM memories ORDER BY updated_at DESC LIMIT $1", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []service.Memory
	for rows.Next() {
		var m gen.Memory
		if err := rows.Scan(&m.ID, &m.ArchiveID, &m.ParentID, &m.Kind, &m.Title, &m.Content, &m.Summary, &m.Type, &m.Tags, &m.Author, &m.Visibility, &m.SharedWith, &m.Attachments, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, *convertMemory(m))
	}
	return result, nil
}

// --- Helpers ---

func convertArchive(row gen.Archive) *service.Archive {
	return &service.Archive{
		ID:          formatUUID(row.ID),
		Name:        row.Name,
		Description: row.Description,
		Icon:        row.Icon,
		Author:      row.Author,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}

func convertMemory(row gen.Memory) *service.Memory {
	var parentID *string
	if row.ParentID.Valid {
		s := formatUUID(row.ParentID)
		parentID = &s
	}
	return &service.Memory{
		ID:          formatUUID(row.ID),
		ArchiveID:   formatUUID(row.ArchiveID),
		ParentID:    parentID,
		Kind:        row.Kind,
		Title:       row.Title,
		Content:     row.Content,
		Summary:     row.Summary,
		Type:        row.Type,
		Tags:        jsonToStringSlice(row.Tags),
		Author:      row.Author,
		Visibility:  row.Visibility,
		SharedWith:  jsonToStringSlice(row.SharedWith),
		Attachments: jsonToAttachments(row.Attachments),
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}

func parseUUID(s string) (pgtype.UUID, error) {
	var uid pgtype.UUID
	if err := uid.Scan(s); err != nil {
		return uid, fmt.Errorf("invalid UUID %q: %w", s, err)
	}
	return uid, nil
}

func formatUUID(uid pgtype.UUID) string {
	if !uid.Valid {
		return ""
	}
	return uid.String()
}

func jsonToStringSlice(data []byte) []string {
	if data == nil {
		return []string{}
	}
	var result []string
	if err := json.Unmarshal(data, &result); err != nil {
		log.Printf("[knowledge-repo] unmarshal string slice: %v", err)
	}
	if result == nil {
		return []string{}
	}
	return result
}

func jsonToAttachments(data []byte) []service.Attachment {
	if data == nil {
		return []service.Attachment{}
	}
	var result []service.Attachment
	if err := json.Unmarshal(data, &result); err != nil {
		log.Printf("[knowledge-repo] unmarshal attachments: %v", err)
	}
	if result == nil {
		return []service.Attachment{}
	}
	return result
}

func derefString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// Ensure interface compliance.
var _ service.KnowledgeRepository = (*KnowledgeRepository)(nil)
