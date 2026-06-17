// STUB FEATURE — delete internal/features/tasks to start your project.

package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/zercle/go-angular-spa-template/internal/features/tasks/domain"
	sqlcdb "github.com/zercle/go-angular-spa-template/internal/infrastructure/db/sqlc"
)

// Repository is a pgx + sqlc implementation of the domain.Repository port.
type Repository struct {
	queries *sqlcdb.Queries
}

// NewRepository returns a Repository backed by the provided sqlc queries.
func NewRepository(queries *sqlcdb.Queries) *Repository {
	return &Repository{queries: queries}
}

// Create persists a new task.
// nolint:wrapcheck // sqlc exec error is propagated without added context.
func (r *Repository) Create(ctx context.Context, task *domain.Task) error {
	return r.queries.CreateTask(ctx, sqlcdb.CreateTaskParams{
		ID:        task.ID,
		Title:     task.Title,
		Done:      task.Done,
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
	})
}

// GetByID retrieves a task by its UUID. It maps pgx.ErrNoRows to
// domain.ErrTaskNotFound and wraps other errors.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	row, err := r.queries.GetTask(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTaskNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}

	return mapRowToDomain(row), nil
}

// List returns a paginated slice of tasks ordered by created_at descending.
func (r *Repository) List(ctx context.Context, limit, offset int32) ([]domain.Task, error) {
	rows, err := r.queries.ListTasks(ctx, sqlcdb.ListTasksParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	tasks := make([]domain.Task, len(rows))
	for i, row := range rows {
		tasks[i] = *mapRowToDomain(row)
	}

	return tasks, nil
}

// Update modifies an existing task's title, done flag, and updated_at.
func (r *Repository) Update(ctx context.Context, task *domain.Task) error {
	rows, err := r.queries.UpdateTask(ctx, sqlcdb.UpdateTaskParams{
		ID:        task.ID,
		Title:     task.Title,
		Done:      task.Done,
		UpdatedAt: task.UpdatedAt,
	})
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	if rows == 0 {
		return domain.ErrTaskNotFound
	}
	return nil
}

// Delete removes a task, returning domain.ErrTaskNotFound when nothing matched.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	rows, err := r.queries.DeleteTask(ctx, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if rows == 0 {
		return domain.ErrTaskNotFound
	}
	return nil
}

func mapRowToDomain(row sqlcdb.Task) *domain.Task {
	return &domain.Task{
		ID:        row.ID,
		Title:     row.Title,
		Done:      row.Done,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
