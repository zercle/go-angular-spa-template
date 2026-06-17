// Package postgres is the infrastructure adapter implementing the domain ports
// on top of Postgres, using sqlc-generated queries and a pgx connection pool.
package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cerzzlive/gofiber-angular-spa/backend/internal/domain"
	"github.com/cerzzlive/gofiber-angular-spa/backend/internal/repository/postgres/gen"
)

// TaskRepository implements domain.TaskRepository.
type TaskRepository struct {
	q *gen.Queries
}

// NewTaskRepository builds a repository backed by the given pool.
func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{q: gen.New(pool)}
}

// Compile-time assertion that the adapter satisfies the domain port.
var _ domain.TaskRepository = (*TaskRepository)(nil)

func toDomain(t gen.Task) domain.Task {
	return domain.Task{
		ID:        t.ID,
		Title:     t.Title,
		Done:      t.Done,
		CreatedAt: t.CreatedAt.Time,
		UpdatedAt: t.UpdatedAt.Time,
	}
}

// Create inserts a new task.
func (r *TaskRepository) Create(ctx context.Context, p domain.CreateTaskParams) (domain.Task, error) {
	t, err := r.q.CreateTask(ctx, p.Title)
	if err != nil {
		return domain.Task{}, err
	}
	return toDomain(t), nil
}

// Get fetches a task by id, translating "no rows" into domain.ErrNotFound.
func (r *TaskRepository) Get(ctx context.Context, id int64) (domain.Task, error) {
	t, err := r.q.GetTask(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Task{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Task{}, err
	}
	return toDomain(t), nil
}

// List returns all tasks, newest first.
func (r *TaskRepository) List(ctx context.Context) ([]domain.Task, error) {
	rows, err := r.q.ListTasks(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Task, 0, len(rows))
	for _, t := range rows {
		out = append(out, toDomain(t))
	}
	return out, nil
}

// Update modifies an existing task.
func (r *TaskRepository) Update(ctx context.Context, id int64, p domain.UpdateTaskParams) (domain.Task, error) {
	t, err := r.q.UpdateTask(ctx, gen.UpdateTaskParams{Title: p.Title, Done: p.Done, ID: id})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Task{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Task{}, err
	}
	return toDomain(t), nil
}

// Delete removes a task, returning domain.ErrNotFound when nothing was deleted.
func (r *TaskRepository) Delete(ctx context.Context, id int64) error {
	n, err := r.q.DeleteTask(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
