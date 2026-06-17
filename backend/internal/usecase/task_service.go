// Package usecase contains application business logic. It depends only on the
// domain ports, never on frameworks or infrastructure.
package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/cerzzlive/gofiber-angular-spa/backend/internal/domain"
)

// TaskService implements the task-related use cases.
type TaskService struct {
	repo domain.TaskRepository
}

// NewTaskService wires a TaskService to its repository port.
func NewTaskService(repo domain.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

// List returns all tasks.
func (s *TaskService) List(ctx context.Context) ([]domain.Task, error) {
	return s.repo.List(ctx)
}

// Get returns a single task by id.
func (s *TaskService) Get(ctx context.Context, id int64) (domain.Task, error) {
	return s.repo.Get(ctx, id)
}

// Create validates and persists a new task.
func (s *TaskService) Create(ctx context.Context, p domain.CreateTaskParams) (domain.Task, error) {
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return domain.Task{}, fmt.Errorf("%w: title is required", domain.ErrInvalidInput)
	}
	return s.repo.Create(ctx, domain.CreateTaskParams{Title: title})
}

// Update validates and persists changes to an existing task.
func (s *TaskService) Update(ctx context.Context, id int64, p domain.UpdateTaskParams) (domain.Task, error) {
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return domain.Task{}, fmt.Errorf("%w: title is required", domain.ErrInvalidInput)
	}
	return s.repo.Update(ctx, id, domain.UpdateTaskParams{Title: title, Done: p.Done})
}

// Delete removes a task by id.
func (s *TaskService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
