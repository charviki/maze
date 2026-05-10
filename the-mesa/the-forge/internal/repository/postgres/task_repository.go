package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/charviki/maze/the-mesa/the-forge/internal/repository/postgres/sqlc"
	"github.com/charviki/maze/the-mesa/the-forge/internal/service"
)

// TaskRepository 实现 service.TaskRepository。
type TaskRepository struct {
	pool *pgxpool.Pool
	q    *gen.Queries
}

// NewTaskRepository 创建 TaskRepository。
func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool, q: gen.New(pool)}
}

// CreateDirective inserts a new directive into the database.
func (r *TaskRepository) CreateDirective(ctx context.Context, title, description, status, priority, assignee, author string, requireDocIDs []string, narrativeID string, archiveID *string, visibility string) (*service.Directive, error) {
	var aid pgtype.UUID
	if archiveID != nil {
		aid, _ = parseUUID(*archiveID)
	}
	docIDsJSON, _ := json.Marshal(requireDocIDs)

	row, err := r.q.CreateDirective(ctx, gen.CreateDirectiveParams{
		Title: title, Description: description, Status: status, Priority: priority,
		Assignee: assignee, Author: author, RequireDocIds: docIDsJSON,
		NarrativeID: narrativeID, ArchiveID: aid, Visibility: visibility,
	})
	if err != nil {
		return nil, err
	}
	return convertDirective(row), nil
}

// GetDirective retrieves a single directive by its ID.
func (r *TaskRepository) GetDirective(ctx context.Context, id string) (*service.Directive, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetDirective(ctx, uid)
	if err != nil {
		return nil, err
	}
	return convertDirective(row), nil
}

// ListDirectives returns a paginated list of directives matching the given filters.
func (r *TaskRepository) ListDirectives(ctx context.Context, status, assignee, priority, visibility string, limit, offset int32) ([]service.Directive, error) {
	rows, err := r.q.ListDirectives(ctx, gen.ListDirectivesParams{
		Status: status, Assignee: assignee, Priority: priority, Visibility: visibility,
		Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	result := make([]service.Directive, len(rows))
	for i, row := range rows {
		result[i] = *convertDirective(row)
	}
	return result, nil
}

// CountDirectives returns the total number of directives matching the given filters.
func (r *TaskRepository) CountDirectives(ctx context.Context, status, assignee, priority, visibility string) (int64, error) {
	return r.q.CountDirectives(ctx, gen.CountDirectivesParams{
		Status: status, Assignee: assignee, Priority: priority, Visibility: visibility,
	})
}

// UpdateDirective updates the fields of an existing directive.
func (r *TaskRepository) UpdateDirective(ctx context.Context, id string, params service.UpdateDirectiveParams) (*service.Directive, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	docIDsJSON, _ := json.Marshal(params.RequireDocIDs)
	var aid pgtype.UUID
	if params.ArchiveID != nil {
		aid, _ = parseUUID(*params.ArchiveID)
	}

	row, err := r.q.UpdateDirective(ctx, gen.UpdateDirectiveParams{
		ID:            uid,
		Title:         derefString(params.Title),
		Description:   derefString(params.Description),
		Status:        derefString(params.Status),
		Priority:      derefString(params.Priority),
		Assignee:      derefString(params.Assignee),
		Author:        derefString(params.Author),
		RequireDocIds: docIDsJSON,
		NarrativeID:   derefString(params.NarrativeID),
		ArchiveID:     aid,
		Visibility:    derefString(params.Visibility),
	})
	if err != nil {
		return nil, err
	}
	return convertDirective(row), nil
}

// DeleteDirective removes a directive by its ID.
func (r *TaskRepository) DeleteDirective(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return err
	}
	return r.q.DeleteDirective(ctx, uid)
}

// ListDirectivesByDocID returns all directives that reference the given document ID.
func (r *TaskRepository) ListDirectivesByDocID(ctx context.Context, docID string) ([]service.Directive, error) {
	rows, err := r.q.ListDirectivesByDocID(ctx, docID)
	if err != nil {
		return nil, err
	}
	result := make([]service.Directive, len(rows))
	for i, row := range rows {
		result[i] = *convertDirective(row)
	}
	return result, nil
}

// GetDirectivesByStatus returns a count of directives grouped by status.
func (r *TaskRepository) GetDirectivesByStatus(ctx context.Context) (map[string]int, error) {
	rows, err := r.pool.Query(ctx, "SELECT status, count(*) FROM directives GROUP BY status")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		result[status] = count
	}
	return result, nil
}

func convertDirective(row gen.Directive) *service.Directive {
	var archiveID *string
	if row.ArchiveID.Valid {
		s := formatUUID(row.ArchiveID)
		archiveID = &s
	}
	return &service.Directive{
		ID:            formatUUID(row.ID),
		Title:         row.Title,
		Description:   row.Description,
		Status:        row.Status,
		Priority:      row.Priority,
		Assignee:      row.Assignee,
		Author:        row.Author,
		RequireDocIDs: jsonToStringSlice(row.RequireDocIds),
		NarrativeID:   row.NarrativeID,
		ArchiveID:     archiveID,
		Visibility:    row.Visibility,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}
}

var _ service.TaskRepository = (*TaskRepository)(nil)
