package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/charviki/maze/fabrication/cradle/logutil"
	gen "github.com/charviki/maze/the-mesa/the-forge/internal/repository/postgres/sqlc"
	"github.com/charviki/maze/the-mesa/the-forge/internal/service"
)

// DocRepository 实现 service.DocRepository。
type DocRepository struct {
	pool   *pgxpool.Pool
	txm    *TxManager
	logger logutil.Logger
}

// NewDocRepository 创建 DocRepository。
func NewDocRepository(pool *pgxpool.Pool, txm *TxManager, logger logutil.Logger) *DocRepository {
	return &DocRepository{pool: pool, txm: txm, logger: logger}
}

func (r *DocRepository) queries(ctx context.Context) *gen.Queries {
	return gen.New(docExecutorFromContext(ctx, r.pool))
}

// WithinTx 在事务中执行 fn。
func (r *DocRepository) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.txm.WithinTx(ctx, fn)
}

// --- Archive ---

// CreateArchive inserts a new knowledge archive into the database.
func (r *DocRepository) CreateArchive(ctx context.Context, name, description, icon, author string) (*service.Archive, error) {
	row, err := r.queries(ctx).CreateArchive(ctx, gen.CreateArchiveParams{Name: name, Description: description, Icon: icon, Author: author})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, service.ErrAlreadyExists
		}
		return nil, err
	}
	return convertArchive(row), nil
}

// GetArchive retrieves a single archive by its ID.
func (r *DocRepository) GetArchive(ctx context.Context, id string) (*service.Archive, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	row, err := r.queries(ctx).GetArchive(ctx, uid)
	if err != nil {
		return nil, mapNotFoundError(err, service.ErrArchiveNotFound)
	}
	return convertArchive(row), nil
}

// ListArchives returns all knowledge archives.
func (r *DocRepository) ListArchives(ctx context.Context) ([]service.Archive, error) {
	rows, err := r.queries(ctx).ListArchives(ctx)
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
func (r *DocRepository) UpdateArchive(ctx context.Context, id, name, description, icon string) (*service.Archive, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	row, err := r.queries(ctx).UpdateArchive(ctx, gen.UpdateArchiveParams{ID: uid, Name: name, Description: description, Icon: icon})
	if err != nil {
		return nil, mapNotFoundError(err, service.ErrArchiveNotFound)
	}
	return convertArchive(row), nil
}

// DeleteArchive removes an archive by its ID.
func (r *DocRepository) DeleteArchive(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return err
	}
	err = r.queries(ctx).DeleteArchive(ctx, uid)
	return mapNotFoundError(err, service.ErrArchiveNotFound)
}

// --- Doc ---

// CreateDoc inserts a new doc into the database.
func (r *DocRepository) CreateDoc(ctx context.Context, params service.CreateDocParams) (*service.Doc, error) {
	aid, err := parseUUID(params.ArchiveID)
	if err != nil {
		return nil, err
	}
	var pid pgtype.UUID
	if params.ParentID != nil {
		pid, err = parseUUID(*params.ParentID)
		if err != nil {
			return nil, err
		}
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

	row, err := r.queries(ctx).CreateDoc(ctx, gen.CreateDocParams{
		ArchiveID:   aid,
		ParentID:    pid,
		Title:       params.Title,
		Content:     params.Content,
		Summary:     params.Summary,
		Status:      params.Status,
		Priority:    params.Priority,
		Assignee:    params.Assignee,
		Tags:        tagsJSON,
		Author:      params.Author,
		Visibility:  params.Visibility,
		SharedWith:  sharedJSON,
		Attachments: attachJSON,
	})
	if err != nil {
		return nil, err
	}
	return convertDoc(row, r.logger)
}

// GetDoc retrieves a single doc by its ID.
func (r *DocRepository) GetDoc(ctx context.Context, id string) (*service.Doc, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	row, err := r.queries(ctx).GetDoc(ctx, uid)
	if err != nil {
		return nil, mapNotFoundError(err, service.ErrDocNotFound)
	}
	return convertDoc(row, r.logger)
}

// ListDocs returns a paginated list of docs matching the given filters.
func (r *DocRepository) ListDocs(ctx context.Context, archiveID *string, parentID *string, status, visibility, author string, limit, offset int32) ([]service.Doc, error) {
	var aid pgtype.UUID
	if archiveID != nil {
		var err error
		if aid, err = parseUUID(*archiveID); err != nil {
			return nil, err
		}
	}
	var pid pgtype.UUID
	if parentID != nil {
		var err error
		if pid, err = parseUUID(*parentID); err != nil {
			return nil, err
		}
	}
	rows, err := r.queries(ctx).ListDocs(ctx, gen.ListDocsParams{
		Column1: aid,
		Column2: pid,
		Column3: status,
		Column4: visibility,
		Column5: author,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, err
	}
	result := make([]service.Doc, len(rows))
	for i, row := range rows {
		doc, err := convertDoc(row, r.logger)
		if err != nil {
			return nil, err
		}
		result[i] = *doc
	}
	return result, nil
}

// ListDocsHasStatus returns a paginated list of docs with non-null status matching the given filters.
func (r *DocRepository) ListDocsHasStatus(ctx context.Context, archiveID *string, parentID *string, status, visibility, author string, limit, offset int32) ([]service.Doc, error) {
	var aid pgtype.UUID
	if archiveID != nil {
		var err error
		if aid, err = parseUUID(*archiveID); err != nil {
			return nil, err
		}
	}
	var pid pgtype.UUID
	if parentID != nil {
		var err error
		if pid, err = parseUUID(*parentID); err != nil {
			return nil, err
		}
	}
	rows, err := r.queries(ctx).ListDocsHasStatus(ctx, gen.ListDocsHasStatusParams{
		Column1: aid,
		Column2: pid,
		Column3: status,
		Column4: visibility,
		Column5: author,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, err
	}
	result := make([]service.Doc, len(rows))
	for i, row := range rows {
		doc, err := convertDoc(row, r.logger)
		if err != nil {
			return nil, err
		}
		result[i] = *doc
	}
	return result, nil
}

// CountDocs returns the total number of docs matching the given filters.
func (r *DocRepository) CountDocs(ctx context.Context, archiveID *string, parentID *string, status, visibility, author string) (int64, error) {
	var aid pgtype.UUID
	if archiveID != nil {
		var err error
		if aid, err = parseUUID(*archiveID); err != nil {
			return 0, err
		}
	}
	var pid pgtype.UUID
	if parentID != nil {
		var err error
		if pid, err = parseUUID(*parentID); err != nil {
			return 0, err
		}
	}
	return r.queries(ctx).CountDocs(ctx, gen.CountDocsParams{
		Column1: aid,
		Column2: pid,
		Column3: status,
		Column4: visibility,
		Column5: author,
	})
}

// CountDocsHasStatus returns the total number of docs with non-null status matching the given filters.
func (r *DocRepository) CountDocsHasStatus(ctx context.Context, archiveID *string, parentID *string, status, visibility, author string) (int64, error) {
	var aid pgtype.UUID
	if archiveID != nil {
		var err error
		if aid, err = parseUUID(*archiveID); err != nil {
			return 0, err
		}
	}
	var pid pgtype.UUID
	if parentID != nil {
		var err error
		if pid, err = parseUUID(*parentID); err != nil {
			return 0, err
		}
	}
	return r.queries(ctx).CountDocsHasStatus(ctx, gen.CountDocsHasStatusParams{
		Column1: aid,
		Column2: pid,
		Column3: status,
		Column4: visibility,
		Column5: author,
	})
}

// UpdateDoc updates the fields of an existing doc.
func (r *DocRepository) UpdateDoc(ctx context.Context, id string, params service.UpdateDocParams) (*service.Doc, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	var parentID pgtype.UUID
	if params.ParentID != nil {
		parentID, _ = parseUUID(*params.ParentID)
	}

	row, err := r.queries(ctx).UpdateDoc(ctx, gen.UpdateDocParams{
		ID:            uid,
		Title:         params.Title,
		Content:       params.Content,
		Summary:       params.Summary,
		ClearStatus:   boolPtr(params.ClearStatus),
		Status:        params.Status,
		ClearPriority: boolPtr(params.ClearPriority),
		Priority:      params.Priority,
		Assignee:      params.Assignee,
		Tags:          params.Tags,
		Visibility:    params.Visibility,
		SharedWith:    params.SharedWith,
		Attachments:   params.Attachments,
		ParentID:      parentID,
	})
	if err != nil {
		return nil, mapNotFoundError(err, service.ErrDocNotFound)
	}
	return convertDoc(row, r.logger)
}

// DeleteDoc removes a doc by its ID.
func (r *DocRepository) DeleteDoc(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return err
	}
	err = r.queries(ctx).DeleteDoc(ctx, uid)
	return mapNotFoundError(err, service.ErrDocNotFound)
}

// SearchDocs performs a full-text search for docs matching the query.
func (r *DocRepository) SearchDocs(ctx context.Context, query string, archiveID *string, visibility, author string) ([]service.Doc, error) {
	var aid pgtype.UUID
	if archiveID != nil {
		aid, _ = parseUUID(*archiveID)
	}
	rows, err := r.queries(ctx).SearchDocs(ctx, gen.SearchDocsParams{
		PlaintoTsquery: query,
		ArchiveID:      aid,
		Visibility:     visibility,
		Author:         author,
	})
	if err != nil {
		return nil, err
	}
	result := make([]service.Doc, len(rows))
	for i, row := range rows {
		doc, err := convertDoc(row, r.logger)
		if err != nil {
			return nil, err
		}
		result[i] = *doc
	}
	return result, nil
}

// GetDocChildren returns all immediate child docs of the given parent.
func (r *DocRepository) GetDocChildren(ctx context.Context, parentID string) ([]service.Doc, error) {
	uid, err := parseUUID(parentID)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries(ctx).GetDocChildren(ctx, uid)
	if err != nil {
		return nil, err
	}
	result := make([]service.Doc, len(rows))
	for i, row := range rows {
		doc, err := convertDoc(row, r.logger)
		if err != nil {
			return nil, err
		}
		result[i] = *doc
	}
	return result, nil
}

// GetDocRootChildren returns root-level docs (parent_id IS NULL) for the given archive.
func (r *DocRepository) GetDocRootChildren(ctx context.Context, archiveID string) ([]service.Doc, error) {
	aid, err := parseUUID(archiveID)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries(ctx).GetDocRootChildren(ctx, aid)
	if err != nil {
		return nil, err
	}
	result := make([]service.Doc, len(rows))
	for i, row := range rows {
		doc, err := convertDoc(row, r.logger)
		if err != nil {
			return nil, err
		}
		result[i] = *doc
	}
	return result, nil
}

// GetAllDocsByArchive returns all docs in the given archive for tree building.
func (r *DocRepository) GetAllDocsByArchive(ctx context.Context, archiveID string) ([]service.Doc, error) {
	aid, err := parseUUID(archiveID)
	if err != nil {
		return nil, err
	}
	rows, err := docExecutorFromContext(ctx, r.pool).Query(ctx, `
		SELECT id, archive_id, parent_id, title, content, summary, status, priority, assignee, tags, author, visibility, shared_with, attachments, created_at, updated_at
		FROM docs WHERE archive_id = $1
		ORDER BY title ASC`, aid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []service.Doc
	for rows.Next() {
		var d gen.Doc
		if err := rows.Scan(&d.ID, &d.ArchiveID, &d.ParentID, &d.Title, &d.Content, &d.Summary, &d.Status, &d.Priority, &d.Assignee, &d.Tags, &d.Author, &d.Visibility, &d.SharedWith, &d.Attachments, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		converted, err := convertDoc(d, r.logger)
		if err != nil {
			return nil, err
		}
		result = append(result, *converted)
	}
	return result, nil
}

// GetDocAncestors returns the ancestor chain of a doc using a recursive CTE.
func (r *DocRepository) GetDocAncestors(ctx context.Context, id string) ([]service.Doc, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	rows, err := docExecutorFromContext(ctx, r.pool).Query(ctx, `
		WITH RECURSIVE ancestor_path(id, archive_id, parent_id, title, content, summary, status, priority, assignee, tags, author, visibility, shared_with, attachments, created_at, updated_at) AS (
			SELECT id, archive_id, parent_id, title, content, summary, status, priority, assignee, tags, author, visibility, shared_with, attachments, created_at, updated_at
			FROM docs WHERE id = $1
			UNION
			SELECT d.id, d.archive_id, d.parent_id, d.title, d.content, d.summary, d.status, d.priority, d.assignee, d.tags, d.author, d.visibility, d.shared_with, d.attachments, d.created_at, d.updated_at
			FROM docs d JOIN ancestor_path ap ON d.id = ap.parent_id
		)
		SELECT id, archive_id, parent_id, title, content, summary, status, priority, assignee, tags, author, visibility, shared_with, attachments, created_at, updated_at
		FROM ancestor_path WHERE id != $1 ORDER BY created_at ASC`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []service.Doc
	for rows.Next() {
		var d gen.Doc
		if err := rows.Scan(&d.ID, &d.ArchiveID, &d.ParentID, &d.Title, &d.Content, &d.Summary, &d.Status, &d.Priority, &d.Assignee, &d.Tags, &d.Author, &d.Visibility, &d.SharedWith, &d.Attachments, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		converted, err := convertDoc(d, r.logger)
		if err != nil {
			return nil, err
		}
		result = append(result, *converted)
	}
	return result, nil
}

// DeleteDocSubtree deletes a doc and all its descendants in a single recursive SQL statement.
func (r *DocRepository) DeleteDocSubtree(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return err
	}
	tag, err := docExecutorFromContext(ctx, r.pool).Exec(ctx, `
		WITH RECURSIVE subtree(id) AS (
			SELECT id FROM docs WHERE id = $1
			UNION
			SELECT d.id FROM docs d JOIN subtree st ON d.parent_id = st.id
		)
		DELETE FROM docs WHERE id IN (SELECT id FROM subtree)`, uid)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return service.ErrDocNotFound
	}
	return nil
}

// --- Doc Link ---

// CreateLink creates a doc link between two docs.
func (r *DocRepository) CreateLink(ctx context.Context, sourceID, targetID, relationType string) (*service.DocLink, error) {
	sid, err := parseUUID(sourceID)
	if err != nil {
		return nil, err
	}
	tid, err := parseUUID(targetID)
	if err != nil {
		return nil, err
	}
	row, err := r.queries(ctx).CreateLink(ctx, gen.CreateLinkParams{SourceID: sid, TargetID: tid, RelationType: relationType})
	if err != nil {
		return nil, err
	}
	return &service.DocLink{
		ID:           formatUUID(row.ID),
		SourceID:     formatUUID(row.SourceID),
		TargetID:     formatUUID(row.TargetID),
		RelationType: row.RelationType,
		CreatedAt:    row.CreatedAt.Time,
	}, nil
}

// GetOutLinks returns all outgoing doc links from the given source doc.
func (r *DocRepository) GetOutLinks(ctx context.Context, sourceID, relationType string) ([]service.DocLink, error) {
	sid, err := parseUUID(sourceID)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries(ctx).GetLinksBySource(ctx, gen.GetLinksBySourceParams{SourceID: sid, RelationType: relationType})
	if err != nil {
		return nil, err
	}
	result := make([]service.DocLink, len(rows))
	for i, row := range rows {
		result[i] = service.DocLink{
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

// GetInLinks returns all incoming doc links to the given target doc.
func (r *DocRepository) GetInLinks(ctx context.Context, targetID, relationType string) ([]service.DocLink, error) {
	tid, err := parseUUID(targetID)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries(ctx).GetLinksByTarget(ctx, gen.GetLinksByTargetParams{TargetID: tid, RelationType: relationType})
	if err != nil {
		return nil, err
	}
	result := make([]service.DocLink, len(rows))
	for i, row := range rows {
		result[i] = service.DocLink{
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

// DeleteLink removes a doc link by its ID.
func (r *DocRepository) DeleteLink(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return err
	}
	err = r.queries(ctx).DeleteLink(ctx, uid)
	return mapNotFoundError(err, service.ErrLinkNotFound)
}

// --- Helpers ---

func mapNotFoundError(err error, target error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return target
	}
	return err
}

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

func convertDoc(row gen.Doc, logger logutil.Logger) (*service.Doc, error) {
	var parentID *string
	if row.ParentID.Valid {
		s := formatUUID(row.ParentID)
		parentID = &s
	}
	tags, err := jsonToStringSlice(row.Tags, logger)
	if err != nil {
		return nil, fmt.Errorf("unmarshal tags: %w", err)
	}
	sharedWith, err := jsonToStringSlice(row.SharedWith, logger)
	if err != nil {
		return nil, fmt.Errorf("unmarshal sharedWith: %w", err)
	}
	attachments, err := jsonToAttachments(row.Attachments, logger)
	if err != nil {
		return nil, fmt.Errorf("unmarshal attachments: %w", err)
	}
	return &service.Doc{
		ID:          formatUUID(row.ID),
		ArchiveID:   formatUUID(row.ArchiveID),
		ParentID:    parentID,
		Title:       row.Title,
		Content:     row.Content,
		Summary:     row.Summary,
		Status:      row.Status,
		Priority:    row.Priority,
		Assignee:    row.Assignee,
		Tags:        tags,
		Author:      row.Author,
		Visibility:  row.Visibility,
		SharedWith:  sharedWith,
		Attachments: attachments,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
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

func jsonToStringSlice(data []byte, logger logutil.Logger) ([]string, error) {
	if data == nil {
		return []string{}, nil
	}
	var result []string
	if err := json.Unmarshal(data, &result); err != nil {
		logger.Warnf("[doc-repo] unmarshal string slice: %v", err)
		return nil, err
	}
	if result == nil {
		return []string{}, nil
	}
	return result, nil
}

func jsonToAttachments(data []byte, logger logutil.Logger) ([]service.Attachment, error) {
	if data == nil {
		return []service.Attachment{}, nil
	}
	var result []service.Attachment
	if err := json.Unmarshal(data, &result); err != nil {
		logger.Warnf("[doc-repo] unmarshal attachments: %v", err)
		return nil, err
	}
	if result == nil {
		return []service.Attachment{}, nil
	}
	return result, nil
}

func boolPtr(b bool) *bool { return &b }

// Ensure interface compliance.
var _ service.DocRepository = (*DocRepository)(nil)
